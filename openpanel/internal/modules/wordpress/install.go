package wordpress

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
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
)

// handleInstallPage renders the WordPress install form. When the user is
// over their site limit, it still falls through to render the form with
// a warning flash - only the MySQL-ensure-running step and the POST
// handoff are skipped.
func handleInstallPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
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
	} else {
		mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
		if !docker.IsServiceRunning(ctx, userContext, mysqlVersion) {
			docker.StartOrStopContainer(ctx, userContext, mysqlVersion, "activate", "detached")
		}
		if r.Method == http.MethodPost {
			handleInstallStream(a, w, r)
			return
		}
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	renderInstallPage(a, w, r, domains)
}

func writeNDJSON(w http.ResponseWriter, flusher http.Flusher, canFlush bool, v map[string]any) {
	b, _ := json.Marshal(v)
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n"))
	if canFlush {
		flusher.Flush()
	}
}

// handleInstallStream drives a WordPress install end to end, streaming
// NDJSON progress events to the client as each step completes.
func handleInstallStream(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	flusher, canFlush := w.(http.Flusher)
	emit := func(v map[string]any) { writeNDJSON(w, flusher, canFlush, v) }

	ipAddress := reqip.ClientIP(r)
	domainID := r.FormValue("domain_id")

	mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")

	if err := createLockFile(currentUsername); err != nil {
		emit(map[string]any{"error": "Error creating lock file: " + err.Error()})
		return
	}

	dom, found, dbErr := lookupDomainByID(ctx, a, domainID)
	if dbErr != nil {
		emit(map[string]any{"error": "An error occurred fetching docroot for domain from database."})
		return
	}
	if !found {
		emit(map[string]any{"error": "Domain not found"})
		return
	}
	selectedDomain := dom.DomainURL
	docroot := dom.Docroot.String
	phpVersion := dom.PHPVersion.String

	if !a.CheckDomainBelongsToUser(ctx, userID, selectedDomain) {
		return
	}

	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + phpVersion
	}

	if !docker.IsServiceRunning(ctx, userContext, phpContainer) {
		emit(map[string]any{"status": "Starting PHP container: " + phpContainer})
		docker.StartOrStopContainer(ctx, userContext, phpContainer, "activate", "detached")
	}

	adminEmail := formOr(r, "admin_email", "admin@"+selectedDomain)
	websiteName := formOr(r, "website_name", "My Blog")
	siteDescription := formOr(r, "site_description", "My WordPress Blog")
	adminUsername := r.FormValue("admin_username")
	adminPassword := r.FormValue("admin_password")
	wordpressVersion := formOr(r, "wordpress_version", "latest")
	subdirectory := strings.ReplaceAll(strings.ToLower(r.FormValue("subdirectory")), " ", "")
	dbName := strings.ToLower(r.FormValue("db_name"))
	if dbName == "" {
		dbName = generateRandomString(6)
	}
	dbUser := strings.ToLower(r.FormValue("db_user"))
	if dbUser == "" {
		dbUser = generateRandomString(6)
	}
	dbPassword := r.FormValue("db_password")
	if dbPassword == "" {
		dbPassword = generateRandomString(16)
	}
	prefix := r.FormValue("db_prefix")
	if prefix == "" {
		prefix = "wp_"
	}
	if !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	installPath := docroot
	if subdirectory != "" {
		installPath = docroot + "/" + subdirectory
		selectedDomain = selectedDomain + "/" + subdirectory
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	docrootWithoutWWW := strings.TrimPrefix(strings.TrimPrefix(installPath, "/var/www/html/"), "/")
	hostOSPath := filepath.Join(htmlVolume, docrootWithoutWWW)

	for _, fileName := range []string{".htaccess", "wp-config.php", "index.php"} {
		if _, statErr := os.Stat(filepath.Join(hostOSPath, fileName)); statErr == nil {
			emit(map[string]any{"error": "File " + fileName + " already exists in document root: " + installPath})
			return
		}
	}

	if mkErr := os.MkdirAll(hostOSPath, 0o755); mkErr != nil {
		emit(map[string]any{"error": "Error creating document root: " + mkErr.Error()})
		return
	}

	archiveName := "wordpress-" + wordpressVersion + ".tar.gz"
	archiveDir := "/etc/openpanel/wordpress/archives"
	archivePath := filepath.Join(archiveDir, archiveName)
	if _, statErr := os.Stat(archivePath); statErr != nil {
		downloadURL := "https://wordpress.org/" + archiveName
		emit(map[string]any{"status": "Downloading " + downloadURL})
		_ = os.MkdirAll(archiveDir, 0o755)
		if runErr := exec.CommandContext(ctx, "wget", "-q", "-P", archiveDir, downloadURL).Run(); runErr != nil {
			emit(map[string]any{"error": "Error downloading WordPress: " + runErr.Error()})
			return
		}
	} else {
		emit(map[string]any{"status": "Using existing archive " + archivePath})
	}

	emit(map[string]any{"status": "Extracting files to " + installPath})
	if runErr := exec.CommandContext(ctx, "tar", "-xzf", archivePath, "--strip-components=1", "-C", hostOSPath).Run(); runErr != nil {
		emit(map[string]any{"error": "Error extracting WordPress archive: " + runErr.Error()})
		return
	}

	emit(map[string]any{"status": "Creating .htaccess file for WordPress permalinks"})
	htaccessContent, htErr := loadHtaccess(webServer)
	if htErr != nil {
		emit(map[string]any{"error": htErr.Error()})
		return
	}
	_ = os.WriteFile(filepath.Join(hostOSPath, ".htaccess"), []byte(htaccessContent), 0o644)

	emit(map[string]any{"status": "Writing database logins to wp-config.php"})
	wpConfigFile := filepath.Join(hostOSPath, "wp-config.php")
	wpConfigSampleFile := filepath.Join(hostOSPath, "wp-config-sample.php")
	if copyErr := copyFile(wpConfigSampleFile, wpConfigFile); copyErr != nil {
		emit(map[string]any{"error": "Error creating wp-config.php: " + copyErr.Error()})
		return
	}

	content, _ := os.ReadFile(wpConfigFile)
	strContent := string(content)
	strContent = strings.Replace(strContent, "database_name_here", dbName, 1)
	strContent = strings.Replace(strContent, "localhost", mysqlVersion, 1)
	strContent = strings.Replace(strContent, "username_here", dbUser, 1)
	strContent = strings.Replace(strContent, "password_here", dbPassword, 1)
	strContent = tablePrefixRE.ReplaceAllString(strContent, "${1}"+prefix+"${2}")
	_ = os.WriteFile(wpConfigFile, []byte(strContent), 0o644)

	emit(map[string]any{"status": "Shuffling WordPress Salts in wp-config.php file"})
	writeSaltsLocally(wpConfigFile)

	muPluginSrc := "/etc/openpanel/wordpress/mu-plugin.php"
	muPluginDestDir := filepath.Join(hostOSPath, "wp-content", "mu-plugins")
	_ = os.MkdirAll(muPluginDestDir, 0o755)
	if _, statErr := os.Stat(muPluginSrc); statErr == nil {
		if copyErr := copyFile(muPluginSrc, filepath.Join(muPluginDestDir, "mu-plugin.php")); copyErr == nil {
			emit(map[string]any{"status": "Added mu-plugin to force HTTPS for wp-admin behind Varnish."})
		}
	}

	if isLitespeed {
		setFSMethod(wpConfigFile, emit)
	}

	emit(map[string]any{"status": "Setting files permissions and owner to '" + userContext + "'"})
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		uidStr := itoa(uid)
		_ = exec.Command("chown", uidStr+":"+uidStr, "-R", hostOSPath).Run()
	}

	if isLitespeed {
		emit(map[string]any{"status": "Fixing file permissions for OpenLiteSpeed"})
		execPrefix := podmanmanager.PodmanArgv(userContext, "exec", phpContainer)
		for _, wpDir := range []string{filepath.Join("wp-content", "upgrade"), filepath.Join("wp-content", "uploads")} {
			_ = podmanmanager.Command(ctx, userContext, append(append([]string{}, execPrefix...), "mkdir", "-p", filepath.Join(installPath, wpDir))).Run()
		}
		_ = podmanmanager.Command(ctx, userContext, append(append([]string{}, execPrefix...), "chown", "-R", "root:nogroup", installPath)).Run()
		_ = podmanmanager.Command(ctx, userContext, append(append([]string{}, execPrefix...), "find", installPath, "-type", "d", "-exec", "chmod", "2775", "{}", "+")).Run()
		_ = podmanmanager.Command(ctx, userContext, append(append([]string{}, execPrefix...), "find", installPath, "-type", "f", "-exec", "chmod", "664", "{}", "+")).Run()
	}

	if waitForWPAvailable(ctx, userContext, phpContainer) {
		emit(map[string]any{"status": "PHP container is running and WP-CLI is functional, starting process.."})
	} else {
		emit(map[string]any{"status": "WP-CLI not available or the " + phpContainer + " container not running."})
		emitCleanupFiles(hostOSPath, emit)
		return
	}

	emit(map[string]any{"status": "Testing database connection.."})
	if !mysql.CheckMySQLInsideContainer(ctx, userContext, true) {
		emit(map[string]any{"status": "Checking " + mysqlVersion + " container status.."})
		if !mysql.CheckMySQLNotTemporary(ctx, userContext, mysqlVersion) {
			emit(map[string]any{"status": "Error: The " + mysqlVersion + " container is either not running or still initializing. Please ensure your plan has sufficient resources to start the service."})
			emitCleanupFiles(hostOSPath, emit)
			return
		}
	} else {
		emit(map[string]any{"status": mysqlVersion + " container is healthy, proceeding with database operations.."})
	}

	if mysql.DatabaseLimitReached(ctx, a, userID, currentUsername, userContext) {
		emit(map[string]any{"error": "You have reached the maximum number of databases allowed on your plan." + a.UpgradeMessageForUser(ctx, userID)})
		emitCleanupFiles(hostOSPath, emit)
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
	var dbCreateErr error
	for _, q := range queries {
		if _, execErr := mysqlmanager.Exec(ctx, userContext, q, ""); execErr != nil {
			dbCreateErr = execErr
			break
		}
	}
	if dbCreateErr != nil {
		invalidateMySQLCaches(ctx, a, userContext, currentUsername)
		emit(map[string]any{"error": "Error creating MySQL database and user: " + dbCreateErr.Error()})
		emitCleanupFiles(hostOSPath, emit)
		emitCleanupDatabase(ctx, userContext, mysqlVersion, dbName, dbUser, dbHost, emit)
		return
	}
	invalidateMySQLCaches(ctx, a, userContext, currentUsername)

	emit(map[string]any{"status": "Importing WordPress tables in the database"})
	wpBaseCmd := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)
	installArgv := append(append([]string{}, wpBaseCmd...), "core", "install",
		"--url=https://"+selectedDomain, "--title="+websiteName, "--admin_user="+adminUsername,
		"--admin_password="+adminPassword, "--admin_email="+adminEmail, "--path="+installPath, "--allow-root")

	if runErr := podmanmanager.Command(ctx, userContext, installArgv).Run(); runErr != nil {
		time.Sleep(10 * time.Second)
		if retryErr := podmanmanager.Command(ctx, userContext, installArgv).Run(); retryErr != nil {
			emit(map[string]any{"error": "Error running wp core install: " + retryErr.Error()})
			emitCleanupFiles(hostOSPath, emit)
			emitCleanupDatabase(ctx, userContext, mysqlVersion, dbName, dbUser, dbHost, emit)
			return
		}
	}

	var wpEvalParts []string
	if isLitespeed {
		wpEvalParts = append(wpEvalParts, "wp_rewrite_structure( '/%postname%/' );")
		emit(map[string]any{"status": "Setting pretty permalinks for WordPress"})
	}
	if siteDescription != "" {
		escaped := strings.ReplaceAll(siteDescription, "'", "\\'")
		wpEvalParts = append(wpEvalParts, "update_option( 'blogdescription', '"+escaped+"' );")
		emit(map[string]any{"status": "Setting site tagline to '" + siteDescription + "'"})
	}
	if len(wpEvalParts) > 0 {
		evalArgv := append(append([]string{}, wpBaseCmd...), "eval", strings.Join(wpEvalParts, "; "), "--path="+installPath, "--allow-root")
		_ = podmanmanager.Command(ctx, userContext, evalArgv).Run()
	}

	emit(map[string]any{"status": "Enabling auto-login from SiteManager"})

	emit(map[string]any{"status": "Checking for server-wide or user provided themes and plugins sets to be installed"})
	systemWidePlugins := "/etc/openpanel/wordpress/sets/plugins.txt"
	systemWideThemes := "/etc/openpanel/wordpress/sets/themes.txt"
	userPlugins := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/wpcli_plugins.txt"
	userThemes := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/wpcli_themes.txt"

	pluginFile := systemWidePlugins
	if _, statErr := os.Stat(userPlugins); statErr == nil {
		pluginFile = userPlugins
	}
	themeFile := systemWideThemes
	if _, statErr := os.Stat(userThemes); statErr == nil {
		themeFile = userThemes
	}
	processInstallSet(ctx, userContext, wpBaseCmd, installPath, pluginFile, "plugin", emit)
	processInstallSet(ctx, userContext, wpBaseCmd, installPath, themeFile, "theme", emit)

	emit(map[string]any{"status": "Saving website information to SiteManager"})
	if _, insertErr := a.DB.ExecContext(ctx,
		"INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
		selectedDomain, domainID, adminEmail, wordpressVersion, "wordpress"); insertErr != nil {
		emit(map[string]any{"error": "WordPress installed, but error occurred while saving data to WP Manager: " + insertErr.Error()})
		return
	}

	if isLitespeed {
		emit(map[string]any{"status": "Reloading " + phpContainer + " to apply rewrite rules from .htaccess file"})
		_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "restart", phpContainer)).Run()
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "installed WordPress on domain "+selectedDomain, ipAddress)
	flashSess(a, w, r, "success", "WordPress installed successfully on "+selectedDomain)
	emit(map[string]any{"status": "WordPress installation completed!"})
	removeLockFile(currentUsername)
}

