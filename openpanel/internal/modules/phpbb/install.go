package phpbb

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
)

// phpbbVersion is the fallback used when the install form doesn't supply
// one (e.g. a direct API call) - confirmed live against download.phpbb.com
// at the time this module was written. The install page itself offers a
// real version picker populated from https://api.github.com/repos/phpbb/
// phpbb/tags (tag names are "release-x.y.z"), same pattern as Drupal/
// Joomla/Moodle/OpenCart's install pages, so this constant only matters
// as a last-resort default, not as the sole source of truth.
const phpbbVersion = "3.3.17"

// phpbbVersionRE validates a user-supplied phpbb_version form value before
// it's interpolated into the download URL/extract shell script - phpBB's
// releases are plain x.y.z (no "v" prefix, no pre-release suffixes for
// stable tags), confirmed against the real tags feed.
var phpbbVersionRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// phpbbSourceTarball builds the release download URL for version - phpBB
// tags are date-less x.y.z releases at
// download.phpbb.com/pub/release/<x.y>/<x.y.z>/phpBB-<x.y.z>.tar.bz2,
// confirmed live. The tarball's top-level directory is always literally
// "phpBB3/" (not version-suffixed), regardless of release.
func phpbbSourceTarball(version string) string {
	majorMinor := version
	if idx := strings.LastIndex(version, "."); idx != -1 {
		majorMinor = version[:idx]
	}
	return "https://download.phpbb.com/pub/release/" + majorMinor + "/" + version + "/phpBB-" + version + ".tar.bz2"
}

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

func isValidSubdirectory(subdirectory string) bool {
	if subdirectory == "" {
		return true
	}
	return !strings.Contains(subdirectory, "..") && !strings.HasPrefix(subdirectory, "/")
}

// handleInstallStream drives a phpBB install end to end, streaming NDJSON
// progress events: download+extract the release tarball, create a MySQL
// database, run phpBB's own dedicated CLI installer
// (install/phpbbcli.php install <yaml>) against it, delete the install/
// directory (phpBB's own documented post-install security step), then
// record the site.
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

	version := strings.TrimPrefix(strings.TrimSpace(r.FormValue("phpbb_version")), "release-")
	if version == "" {
		version = phpbbVersion
		if latest, verErr := latestPhpbbVersion(ctx); verErr == nil {
			version = latest
		}
	}
	if !phpbbVersionRE.MatchString(version) {
		emit(map[string]any{"error": "Invalid phpBB version."})
		return
	}

	boardName := formOr(r, "board_name", "My Board")
	boardDescription := formOr(r, "board_description", "Powered by phpBB")
	adminUsername := formOr(r, "admin_username", "admin")
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = generateRandomString(16)
	}
	adminEmail := formOr(r, "admin_email", "admin@"+dom.DomainURL)

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "phpbb_" + strings.ToLower(generateRandomString(6))
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

	emit(map[string]any{"status": "Downloading and extracting phpBB"})
	out, runErr := runPhpbbExtract(ctx, userContext, phpContainer, installPath, version)
	if runErr != nil {
		emit(map[string]any{"error": "Download/extract failed: " + strings.TrimSpace(string(out))})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "chown", "-R", uidStr+":"+uidStr, installPath)).Run()
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

	emit(map[string]any{"status": "Running phpBB installer"})
	out, runErr = runPhpbbInstaller(ctx, userContext, phpContainer, installPath, phpbbInstallParams{
		dbHost: mysqlVersion, dbName: dbName, dbUser: dbUser, dbPassword: dbPassword,
		boardName: boardName, boardDescription: boardDescription,
		adminUsername: adminUsername, adminPassword: adminPassword, adminEmail: adminEmail,
		serverName: selectedDomain,
	})
	if runErr != nil {
		emit(map[string]any{"error": "phpBB installer failed: " + strings.TrimSpace(string(out))})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}
	if !strings.Contains(string(out), "finished successfully") {
		emit(map[string]any{"error": "phpBB installer did not report success: " + strings.TrimSpace(string(out))})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	// phpBB's own docs recommend deleting install/ once setup is done -
	// it's dead weight afterward (the regular bin/phpbbcli.php refuses to
	// run until installed, and the browser installer explicitly refuses
	// to run again once PHPBB_INSTALLED is set in config.php, so nothing
	// in the running site ever needs this directory again).
	emit(map[string]any{"status": "Removing install/ directory"})
	rmInstallArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "rm", "-rf", installPath+"/install")
	_ = podmanmanager.Command(ctx, userContext, rmInstallArgv).Run()

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "phpbb"); insertErr != nil {
		emit(map[string]any{"error": "phpBB installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed phpBB on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "phpBB installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "phpBB installation completed!", "admin_user": adminUsername, "admin_password": adminPassword})
}

