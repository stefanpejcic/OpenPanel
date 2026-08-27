package opencart

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/mysql"
)

// handleInstallPage renders the install form / checks the plan's site
// limit for a GET, and hands POST off to handleInstallStream.
func handleInstallPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	plan, _ := a.QueryPlanDetailsByID(ctx, planID)
	websitesLimit := atoiDefault(plan.WebsitesLimit, 0)
	websiteCount, _ := countUserWebsites(a, userID)

	if websitesLimit != 0 && websiteCount >= websitesLimit {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/x-ndjson")
			flusher, canFlush := w.(http.Flusher)
			writeNDJSON(w, flusher, canFlush, map[string]any{"error": "You have reached the maximum number of sites allowed." + plan.UpgradeMessage()})
			return
		}
		flashSess(a, w, r, "warning", "You have reached the maximum number of sites allowed."+plan.UpgradeMessage())
	} else if r.Method == http.MethodPost {
		handleInstallStream(a, w, r)
		return
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	renderInstallPage(a, w, r, domains)
}

func formOr(r *http.Request, key, def string) string {
	if v := r.FormValue(key); v != "" {
		return v
	}
	return def
}

// ensureContainerRunning starts the container if it isn't already running,
// polling briefly for it to come up (mirrors joomla/drupal/phpapp's
// identical helper).
func ensureContainerRunning(ctx context.Context, userContext, container string) bool {
	if docker.IsServiceRunning(ctx, userContext, container) {
		return true
	}
	docker.StartOrStopContainer(ctx, userContext, container, "activate", "detached")
	const attempts = 15
	for i := 0; i < attempts; i++ {
		time.Sleep(2 * time.Second)
		if docker.IsServiceRunning(ctx, userContext, container) {
			return true
		}
	}
	return false
}

// unpackOpenCartArchive extracts only the zip's upload/ subtree (the actual
// OpenCart application - the rest of the release zip is repo scaffolding:
// docs/, .github/, composer.json, tools/) directly into destDir, flattened
// (upload/index.php -> destDir/index.php, not destDir/upload/index.php).
func unpackOpenCartArchive(ctx context.Context, archivePath, destDir string) error {
	tmpDir := destDir + ".extract-tmp"
	script := `set -e
rm -rf "$2"
mkdir -p "$2"
unzip -q "$1" "upload/*" -d "$2"
cd "$2/upload"
for f in .[!.]* ..?* *; do
  [ -e "$f" ] || continue
  mv "$f" "$3/"
done
rm -rf "$2"
`
	cmd := exec.CommandContext(ctx, "sh", "-c", script, "sh", archivePath, tmpDir, destDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return &execError{msg: strings.TrimSpace(string(out)), err: err}
	}
	return nil
}

type execError struct {
	msg string
	err error
}

func (e *execError) Error() string { return e.msg }
func (e *execError) Unwrap() error { return e.err }

