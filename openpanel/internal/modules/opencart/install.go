package opencart

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/installmanifest"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/mysql"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// unpackOpenCartArchive extracts only the zip's upload/ subtree (the actual app, the rest is repo scaffolding) flattened directly into destDir
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

// handleInstallStream drives an OpenCart install end to end over NDJSON: download from GitHub, extract upload/ into the docroot, create the DB, copy config-dist.php to config.php (both root and admin/, since the CLI installer needs them to already exist), then run install/cli_install.php and remove the install/ folder afterward
func handleInstallStream(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	in, ok := cmsapp.StartInstall(a, w, r)
	if !ok {
		return
	}
	defer appkit.RemoveLockFile(in.Username)
	ctx := in.Ctx
	userID := in.UserID
	currentUsername := in.Username
	userContext := in.UserContext
	emit := in.Emit
	ipAddress := in.IP
	domainID := in.DomainID
	dom := in.Domain
	subdirectory := in.Subdirectory

	adminUsername := web.FormOr(r, "admin_username", "admin")
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = appkit.RandomString(16) + "!A1"
	}
	adminEmail := web.FormOr(r, "admin_email", "admin@"+dom.DomainURL)

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "opencart_" + strings.ToLower(appkit.RandomString(6))
	}
	dbUser := strings.ToLower(r.FormValue("db_user"))
	if dbUser == "" {
		dbUser = strings.ToLower(appkit.RandomString(10))
	}
	dbName = mysqlmanager.WithPrefix(a.Config, userContext, dbName)
	dbUser = mysqlmanager.WithPrefix(a.Config, userContext, dbUser)
	dbPassword := r.FormValue("db_password")
	if dbPassword == "" {
		dbPassword = appkit.RandomString(16)
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
	if !cmsapp.EnsureContainerRunning(ctx, userContext, phpContainer) {
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
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Preparing config.php files"})
	if copyErr := copyFile(filepath.Join(hostOSPath, "config-dist.php"), filepath.Join(hostOSPath, "config.php")); copyErr != nil {
		emit(map[string]any{"error": "Error creating config.php: " + copyErr.Error()})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}
	if copyErr := copyFile(filepath.Join(hostOSPath, "admin", "config-dist.php"), filepath.Join(hostOSPath, "admin", "config.php")); copyErr != nil {
		emit(map[string]any{"error": "Error creating admin/config.php: " + copyErr.Error()})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	// host-side chown, not podman exec chown - a rootless container's root can't chown files outside its UID mapping (silently failed with "Operation not permitted" when building Joomla)
	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	uid, uidErr := podmanmanager.GetUID(userContext)
	if uidErr != nil {
		emit(map[string]any{"error": "Could not determine file owner: " + uidErr.Error()})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}
	uidStr := strconv.Itoa(uid)
	if chownErr := exec.Command("chown", "-R", uidStr+":"+uidStr, hostOSPath).Run(); chownErr != nil {
		emit(map[string]any{"error": "Could not set file ownership: " + chownErr.Error()})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	mysqlVersion := mysql.GetMySQLVersion(ctx, a, userContext)
	if !docker.IsServiceRunning(ctx, userContext, mysqlVersion) {
		emit(map[string]any{"status": "Starting " + mysqlVersion + " container.."})
		docker.StartOrStopContainer(ctx, userContext, mysqlVersion, "activate", "detached")
	}
	emit(map[string]any{"status": "Testing database connection.."})
	if !mysql.CheckMySQLInsideContainer(ctx, userContext, true) {
		emit(map[string]any{"status": "Checking " + mysqlVersion + " container status.."})
		if !mysql.CheckMySQLNotTemporary(ctx, userContext, mysqlVersion) {
			emit(map[string]any{"error": "The " + mysqlVersion + " container is either not running or still initializing. Please ensure your plan has sufficient resources to start the service."})
			cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
			return
		}
	}

	if mysql.DatabaseLimitReached(ctx, a, userID, currentUsername, userContext) {
		emit(map[string]any{"error": "You have reached the maximum number of databases allowed on your plan." + a.UpgradeMessageForUser(ctx, userID)})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
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
			appkit.InvalidateMySQLCaches(ctx, a, userContext, currentUsername)
			emit(map[string]any{"error": "Error creating MySQL database and user: " + execErr.Error()})
			cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
			cmsapp.EmitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
			return
		}
	}
	appkit.InvalidateMySQLCaches(ctx, a, userContext, currentUsername)

	emit(map[string]any{"status": "Running OpenCart CLI installer"})
	httpServer := "https://" + selectedDomain + "/"
	// cli_install.php only recognizes space-separated "--flag value" pairs - "--flag=value" silently fails to populate anything
	installArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
		installPath+"/install/cli_install.php", "install",
		"--username", adminUsername,
		"--email", adminEmail,
		"--password", adminPassword,
		"--http_server", httpServer,
		"--language", "en-gb",
		"--db_driver", "mysqli",
		"--db_hostname", mysql.AppDBHost(userContext, mysqlVersion),
		"--db_username", dbUser,
		"--db_password", dbPassword,
		"--db_database", dbName,
		"--db_port", "3306",
		"--db_prefix", "oc_")
	out, runErr := podmanmanager.Command(ctx, userContext, installArgv).CombinedOutput()
	if runErr != nil || !strings.Contains(string(out), "SUCCESS!") {
		emit(map[string]any{"error": "OpenCart CLI installer failed: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		cmsapp.EmitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	emit(map[string]any{"status": "Removing installer folder"})
	_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "rm", "-rf", installPath+"/install")).Run()

	emit(map[string]any{"status": "Deploying admin login helper"})
	loginFilePath := filepath.Join(hostOSPath, openpanelLoginFileName)
	if writeErr := os.WriteFile(loginFilePath, []byte(openpanelLoginPHP), 0o644); writeErr == nil {
		_ = os.Chown(loginFilePath, uid, uid)
	}

	// record exactly what this install created so a later uninstall can remove precisely that, not the whole docroot
	_ = installmanifest.Record(hostOSPath)

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "opencart"); insertErr != nil {
		emit(map[string]any{"error": "OpenCart installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed OpenCart on domain "+selectedDomain, ipAddress)
	web.Flash(a, w, r, "success", web.Tr(a, r, "OpenCart installed successfully on %(selected_domain)s", "selected_domain", selectedDomain))
	emit(map[string]any{"status": "OpenCart installation completed!"})
}

func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0o644)
}
