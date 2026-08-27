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
// polling briefly for it to come up.
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

// handleInstallStream drives a Flarum install end to end, streaming NDJSON
// progress events to the client as each step completes: create a Composer
// project (flarum/flarum), create a MySQL database, run Flarum's own
// console installer (see runFlarumInstaller's doc comment), then record
// the site.
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

	forumTitle := formOr(r, "site_name", "Flarum")
	flarumVersion := strings.TrimSpace(r.FormValue("flarum_version"))
	adminUsername := formOr(r, "admin_username", "admin")
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = generateRandomString(16)
	}
	adminEmail := formOr(r, "admin_email", "admin@"+dom.DomainURL)

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "flarum_" + strings.ToLower(generateRandomString(6))
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

	// Flarum 2.x (what projectConstraint below asks composer for) requires
	// PHP 8.1+. Composer silently falls back to the newest flarum/flarum
	// release that DOES satisfy the domain's actual PHP version instead of
	// failing outright, so an old PHP version doesn't surface as an
	// install error here - it surfaces later as flarum/core 1.x getting
	// installed with a completely different (and untested by this module)
	// console installer, leaving config.php never written. Failing fast
	// here with a clear message beats that silent, confusing downgrade.
	if !isLitespeed && phpVersionBelow(phpVersion, 8, 1) {
		emit(map[string]any{"error": "Flarum requires PHP 8.1 or newer, but this domain is set to PHP " + phpVersion + ". Change the domain's PHP version (or install into a subdirectory using PHP 8.1+) and try again."})
		return
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

	projectConstraint := "flarum/flarum:^2.0.0"
	if flarumVersion != "" && flarumVersion != "latest" {
		projectConstraint = "flarum/flarum:^" + strings.TrimPrefix(flarumVersion, "v")
	}

	// Deliberately not pre-creating hostOSPath: composer create-project
	// makes its own target directory (see drupal/install.go's identical
	// comment about a host-side mkdir racing composer's own directory
	// creation over the rootless bind mount).
	emit(map[string]any{"status": "Creating Composer project " + projectConstraint})
	composerArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "composer"),
		"create-project", projectConstraint, installPath, "--stability=beta", "--no-interaction")
	out, runErr := podmanmanager.Command(ctx, userContext, composerArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "composer create-project failed: " + strings.TrimSpace(string(out))})
		return
	}

	// flarum/flarum's docroot is the public/ subdirectory, not the composer
	// project root - same shape as Drupal's web/ subdirectory quirk. We
	// symlink public/'s entries up into installPath so installPath itself
	// is servable, EXCEPT index.php: it does a literal `require
	// '../site.php'`, which PHP resolves relative to the entry script's
	// path as given by SCRIPT_FILENAME (the symlink's own location, not
	// the symlinked-to real path) - so a plain symlink makes it look one
	// directory too high and 500s on every request (confirmed live via a
	// real install: "Failed opening required '../site.php'"). Instead we
	// write a tiny wrapper at installPath/index.php that requires the real
	// public/index.php by its real absolute path, so PHP resolves that
	// file's own relative require correctly against public/.
	emit(map[string]any{"status": "Linking public root into docroot"})
	linkScript := `cd "$1/public" && for f in .[!.]* ..?* *; do [ "$f" = "index.php" ] && continue; [ -e "$f" ] || continue; ln -sfn "public/$f" "$1/$f"; done
printf '%s\n' '<?php' 'chdir(__DIR__ . "/public"); require __DIR__ . "/public/index.php";' > "$1/index.php"`
	linkArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", linkScript, "sh"), installPath)
	out, runErr = podmanmanager.Command(ctx, userContext, linkArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "Linking public root into docroot failed: " + strings.TrimSpace(string(out))})
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

	emit(map[string]any{"status": "Running Flarum installer"})
	baseURL := "https://" + selectedDomain
	out, runErr = runFlarumInstaller(ctx, userContext, phpContainer, installPath, flarumInstallParams{
		dbHost: mysqlVersion, dbPort: 3306, dbName: dbName, dbUser: dbUser, dbPassword: dbPassword,
		baseURL: baseURL, forumTitle: forumTitle,
		adminUsername: adminUsername, adminPassword: adminPassword, adminEmail: adminEmail,
	})
	log.Printf("FLARUM - installer for %s exited (err=%v), output: %s", installPath, runErr, strings.TrimSpace(string(out)))
	if runErr != nil {
		emit(map[string]any{"error": "Flarum installer failed: " + strings.TrimSpace(string(out))})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}
	if !strings.Contains(string(out), "DONE") {
		emit(map[string]any{"error": "Flarum installer did not report success: " + strings.TrimSpace(string(out))})
		emitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		emitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	version := flarumVersion
	if version == "" {
		version = "latest"
	}

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "flarum"); insertErr != nil {
		emit(map[string]any{"error": "Flarum installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed Flarum on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "Flarum installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "Flarum installation completed!"})
}

