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

// handleInstallPage renders the install form / checks the plan's site limit for a GET, and hands POST off to handleInstallStream
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

// ensureContainerRunning starts the container if it isn't already running and polls briefly for it to come up, mirrors every other CMS module's identical helper
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

// unpackOJSArchive extracts PKP's packaged tarball into destDir, stripping the top-level "ojs-{version}/" wrapper so destDir itself becomes the app root/web root
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

// buildOJSInstallAnswers builds the newline-joined stdin block tools/install.php expects, in prompt order - there's no --flag CLI form like Moodle/WordPress's installers, only interactive stdin, and no timeZone prompt exists here even though the web form has one
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

// runOJSInstaller execs tools/install.php inside phpContainer, feeding it buildOJSInstallAnswers' stdin block - needs "-i" on podman exec so stdin actually pipes into the container, same as backups.go's mysql import
func runOJSInstaller(ctx context.Context, userContext, phpContainer, approotContainerPath, answers string) (string, error) {
	argv := podmanmanager.PodmanArgv(userContext, "exec", "-i", phpContainer, "php", approotContainerPath+"/tools/install.php")
	cmd := podmanmanager.Command(ctx, userContext, argv)
	cmd.Stdin = strings.NewReader(answers)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// fixUpOJSConfig rewrites base_url and time_zone in the freshly-installed config.inc.php - both are left wrong/blank by tools/install.php when run non-interactively (base_url has no real HTTP request to derive from, timeZone is never prompted for on the CLI path)
func fixUpOJSConfig(configPath, selectedDomain string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	text := string(content)
	text = iniBaseURLRE.ReplaceAllString(text, iniQuoted("base_url", "https://"+selectedDomain))
	text = iniTimeZoneRE.ReplaceAllString(text, iniQuoted("time_zone", "UTC"))
	// the CLI installer always writes allowed_hosts non-empty, which locks the site out with "400 Server host not allowed" unless reset back to the empty-string default
	text = iniAllowedHostsRE.ReplaceAllString(text, iniQuoted("allowed_hosts", ""))
	return os.WriteFile(configPath, []byte(text), 0o644)
}

// handleInstallStream drives an OJS install end to end over NDJSON: download the tarball, extract into a sibling app-root dir (see ojs.go), symlink docroot to it, create a sibling files dir, create the DB, run tools/install.php via piped stdin, fix up base_url/time_zone, deploy the autologin helper, and register the per-minute scheduler.php cron job
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

	// host-side chown, not podman exec chown - a rootless container's root can't chown files outside its UID mapping, same gotcha as moodle/joomla install.go
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
	// the symlink target must be the container-visible path, not the host filesystem path, since php-fpm only sees its own /var/www/html/ bind mount - same gotcha as moodle/install.go
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

	// OJS's Composer platform check fails hard if the "ftp" PHP extension isn't present, so install it here rather than adding a precheck/error path to the form
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

// emitCleanupFiles removes a failed install's docroot symlink and its backing approot/files directories, always safe since install.go just created all three
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
