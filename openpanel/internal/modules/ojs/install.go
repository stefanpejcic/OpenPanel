package ojs

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
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
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

// unpackOJSArchive extracts PKP's packaged tarball (a single top-level
// "ojs-{dotted-version}/" directory containing index.php, tools/, lib/,
// public/, etc. directly at its root - confirmed live against the 3.5.0-5
// release tarball) directly into destDir, stripping that wrapper directory
// so destDir itself becomes the OJS app root/web root.
func unpackOJSArchive(ctx context.Context, archivePath, destDir string) error {
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

// buildOJSInstallAnswers builds the newline-joined stdin block
// tools/install.php's OJSInstallTool::readParams() (which wraps
// PKP\cliTool\InstallTool::readParams()) expects, in the exact order it
// prompts for them - confirmed by reading
// lib/pkp/classes/cliTool/InstallTool.php and tools/install.php directly:
// there is no --flag CLI form the way Moodle/WordPress's installers have,
// only interactive stdin prompts, and (confirmed) no timeZone prompt exists
// in this CLI path even though the web install form has one.
func buildOJSInstallAnswers(filesDirContainerPath, adminUsername, adminPassword, adminEmail, dbHost, dbUser, dbPassword, dbName, oaiRepositoryID string) string {
	lines := []string{
		"en",                     // locale
		"",                       // additionalLocales
		filesDirContainerPath,    // filesDir
		adminUsername,            // adminUsername
		adminPassword,            // adminPassword
		adminPassword,            // adminPassword2
		adminEmail,               // adminEmail
		"mysqli",                 // databaseDriver
		dbHost,                   // databaseHost
		dbUser,                   // databaseUsername
		dbPassword,               // databasePassword
		dbName,                   // databaseName
		oaiRepositoryID,          // oaiRepositoryId
		"n",                      // enableBeacon
		"y",                      // install (OJSInstallTool's own extra confirmation prompt)
	}
	return strings.Join(lines, "\n") + "\n"
}

// runOJSInstaller execs tools/install.php inside phpContainer, feeding it
// buildOJSInstallAnswers' stdin block. Needs "-i" on the podman exec (not
// just "exec ... php ...") so stdin is actually piped into the container
// process - the same flag backups.go already relies on for the mysql
// import's `podman exec -i`.
func runOJSInstaller(ctx context.Context, userContext, phpContainer, approotContainerPath, answers string) (string, error) {
	argv := podmanmanager.PodmanArgv(userContext, "exec", "-i", phpContainer, "php", approotContainerPath+"/tools/install.php")
	cmd := podmanmanager.Command(ctx, userContext, argv)
	cmd.Stdin = strings.NewReader(answers)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// fixUpOJSConfig rewrites base_url and time_zone in the freshly-installed
// config.inc.php. Both are left wrong/blank by tools/install.php when run
// non-interactively: PKPInstall::updateConfig() sets base_url from
// $request->getBaseUrl() (there is no real HTTP request in a CLI process,
// so this resolves to something useless) and time_zone from
// $this->getParam('timeZone') (a param InstallTool::readParams() never
// actually prompts for on the CLI path, so it's always empty) - confirmed
// by reading lib/pkp/classes/install/PKPInstall.php directly. Both matter:
// base_url drives every URL OJS itself generates, and an empty time_zone
// leaves PHP's timezone unset.
func fixUpOJSConfig(configPath, selectedDomain string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	text := string(content)
	text = iniBaseURLRE.ReplaceAllString(text, iniQuoted("base_url", "https://"+selectedDomain))
	text = iniTimeZoneRE.ReplaceAllString(text, iniQuoted("time_zone", "UTC"))
	// The CLI installer (unlike the web form, which collects a real
	// hostname list) always writes allowed_hosts as a JSON array containing
	// one empty string (`["";]`) - AllowedHostsPolicy treats any non-""
	// value here as "restrict access to exactly these hosts" (see
	// lib/pkp/classes/security/authorization/AllowedHostsPolicy.php's
	// applies()/effect()), so a CLI install locks itself out of its own
	// site with "400 Server host not allowed" unless this is reset back to
	// the template's actual empty-string default (confirmed live).
	text = iniAllowedHostsRE.ReplaceAllString(text, iniQuoted("allowed_hosts", ""))
	return os.WriteFile(configPath, []byte(text), 0o644)
}

// handleInstallStream drives an OJS install end to end, streaming NDJSON
// progress events: download the packaged tarball from pkp.sfu.ca, extract
// it into a sibling "app root" directory (see ojs.go's package doc comment
// for why), symlink the domain's docroot to it, create a separate sibling
// "files" directory outside the docroot for OJS's file storage, create a
// MySQL database, run OJS's own `tools/install.php` non-interactively (via
// piped stdin - see buildOJSInstallAnswers), fix up base_url/time_zone,
// deploy the autologin helper PHP file, and register a per-minute
// lib/pkp/tools/scheduler.php cron job (OJS's documented way to run
// scheduled tasks in production instead of its discouraged built-in
// end-of-request task runner).
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

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "ojs_" + strings.ToLower(generateRandomString(6))
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
	approotHostPath := filepath.Join(htmlVolume, slug+"_ojsapp")
	approotContainerPath := "/var/www/html/" + slug + "_ojsapp"
	filesHostPath := filepath.Join(htmlVolume, slug+"_ojsfiles")
	filesContainerPath := "/var/www/html/" + slug + "_ojsfiles"

	version := strings.TrimSpace(r.FormValue("ojs_version"))
	var dotted string
	if version == "" {
		latest, latestErr := latestOJSVersion(ctx)
		if latestErr != nil {
			emit(map[string]any{"error": "Could not determine latest OJS version: " + latestErr.Error()})
			return
		}
		dotted = latest.Dotted
	} else {
		resolved, resolveErr := findOJSVersion(ctx, version)
		if resolveErr != nil {
			emit(map[string]any{"error": resolveErr.Error()})
			return
		}
		dotted = resolved.Dotted
	}

	archiveName := "ojs-" + dotted + ".tar.gz"
	archiveDir := "/etc/openpanel/ojs/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := ojsDownloadURL(dotted)
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-O", archivePath, downloadURL).Run(); runErr != nil {
			_ = os.Remove(archivePath)
			emit(map[string]any{"error": "Error downloading OJS " + dotted + ": " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Extracting files to " + approotContainerPath})
	if unpackErr := unpackOJSArchive(ctx, archivePath, approotHostPath); unpackErr != nil {
		emit(map[string]any{"error": "Error extracting OJS archive: " + unpackErr.Error()})
		_ = os.RemoveAll(approotHostPath)
		return
	}

	if mkErr := os.MkdirAll(filesHostPath, 0o755); mkErr != nil {
		emit(map[string]any{"error": "Error creating OJS files directory: " + mkErr.Error()})
		_ = os.RemoveAll(approotHostPath)
		return
	}

	// Host-side chown, not `podman exec ... chown` - the archive was
	// extracted host-side, so the files are owned by this process's own
	// (real, unmapped) UID, and a rootless container's own "root" cannot
	// chown files it doesn't already own outside its own user-namespace UID
	// range (same gotcha documented in moodle/install.go and
	// joomla/install.go, both fixed the same way).
	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	uid, uidErr := podmanmanager.GetUID(userContext)
	if uidErr != nil {
		emit(map[string]any{"error": "Could not determine file owner: " + uidErr.Error()})
		_ = os.RemoveAll(approotHostPath)
		_ = os.RemoveAll(filesHostPath)
		return
	}
	uidStr := strconv.Itoa(uid)
	_ = exec.Command("chown", "-R", uidStr+":"+uidStr, approotHostPath).Run()
	_ = exec.Command("chown", "-R", uidStr+":"+uidStr, filesHostPath).Run()

	emit(map[string]any{"status": "Linking web root to OJS app directory"})
	_ = os.Remove(hostOSPath)
	// The symlink target must be the container-visible path
	// (/var/www/html/...), not approotHostPath's host filesystem path - the
	// symlink is created host-side but read by the php-fpm container, which
	// only sees its own /var/www/html/ bind mount (same gotcha documented
	// live in moodle/install.go's identical symlink call).
	if symErr := os.Symlink(approotContainerPath, hostOSPath); symErr != nil {
		emit(map[string]any{"error": "Error creating web root symlink: " + symErr.Error()})
		_ = os.RemoveAll(approotHostPath)
		_ = os.RemoveAll(filesHostPath)
		return
	}

	mysqlVersion := mysql.GetMySQLVersion(ctx, a, userContext)
	emit(map[string]any{"status": "Testing database connection.."})
	if !mysql.CheckMySQLInsideContainer(ctx, userContext, true) {
		emit(map[string]any{"status": "Checking " + mysqlVersion + " container status.."})
		if !mysql.CheckMySQLNotTemporary(ctx, userContext, mysqlVersion) {
			emit(map[string]any{"error": "The " + mysqlVersion + " container is either not running or still initializing. Please ensure your plan has sufficient resources to start the service."})
			emitCleanupFiles(hostOSPath, approotHostPath, filesHostPath, emit)
			return
		}
	}

	if mysql.DatabaseLimitReached(ctx, a, userID, currentUsername, userContext) {
		emit(map[string]any{"error": "You have reached the maximum number of databases allowed on your plan." + a.UpgradeMessageForUser(ctx, userID)})
		emitCleanupFiles(hostOSPath, approotHostPath, filesHostPath, emit)
		return
	}

	emit(map[string]any{"status": "Creating database " + dbName + " and user " + dbUser})
	const dbHostGrant = "%"
	queries := []string{
		"CREATE DATABASE IF NOT EXISTS `" + dbName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		"CREATE USER IF NOT EXISTS '" + dbUser + "'@'" + dbHostGrant + "' IDENTIFIED BY '" + dbPassword + "'",
		"GRANT ALL PRIVILEGES ON `" + dbName + "`.* TO '" + dbUser + "'@'%'",
		"FLUSH PRIVILEGES",
	}
	for _, q := range queries {
		if _, execErr := mysqlmanager.Exec(ctx, userContext, q, ""); execErr != nil {
			invalidateMySQLCaches(ctx, a, userContext, currentUsername)
			emit(map[string]any{"error": "Error creating MySQL database and user: " + execErr.Error()})
			emitCleanupFiles(hostOSPath, approotHostPath, filesHostPath, emit)
			emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHostGrant, emit)
			return
		}
	}
	invalidateMySQLCaches(ctx, a, userContext, currentUsername)

	// OJS's Composer platform check fails hard (before install.php even
	// reads its stdin answers) if the "ftp" PHP extension isn't present -
	// confirmed live against a real php-fpm-8.5 image that didn't ship it.
	// Installing it here, right before the CLI installer runs, means the
	// install form doesn't need its own "missing extension" precheck/error
	// path - it just self-heals on first install.
	emit(map[string]any{"status": "Ensuring PHP 'ftp' extension is installed"})
	if extErr := php.EnsureExtensionInstalled(ctx, userContext, phpContainer, "ftp"); extErr != nil {
		emit(map[string]any{"error": "Could not install required PHP extension 'ftp': " + extErr.Error()})
		emitCleanupFiles(hostOSPath, approotHostPath, filesHostPath, emit)
		emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHostGrant, emit)
		return
	}

	emit(map[string]any{"status": "Running OJS CLI installer (tools/install.php)"})
	answers := buildOJSInstallAnswers(filesContainerPath, adminUser, adminPassword, adminEmail, mysqlVersion, dbUser, dbPassword, dbName, "oai:"+selectedDomain)
	out, runErr := runOJSInstaller(ctx, userContext, phpContainer, approotContainerPath, answers)
	if runErr != nil {
		emit(map[string]any{"error": "OJS CLI installer failed: " + strings.TrimSpace(out)})
		emitCleanupFiles(hostOSPath, approotHostPath, filesHostPath, emit)
		emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHostGrant, emit)
		return
	}

	emit(map[string]any{"status": "Finalizing configuration (base_url, time_zone)"})
	if fixErr := fixUpOJSConfig(filepath.Join(approotHostPath, "config.inc.php"), selectedDomain); fixErr != nil {
		emit(map[string]any{"status": "Warning: OJS installed, but base_url could not be finalized: " + fixErr.Error()})
	}

	emit(map[string]any{"status": "Setting permissions on OJS files directory"})
	_ = exec.Command("chmod", "-R", "777", filesHostPath).Run()

	emit(map[string]any{"status": "Deploying admin login helper"})
	loginFilePath := filepath.Join(approotHostPath, openpanelLoginFileName)
	if writeErr := os.WriteFile(loginFilePath, []byte(openpanelLoginPHP), 0o644); writeErr == nil {
		_ = os.Chown(loginFilePath, uid, uid)
	}

	emit(map[string]any{"status": "Registering cron job (lib/pkp/tools/scheduler.php, every minute)"})
	cronComment := ojsCronComment(selectedDomain)
	cronCommand := "php " + approotContainerPath + "/lib/pkp/tools/scheduler.php run"
	if cronErr := crons.AddJob(ctx, userContext, cronComment, "0 * * * * *", phpContainer, cronCommand, true); cronErr != nil {
		emit(map[string]any{"status": "Warning: OJS installed, but the scheduled-tasks cron job could not be registered automatically: " + cronErr.Error() + " - add it manually from Cron Jobs: schedule '0 * * * * *', container '" + phpContainer + "', command '" + cronCommand + "'."})
	}

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, dotted, "ojs"); insertErr != nil {
		emit(map[string]any{"error": "OJS installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed OJS on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "OJS installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "OJS installation completed!"})
}

func invalidateMySQLCaches(ctx context.Context, a *appctx.App, userContext, currentUsername string) {
	_ = a.Cache.Delete(ctx, "databases_info:"+userContext)
	_ = a.Cache.Delete(ctx, "get_database_count:"+currentUsername)
}

// emitCleanupFiles removes a failed install's docroot symlink and its
// backing approot/files directories - always safe since install.go just
// created all three.
func emitCleanupFiles(hostOSPath, approotHostPath, filesHostPath string, emit func(map[string]any)) {
	_ = os.Remove(hostOSPath)
	_ = os.RemoveAll(approotHostPath)
	_ = os.RemoveAll(filesHostPath)
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