// flarumInstallParams is runFlarumInstaller's input.
type flarumInstallParams struct {
	dbHost, dbUser, dbPassword, dbName       string
	dbPort                                   int
	baseURL, forumTitle                      string
	adminUsername, adminPassword, adminEmail string
}

// flarumInstallConfig is the --file=<json> schema Flarum's own
// Install\Console\FileDataProvider expects (confirmed against its source:
// it reads .debug, .baseUrl, .databaseConfiguration.{driver,host,port,
// database,username,password,prefix}, .adminUser.{username,password,email},
// .settings).
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

// runFlarumInstaller finishes a Flarum install via its own console
// installer: `php <installPath>/flarum install --file=<json> --config=...`.
//
// Two things this had to work around, found by testing against a live
// container rather than trusting docs.flarum.org (which only documents
// the browser wizard) or a stale doc summary (which missed a parameter):
//   - `install` IS a real, supported non-interactive command
//     (Flarum\Install\Console\InstallCommand, using FileDataProvider for
//     --file) - it's just not registered in the ConsoleServiceProvider
//     list that (only) applies to an already-installed site, so grepping
//     that list alone misses it.
//   - The container's `php` on PATH is a wrapper
//     (/usr/local/aliases/php in the shinsenter/php image) that does not
//     reliably preserve cwd for relative-path script resolution, so the
//     entry script must be invoked as an absolute path
//     (installPath+"/flarum", not a bare "flarum" after cd).
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
	// --config is resolved as base+"/"+value internally (confirmed by
	// testing against a live container: passing the already-absolute
	// installPath+"/config.php" here produced a doubled path,
	// ".../config.php//var/www/html/.../config.php", and StoreConfig's
	// file_put_contents failed on that bogus path) - so this must be a
	// bare relative filename, not the absolute path install.go uses
	// everywhere else.
	const configRelPath = "config.php"
	flarumScript := installPath + "/flarum"
	// The JSON is passed as "$1" (a discrete argv element, not interpolated
	// into the shell script text) so its content never needs shell-quoting.
	installScript := `set -e; printf '%s' "$1" > ` + jsonPath + `; php ` + flarumScript +
		` install --file=` + jsonPath + ` --config=` + configRelPath + `; rc=$?; rm -f ` + jsonPath + `; exit $rc`

	argv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", installScript, "sh"), string(cfgJSON))
	return podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
}

func invalidateMySQLCaches(ctx context.Context, a *appctx.App, userContext, currentUsername string) {
	_ = a.Cache.Delete(ctx, "databases_info:"+userContext)
	_ = a.Cache.Delete(ctx, "get_database_count:"+currentUsername)
}

// emitCleanupFiles removes a failed install's partially-created directory.
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

// phpVersionBelow reports whether version (e.g. "7.2", "8.5") is older
// than wantMajor.wantMinor. An unparseable version is treated as too old,
// so an unexpected format fails safe (blocks the install with a clear
// message) rather than silently proceeding.
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

func isValidSubdirectory(subdirectory string) bool {
	if subdirectory == "" {
		return true
	}
	return !strings.Contains(subdirectory, "..") && !strings.HasPrefix(subdirectory, "/")
}
