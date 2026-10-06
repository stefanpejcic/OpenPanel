package flarum

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

// handleInstallStream drives a Flarum install end to end over NDJSON: create a Composer project (flarum/flarum), create a MySQL database, run Flarum's own console installer (see runFlarumInstaller's doc comment), then record the site
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

	forumTitle := web.FormOr(r, "site_name", "Flarum")
	flarumVersion := strings.TrimSpace(r.FormValue("flarum_version"))
	adminUsername := web.FormOr(r, "admin_username", "admin")
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = appkit.RandomString(16)
	}
	adminEmail := web.FormOr(r, "admin_email", "admin@"+dom.DomainURL)

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "flarum_" + strings.ToLower(appkit.RandomString(6))
	}
	dbUser := strings.ToLower(r.FormValue("db_user"))
	if dbUser == "" {
		dbUser = strings.ToLower(appkit.RandomString(10))
	}
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

	// only flarum/core 2.x requires PHP 8.1+ (1.x, what "latest" resolves to since no stable 2.0.0 has shipped, needs only PHP >=7.3) - this guard only applies when the user explicitly types a 2.x version, since otherwise composer would silently fall back to 1.x on old PHP instead of failing
	if !isLitespeed && strings.HasPrefix(strings.TrimPrefix(flarumVersion, "v"), "2.") && phpVersionBelow(phpVersion, 8, 1) {
		emit(map[string]any{"error": "Flarum 2.x requires PHP 8.1 or newer, but this domain is set to PHP " + phpVersion + ". Change the domain's PHP version (or install into a subdirectory using PHP 8.1+) and try again."})
		return
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

	projectConstraint := "flarum/flarum"
	if flarumVersion != "" && flarumVersion != "latest" {
		projectConstraint = "flarum/flarum:^" + strings.TrimPrefix(flarumVersion, "v")
	} else {
		// no explicit pin for "latest" - fetch the real latest *stable* numeric release rather than letting composer's own resolution run unconstrained, so this never silently picks up a future 2.0.0 pre-release the moment one ships
		if version, verErr := latestFlarumVersion(ctx); verErr == nil {
			// that's flarum/core's version, the flarum/flarum skeleton lags behind on patch releases so only pin major.minor
			if parts := strings.SplitN(version, ".", 3); len(parts) >= 2 {
				version = parts[0] + "." + parts[1]
			}
			projectConstraint = "flarum/flarum:^" + version
		}
	}

	// deliberately not pre-creating hostOSPath, composer create-project makes its own target dir (see drupal/install.go's identical comment about a host-side mkdir racing composer over the rootless bind mount)
	emit(map[string]any{"status": "Creating Composer project " + projectConstraint})
	composerArgv := append(composerExec(userContext, phpContainer, projectConstraint),
		"create-project", projectConstraint, installPath, "--no-interaction")
	out, runErr := podmanmanager.Command(ctx, userContext, composerArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "composer create-project failed: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	// flarum/flarum's docroot is public/, not the project root, same shape as Drupal's web/ quirk - we symlink public/'s entries up into installPath EXCEPT index.php, since its literal `require '../site.php'` resolves against the symlink's own location (SCRIPT_FILENAME) not the real path, landing one directory too high and 500ing on every request; instead we write a tiny wrapper at installPath/index.php that requires the real public/index.php by its absolute path, so its own relative require resolves correctly against public/
	emit(map[string]any{"status": "Linking public root into docroot"})
	linkScript := `cd "$1/public" && for f in .[!.]* ..?* *; do [ "$f" = "index.php" ] && continue; [ -e "$f" ] || continue; ln -sfn "public/$f" "$1/$f"; done
printf '%s\n' '<?php' 'chdir(__DIR__ . "/public"); require __DIR__ . "/public/index.php";' > "$1/index.php"`
	linkArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", linkScript, "sh"), installPath)
	out, runErr = podmanmanager.Command(ctx, userContext, linkArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "Linking public root into docroot failed: " + strings.TrimSpace(string(out))})
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

	emit(map[string]any{"status": "Running Flarum installer"})
	baseURL := "https://" + selectedDomain
	out, runErr = runFlarumInstaller(ctx, userContext, phpContainer, installPath, flarumInstallParams{
		dbHost: mysql.AppDBHost(userContext, mysqlVersion), dbPort: 3306, dbName: dbName, dbUser: dbUser, dbPassword: dbPassword,
		baseURL: baseURL, forumTitle: forumTitle,
		adminUsername: adminUsername, adminPassword: adminPassword, adminEmail: adminEmail,
	})
	log.Printf("FLARUM - installer for %s exited (err=%v), output: %s", installPath, runErr, strings.TrimSpace(string(out)))
	if runErr != nil {
		emit(map[string]any{"error": "Flarum installer failed: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		cmsapp.EmitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}
	if !strings.Contains(string(out), "DONE") {
		emit(map[string]any{"error": "Flarum installer did not report success: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		cmsapp.EmitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	emit(map[string]any{"status": "Registering cron job (flarum schedule:run, every minute)"})
	if cronErr := addScheduler(ctx, userContext, selectedDomain, phpContainer, installPath); cronErr != nil {
		emit(map[string]any{"status": "Warning: Flarum installed, but the scheduler cron job could not be registered: " + cronErr.Error() + " - turn it on from the Scheduler card on the site page."})
	}

	version := flarumVersion
	if version == "" {
		version = "latest"
	}

	// record exactly what this install created so a later uninstall can remove precisely that, not the whole docroot
	_ = installmanifest.Record(hostOSPath)

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "flarum"); insertErr != nil {
		emit(map[string]any{"error": "Flarum installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed Flarum on domain "+selectedDomain, ipAddress)
	web.Flash(a, w, r, "success", web.Tr(a, r, "Flarum installed successfully on %(selected_domain)s", "selected_domain", selectedDomain))
	emit(map[string]any{"status": "Flarum installation completed!"})
}

// flarumInstallParams is runFlarumInstaller's input.
type flarumInstallParams struct {
	dbHost, dbUser, dbPassword, dbName       string
	dbPort                                   int
	baseURL, forumTitle                      string
	adminUsername, adminPassword, adminEmail string
}

// flarumInstallConfig is the --file=<json> schema Flarum's own Install\Console\FileDataProvider expects (reads .debug, .baseUrl, .databaseConfiguration.{driver,host,port,database,username,password,prefix}, .adminUser.{username,password,email}, .settings)
type flarumInstallConfig struct {
	Debug                 bool                        `json:"debug"`
	BaseURL               string                      `json:"baseUrl"`
	DatabaseConfiguration flarumInstallDatabaseConfig `json:"databaseConfiguration"`
	AdminUser             flarumInstallAdminUser      `json:"adminUser"`
	Settings              flarumInstallSettings       `json:"settings"`
}

type flarumInstallDatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	Prefix   string `json:"prefix"`
}

type flarumInstallAdminUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type flarumInstallSettings struct {
	ForumTitle string `json:"forum_title"`
}

// runFlarumInstaller finishes a Flarum install via its own console installer: `php <installPath>/flarum install --file=<json> --config=...` - two things had to be worked around by testing against a live container rather than trusting docs.flarum.org's browser-wizard-only docs: `install` IS a real supported non-interactive command (Flarum\Install\Console\InstallCommand via FileDataProvider) but isn't registered in the ConsoleServiceProvider list that only applies to an already-installed site, so grepping that list alone misses it; and the container's `php` on PATH is a wrapper (/usr/local/aliases/php in shinsenter/php) that doesn't reliably preserve cwd for relative-path resolution, so the entry script must be an absolute path
func runFlarumInstaller(ctx context.Context, userContext, phpContainer, installPath string, p flarumInstallParams) ([]byte, error) {
	cfg := flarumInstallConfig{
		BaseURL: p.baseURL,
		DatabaseConfiguration: flarumInstallDatabaseConfig{
			Driver: "mysql", Host: p.dbHost, Port: p.dbPort, Database: p.dbName,
			Username: p.dbUser, Password: p.dbPassword,
		},
		AdminUser: flarumInstallAdminUser{
			Username: p.adminUsername, Password: p.adminPassword, Email: p.adminEmail,
		},
		Settings: flarumInstallSettings{ForumTitle: p.forumTitle},
	}
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	const jsonPath = "/tmp/openpanel-flarum-install.json"
	// --config is resolved as base+"/"+value internally - passing the already-absolute installPath+"/config.php" produced a doubled bogus path and StoreConfig's file_put_contents failed, so this must be a bare relative filename, not the absolute path install.go uses everywhere else
	const configRelPath = "config.php"
	flarumScript := installPath + "/flarum"
	// the JSON is passed as "$1" (a discrete argv element, not interpolated into the shell script text) so its content never needs shell-quoting
	installScript := `set -e; printf '%s' "$1" > ` + jsonPath + `; php ` + flarumScript +
		` install --file=` + jsonPath + ` --config=` + configRelPath + `; rc=$?; rm -f ` + jsonPath + `; exit $rc`

	argv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", installScript, "sh"), string(cfgJSON))
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
}

// phpVersionBelow reports whether version (e.g. "7.2") is older than wantMajor.wantMinor - an unparseable version fails safe as "too old" (blocks the install with a clear message) rather than silently proceeding
func phpVersionBelow(version string, wantMajor, wantMinor int) bool {
	major, minor := 0, 0
	if _, err := fmt.Sscanf(version, "%d.%d", &major, &minor); err != nil {
		return true
	}
	if major != wantMajor {
		return major < wantMajor
	}
	return minor < wantMinor
}

// composerExec is `podman exec <php> composer`, letting Flarum 1.x through composer's security blocking since every 1.x release requires league/flysystem 1.x, which carries advisories upstream won't fix
func composerExec(userContext, phpContainer, constraint string) []string {
	if strings.Contains(constraint, ":^2") {
		return podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "composer")
	}
	return podmanmanager.PodmanArgv(userContext, "exec", "-e", "COMPOSER_NO_SECURITY_BLOCKING=1", phpContainer, "composer")
}
