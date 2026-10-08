package prestashop

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

// ensureContainerTmpOnSameFilesystem makes the container's /tmp resolve onto the same filesystem as /var/www/html by symlinking it into a shared sticky-bit dir there (idempotent) - PHP's sys_temp_dir can't be overridden per-directory (PHP_INI_SYSTEM scope, .user.ini ignored), so a filesystem-level redirect of /tmp itself is the only thing that works; applies container-wide, which is intentional and safe
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

// unpackPrestashopArchive extracts the real application from a GitHub release asset, which is a two-layer package: prestashop_<version>.zip contains an inner prestashop.zip, itself the actual flat-root application (no wrapper folder) - this extracts the inner zip to a temp file, then unzips that directly into destDir
func unpackPrestashopArchive(ctx context.Context, archivePath, destDir string) error {
	tmpDir := destDir + ".extract-tmp"
	script := `set -e
rm -rf "$2"
mkdir -p "$2" "$3"
unzip -q "$1" "prestashop.zip" -d "$2"
unzip -q "$2/prestashop.zip" -d "$3"
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

// handleInstallStream drives a PrestaShop install end to end over NDJSON: download the release asset, extract inner prestashop.zip into the docroot, create the DB, run install/index_cli.php install (its argv parser only recognizes "--flag=value", the opposite of OpenCart's space-separated form)
// immediately after a successful install, admin/ is renamed to a random name and install/ is removed - PrestaShop's own AdminLoginController does the rename automatically on first browser load, but doing it here closes the window where admin/ sits at its guessable default name
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

	adminEmail := web.FormOr(r, "admin_email", "admin@"+dom.DomainURL)
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = appkit.RandomString(16) + "!A1"
	}
	adminFirstname := web.FormOr(r, "admin_firstname", "Admin")
	adminLastname := web.FormOr(r, "admin_lastname", "User")

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "prestashop_" + strings.ToLower(appkit.RandomString(6))
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
	baseURI := "/"
	if subdirectory != "" {
		installPath = strings.TrimSuffix(docroot, "/") + "/" + subdirectory
		selectedDomain = selectedDomain + "/" + subdirectory
		baseURI = "/" + subdirectory + "/"
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

	version := strings.TrimSpace(r.FormValue("prestashop_version"))
	if version == "" {
		var latestErr error
		version, latestErr = latestPrestashopVersion(ctx)
		if latestErr != nil {
			emit(map[string]any{"error": "Could not determine latest PrestaShop version: " + latestErr.Error()})
			return
		}
	}

	archiveName := "prestashop_" + version + ".zip"
	archiveDir := "/etc/openpanel/prestashop/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := "https://github.com/PrestaShop/PrestaShop/releases/download/" + version + "/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-P", archiveDir, downloadURL).Run(); runErr != nil {
			emit(map[string]any{"error": "Error downloading PrestaShop " + version + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Extracting files to " + installPath})
	if unpackErr := unpackPrestashopArchive(ctx, archivePath, hostOSPath); unpackErr != nil {
		emit(map[string]any{"error": "Error extracting PrestaShop archive: " + unpackErr.Error()})
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

	// PrestaShop's Symfony Filesystem component writes cache files via tmp-file-then-rename(), which fails with EXDEV since the container's /tmp is a different filesystem than the docroot's bind mount - sys_temp_dir can't fix this (PHP_INI_SYSTEM scope), so ensureContainerTmpOnSameFilesystem redirects /tmp itself at the kernel level instead
	emit(map[string]any{"status": "Ensuring PHP container temp directory is writable across filesystems"})
	ensureContainerTmpOnSameFilesystem(ctx, userContext, phpContainer)

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

	emit(map[string]any{"status": "Running PrestaShop CLI installer"})
	// index_cli.php's argv parser only recognizes "--flag=value" single-element pairs - space-separated "--flag value" silently drops the value
	installArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
		installPath+"/install/index_cli.php",
		"--domain="+dom.DomainURL,
		"--base_uri="+baseURI,
		"--db_server="+mysql.AppDBHost(userContext, mysqlVersion),
		"--db_name="+dbName,
		"--db_user="+dbUser,
		"--db_password="+dbPassword,
		"--db_create=0",
		"--db_clear=1",
		"--prefix=ps_",
		"--firstname="+adminFirstname,
		"--lastname="+adminLastname,
		"--email="+adminEmail,
		"--password="+adminPassword,
		"--language=en",
		"--country=us",
		"--timezone=UTC",
		"--fixtures=0",
		"--ssl=1")
	out, runErr := podmanmanager.Command(ctx, userContext, installArgv).CombinedOutput()
	if runErr != nil || !strings.Contains(string(out), "Installation successful") {
		emit(map[string]any{"error": "PrestaShop CLI installer failed: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		cmsapp.EmitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	emit(map[string]any{"status": "Warming up production cache"})
	// the CLI installer doesn't warm the Symfony prod cache the way the browser installer's final redirect does - without this, the first real HTTP request fatals opening var/cache/prod/appParameters.php, and a stray concurrent request during that window can leave the cache half-written and permanently broken, so warm it here once in a controlled step
	warmupArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
		"-d", "sys_temp_dir="+installPath+"/var/tmp",
		installPath+"/bin/console", "cache:warmup", "--env=prod")
	_, _ = podmanmanager.Command(ctx, userContext, warmupArgv).CombinedOutput()

	emit(map[string]any{"status": "Removing installer folder"})
	_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "rm", "-rf", installPath+"/install")).Run()

	// the front/admin controllers need runtime write access to cache/asset dirs, written by the php-fpm worker's uid rather than the mapped root the docroot is chowned to - applied after the installer runs since the installer recreates these dirs with restrictive permissions, undoing an earlier chmod; world-write on just these known-writable folders is what PrestaShop's own hosting docs recommend
	emit(map[string]any{"status": "Setting permissions on writable directories"})
	writableDirs := []string{"var", "config", "img", "mails", "modules", "translations", "upload", "download", "themes"}
	for _, d := range writableDirs {
		_ = exec.Command("chmod", "-R", "777", filepath.Join(hostOSPath, d)).Run()
	}

	emit(map[string]any{"status": "Securing admin directory"})
	// see handleInstallStream's doc comment: doing this right after install avoids ever leaving admin/ at its guessable default name
	adminDirName := "admin" + appkit.RandomString(20)
	if renameErr := os.Rename(filepath.Join(hostOSPath, "admin"), filepath.Join(hostOSPath, adminDirName)); renameErr != nil {
		emit(map[string]any{"status": "Warning: could not rename admin directory: " + renameErr.Error()})
		adminDirName = "admin"
	}
	_ = exec.Command("chmod", "-R", "777", filepath.Join(hostOSPath, adminDirName, "autoupgrade")).Run()
	_ = exec.Command("chmod", "-R", "777", filepath.Join(hostOSPath, adminDirName, "themes")).Run()

	emit(map[string]any{"status": "Deploying admin login helper"})
	loginFilePath := filepath.Join(hostOSPath, adminDirName, openpanelLoginFileName)
	if writeErr := os.WriteFile(loginFilePath, []byte(openpanelLoginPHP), 0o644); writeErr == nil {
		_ = os.Chown(loginFilePath, uid, uid)
	}

	// record exactly what this install created so a later uninstall can remove precisely that, not the whole docroot
	_ = installmanifest.Record(hostOSPath)

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "prestashop"); insertErr != nil {
		emit(map[string]any{"error": "PrestaShop installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed PrestaShop on domain "+selectedDomain, ipAddress)
	web.Flash(a, w, r, "success", web.Tr(a, r, "PrestaShop installed successfully on %(selected_domain)s", "selected_domain", selectedDomain))
	emit(map[string]any{"status": "PrestaShop installation completed!"})
}