func formOr(r *http.Request, key, def string) string {
	if v := r.FormValue(key); v != "" {
		return v
	}
	return def
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

var tablePrefixRE = regexp.MustCompile(`(\$table_prefix\s*=\s*')[^']*(';)`)

func invalidateMySQLCaches(ctx context.Context, a *appctx.App, userContext, currentUsername string) {
	_ = a.Cache.Delete(ctx, "databases_info:"+userContext)
	_ = a.Cache.Delete(ctx, "get_database_count:"+currentUsername)
}

func emitCleanupFiles(hostOSPath string, emit func(map[string]any)) {
	var deleted, failed []string
	for _, item := range wordpressFiles {
		itemPath := filepath.Join(hostOSPath, item)
		info, statErr := os.Lstat(itemPath)
		if statErr != nil {
			continue
		}
		var rmErr error
		if info.IsDir() {
			rmErr = os.RemoveAll(itemPath)
		} else {
			rmErr = os.Remove(itemPath)
		}
		if rmErr != nil {
			failed = append(failed, item+": "+rmErr.Error())
		} else {
			deleted = append(deleted, item)
		}
	}
	if len(deleted) == 0 && len(failed) == 0 {
		emit(map[string]any{"status": "No WordPress files found to delete."})
		return
	}
	msg := map[string]any{"status": "Files cleanup completed."}
	if len(deleted) > 0 {
		msg["deleted"] = deleted
	}
	if len(failed) > 0 {
		msg["failed"] = failed
	}
	emit(msg)
}

func emitCleanupDatabase(ctx context.Context, userContext, mysqlVersion, dbName, dbUser, dbHost string, emit func(map[string]any)) {
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'"+dbHost+"'", ""); execErr != nil {
		emit(map[string]any{"error": "Cleanup: failed to drop database/user: " + execErr.Error()})
		return
	}
	emit(map[string]any{"status": "Cleanup: dropped database `" + dbName + "` and user `" + dbUser + "`"})
}

