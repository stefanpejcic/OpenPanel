package moodle

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
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
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
// polling briefly for it to come up (mirrors every other CMS module's
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

// unpackMoodleArchive extracts download.moodle.org's packaged tarball
// (a single top-level "moodle/" directory containing config-dist.php,
// admin/, lib/, and a public/ subdirectory that is the actual web root -
// confirmed live against the 5.0.2 stable branch tarball) directly into
// destDir, stripping that wrapper directory so destDir itself becomes the
// Moodle "app root" (approot) - config.php ends up at destDir/config.php,
// the web-served files at destDir/public/.
func unpackMoodleArchive(ctx context.Context, archivePath, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "tar", "-xzf", archivePath, "-C", destDir, "--strip-components=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
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

// moodleBranch converts a "vX.Y.Z"-style GitHub tag into
// download.moodle.org's packaging branch token ("stableXY0" -
// e.g. v5.0.2 -> "500", v4.5.3 -> "405" - confirmed live: packaging.moodle.org
// serves /stable{major}{minor:02d}/moodle-latest-{major}{minor:02d}.tgz).
func moodleBranch(version string) string {
	v := strings.TrimPrefix(version, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return ""
	}
	major := parts[0]
	minor := parts[1]
	if len(minor) == 1 {
		minor = "0" + minor
	}
	return major + minor
}

// handleInstallStream drives a Moodle install end to end, streaming NDJSON
// progress events: download the packaged tarball from download.moodle.org,
// extract it into a sibling "app root" directory (see moodle.go's package
// doc comment for why - Moodle 5.x's public/ split), symlink the domain's
// actual docroot to <approot>/public, create a MySQL database, run
// Moodle's own `admin/cli/install.php` non-interactively, then register a
// per-minute admin/cli/cron.php job via crons.AddJob (Moodle does nothing -
// no mail, no enrolments, no scheduled tasks - without this running).
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
	adminUser := formOr(r, "admin_username", "admin")
	siteFullName := formOr(r, "site_name", "Moodle")
	siteShortName := formOr(r, "site_shortname", "moodle")

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "moodle_" + strings.ToLower(generateRandomString(6))
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

	if info, statErr := os.Lstat(hostOSPath); statErr == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			if entries, readErr := os.ReadDir(hostOSPath); readErr == nil && len(entries) > 0 {
				emit(map[string]any{"error": "Directory " + installPath + " already exists and is not empty."})
				return
			}
		} else {
			emit(map[string]any{"error": "Directory " + installPath + " already exists."})
			return
		}
	}

	slug := siteSlug(selectedDomain)
	approotHostPath := filepath.Join(htmlVolume, slug+"_moodleapp")
	approotContainerPath := "/var/www/html/" + slug + "_moodleapp"
	datarootHostPath := filepath.Join(htmlVolume, slug+"_moodledata")
	datarootContainerPath := "/var/www/html/" + slug + "_moodledata"

	version := strings.TrimSpace(r.FormValue("moodle_version"))
	if version == "" {
		var latestErr error
		version, latestErr = latestMoodleVersion(ctx)
		if latestErr != nil {
			emit(map[string]any{"error": "Could not determine latest Moodle version: " + latestErr.Error()})
			return
		}
	}
	branch := moodleBranch(version)
	if branch == "" {
		emit(map[string]any{"error": "Could not determine Moodle release branch for version " + version})
		return
	}

	archiveName := "moodle-latest-" + branch + ".tgz"
	archiveDir := "/etc/openpanel/moodle/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := "https://download.moodle.org/download.php/direct/stable" + branch + "/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-O", archivePath, downloadURL).Run(); runErr != nil {
			_ = os.Remove(archivePath)
			emit(map[string]any{"error": "Error downloading Moodle " + version + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Extracting files to " + approotContainerPath})
	if unpackErr := unpackMoodleArchive(ctx, archivePath, approotHostPath); unpackErr != nil {
		emit(map[string]any{"error": "Error extracting Moodle archive: " + unpackErr.Error()})
		_ = os.RemoveAll(approotHostPath)
		return
	}

	if mkErr := os.MkdirAll(datarootHostPath, 0o755); mkErr != nil {
		emit(map[string]any{"error": "Error creating moodledata directory: " + mkErr.Error()})
		_ = os.RemoveAll(approotHostPath)
		return
	}

	// Host-side chown, not `podman exec ... chown` - the archive was
	// extracted host-side, so the files are owned by this process's own
	// (real, unmapped) UID. A rootless container's own "root" is confined
	// to its user-namespace's UID range and cannot chown files it doesn't
	// already own outside that mapping (confirmed live while building the
	// Joomla module).
	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	uid, uidErr := podmanmanager.GetUID(userContext)
	if uidErr != nil {
		emit(map[string]any{"error": "Could not determine file owner: " + uidErr.Error()})
		_ = os.RemoveAll(approotHostPath)
		_ = os.RemoveAll(datarootHostPath)
		return
	}
	uidStr := strconv.Itoa(uid)
	_ = exec.Command("chown", "-R", uidStr+":"+uidStr, approotHostPath).Run()
	_ = exec.Command("chown", "-R", uidStr+":"+uidStr, datarootHostPath).Run()

	emit(map[string]any{"status": "Linking web root to Moodle's public/ directory"})
	_ = os.Remove(hostOSPath)
	// The symlink target must be the container-visible path
	// (/var/www/html/...), not approotHostPath's host filesystem path -
	// the symlink is created host-side but read by the php-fpm container,
	// which only sees its own /var/www/html/ bind mount, not the host's
	// /home/<user>/docker-data/... tree. Confirmed live: a host-path
	// symlink resolves fine via `ls` on the host but is broken (ENOENT)
	// from inside the container, causing a 403 on every request.
	if symErr := os.Symlink(approotContainerPath+"/public", hostOSPath); symErr != nil {
		emit(map[string]any{"error": "Error creating web root symlink: " + symErr.Error()})
		_ = os.RemoveAll(approotHostPath)
		_ = os.RemoveAll(datarootHostPath)
		return
	}

	mysqlVersion := mysql.GetMySQLVersion(ctx, a, userContext)
	emit(map[string]any{"status": "Testing database connection.."})
	if !mysql.CheckMySQLInsideContainer(ctx, userContext, true) {
		emit(map[string]any{"status": "Checking " + mysqlVersion + " container status.."})
		if !mysql.CheckMySQLNotTemporary(ctx, userContext, mysqlVersion) {
			emit(map[string]any{"error": "The " + mysqlVersion + " container is either not running or still initializing. Please ensure your plan has sufficient resources to start the service."})
			emitCleanupFiles(hostOSPath, approotHostPath, datarootHostPath, emit)
			return
		}
	}

	if mysql.DatabaseLimitReached(ctx, a, userID, currentUsername, userContext) {
		emit(map[string]any{"error": "You have reached the maximum number of databases allowed on your plan."})
		emitCleanupFiles(hostOSPath, approotHostPath, datarootHostPath, emit)
		return
	}

	emit(map[string]any{"status": "Creating database " + dbName + " and user " + dbUser})
	const dbHost = "%"
	queries := []string{
		"CREATE DATABASE IF NOT EXISTS `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		"CREATE USER IF NOT EXISTS '" + dbUser + "'@'" + dbHost + "' IDENTIFIED BY '" + dbPassword + "'",
		"GRANT ALL PRIVILEGES ON `" + dbName + "`.* TO '" + dbUser + "'@'%'",
		"FLUSH PRIVILEGES",
	}
	for _, q := range queries {
		if _, execErr := mysqlmanager.Exec(ctx, userContext, q, ""); execErr != nil {
			invalidateMySQLCaches(ctx, a, userContext, currentUsername)
			emit(map[string]any{"error": "Error creating MySQL database and user: " + execErr.Error()})
			emitCleanupFiles(hostOSPath, approotHostPath, datarootHostPath, emit)
			emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
			return
		}
	}
	invalidateMySQLCaches(ctx, a, userContext, currentUsername)

	dbType := "mariadb"
	if mysqlVersion == "mysql" {
		dbType = "mysqli"
	}

	emit(map[string]any{"status": "Running Moodle CLI installer (admin/cli/install.php)"})
	installArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
		approotContainerPath+"/admin/cli/install.php",
		"--chmod=0777",
		"--lang=en",
		"--wwwroot=https://"+selectedDomain,
		"--dataroot="+datarootContainerPath,
		"--dbtype="+dbType,
		"--dbhost="+mysqlVersion,
		"--dbname="+dbName,
		"--dbuser="+dbUser,
		"--dbpass="+dbPassword,
		"--prefix=mdl_",
		"--fullname="+siteFullName,
		"--shortname="+siteShortName,
		"--adminuser="+adminUser,
		"--adminpass="+adminPassword,
		"--adminemail="+adminEmail,
		"--non-interactive",
		"--agree-license")
	out, runErr := podmanmanager.Command(ctx, userContext, installArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "Moodle CLI installer failed: " + strings.TrimSpace(string(out))})
		emitCleanupFiles(hostOSPath, approotHostPath, datarootHostPath, emit)
		emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	// Moodle writes cache/session data under dataroot only (config.php and
	// the approot's PHP source stay read-only at runtime) - but the
	// php-fpm worker process's UID is not the same UID that just ran the
	// CLI installer above (that ran as this container's mapped "root" via
	// podman exec), the same worker-vs-install UID mismatch class of bug
	// found and fixed for PrestaShop earlier - so dataroot needs to be
	// writable by whatever UID actually is, which --chmod=0777 above
	// already asked Moodle's own installer to apply to everything it
	// creates under dataroot, but the dataroot directory itself (created
	// by this Go process above, not by the installer) needs the same
	// treatment explicitly.
	emit(map[string]any{"status": "Setting permissions on moodledata directory"})
	_ = exec.Command("chmod", "-R", "777", datarootHostPath).Run()

	emit(map[string]any{"status": "Registering cron job (admin/cli/cron.php, every minute)"})
	cronComment := moodleCronComment(selectedDomain)
	cronCommand := "php " + approotContainerPath + "/admin/cli/cron.php"
	if cronErr := crons.AddJob(ctx, userContext, cronComment, "0 * * * * *", phpContainer, cronCommand, true); cronErr != nil {
		emit(map[string]any{"status": "Warning: Moodle installed, but the cron job could not be registered automatically: " + cronErr.Error() + " - add it manually from Cron Jobs: schedule '0 * * * * *', container '" + phpContainer + "', command '" + cronCommand + "'."})
	}

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "moodle"); insertErr != nil {
		emit(map[string]any{"error": "Moodle installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed Moodle on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "Moodle installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "Moodle installation completed!"})
}

func invalidateMySQLCaches(ctx context.Context, a *appctx.App, userContext, currentUsername string) {
	_ = a.Cache.Delete(ctx, "databases_info:"+userContext)
	_ = a.Cache.Delete(ctx, "get_database_count:"+currentUsername)
}

// emitCleanupFiles removes a failed install's docroot symlink and its
// backing approot/dataroot directories - always safe since install.go just
// created all three.
func emitCleanupFiles(hostOSPath, approotHostPath, datarootHostPath string, emit func(map[string]any)) {
	_ = os.Remove(hostOSPath)
	_ = os.RemoveAll(approotHostPath)
	_ = os.RemoveAll(datarootHostPath)
	emit(map[string]any{"status": "Cleanup: removed partially-installed files"})
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