// handleInstallStream drives an OpenCart install end to end, streaming
// NDJSON progress events: download the release archive from GitHub (mirrors
// wordpress/install.go's wordpress.org download and joomla/install.go's
// GitHub download), extract its upload/ subtree directly into the docroot,
// create a MySQL database, copy config-dist.php -> config.php (and the
// admin/ equivalent - OpenCart's CLI installer requires both to already
// exist and be writable, it doesn't create them), then run OpenCart's own
// `install/cli_install.php install` CLI installer (confirmed against a live
// OpenCart 4 release: no --root/--uri flags needed beyond --http_server,
// and the install/ folder is left behind afterward - we remove it
// ourselves, matching what a manual install always recommends).
func handleInstallStream(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	flusher, canFlush := w.(http.Flusher)
	emit := func(v map[string]any) { writeNDJSON(w, flusher, canFlush, v) }

	ipAddress := reqip.ClientIP(r)
	domainID := r.FormValue("domain_id")
	if domainID == "" {
		emit(map[string]any{"error": "Missing required field: domain"})
		return
	}

	emit(map[string]any{"status": "Checking if existing installation processes are running.."})
	if err := createLockFile(currentUsername); err != nil {
		emit(map[string]any{"error": "Error creating lock file: " + err.Error()})
		return
	}
	defer removeLockFile(currentUsername)

	dom, found, dbErr := lookupDomainByID(ctx, a, domainID)
	if dbErr != nil {
		emit(map[string]any{"error": "An error occurred fetching docroot for domain from database."})
		return
	}
	if !found {
		emit(map[string]any{"error": "Domain not found"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, dom.DomainURL) {
		return
	}

	emit(map[string]any{"status": "Validating provided data"})
	subdirectory := strings.ToLower(strings.ReplaceAll(r.FormValue("subdirectory"), " ", ""))
	if !isValidSubdirectory(subdirectory) {
		emit(map[string]any{"error": "Invalid subdirectory."})
		return
	}

	adminUsername := formOr(r, "admin_username", "admin")
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = generateRandomString(16) + "!A1"
	}
	adminEmail := formOr(r, "admin_email", "admin@"+dom.DomainURL)

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "opencart_" + strings.ToLower(generateRandomString(6))
	}
	dbUser := strings.ToLower(r.FormValue("db_user"))
	if dbUser == "" {
		dbUser = strings.ToLower(generateRandomString(10))
	}
	dbPassword := r.FormValue("db_password")
	if dbPassword == "" {
		dbPassword = generateRandomString(16)
	}

	docroot := dom.Docroot.String
	selectedDomain := dom.DomainURL
	installPath := docroot
	if subdirectory != "" {
		installPath = strings.TrimSuffix(docroot, "/") + "/" + subdirectory
		selectedDomain = selectedDomain + "/" + subdirectory
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpVersion := dom.PHPVersion.String
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + phpVersion
	}

	emit(map[string]any{"status": "Starting PHP container: " + phpContainer})
	if !ensureContainerRunning(ctx, userContext, phpContainer) {
		emit(map[string]any{"error": "PHP container failed to start. Please check it from Services."})
		return
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	docrootWithoutWWW := strings.TrimPrefix(strings.TrimPrefix(installPath, "/var/www/html/"), "/")
	hostOSPath := filepath.Join(htmlVolume, docrootWithoutWWW)

	if _, statErr := os.Stat(hostOSPath); statErr == nil {
		if entries, readErr := os.ReadDir(hostOSPath); readErr == nil && len(entries) > 0 {
			emit(map[string]any{"error": "Directory " + installPath + " already exists and is not empty."})
			return
		}
	}
	if mkErr := os.MkdirAll(hostOSPath, 0o755); mkErr != nil {
		emit(map[string]any{"error": "Error creating document root: " + mkErr.Error()})
		return
	}

	version := strings.TrimSpace(r.FormValue("opencart_version"))
	if version == "" {
		var latestErr error
		version, latestErr = latestOpenCartVersion(ctx)
		if latestErr != nil {
			emit(map[string]any{"error": "Could not determine latest OpenCart version: " + latestErr.Error()})
			return
		}
	}

	archiveName := "opencart-" + version + ".zip"
	archiveDir := "/etc/openpanel/opencart/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := "https://github.com/opencart/opencart/releases/download/" + version + "/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-P", archiveDir, downloadURL).Run(); runErr != nil {
			emit(map[string]any{"error": "Error downloading OpenCart " + version + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Extracting files to " + installPath})
	if unpackErr := unpackOpenCartArchive(ctx, archivePath, hostOSPath); unpackErr != nil {
		emit(map[string]any{"error": "Error extracting OpenCart archive: " + unpackErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Preparing config.php files"})
	if copyErr := copyFile(filepath.Join(hostOSPath, "config-dist.php"), filepath.Join(hostOSPath, "config.php")); copyErr != nil {
		emit(map[string]any{"error": "Error creating config.php: " + copyErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}
	if copyErr := copyFile(filepath.Join(hostOSPath, "admin", "config-dist.php"), filepath.Join(hostOSPath, "admin", "config.php")); copyErr != nil {
		emit(map[string]any{"error": "Error creating admin/config.php: " + copyErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	// Host-side chown, not `podman exec ... chown` - the archive was
	// extracted host-side, so the files are owned by this process's own
	// (real, unmapped) UID. A rootless container's own "root" is confined
	// to its user-namespace's UID range and cannot chown files it doesn't
	// already own outside that mapping (confirmed live while building the
	// Joomla module: that silently failed with "Operation not permitted",
	// leaving the docroot unwritable by the container's PHP process).
	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	uid, uidErr := podmanmanager.GetUID(userContext)
	if uidErr != nil {
		emit(map[string]any{"error": "Could not determine file owner: " + uidErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}
	uidStr := strconv.Itoa(uid)
	if chownErr := exec.Command("chown", "-R", uidStr+":"+uidStr, hostOSPath).Run(); chownErr != nil {
		emit(map[string]any{"error": "Could not set file ownership: " + chownErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	mysqlVersion := mysql.GetMySQLVersion(ctx, a, userContext)
	emit(map[string]any{"status": "Testing database connection.."})
	if !mysql.CheckMySQLInsideContainer(ctx, userContext, true) {
		emit(map[string]any{"status": "Checking " + mysqlVersion + " container status.."})
		if !mysql.CheckMySQLNotTemporary(ctx, userContext, mysqlVersion) {
			emit(map[string]any{"error": "The " + mysqlVersion + " container is either not running or still initializing. Please ensure your plan has sufficient resources to start the service."})
			emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
			return
		}
	}

	if mysql.DatabaseLimitReached(ctx, a, userID, currentUsername, userContext) {
		emit(map[string]any{"error": "You have reached the maximum number of databases allowed on your plan." + a.UpgradeMessageForUser(ctx, userID)})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Creating database " + dbName + " and user " + dbUser})
	const dbHost = "%"
	queries := []string{
		"CREATE DATABASE IF NOT EXISTS `" + dbName + "`",
		"CREATE USER IF NOT EXISTS '" + dbUser + "'@'" + dbHost + "' IDENTIFIED BY '" + dbPassword + "'",
		"GRANT ALL PRIVILEGES ON `" + dbName + "`.* TO '" + dbUser + "'@'%'",
		"FLUSH PRIVILEGES",
	}
	for _, q := range queries {
		if _, execErr := mysqlmanager.Exec(ctx, userContext, q, ""); execErr != nil {
			invalidateMySQLCaches(ctx, a, userContext, currentUsername)
			emit(map[string]any{"error": "Error creating MySQL database and user: " + execErr.Error()})
			emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
			emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
			return
		}
	}
	invalidateMySQLCaches(ctx, a, userContext, currentUsername)

	emit(map[string]any{"status": "Running OpenCart CLI installer"})
	httpServer := "https://" + selectedDomain + "/"
	// cli_install.php's own argv parser only recognises space-separated
	// `--flag value` pairs (substr($argv[$i], 2) as the key) - `--flag=value`
	// silently fails to populate anything, confirmed live: it produced a
	// "missing or invalid" error for every field passed that way.
	installArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
		installPath+"/install/cli_install.php", "install",
		"--username", adminUsername,
		"--email", adminEmail,
		"--password", adminPassword,
		"--http_server", httpServer,
		"--language", "en-gb",
		"--db_driver", "mysqli",
		"--db_hostname", mysqlVersion,
		"--db_username", dbUser,
		"--db_password", dbPassword,
		"--db_database", dbName,
		"--db_port", "3306",
		"--db_prefix", "oc_")
	out, runErr := podmanmanager.Command(ctx, userContext, installArgv).CombinedOutput()
	if runErr != nil || !strings.Contains(string(out), "SUCCESS!") {
		emit(map[string]any{"error": "OpenCart CLI installer failed: " + strings.TrimSpace(string(out))})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	emit(map[string]any{"status": "Removing installer folder"})
	_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "rm", "-rf", installPath+"/install")).Run()

	emit(map[string]any{"status": "Deploying admin login helper"})
	loginFilePath := filepath.Join(hostOSPath, openpanelLoginFileName)
	if writeErr := os.WriteFile(loginFilePath, []byte(openpanelLoginPHP), 0o644); writeErr == nil {
		_ = os.Chown(loginFilePath, uid, uid)
	}

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "opencart"); insertErr != nil {
		emit(map[string]any{"error": "OpenCart installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed OpenCart on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "OpenCart installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "OpenCart installation completed!"})
}

func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0o644)
}

func invalidateMySQLCaches(ctx context.Context, a *appctx.App, userContext, currentUsername string) {
	_ = a.Cache.Delete(ctx, "databases_info:"+userContext)
	_ = a.Cache.Delete(ctx, "get_database_count:"+currentUsername)
}

// emitCleanupFiles removes a failed install's partially-created directory -
// like joomla/drupal's identical helper, always safe to blow away entirely
// since it's always a directory install.go just created.
func emitCleanupFiles(ctx context.Context, userContext, phpContainer, installPath string, emit func(map[string]any)) {
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "rm", "-rf", installPath)
	if err := podmanmanager.Command(ctx, userContext, argv).Run(); err != nil {
		emit(map[string]any{"status": "Cleanup: failed to remove " + installPath + ": " + err.Error()})
		return
	}
	emit(map[string]any{"status": "Cleanup: removed " + installPath})
}

func emitCleanupDatabase(ctx context.Context, userContext, dbName, dbUser, dbHost string, emit func(map[string]any)) {
	_, _ = mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+dbName+"`", "")
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'"+dbHost+"'", ""); execErr != nil {
		emit(map[string]any{"error": "Cleanup: failed to drop database/user: " + execErr.Error()})
		return
	}
	emit(map[string]any{"status": "Cleanup: dropped database `" + dbName + "` and user `" + dbUser + "`"})
}

func isValidSubdirectory(subdirectory string) bool {
	if subdirectory == "" {
		return true
	}
	return !strings.Contains(subdirectory, "..") && !strings.HasPrefix(subdirectory, "/")
}