// runPhpbbExtract downloads the release tarball and extracts it into
// installPath. The tarball wraps everything in a fixed "phpBB3/" top-level
// directory (confirmed live, not version-suffixed like DokuWiki's), so
// this extracts to a scratch location first and copies that directory's
// *contents* into installPath - same `cp -a` reasoning as every other
// module here (a root, no-subdirectory install's installPath already
// exists as the domain's docroot, even though empty).
func runPhpbbExtract(ctx context.Context, userContext, phpContainer, installPath, version string) ([]byte, error) {
	scratch := "/tmp/openpanel-phpbb-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	script := `set -e
rm -rf ` + scratch + ` ` + scratch + `.tar.bz2
mkdir -p ` + scratch + `
curl -sL -o ` + scratch + `.tar.bz2 ` + phpbbSourceTarball(version) + `
tar -xjf ` + scratch + `.tar.bz2 -C ` + scratch + `
mkdir -p ` + installPath + `
cp -a ` + scratch + `/phpBB3/. ` + installPath + `/
rm -rf ` + scratch + ` ` + scratch + `.tar.bz2`

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script)
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
}

type phpbbInstallParams struct {
	dbHost, dbName, dbUser, dbPassword       string
	boardName, boardDescription              string
	adminUsername, adminPassword, adminEmail string
	serverName                               string
}

// escapeYAMLSingleQuoted doubles single quotes, the YAML 1.1 single-quoted
// scalar escape - safe for any of the plain string values written into
// the installer config below (no other characters need escaping inside a
// single-quoted YAML scalar).
func escapeYAMLSingleQuoted(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

// runPhpbbInstaller finishes a phpBB install via phpBB's own dedicated CLI
// installer app (install/phpbbcli.php's "install" command - NOT the
// regular bin/phpbbcli.php, which refuses to run at all until phpBB is
// already installed). Confirmed live against a real extracted copy: this
// is a full, non-interactive equivalent of the browser setup wizard - it
// creates the schema, writes config.php, creates the admin user, and
// reports "The installer has finished successfully" on completion. Takes
// a YAML config matching phpbb\install\installer_configuration's schema
// (verified against that class's source rather than guessed).
func runPhpbbInstaller(ctx context.Context, userContext, phpContainer, installPath string, p phpbbInstallParams) ([]byte, error) {
	yamlConfig := "installer:\n" +
		"    admin:\n" +
		"        name: '" + escapeYAMLSingleQuoted(p.adminUsername) + "'\n" +
		"        password: '" + escapeYAMLSingleQuoted(p.adminPassword) + "'\n" +
		"        email: '" + escapeYAMLSingleQuoted(p.adminEmail) + "'\n" +
		"    board:\n" +
		"        lang: en\n" +
		"        name: '" + escapeYAMLSingleQuoted(p.boardName) + "'\n" +
		"        description: '" + escapeYAMLSingleQuoted(p.boardDescription) + "'\n" +
		"    database:\n" +
		"        dbms: mysqli\n" +
		"        dbhost: '" + escapeYAMLSingleQuoted(p.dbHost) + "'\n" +
		"        dbport: null\n" +
		"        dbuser: '" + escapeYAMLSingleQuoted(p.dbUser) + "'\n" +
		"        dbpasswd: '" + escapeYAMLSingleQuoted(p.dbPassword) + "'\n" +
		"        dbname: '" + escapeYAMLSingleQuoted(p.dbName) + "'\n" +
		"        table_prefix: phpbb_\n" +
		"    server:\n" +
		"        cookie_secure: true\n" +
		"        server_protocol: 'https://'\n" +
		"        force_server_vars: true\n" +
		"        server_name: '" + escapeYAMLSingleQuoted(p.serverName) + "'\n" +
		"        server_port: 443\n" +
		"        script_path: /\n" +
		"    extensions: []\n"

	const yamlPath = "/tmp/openpanel-phpbb-install.yml"
	installerScript := installPath + "/install/phpbbcli.php"
	// The YAML is passed as "$1" (a discrete argv element, not
	// interpolated into the shell script text) so its content never needs
	// shell-quoting.
	script := `set -e; printf '%s' "$1" > ` + yamlPath + `; php ` + installerScript +
		` install ` + yamlPath + ` -n --no-ansi; rc=$?; rm -f ` + yamlPath + `; exit $rc`

	argv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script, "sh"), yamlConfig)
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
}

func invalidateMySQLCaches(ctx context.Context, a *appctx.App, userContext, currentUsername string) {
	_ = a.Cache.Delete(ctx, "databases_info:"+userContext)
	_ = a.Cache.Delete(ctx, "get_database_count:"+currentUsername)
}

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
