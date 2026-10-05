package drupal

import (
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
	"gist.github.com/stefanpejcic/openpanel/internal/modules/waf"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleInstallStream drives a Drupal install end to end over NDJSON: create a Composer project (drupal/recommended-project), create a MySQL database, run `drush site:install`, then record the site
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

	siteName := web.FormOr(r, "site_name", "Drupal Site")
	drupalVersion := strings.TrimSpace(r.FormValue("drupal_version"))
	adminUsername := web.FormOr(r, "admin_username", "admin")
	adminPassword := r.FormValue("admin_password")
	if adminPassword == "" {
		adminPassword = appkit.RandomString(16)
	}
	adminEmail := web.FormOr(r, "admin_email", "admin@"+dom.DomainURL)

	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = "drupal_" + strings.ToLower(appkit.RandomString(6))
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

	projectConstraint := "drupal/recommended-project"
	if drupalVersion != "" && drupalVersion != "latest" {
		projectConstraint = "drupal/recommended-project:" + drupalVersion
	}

	// deliberately not pre-creating hostOSPath, composer create-project makes its own target dir (see phpapp/install.go's identical comment about a host-side mkdir racing composer over the rootless bind mount)
	emit(map[string]any{"status": "Creating Composer project " + projectConstraint})
	composerArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "composer"),
		"create-project", projectConstraint, installPath, "--no-interaction")
	out, runErr := podmanmanager.Command(ctx, userContext, composerArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "composer create-project failed: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	// drupal/recommended-project doesn't bundle Drush (hasn't since Drupal 9), has to be required explicitly before site:install below
	emit(map[string]any{"status": "Requiring drush/drush"})
	requireDrushArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "composer"),
		"--working-dir="+installPath, "require", "drush/drush", "--no-interaction")
	out, runErr = podmanmanager.Command(ctx, userContext, requireDrushArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "composer require drush/drush failed: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		return
	}

	// composer doesn't reliably leave vendor/bin/* executable here (came out 644 not 755, breaking every later drush call including the manager page's autologin button), so force chmod the whole dir plus drush's real binary rather than trusting composer's bin-dir handling
	chmodDrushArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "find",
		installPath+"/vendor/bin", installPath+"/vendor/drush/drush/drush", "-type", "f", "-exec", "chmod", "+x", "{}", "+")
	_, _ = podmanmanager.Command(ctx, userContext, chmodDrushArgv).CombinedOutput()

	// drupal/recommended-project's docroot is web/, not the project root, but OpenPanel's docroot always maps to installPath - symlinking web/'s entries up into installPath (rather than moving them) makes it servable without breaking web/'s own relative includes, since PHP resolves __DIR__/__FILE__ against the symlink target
	emit(map[string]any{"status": "Linking web root into docroot"})
	linkScript := `cd "$1/web" && for f in .[!.]* ..?* *; do [ -e "$f" ] || continue; ln -sfn "web/$f" "$1/$f"; done`
	linkArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", linkScript, "sh"), installPath)
	out, runErr = podmanmanager.Command(ctx, userContext, linkArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "Linking web root into docroot failed: " + strings.TrimSpace(string(out))})
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

	emit(map[string]any{"status": "Running drush site:install"})
	dbURL := "mysql://" + dbUser + ":" + dbPassword + "@" + mysql.AppDBHost(userContext, mysqlVersion) + "/" + dbName
	// absolute path, not "vendor/bin/drush" - podman exec's cwd is the container's default workdir, not installPath, so a relative path resolves wrong for subdirectory/non-root installs
	drushArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, installPath+"/vendor/bin/drush"),
		"site:install", "standard",
		"--db-url="+dbURL,
		"--site-name="+siteName,
		"--account-name="+adminUsername,
		"--account-pass="+adminPassword,
		"--account-mail="+adminEmail,
		"--root="+installPath,
		"-y")
	out, runErr = podmanmanager.Command(ctx, userContext, drushArgv).CombinedOutput()
	if runErr != nil {
		emit(map[string]any{"error": "drush site:install failed: " + strings.TrimSpace(string(out))})
		cmsapp.EmitCleanupFiles(ctx, userContext, phpContainer, installPath, emit)
		cmsapp.EmitCleanupDatabase(ctx, userContext, dbName, dbUser, dbHost, emit)
		return
	}

	version := drupalVersion
	if version == "" {
		version = "latest"
	}

	// record exactly what this install created so a later uninstall can remove precisely that, not the whole docroot
	_ = installmanifest.Record(hostOSPath)

	emit(map[string]any{"status": "Saving website information to Site Manager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, version, "drupal"); insertErr != nil {
		emit(map[string]any{"error": "Drupal installed, but an error occurred while saving to Site Manager: " + insertErr.Error()})
		return
	}
	websites.TriggerScreenshotGeneration(a, selectedDomain)
	waf.EnableProfileForNewSite(a, selectedDomain, "drupal")

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed Drupal on domain "+selectedDomain, ipAddress)
	web.Flash(a, w, r, "success", web.Tr(a, r, "Drupal installed successfully on %(selected_domain)s", "selected_domain", selectedDomain))
	emit(map[string]any{"status": "Drupal installation completed!"})
}
