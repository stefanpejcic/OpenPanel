package prestashop

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
// polling briefly for it to come up (mirrors joomla/drupal/opencart/nextcloud's
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

// ensureContainerTmpOnSameFilesystem makes the given PHP container's own
// /tmp resolve onto the same filesystem as /var/www/html, by replacing it
// with a symlink into a shared, sticky-bit directory under /var/www/html
// (idempotent - a no-op once already applied). Confirmed live: PHP's
// sys_temp_dir cannot be overridden per-directory (it's PHP_INI_SYSTEM
// scope, .user.ini is silently ignored), so redirecting the literal /tmp
// path at the filesystem level is the only mechanism that actually works;
// every PHP tempnam()/rename() call using the hardcoded "/tmp" string
// transparently lands on the right device with no PHP-level configuration
// needed. This affects every domain sharing this PHP version's container,
// which is intentional and always safe - it fixes an existing filesystem
// mismatch, it doesn't introduce one.
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

// unpackPrestashopArchive extracts the real application from a GitHub
// release asset, which (confirmed live against 8.2.7) is a two-layer
// package: prestashop_<version>.zip contains index.php,
// Install_PrestaShop.html, and an INNER prestashop.zip - that inner zip is
// the actual flat-root application (no wrapper folder, unlike
// OpenCart's upload/ or Nextcloud's nextcloud/ subtree). This extracts the
// outer zip's prestashop.zip member to a temp file, then unzips that
// directly into destDir.
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

// handleInstallStream drives a PrestaShop install end to end, streaming
// NDJSON progress events: download the release asset from GitHub, extract
// its inner prestashop.zip directly into the docroot, create a MySQL
// database, then run PrestaShop's own `install/index_cli.php install` CLI
// installer (confirmed live against a real 8.2.7 release - its argv parser
// only recognises `--flag=value` single-element pairs, the opposite of
// OpenCart's space-separated form, so don't copy that pattern here).
//
// Immediately after a successful install, the admin/ directory is renamed
// to a random name and install/ is removed - PrestaShop's own
// AdminLoginController does the admin/ rename itself automatically the
// first time any admin controller loads in a browser (confirmed live), but
// leaving that to chance means the directory sits at the guessable default
// "admin" until some visitor happens to trigger it; doing it here removes
// that window and matches doing the rename before any credentials or
// login links reference the path.
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

	adminEmail := formOr(r, "admin_email", "admin@"+dom.DomainURL)
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = generateRandomString(16) + "!A1"
	}
	adminFirstname := formOr(r, "admin_firstname", "Admin")
	adminLastname := formOr(r, "admin_lastname", "User")

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "prestashop_" + strings.ToLower(generateRandomString(6))
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

	// PrestaShop's vendored Symfony Filesystem component (and its own
	// LegacyClassLoader) write cache files atomically via a tmp-file-then-
	// rename() pattern. Confirmed live: this container's /tmp is a
	// different filesystem (fuseblk) from the docroot's bind-mounted volume
	// (ext2/ext3) - rename() across filesystems always fails with EXDEV.
	// A `sys_temp_dir` override in .user.ini does NOT fix this - confirmed
	// live that ini_get('sys_temp_dir') stays empty and sys_get_temp_dir()
	// keeps returning /tmp regardless, because sys_temp_dir is a
	// PHP_INI_SYSTEM directive (php.ini/pool-level only, never
	// per-directory). The only mechanism that actually works is a
	// filesystem-level fix: point the container's own /tmp at a directory
	// that lives under /var/www/html (the same bind-mounted filesystem
	// every domain's docroot is on), so PHP's hardcoded "/tmp" path string
	// transparently resolves onto the right device at the kernel level -
	// no ini setting involved. This is a container-wide fix (shared by
	// every domain on this PHP version), not per-install, so it's applied
	// once, idempotently, rather than per-domain.
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

	emit(map[string]any{"status": "Running PrestaShop CLI installer"})
	// index_cli.php's own argv parser (install/classes/datas.php,
	// getAndCheckArgs()) only recognises `--flag=value` single-element
	// pairs via a regex on each argv entry - space-separated `--flag value`
	// silently drops the value, confirmed live against 8.2.7.
	installArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
		installPath+"/install/index_cli.php",
		"--domain="+dom.DomainURL,
		"--base_uri="+baseURI,
		"--db_server="+mysqlVersion,
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
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	emit(map[string]any{"status": "Warming up production cache"})
	// PrestaShop's CLI installer does NOT compile/warm the Symfony prod
	// container cache the way the browser-based installer's own final
	// redirect does - confirmed live: without this, the very first real
	// HTTP request to the site (front office OR admin) fatals with
	// "require_once(): Failed opening required '.../var/cache/prod/
	// appParameters.php'", because that file only gets generated on first
	// access, and a stray concurrent request during that window can leave
	// the cache half-written and permanently broken. Warming it here, once,
	// in a single controlled step right after install avoids that race
	// entirely - this mirrors what PrestaShop's own documentation
	// recommends running by hand after a CLI install.
	warmupArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
		"-d", "sys_temp_dir="+installPath+"/var/tmp",
		installPath+"/bin/console", "cache:warmup", "--env=prod")
	_, _ = podmanmanager.Command(ctx, userContext, warmupArgv).CombinedOutput()

	emit(map[string]any{"status": "Removing installer folder"})
	_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "rm", "-rf", installPath+"/install")).Run()

	// PrestaShop's front/admin controllers need real runtime write access
	// to several cache/asset directories (var/cache, var/logs, per-theme
	// assets/cache for CCC JS/CSS compilation, config/, etc) - confirmed
	// live these get written by the actual PHP-FPM worker process (uid of
	// the "www-data" pool user), which is NOT the same uid the docroot is
	// chowned to above (that's the container's own mapped "root", used so
	// the CLI installer - run via `podman exec` as that mapped root - can
	// write everything at install time). Applied AFTER the installer runs,
	// not before: the installer itself recreates var/cache/config/etc as
	// the mapped root with default restrictive permissions, which would
	// silently undo an earlier chmod. Since the worker's exact uid isn't
	// reliably knowable in advance, world-write on just these known-
	// writable folders is what PrestaShop's own hosting docs recommend,
	// scoped to cache/asset/config directories only - not the PHP source
	// tree.
	emit(map[string]any{"status": "Setting permissions on writable directories"})
	writableDirs := []string{"var", "config", "img", "mails", "modules", "translations", "upload", "download", "themes"}
	for _, d := range writableDirs {
		_ = exec.Command("chmod", "-R", "777", filepath.Join(hostOSPath, d)).Run()
	}

	emit(map[string]any{"status": "Securing admin directory"})
	// See the comment above handleInstallStream: doing this ourselves,
	// right after install, avoids ever leaving the admin/ directory at its
	// guessable default name.
	adminDirName := "admin" + generateRandomString(20)
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

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "prestashop"); insertErr != nil {
		emit(map[string]any{"error": "PrestaShop installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed PrestaShop on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "PrestaShop installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "PrestaShop installation completed!"})
}

func invalidateMySQLCaches(ctx context.Context, a *appctx.App, userContext, currentUsername string) {
	_ = a.Cache.Delete(ctx, "databases_info:"+userContext)
	_ = a.Cache.Delete(ctx, "get_database_count:"+currentUsername)
}

// emitCleanupFiles removes a failed install's partially-created directory -
// like joomla/opencart/nextcloud's identical helper, always safe to blow
// away entirely since it's always a directory install.go just created.
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
