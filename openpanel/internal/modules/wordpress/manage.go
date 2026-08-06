package wordpress

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
)

// escapeMySQLString mirrors escape_mysql_string() (duplicated per-package,
// same as mysql.escapeMySQLString - see that function's doc comment).
func escapeMySQLString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}

// ---------------------- CLONE ---------------------- //

// handleCloneWordPress mirrors clone_wordpress().
func handleCloneWordPress(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	lockPath := lockFilePath(currentUsername)
	if info, statErr := os.Stat(lockPath); statErr == nil {
		if time.Since(info.ModTime()) < 5*time.Minute {
			writeJSON(w, http.StatusOK, map[string]string{"message": "Abort, another WordPress clone process is currently running."})
			return
		}
		_ = os.Remove(lockPath)
	}

	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	plan, _ := a.QueryPlanDetailsByID(ctx, planID)
	websitesLimit := atoiDefault(plan.WebsitesLimit, 0)
	websiteCount, _ := countUserWebsites(a, userID)
	if websitesLimit != 0 && websiteCount >= websitesLimit {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "You have reached the maximum number of sites allowed"})
		return
	}

	providedDomain := r.FormValue("source_domain")
	dstDomain := r.FormValue("target_domain")
	srcDB := r.FormValue("source_db")
	srcFolder := r.FormValue("source_folder")
	dstFolder := r.FormValue("subdirectory")

	dstDB := strings.ToLower(formOr(r, "target_db", "wp_clone_"+generateRandomString(6)))
	dstDBUser := strings.ToLower(formOr(r, "target_db_user", dstDB))
	dstDBUserPassword := formOr(r, "target_db_user_password", generateRandomString(16))

	if providedDomain == "" || dstDomain == "" || srcDB == "" || srcFolder == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing required form fields"})
		return
	}

	var domainID int
	var docroot, phpVersion string
	row := a.DB.QueryRowContext(ctx, "SELECT domain_id, docroot, php_version FROM domains WHERE domain_url = ?", dstDomain)
	if scanErr := row.Scan(&domainID, &docroot, &phpVersion); scanErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Destination domain not found in database"})
		return
	}

	dstDomainWithSubdir := dstDomain
	if dstFolder != "" {
		docroot = filepath.Join(docroot, dstFolder)
		dstDomainWithSubdir = dstDomain + "/" + dstFolder
	}

	srcDomain := strings.Split(providedDomain, "/")[0]

	if !validateDomain(srcDomain) || !validateDomain(dstDomain) || !validateDB(srcDB) || !validateDB(dstDB) ||
		!validateDB(dstDBUser) || !validateDocroot(srcFolder) || !validateDocroot(docroot) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid input or unsafe docroot"})
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + phpVersion
	}

	mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
	var dumpCmd string
	switch mysqlVersion {
	case "mysql":
		dumpCmd = "mysqldump --column-statistics=0 --set-gtid-purged=OFF"
	case "mariadb":
		dumpCmd = "mariadb-dump --gtid"
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unsupported MYSQL_TYPE: " + mysqlVersion})
		return
	}

	const wwwBaseDirectory = "/var/www/html/"
	baseDirectory := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	srcPath := strings.Replace(filepath.Clean(srcFolder), wwwBaseDirectory, baseDirectory, 1)
	dstPath := strings.Replace(filepath.Clean(docroot), wwwBaseDirectory, baseDirectory, 1)

	if info, statErr := os.Stat(srcPath); statErr != nil || !info.IsDir() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Source folder not found: " + srcFolder})
		return
	}

	if mkErr := os.MkdirAll(dstPath, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy WordPress files: " + mkErr.Error()})
		return
	}

	for _, item := range wordpressFiles {
		srcItem := filepath.Join(srcPath, item)
		dstItem := filepath.Join(dstPath, item)
		info, statErr := os.Stat(srcItem)
		if statErr != nil {
			continue
		}
		var cpErr error
		if info.IsDir() {
			_ = os.MkdirAll(dstItem, 0o755)
			cpErr = exec.CommandContext(ctx, "cp", "-a", srcItem+"/.", dstItem+"/").Run()
		} else {
			_ = os.MkdirAll(filepath.Dir(dstItem), 0o755)
			cpErr = exec.CommandContext(ctx, "cp", srcItem, dstItem).Run()
		}
		if cpErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy WordPress files: " + cpErr.Error()})
			return
		}
	}

	escapedPassword := escapeMySQLString(dstDBUserPassword)
	cloneQueries := []string{
		"CREATE DATABASE IF NOT EXISTS `" + dstDB + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		"CREATE USER IF NOT EXISTS '" + dstDBUser + "'@'%' IDENTIFIED BY '" + escapedPassword + "'",
		"GRANT ALL PRIVILEGES ON `" + dstDB + "`.* TO '" + dstDBUser + "'@'%'",
		"FLUSH PRIVILEGES",
	}
	for _, q := range cloneQueries {
		if _, execErr := mysqlmanager.Exec(ctx, userContext, q, ""); execErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": execErr.Error()})
			return
		}
	}

	dumpTablesCmd := dumpCmd + " --single-transaction --quick `" + srcDB + "` | " + mysqlVersion + " `" + dstDB + "`"
	fullDBArgv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, "bash", "-c", dumpTablesCmd)
	if runErr := podmanmanager.Command(ctx, userContext, fullDBArgv).Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "command_failed", "command": strings.Join(fullDBArgv, " ")})
		return
	}

	wpConfigFile := filepath.Join(dstPath, "wp-config.php")
	content, readErr := os.ReadFile(wpConfigFile)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": readErr.Error()})
		return
	}
	strContent := string(content)
	strContent = cloneDBNameRE.ReplaceAllString(strContent, "define('DB_NAME', '"+escapePHPSingleQuoted(dstDB)+"');")
	strContent = cloneDBUserRE.ReplaceAllString(strContent, "define('DB_USER', '"+escapePHPSingleQuoted(dstDBUser)+"');")
	strContent = cloneDBPasswordRE.ReplaceAllString(strContent, "define('DB_PASSWORD', '"+escapePHPSingleQuoted(dstDBUserPassword)+"');")
	if writeErr := os.WriteFile(wpConfigFile, []byte(strContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": writeErr.Error()})
		return
	}

	if !strings.Contains(strContent, dstDB) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "Failed to set DB_NAME constant in " + docroot})
		return
	}

	wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)
	searchReplaceArgv := append(append([]string{}, wpBase...), "search-replace",
		"https://"+providedDomain, "https://"+dstDomainWithSubdir, "--all-tables", "--skip-columns=guid",
		"--path="+docroot, "--skip-themes", "--allow-root")
	_ = podmanmanager.Command(ctx, userContext, searchReplaceArgv).Run()

	rewriteFlushArgv := append(append([]string{}, wpBase...), "rewrite", "flush", "--hard", "--path="+docroot, "--skip-themes", "--allow-root")
	_ = podmanmanager.Command(ctx, userContext, rewriteFlushArgv).Run()

	cacheFlushArgv := append(append([]string{}, wpBase...), "cache", "flush", "--path="+docroot, "--skip-themes", "--allow-root")
	_ = podmanmanager.Command(ctx, userContext, cacheFlushArgv).Run()

	if isLitespeed {
		restartCmd := podmanmanager.Command(context.Background(), userContext, podmanmanager.PodmanArgv(userContext, "restart", webServer))
		if startErr := restartCmd.Start(); startErr == nil {
			go func() { _ = restartCmd.Wait() }()
		}
	}

	_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))
	_ = os.Remove(lockPath)

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	wpVersion := formOr(r, "wordpress_version", "latest")
	if _, insertErr := a.DB.ExecContext(ctx, `
		INSERT INTO sites (site_name, domain_id, admin_email, version, type)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE domain_id = VALUES(domain_id), admin_email = VALUES(admin_email), version = VALUES(version), type = VALUES(type)`,
		dstDomainWithSubdir, domainID, adminEmail, wpVersion, "wordpress"); insertErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cloned WordPress website from "+providedDomain+" to "+dstDomainWithSubdir, reqip.ClientIP(r))

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "success", "source": providedDomain, "target": dstDomainWithSubdir,
		"source_path": srcPath, "target_path": dstPath, "target_db": dstDB,
	})
}