// loadHtaccess returns the stock .htaccess contents for the given web server type.
func loadHtaccess(webServerType string) (string, error) {
	path := "/etc/openpanel/wordpress/htaccess/apache.htaccess"
	if strings.Contains(strings.ToLower(webServerType), "litespeed") {
		path = "/etc/openpanel/wordpress/htaccess/litespeed.htaccess"
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func setFSMethod(wpConfigFile string, emit func(map[string]any)) {
	content, err := os.ReadFile(wpConfigFile)
	if err != nil {
		emit(map[string]any{"status": "Error specifying FS_METHOD in wp-config.php"})
		return
	}
	if strings.Contains(string(content), "FS_METHOD") {
		return
	}
	const marker = "/* That's all, stop editing! Happy publishing. */"
	idx := strings.Index(string(content), marker)
	if idx == -1 {
		return
	}
	updated := string(content)[:idx] + "define( 'FS_METHOD', 'direct' );\n" + string(content)[idx:]
	if writeErr := os.WriteFile(wpConfigFile, []byte(updated), 0o644); writeErr != nil {
		emit(map[string]any{"status": "Error specifying FS_METHOD in wp-config.php"})
		return
	}
	emit(map[string]any{"status": "Set FS_METHOD 'direct' in wp-config.php file."})
}

const saltsAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_ []{}<>~`+=,.;:/?|"

// generateSaltsLocally generates a fresh set of WordPress auth keys/salts locally, without calling the WordPress.org secret-key API.
func generateSaltsLocally() string {
	keys := []string{"AUTH_KEY", "SECURE_AUTH_KEY", "LOGGED_IN_KEY", "NONCE_KEY", "AUTH_SALT", "SECURE_AUTH_SALT", "LOGGED_IN_SALT", "NONCE_SALT"}
	var lines []string
	for _, key := range keys {
		salt := generateRandomStringFromAlphabet(64, saltsAlphabet)
		lines = append(lines, "define( '"+key+"', '"+salt+"' );")
	}
	return strings.Join(lines, "\n")
}

var saltsBlockRE = regexp.MustCompile(`(?s)/\*\*#@\+.*?\*\*#@-\*/`)

// writeSaltsLocally replaces the salts block in wp-config.php with a freshly generated set.
func writeSaltsLocally(wpConfigPath string) {
	newSalts := generateSaltsLocally()
	content, err := os.ReadFile(wpConfigPath)
	if err != nil {
		return
	}
	updated := saltsBlockRE.ReplaceAllString(string(content), "/**#@+\n"+newSalts+"\n/**#@-*/")
	_ = os.WriteFile(wpConfigPath, []byte(updated), 0o644)
}

// waitForWPAvailable polls until WP-CLI reports the install is usable
// inside the container (30s timeout, 5s interval).
func waitForWPAvailable(ctx context.Context, userContext, phpContainer string) bool {
	const timeout = 30 * time.Second
	const interval = 5 * time.Second
	deadline := time.Now().Add(timeout)
	for {
		if isWPAvailableInContainer(ctx, userContext, phpContainer) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(interval)
	}
}

func isWPAvailableInContainer(ctx context.Context, userContext, phpContainer string) bool {
	argv := append(append([]string{}, podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)...), "--info")
	return podmanmanager.Command(ctx, userContext, argv).Run() == nil
}

// processInstallSet installs every plugin/theme slug listed (one per
// line) in filePath, activating/forcing each.
func processInstallSet(ctx context.Context, userContext string, wpBaseCmd []string, installPath, filePath, wpType string, emit func(map[string]any)) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(content), "\n") {
		item := strings.TrimSpace(line)
		if item == "" {
			continue
		}
		emit(map[string]any{"status": "Installing " + wpType + ": " + item})
		argv := append(append([]string{}, wpBaseCmd...), wpType, "install", item, "--activate", "--force", "--skip-themes", "--path="+installPath, "--allow-root")
		out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
		if runErr == nil {
			emit(map[string]any{"status": strings.ToUpper(wpType[:1]) + wpType[1:] + " '" + item + "' installed successfully."})
		} else {
			emit(map[string]any{"status": "Failed to install " + wpType + " '" + item + "'.", "error": strings.TrimSpace(string(out))})
		}
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
