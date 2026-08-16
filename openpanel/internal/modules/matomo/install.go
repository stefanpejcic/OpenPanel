package matomo

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
			writeNDJSON(w, flusher, canFlush, map[string]any{"error": "You have reached the maximum number of sites allowed."})
			return
		}
		flashSess(a, w, r, "warning", "You have reached the maximum number of sites allowed.")
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
// polling briefly for it to come up (mirrors joomla/drupal/opencart/
// nextcloud/prestashop's identical helper).
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

// ensureContainerTmpOnSameFilesystem makes the given PHP container's own
// /tmp resolve onto the same filesystem as /var/www/html - identical fix to
// prestashop/install.go's helper of the same name (see that file's comment
// for the full rationale: sys_temp_dir can't be overridden per-directory,
// so redirecting /tmp at the filesystem level is the only mechanism that
// actually works). Applied defensively here too since Matomo's installer
// also writes temp files during the DB setup step.
func ensureContainerTmpOnSameFilesystem(ctx context.Context, userContext, phpContainer string) {
	script := `set -e
if [ -L /tmp ]; then exit 0; fi
mkdir -p /var/www/html/.php-tmp
chmod 1777 /var/www/html/.php-tmp
rm -rf /tmp
ln -s /var/www/html/.php-tmp /tmp
`
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	_ = podmanmanager.Command(ctx, userContext, argv).Run()
}

// unpackMatomoArchive extracts a matomo-X.Y.Z.zip release asset, which
// (confirmed live against 5.12.0) wraps the actual flat application in a
// single top-level "matomo/" directory - same shape as OpenCart's upload/
// or Nextcloud's nextcloud/ subtree (not PrestaShop's double-zip layout).
// Unzips to a throwaway sibling directory then copies that wrapper
// directory's contents into destDir (already created empty by the caller).
func unpackMatomoArchive(ctx context.Context, archivePath, destDir string) error {
	tmpDir := destDir + ".extract-tmp"
	script := `set -e
rm -rf "$2"
mkdir -p "$2"
unzip -q "$1" -d "$2"
cp -a "$2/matomo/." "$3/"
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

// handleInstallStream drives a Matomo install end to end, streaming NDJSON
// progress events: download the release asset from GitHub, extract it into
// the docroot, create a MySQL database, then drive Matomo's own browser
// installation wizard (plugins/Installation/Controller.php) as a plain HTTP
// request sequence - Matomo ships no non-interactive CLI installer
// (confirmed by inspecting a real release's console commands: none of them
// is a core:install equivalent), so this replicates exactly what a browser
// would submit at each step, field-for-field matched against that
// controller's source and verified live end to end. See wizard.go for the
// step-by-step implementation.
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

	adminLogin := formOr(r, "admin_login", "admin")
	adminEmail := formOr(r, "admin_email", "admin@"+dom.DomainURL)
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = generateRandomString(16) + "!A1"
	}

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "matomo_" + strings.ToLower(generateRandomString(6))
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

	version := strings.TrimSpace(r.FormValue("matomo_version"))
	if version == "" {
		var latestErr error
		version, latestErr = latestMatomoVersion(ctx)
		if latestErr != nil {
			emit(map[string]any{"error": "Could not determine latest Matomo version: " + latestErr.Error()})
			return
		}
	}

	archiveName := "matomo-" + version + ".zip"
	archiveDir := "/etc/openpanel/matomo/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := "https://github.com/matomo-org/matomo/releases/download/" + version + "/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-P", archiveDir, downloadURL).Run(); runErr != nil {
			emit(map[string]any{"error": "Error downloading Matomo " + version + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Extracting files to " + installPath})
	if unpackErr := unpackMatomoArchive(ctx, archivePath, hostOSPath); unpackErr != nil {
		emit(map[string]any{"error": "Error extracting Matomo archive: " + unpackErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	// Host-side chown, not `podman exec ... chown` - the archive was
	// extracted host-side, so the files are owned by this process's own
	// (real, unmapped) UID. A rootless container's own "root" cannot chown
	// files it doesn't already own outside its own subuid mapping range
	// (established while building the Joomla/PrestaShop modules).
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

	// Matomo's installer and runtime write cache/session/log files under
	// tmp/ and (once) config/config.ini.php - confirmed live these must be
	// writable by the PHP-FPM worker uid, which is NOT the uid the docroot
	// was just chowned to above (that's the container's own mapped "root",
	// used only so podman-exec'd commands can write as that identity).
	// World-writable on just these two known-writable directories mirrors
	// the same scoped-chmod approach used for PrestaShop/OpenCart's
	// writable dirs - never the PHP source tree itself.
	emit(map[string]any{"status": "Setting permissions on writable directories"})
	_ = exec.Command("chmod", "-R", "777", filepath.Join(hostOSPath, "tmp")).Run()
	_ = exec.Command("chmod", "-R", "777", filepath.Join(hostOSPath, "config")).Run()

	emit(map[string]any{"status": "Ensuring PHP container temp directory is writable across filesystems"})
	ensureContainerTmpOnSameFilesystem(ctx, userContext, phpContainer)

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
		emit(map[string]any{"error": "You have reached the maximum number of databases allowed on your plan."})
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

	emit(map[string]any{"status": "Running Matomo installation wizard"})
	schema := "Mariadb"
	if mysqlVersion == "mysql" {
		schema = "Mysql"
	}
	wizardResult, wizardErr := runMatomoInstallWizard(ctx, matomoWizardParams{
		SiteURL:       "https://" + selectedDomain + "/",
		DBHost:        mysqlVersion,
		DBName:        dbName,
		DBUser:        dbUser,
		DBPassword:    dbPassword,
		DBSchema:      schema,
		AdminLogin:    adminLogin,
		AdminPassword: adminPassword,
		AdminEmail:    adminEmail,
		SiteName:      selectedDomain,
	})
	if wizardErr != nil {
		emit(map[string]any{"error": "Matomo installation wizard failed: " + wizardErr.Error()})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}
	_ = wizardResult

	emit(map[string]any{"status": "Deploying admin login helper"})
	creds, credErr := saveMatomoCredentials(selectedDomain, adminLogin, adminPassword)
	if credErr != nil {
		emit(map[string]any{"status": "Warning: could not store admin credentials for auto-login: " + credErr.Error()})
	} else {
		loginFilePath := filepath.Join(hostOSPath, openpanelLoginFileName)
		if writeErr := os.WriteFile(loginFilePath, []byte(buildOpenpanelLoginPHP(creds.Login, creds.Password, creds.Token, webServer)), 0o644); writeErr == nil {
			_ = os.Chown(loginFilePath, uid, uid)
		}
	}

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "matomo"); insertErr != nil {
		emit(map[string]any{"error": "Matomo installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed Matomo on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "Matomo installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "Matomo installation completed!"})
}

func invalidateMySQLCaches(ctx context.Context, a *appctx.App, userContext, currentUsername string) {
	_ = a.Cache.Delete(ctx, "databases_info:"+userContext)
	_ = a.Cache.Delete(ctx, "get_database_count:"+currentUsername)
}

// emitCleanupFiles removes a failed install's partially-created directory -
// mirrors every other CMS install module's identical helper.
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