var (
	cloneDBNameRE     = regexp.MustCompile(`define\(\s*'DB_NAME'\s*,\s*'.*?'\s*\);`)
	cloneDBUserRE     = regexp.MustCompile(`define\(\s*'DB_USER'\s*,\s*'.*?'\s*\);`)
	cloneDBPasswordRE = regexp.MustCompile(`define\(\s*'DB_PASSWORD'\s*,\s*'.*?'\s*\);`)
)

func escapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}

// ---------------------- REMOVE ---------------------- //

// handleRemoveWordPress mirrors remove_wordpress().
func handleRemoveWordPress(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")

	var siteName, docroot string
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ?`, id)
	if scanErr := row.Scan(&siteName, &docroot); scanErr != nil {
		flashAndRedirect(a, w, r, "error", "No data found for the provided site ID", "/sites")
		return
	}

	parts := strings.Split(siteName, "/")
	selectedDomain := parts[0]
	subdirectory := parts[1:]

	if docroot == "" {
		flashAndRedirect(a, w, r, "error", "WordPress installation not found in the database", "/sites")
		return
	}

	realInstallPath := strings.TrimPrefix(docroot, "/var/www/html/")
	if len(subdirectory) > 0 {
		realInstallPath = filepath.Join(append([]string{realInstallPath}, subdirectory...)...)
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, selectedDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + realInstallPath
	wpConfigFile := filepath.Join(volume, "wp-config.php")
	content, _ := os.ReadFile(wpConfigFile)

	dbNameMatch := removeDBNameRE.FindStringSubmatch(string(content))
	dbUserMatch := removeDBUserRE.FindStringSubmatch(string(content))

	if dbNameMatch != nil && dbUserMatch != nil {
		dbName, dbUser := dbNameMatch[1], dbUserMatch[1]
		_, _ = mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+dbName+"`", "")
		_, _ = mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'%'", "")
		invalidateMySQLCaches(ctx, a, userContext, currentUsername)
	} else {
		flashSess(a, w, r, "warning", "Database name or user not found in wp-config.php")
	}

	var toDelete []string
	for _, item := range wordpressFiles {
		itemPath := filepath.Join(volume, item)
		if _, statErr := os.Stat(itemPath); statErr == nil {
			toDelete = append(toDelete, itemPath)
		}
	}
	if len(toDelete) > 0 {
		_ = exec.CommandContext(ctx, "rm", append([]string{"-rf"}, toDelete...)...).Run()
	}

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", id); delErr != nil {
		message := "An error occurred during WordPress uninstall."
		flashSess(a, w, r, "error", message)
		_, _ = w.Write([]byte(message))
		return
	}

	_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))
	_ = logger.RecordUserAction(a.Config, currentUsername, "uninstalled WordPress website for "+selectedDomain, reqip.ClientIP(r))

	message := "WordPress uninstalled successfully!"
	flashSess(a, w, r, "success", message)
	_, _ = w.Write([]byte(message))
}

