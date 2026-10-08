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
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/installmanifest"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/mysql"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/waf"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// phpbbVersion is the fallback used when the install form doesn't supply one (e.g. a direct API call) - the install page itself offers a real version picker populated from the GitHub tags API, so this only matters as a last-resort default
const phpbbVersion = "3.3.17"

// phpbbVersionRE validates a user-supplied phpbb_version form value before it's interpolated into the download URL/extract shell script - phpBB releases are plain x.y.z, no "v" prefix or pre-release suffixes
var phpbbVersionRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// phpbbSourceTarball builds the release download URL for version - the tarball's top-level directory is always literally "phpBB3/", not version-suffixed, regardless of release
func phpbbSourceTarball(version string) string {
	majorMinor := version
	if idx := strings.LastIndex(version, "."); idx != -1 {
		majorMinor = version[:idx]
	}
	return "https://download.phpbb.com/pub/release/" + majorMinor + "/" + version + "/phpBB-" + version + ".tar.bz2"
}

// handleInstallStream drives a phpBB install end to end over NDJSON: download+extract, create the DB, run install/phpbbcli.php install against it, delete the install/ directory (phpBB's documented post-install security step), then record the site
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

	boardName := web.FormOr(r, "board_name", "My Board")
	boardDescription := web.FormOr(r, "board_description", "Powered by phpBB")
	adminUsername := web.FormOr(r, "admin_username", "admin")
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = appkit.RandomString(16)
	}
	adminEmail := web.FormOr(r, "admin_email", "admin@"+dom.DomainURL)

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "phpbb_" + strings.ToLower(appkit.RandomString(6))
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

	emit(map[string]any{"status": "Downloading and extracting phpBB"})
	out, runErr := runPhpbbExtract(ctx, userContext, phpContainer, installPath, version)
	if runErr != nil {
		emit(map[string]any{"error": "Download/extract failed: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "chown", "-R", uidStr+":"+uidStr, installPath)).Run()
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

	emit(map[string]any{"status": "Running phpBB installer"})
	out, runErr = runPhpbbInstaller(ctx, userContext, phpContainer, installPath, phpbbInstallParams{
		dbHost: mysql.AppDBHost(userContext, mysqlVersion), dbName: dbName, dbUser: dbUser, dbPassword: dbPassword,
		boardName: boardName, boardDescription: boardDescription,
		adminUsername: adminUsername, adminPassword: adminPassword, adminEmail: adminEmail,
		serverName: selectedDomain,
	})
	if runErr != nil {
		emit(map[string]any{"error": "phpBB installer failed: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		cmsapp.EmitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}
	if !strings.Contains(string(out), "finished successfully") {
		emit(map[string]any{"error": "phpBB installer did not report success: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		cmsapp.EmitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	// phpBB's docs recommend deleting install/ once setup is done - it's dead weight afterward since nothing in the running site needs it again
	emit(map[string]any{"status": "Removing install/ directory"})
	rmInstallArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "rm", "-rf", installPath+"/install")
	_ = podmanmanager.Command(ctx, userContext, rmInstallArgv).Run()

	// record exactly what this install created so a later uninstall can remove precisely that, not the whole docroot
	_ = installmanifest.Record(hostOSPath)

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "phpbb"); insertErr != nil {
		emit(map[string]any{"error": "phpBB installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)
	waf.EnableProfileForNewSite(a, selectedDomain, "phpbb")

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed phpBB on domain "+selectedDomain, ipAddress)
	web.Flash(a, w, r, "success", web.Tr(a, r, "phpBB installed successfully on %(selected_domain)s", "selected_domain", selectedDomain))
	emit(map[string]any{"status": "phpBB installation completed!", "admin_user": adminUsername, "admin_password": adminPassword})
}

// runPhpbbExtract downloads the release tarball and extracts it into installPath - the tarball wraps everything in a fixed "phpBB3/" directory, so this extracts to a scratch location first and cp -a's that directory's contents into installPath, same reasoning as every other module here
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

// escapeYAMLSingleQuoted doubles single quotes, the YAML 1.1 single-quoted scalar escape - safe for the plain string values written into the installer config below
func escapeYAMLSingleQuoted(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

// runPhpbbInstaller finishes a phpBB install via install/phpbbcli.php's "install" command (not the regular bin/phpbbcli.php, which refuses to run until phpBB is already installed) - a full non-interactive equivalent of the browser setup wizard, taking a YAML config matching phpbb\install\installer_configuration's schema
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
	// the YAML is passed as "$1", a discrete argv element rather than interpolated into the shell script text, so its content never needs shell-quoting
	script := `set -e; printf '%s' "$1" > ` + yamlPath + `; php ` + installerScript +
		` install ` + yamlPath + ` -n --no-ansi; rc=$?; rm -f ` + yamlPath + `; exit $rc`

	argv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", script, "sh"), yamlConfig)
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
}