var (
	removeDBNameRE = regexp.MustCompile(`define\(\s*'DB_NAME'\s*,\s*'([^']+)'`)
	removeDBUserRE = regexp.MustCompile(`define\(\s*'DB_USER'\s*,\s*'([^']+)'`)
)

// ---------------------- DETACH ---------------------- //

// handleDetachWordPress mirrors detach_wordpress().
func handleDetachWordPress(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		flashAndRedirect(a, w, r, "error", "Missing site ID.", "/sites")
		return
	}

	var siteName string
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ?
		LIMIT 1`, id)
	if scanErr := row.Scan(&siteName); scanErr != nil {
		flashAndRedirect(a, w, r, "error", "No data found for the provided site ID", "/sites")
		return
	}

	domainName := strings.Split(siteName, "/")[0]
	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", id); delErr != nil {
		message := "An error occurred during WordPress detachment. Please try again."
		flashSess(a, w, r, "error", message)
		_, _ = w.Write([]byte("An error occurred during WordPress detachment."))
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "detached WordPress website", reqip.ClientIP(r))
	_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))
	flashSess(a, w, r, "success", "WordPress installation detached from Manager.")
	_, _ = w.Write([]byte("WordPress detachment completed successfully!"))
}

// ---------------------- SHARED SCAN/RELOAD HELPERS ---------------------- //

func getDomainID(ctx context.Context, a *appctx.App, domainName string) (int, bool) {
	var domainID int
	row := a.DB.QueryRowContext(ctx, "SELECT domain_id FROM domains WHERE domain_url = ?", domainName)
	if err := row.Scan(&domainID); err != nil {
		return 0, false
	}
	return domainID, true
}

func checkSiteAlreadyExistsForUser(ctx context.Context, a *appctx.App, siteName string) bool {
	var id int
	row := a.DB.QueryRowContext(ctx, "SELECT id FROM sites WHERE site_name = ?", siteName)
	return row.Scan(&id) == nil
}

// walkForWPConfig walks base looking for wp-config.php files, never
// descending into skipDirs - mirrors os.walk(base_directory) with
// dirs[:] = [d for d in dirs if d not in SKIP_DIRS].
func walkForWPConfig(base string, onFound func(dir string)) {
	_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "wp-config.php" {
			onFound(filepath.Dir(path))
		}
		return nil
	})
}

func phpContainerForUser(userContext string) string {
	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		return webServer
	}
	defaultVersion := webserver.GetEnvFileValue(userContext, "DEFAULT_PHP_VERSION")
	return "php-fpm-" + defaultVersion
}

// ---------------------- RELOAD DATA ---------------------- //

// handleReloadWordPressData mirrors reload_wordpress_data_in_wpmanager().
func handleReloadWordPressData(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	const baseDirectory = "/var/www/html/"
	phpContainer := phpContainerForUser(userContext)
	wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)

	walkForWPConfig(baseDirectory, func(root string) {
		siteURLArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+root, "option", "get", "siteurl")
		out, _ := podmanmanager.Command(ctx, userContext, siteURLArgv).Output()
		siteURL := strings.TrimSpace(string(out))
		siteName := strings.TrimPrefix(strings.TrimPrefix(siteURL, "http://"), "https://")

		domainName := siteName
		if idx := strings.Index(siteName, "/"); idx != -1 {
			domainName = siteName[:idx]
		}

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
			return
		}
		if !checkSiteAlreadyExistsForUser(ctx, a, domainName) {
			return
		}

		emailArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+root, "option", "get", "admin_email")
		emailOut, emailErr := podmanmanager.Command(ctx, userContext, emailArgv).Output()
		adminEmail := strings.TrimSpace(string(emailOut))
		if emailErr != nil || !strings.Contains(adminEmail, "@") {
			return
		}

		versionArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+root, "core", "version")
		versionOut, versionErr := podmanmanager.Command(ctx, userContext, versionArgv).Output()
		if versionErr != nil {
			return
		}
		version := string(versionOut)

		if _, execErr := a.DB.ExecContext(ctx, "UPDATE sites SET admin_email = ?, version = ? WHERE site_name = ?", adminEmail, version, siteName); execErr != nil {
			return
		}
		_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))
	})

	_, _ = w.Write([]byte("Scan completed. Found installations:\n\n"))
}

// ---------------------- SCAN ---------------------- //

// handleScanWordPress mirrors scan_wordpress().
func handleScanWordPress(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	lockPath := lockFilePath(currentUsername)
	if info, statErr := os.Stat(lockPath); statErr == nil {
		if time.Since(info.ModTime()) < time.Minute {
			_, _ = w.Write([]byte("Scan skipped. WordPress installation is currently running."))
			return
		}
		_ = os.Remove(lockPath)
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "initiated scan for WordPress installations", reqip.ClientIP(r))

	wwwBaseDirectory := "/var/www/html/"
	baseDirectory := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	phpContainer := phpContainerForUser(userContext)
	wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)

	type found struct {
		configFile, domain, email, version string
	}
	var installations []found

	walkForWPConfig(baseDirectory, func(root string) {
		containerRoot := strings.Replace(root, baseDirectory, wwwBaseDirectory, 1)
		configFilePath := strings.TrimPrefix(containerRoot, "/")

		siteURLArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+containerRoot, "option", "get", "siteurl")
		out, runErr := podmanmanager.Command(ctx, userContext, siteURLArgv).Output()
		if runErr != nil {
			return
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		siteURL := ""
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				siteURL = strings.TrimSpace(l)
			}
		}
		siteName := strings.TrimPrefix(strings.TrimPrefix(siteURL, "http://"), "https://")
		domainName := siteName
		if idx := strings.Index(siteName, "/"); idx != -1 {
			domainName = siteName[:idx]
		}

		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
			return
		}
		if checkSiteAlreadyExistsForUser(ctx, a, siteName) {
			return
		}

		emailArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+containerRoot, "option", "get", "admin_email")
		emailOut, emailErr := podmanmanager.Command(ctx, userContext, emailArgv).Output()
		adminEmail := strings.TrimSpace(string(emailOut))
		if emailErr != nil || !strings.Contains(adminEmail, "@") {
			return
		}

		versionArgv := append(append([]string{}, wpBase...), "--skip-themes", "--allow-root", "--path="+containerRoot, "core", "version")
		versionOut, versionErr := podmanmanager.Command(ctx, userContext, versionArgv).Output()
		if versionErr != nil {
			return
		}
		version := strings.TrimSpace(string(versionOut))

		domainID, _ := getDomainID(ctx, a, domainName)
		if _, insertErr := a.DB.ExecContext(ctx, "INSERT INTO sites (site_name, domain_id, admin_email, version, type) VALUES (?, ?, ?, ?, ?)",
			siteName, domainID, adminEmail, version, "WordPress"); insertErr != nil {
			return
		}
		_ = a.Cache.Delete(ctx, fmt.Sprintf("get_user_websites:%d", userID))

		installations = append(installations, found{configFile: configFilePath, domain: domainName, email: adminEmail, version: version})
	})

	var summary strings.Builder
	summary.WriteString("Scan completed. Found installations:\n\n")
	for _, inst := range installations {
		summary.WriteString("WordPress installation: " + inst.configFile + "\n")
		summary.WriteString("WordPress Version: " + inst.version + "\n\n")
	}
	_, _ = w.Write([]byte(summary.String()))
}

// ---------------------- SECURE ---------------------- //

var wpManagerRuleRE = regexp.MustCompile(`\bwp_manager_\w+\b`)

// handleWordPressSecure mirrors wordpress_secure().
func handleWordPressSecure(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	providedDomain := r.PathValue("provided_domain")

	if r.Method == http.MethodGet {
		if providedDomain == "" {
			out, runErr := exec.CommandContext(ctx, "opencli", "websites-secure", "--list-available-rules").CombinedOutput()
			if runErr != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list available rules: " + string(out)})
				return
			}
			rules := wpManagerRuleRE.FindAllString(string(out), -1)
			if rules == nil {
				rules = []string{}
			}
			sort.Strings(rules)
			writeJSON(w, http.StatusOK, rules)
			return
		}

		domain := strings.Split(providedDomain, "/")[0]
		if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}

		configPath := "/etc/openpanel/caddy/domains/" + domain + ".conf"
		content, readErr := os.ReadFile(configPath)
		if readErr != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Config file not found"})
			return
		}
		matches := wpManagerRuleRE.FindAllString(string(content), -1)
		seen := map[string]bool{}
		unique := []string{}
		for _, m := range matches {
			if !seen[m] {
				seen[m] = true
				unique = append(unique, m)
			}
		}
		sort.Strings(unique)
		writeJSON(w, http.StatusOK, unique)
		return
	}

	// POST
	domain := strings.Split(providedDomain, "/")[0]
	_ = r.ParseForm()
	var validRules []string
	for key := range r.PostForm {
		if strings.HasPrefix(key, "wp_manager_") {
			validRules = append(validRules, key)
		}
	}

	argv := []string{"opencli", "websites-secure", domain}
	var logAction string
	if len(validRules) > 0 {
		logAction = "enabled hardened rules for domain " + domain + " using WP Manager"
		argv = append(argv, "--rules="+strings.Join(validRules, " "))
	} else {
		logAction = "disabled hardened rules for domain " + domain + " using WP Manager"
		argv = append(argv, "--disable-all")
	}

	if runErr := exec.CommandContext(ctx, argv[0], argv[1:]...).Run(); runErr == nil {
		_ = logger.RecordUserAction(a.Config, currentUsername, logAction, reqip.ClientIP(r))
		flashSess(a, w, r, "success", "Hardening rules have been successfully applied to the website")
	} else {
		flashSess(a, w, r, "error", "Failed to apply hardening rules to the website - please try again")
	}

	http.Redirect(w, r, "/website?domain="+providedDomain, http.StatusFound)
}
