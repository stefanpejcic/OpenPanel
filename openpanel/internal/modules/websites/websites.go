// Package websites implements the /sites listing page, the /website
// CMS-type dispatcher, and the side JSON endpoints (safebrowsing,
// PageSpeed, WP vulnerability scan, pm2 package installs, WP info, and the
// distinct /wordpress/wp-cli/<action> passthrough used by the site-manager
// UI). Drupal and Mautic are not supported CMS types.
package websites

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
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

func injected(a *appctx.App, r *http.Request) (userID int, username, userContext string, err error) {
	userID, _ = auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return userID, "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return userID, username, userContext, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// splitDomainAndFolder splits a "domain/subfolder" path parameter into its
// domain and (possibly empty) folder parts, on the first slash.
func splitDomainAndFolder(param string) (domain, folder string) {
	if idx := strings.Index(param, "/"); idx != -1 {
		return param[:idx], param[idx+1:]
	}
	return param, ""
}

// ---------------------- FAVICON ---------------------- //

// handleFavicon redirects to a domain's favicon, either through a
// configured favicon service or Google's fallback.
func handleFavicon(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")
	apexDomain, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(r.Context(), userID, apexDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	faviconsSetting := a.Config.Get("favicons", "")
	var target string
	if faviconsSetting != "" && faviconsSetting != "local" {
		target = faviconsSetting + domain
	} else {
		target = "https://www.google.com/s2/favicons?domain=" + domain
	}
	http.Redirect(w, r, target, http.StatusFound)
}

// ---------------------- DATABASE SIZE ---------------------- //

// handleDatabaseSize reports either a WordPress install's on-disk size (via
// `wp db size`) or a raw database's size, depending on which query
// parameter was supplied.
func handleDatabaseSize(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := r.URL.Query().Get("domain")
	docroot := r.URL.Query().Get("docroot")
	database := r.URL.Query().Get("database")

	switch {
	case docroot != "" && domain != "":
		baseDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data"
		folder := strings.TrimPrefix(strings.TrimPrefix(docroot, "/"), "var/www/html/")
		docrootInContainer := filepath.Join(baseDir, folder)
		if !strings.HasPrefix(docrootInContainer, baseDir) {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Invalid folder path."})
			return
		}

		phpVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
		phpContainer := "php-fpm-" + phpVersion
		if strings.Contains(strings.ToLower(webServer), "litespeed") {
			phpContainer = webServer
		}

		wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)
		argv := append(append([]string{}, wpBase...), "db", "size", "--human-readable", "--path="+docrootInContainer, "--allow-root")
		out, runErr := podmanmanager.Command(ctx, userContext, argv).Output()
		if runErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": string(out)})
			return
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) < 2 {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to retrieve size"})
			return
		}
		columns := strings.Fields(lines[1])
		if len(columns) < 2 {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unexpected output format"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"size": columns[1]})

	case database != "":
		rows, execErr := mysqlmanager.Exec(ctx, userContext,
			"SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) FROM information_schema.TABLES WHERE table_schema = \""+database+"\"", "")
		if execErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
			return
		}
		if len(rows) == 0 || rows[0][0] == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to retrieve size"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"size": toStringCell(rows[0][0]) + " MB"})

	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request. Please specify a domain or database name."})
	}
}

// ---------------------- SAFE BROWSING ---------------------- //

// safeBrowsingData is a cached Google Safe Browsing lookup, shared by the
// UI's JSON route and the API's GET /api/sites/{domain}/safebrowsing.
func safeBrowsingData(ctx context.Context, a *appctx.App, domain string) (map[string]any, error) {
	return cache.Memoize(ctx, a.Cache, "google_safebrowsing_data:"+domain, 12*time.Hour, func() (map[string]any, error) {
		const apiKey = "AIzaSyBwv4iHQcYGTHOjqg9D4tcLW0TvqWHDbBc"
		urlToCheck := "http://" + domain
		endpoint := "https://safebrowsing.googleapis.com/v4/threatMatches:find?key=" + apiKey

		payload := map[string]any{
			"client": map[string]string{"clientId": "OpenPanel", "clientVersion": "1.7.2"},
			"threatInfo": map[string]any{
				"threatTypes":      []string{"MALWARE", "SOCIAL_ENGINEERING", "UNWANTED_SOFTWARE", "POTENTIALLY_HARMFUL_APPLICATION"},
				"platformTypes":    []string{"ANY_PLATFORM"},
				"threatEntryTypes": []string{"URL"},
				"threatEntries":    []map[string]string{{"url": urlToCheck}},
			},
		}
		body, _ := json.Marshal(payload)

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("Content-Type", "application/json")
		resp, doErr := http.DefaultClient.Do(req)
		if doErr != nil {
			return nil, doErr
		}
		defer resp.Body.Close()

		var result map[string]any
		if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
			return nil, decErr
		}

		if matches, ok := result["matches"]; ok {
			return map[string]any{"status": "danger", "url": urlToCheck, "threats": matches}, nil
		}
		return map[string]any{"status": "safe", "url": urlToCheck, "threats": []any{}}, nil
	})
}

// handleGoogleSafeBrowsing returns the cached Safe Browsing verdict for a
// domain the caller owns.
func handleGoogleSafeBrowsing(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	result, err := safeBrowsingData(ctx, a, domain)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to contact Google Safe Browsing API", "details": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ---------------------- PAGE SPEED ---------------------- //

var websiteParamRE = regexp.MustCompile(`^[a-zA-Z0-9./-]+$`)

// handlePageSpeed serves the cached PageSpeed report on GET, or triggers a
// fresh scan on POST.
func handlePageSpeed(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	websiteParam := r.PathValue("domain")
	domain, _ := splitDomainAndFolder(websiteParam)

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		if !websiteParamRE.MatchString(websiteParam) {
			http.Redirect(w, r, "/sites", http.StatusFound)
			return
		}
		cmd := exec.Command("opencli", "websites-pagespeed", websiteParam)
		if startErr := cmd.Start(); startErr == nil {
			go func() { _ = cmd.Wait() }()
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "initiated refresh for PageSpeed data on website "+websiteParam, reqip.ClientIP(r))
		http.Redirect(w, r, "/sites", http.StatusFound)
		return
	}

	filename := strings.ReplaceAll(strings.ReplaceAll(websiteParam, "://", "_"), "/", "_") + ".json"
	filePath := filepath.Join("/etc/openpanel/openpanel/websites", filename)

	if content, readErr := os.ReadFile(filePath); readErr == nil {
		var data map[string]any
		if json.Unmarshal(content, &data) != nil {
			data = map[string]any{"timestamp": nil, "website": domain, "desktop_speed": map[string]any{}, "mobile_speed": map[string]any{}}
		}
		writeJSON(w, http.StatusOK, data)
		return
	}

	if !websiteParamRE.MatchString(websiteParam) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid website parameter"})
		return
	}
	out, runErr := exec.CommandContext(ctx, "opencli", "websites-pagespeed", websiteParam).CombinedOutput()
	message := strings.TrimSpace(string(out))
	if message == "" {
		message = "No data yet, please allow a few minutes for data gathering.."
	}
	status := http.StatusOK
	if runErr != nil {
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, map[string]string{"message": message})
}

// ---------------------- WP VULNERABILITY ---------------------- //

// handleWPVulnerability serves the cached WordPress vulnerability report
// on GET (running a scan first if none exists yet), or triggers a fresh
// scan on POST.
func handleWPVulnerability(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	websiteParam := r.PathValue("domain")
	domain, _ := splitDomainAndFolder(websiteParam)

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	filename := strings.ReplaceAll(strings.ReplaceAll(websiteParam, "://", "_"), "/", "_") + ".json"
	filePath := filepath.Join("/etc/openpanel/wordpress/vulnerability", filename)

	if r.Method == http.MethodPost {
		_ = exec.CommandContext(ctx, "opencli", "websites-vulnerability", websiteParam).Run()
		_ = logger.RecordUserAction(a.Config, currentUsername, "initiated refresh for WP vulnerability data on website "+websiteParam, reqip.ClientIP(r))

		if redirectTo := r.FormValue("redirect_to"); redirectTo != "" {
			http.Redirect(w, r, redirectTo, http.StatusFound)
			return
		}
		http.Redirect(w, r, "/sites", http.StatusFound)
		return
	}

	data := map[string]any{"core_version_": map[string]any{}, "plugin_": map[string]any{}, "theme_": map[string]any{}}
	if content, readErr := os.ReadFile(filePath); readErr == nil {
		_ = json.Unmarshal(content, &data)
	} else {
		_ = exec.CommandContext(ctx, "opencli", "websites-vulnerability", websiteParam).Run()
		time.Sleep(2 * time.Second)
		if content, readErr2 := os.ReadFile(filePath); readErr2 == nil {
			_ = json.Unmarshal(content, &data)
		}
	}
	writeJSON(w, http.StatusOK, data)
}

// ---------------------- PM2 PACKAGE INSTALL ---------------------- //

var pm2SafeNameRE = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

// installPackagesInContainer runs the install command for the given
// package manager (pip/npm/pnpm) inside a site's container.
func installPackagesInContainer(a *appctx.App, r *http.Request, userContext, siteName, installType string) (bool, string) {
	if !pm2SafeNameRE.MatchString(siteName) {
		return false, "Invalid container name"
	}

	ctx := r.Context()
	var argv []string
	switch installType {
	case "pip":
		argv = podmanmanager.PodmanArgv(userContext, "exec", siteName, "pip", "install", "-r", "requirements.txt")
	case "pnpm":
		pnpmInstall := podmanmanager.PodmanArgv(userContext, "exec", siteName, "npm", "install", "-g", "pnpm@latest-10")
		if out, err := podmanmanager.Command(ctx, userContext, pnpmInstall).CombinedOutput(); err != nil {
			return false, string(out)
		}
		argv = podmanmanager.PodmanArgv(userContext, "exec", siteName, "pnpm", "install")
	case "npm":
		argv = podmanmanager.PodmanArgv(userContext, "exec", siteName, "npm", "install")
	default:
		return false, "Unsupported install type"
	}

	out, err := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if err != nil {
		return false, string(out)
	}
	return true, string(out)
}

// handleInstallPackages installs a Python/Node app's declared dependencies
// inside its container.
func handleInstallPackages(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	siteName := strings.ToLower(r.PathValue("selected_domain"))
	installType := strings.ToLower(r.PathValue("install_type"))

	if installType != "pip" && installType != "npm" && installType != "pnpm" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid install type. Use 'pip' for Python apps and 'npm' or 'pnpm' for NodeJS."})
		return
	}

	success, output := installPackagesInContainer(a, r, userContext, siteName, installType)
	if success {
		successMsg := "NPM packages installed successfully."
		if installType == "pip" {
			successMsg = "Requirements installed successfully."
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "executed "+installType+" install for application "+siteName, reqip.ClientIP(r))
		writeJSON(w, http.StatusOK, map[string]string{"message": successMsg, "output": output})
		return
	}
	errorMsg := "An error occurred while installing NPM packages."
	if installType == "pip" {
		errorMsg = "An error occurred while installing requirements."
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"message": errorMsg, "error_output": output})
}

// ---------------------- WP INFO ---------------------- //

// extractDatabaseInfo parses DB_NAME/DB_USER/DB_PASSWORD/DB_HOST and the
// table prefix out of wp-config.php. Go's RE2 has no backreferences to tie
// the opening/closing quote character together (both single or both
// double), so each quote style is tried separately instead.
func extractDatabaseInfo(userContext, directory string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(directory, wwwPrefix) {
		return nil
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(directory, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "wp-config.php"))
	if err != nil {
		return map[string]string{"error": "wp-config.php not found"}
	}
	text := string(content)

	info := map[string]string{}
	for _, quote := range []string{`'`, `"`} {
		for key, constName := range map[string]string{
			"database_name": "DB_NAME", "database_user": "DB_USER",
			"database_password": "DB_PASSWORD", "database_host": "DB_HOST",
		} {
			if _, exists := info[key]; exists {
				continue
			}
			re := regexp.MustCompile(`define\(\s*` + quote + constName + quote + `\s*,\s*` + quote + `(.+?)` + quote + `\s*\)`)
			if m := re.FindStringSubmatch(text); m != nil {
				info[key] = m[1]
			}
		}
		re := regexp.MustCompile(`\$table_prefix\s*=\s*` + quote + `(.+?)` + quote)
		if _, exists := info["database_table_prefix"]; !exists {
			if m := re.FindStringSubmatch(text); m != nil {
				info["database_table_prefix"] = m[1]
			}
		}
	}
	if len(info) == 0 {
		return map[string]string{"error": "No database information found in wp-config.php"}
	}
	return info
}

var wpVersionRE = regexp.MustCompile(`\$wp_version\s*=\s*['"]([^'"]+)['"]`)

// getWPVersion reads the WordPress version out of wp-includes/version.php.
func getWPVersion(userContext, realPath string) string {
	relPath := strings.TrimPrefix(realPath, "/var/www/html/")
	filePath := filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", relPath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Unknown"
	}
	if m := wpVersionRE.FindStringSubmatch(string(content)); m != nil {
		return strings.TrimSpace(m[1])
	}
	return "Unknown"
}

// getMySQLVersion resolves the running MySQL/MariaDB version, memoized for
// 1 hour since it changes only on upgrade.
func getMySQLVersion(a *appctx.App, r *http.Request, userContext string) string {
	ctx := r.Context()
	version, _ := cache.Memoize(ctx, a.Cache, "get_mysql_version_ws:"+userContext, time.Hour, func() (string, error) {
		mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_VERSION")
		if mysqlVersion != "latest" {
			return mysqlVersion, nil
		}
		mysqlContainer := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
		argv := podmanmanager.PodmanArgv(userContext, "exec", mysqlContainer, mysqlContainer, "-V")
		out, _ := podmanmanager.Command(ctx, userContext, argv).Output()
		if m := regexp.MustCompile(`\d+\.\d+\.\d+`).FindString(string(out)); m != "" {
			return m, nil
		}
		if m := regexp.MustCompile(`(?i)mariadb from (\d+\.\d+\.\d+)`).FindStringSubmatch(string(out)); m != nil {
			return "MariaDB " + m[1], nil
		}
		return "Version not found", nil
	})
	return version
}

// handleWebsiteWPInfo returns a WordPress site's database credentials, WP
// version, PHP version, and MySQL version for the site-manager info panel.
func handleWebsiteWPInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	siteName := r.PathValue("site_name")
	domainNameUsed, folderParam := splitDomainAndFolder(siteName)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainNameUsed) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	var docroot string
	row := a.DB.QueryRowContext(ctx, "SELECT docroot FROM domains WHERE domain_url = ?", domainNameUsed)
	if scanErr := row.Scan(&docroot); scanErr != nil {
		flashAndRedirect(a, w, r, "error", "Unable to detect docroot for the domain.", "/sites")
		return
	}
	if folderParam != "" {
		docroot = docroot + "/" + folderParam
	}

	databaseInfo := extractDatabaseInfo(userContext, docroot)
	wpVersionFile := filepath.Join(docroot, "wp-includes", "version.php")
	wpVersion := getWPVersion(userContext, wpVersionFile)
	phpVersion := php.GetPHPVForDomain(ctx, a, userContext, domainNameUsed)
	mysqlVersion := getMySQLVersion(a, r, userContext)

	writeJSON(w, http.StatusOK, map[string]any{
		"database_info": databaseInfo, "wp_version": wpVersion,
		"php_version": phpVersion, "mysql_version": mysqlVersion,
	})
}

// ---------------------- DISTINCT WP-CLI PASSTHROUGH ---------------------- //
// Distinct from the WordPress module's own /wp-cli/<action>: this one is
// scoped to the site-manager single-page's general/debug/update-preferences
// UI.

var wsAllowedWPCLIActions = map[string]bool{
	"site_info": true, "update_debug": true, "update_site_information": true,
	"update_now": true, "update_update_preferences": true, "debug_info": true, "update_info": true,
}

var wsDocrootSafeRE = regexp.MustCompile(`^[a-zA-Z0-9_/.-]+$`)

// handleWordPressWPCLI dispatches a scoped set of `wp` CLI actions
// (site info, debug flags, update preferences, ...) for the site-manager
// single-page UI.
func handleWordPressWPCLI(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	action := r.PathValue("action")

	website := r.URL.Query().Get("website")
	if website == "" {
		website = r.URL.Query().Get("domain")
	}
	docroot := r.URL.Query().Get("docroot")

	if !wsAllowedWPCLIActions[action] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid action"})
		return
	}
	if website == "" || docroot == "" {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Missing required parameters"})
		return
	}

	domainNameUsed, _ := splitDomainAndFolder(website)
	if !a.CheckDomainBelongsToUser(ctx, userID, domainNameUsed) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if !wsDocrootSafeRE.MatchString(docroot) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid docroot"})
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + php.GetPHPVForDomain(ctx, a, userContext, domainNameUsed)
	}

	wpConfigFileInContainer := docroot + "/wp-config.php"
	wpConfigFile := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" +
		strings.TrimPrefix(wpConfigFileInContainer, "/var/www/html/")

	wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)

	switch action {
	case "site_info":
		handleWPCLISiteInfo(a, w, r, userContext, wpConfigFile, docroot, wpBase)
	case "update_debug":
		handleWPCLIUpdateDebug(a, w, r, currentUsername, userContext, docroot, phpContainer, website)
	case "update_site_information":
		handleWPCLIUpdateSiteInfo(a, w, r, currentUsername, userContext, docroot, wpBase, website)
	case "update_now":
		handleWPCLIUpdateNow(a, w, r, currentUsername, userContext, docroot, wpBase)
	case "update_update_preferences":
		handleWPCLIUpdatePreferences(a, w, r, currentUsername, userContext, docroot, phpContainer)
	case "debug_info":
		handleWPCLIDebugInfo(a, w, r, userContext, wpConfigFile, docroot, wpBase)
	case "update_info":
		handleWPCLIUpdateInfo(a, w, r, userContext, wpConfigFile, docroot, wpBase)
	}
}

func handleWPCLISiteInfo(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext, wpConfigFile, docroot string, wpBase []string) {
	content, readErr := os.ReadFile(wpConfigFile)
	if readErr != nil {
		writeJSON(w, http.StatusOK, wpCLIOptionListFallback(a, r, userContext, docroot, wpBase))
		return
	}
	dbNameMatch := regexp.MustCompile(`define\(\s*['"]DB_NAME['"]\s*,\s*['"]([^'"]+)['"]`).FindStringSubmatch(string(content))
	tablePrefixMatch := regexp.MustCompile(`\$table_prefix\s*=\s*['"]([^'"]+)['"]`).FindStringSubmatch(string(content))
	if dbNameMatch == nil || tablePrefixMatch == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "wp-config parse error"})
		return
	}
	dbName, tablePrefix := dbNameMatch[1], tablePrefixMatch[1]

	rows, execErr := mysqlmanager.Exec(r.Context(), userContext,
		"SELECT option_name, option_value FROM `"+tablePrefix+"options` WHERE option_name IN "+
			"('siteurl','home','blogname','blogdescription','admin_email','users_can_register','blog_public','default_ping_status')", dbName)
	if execErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed reading WordPress options: " + execErr.Error()})
		return
	}
	options := map[string]string{}
	for _, row := range rows {
		options[toStringCell(row[0])] = toStringCell(row[1])
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true, "site_url": options["siteurl"], "home_url": options["home"],
		"site_name": options["blogname"], "tagline": options["blogdescription"], "admin_email": options["admin_email"],
		"registration_enabled": options["users_can_register"] == "1", "seo_indexing_enabled": options["blog_public"] == "1",
		"pingbacks_enabled": options["default_ping_status"] == "open",
	})
}

func wpCLIOptionListFallback(a *appctx.App, r *http.Request, userContext, docroot string, wpBase []string) map[string]any {
	argv := append(append([]string{}, wpBase...), "option", "list", "--format=json", "--path="+docroot, "--skip-themes", "--allow-root")
	out, err := podmanmanager.Command(r.Context(), userContext, argv).Output()
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	var result any
	_ = json.Unmarshal(out, &result)
	return map[string]any{"result": result}
}

func toStringCell(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

const wpCLIPHPEnvFlags = "php -d memory_limit=-1 -d open_basedir=none -d disable_functions= -d display_errors=0 -d error_log=/dev/null /usr/local/bin/wp config set"

func handleWPCLIUpdateDebug(a *appctx.App, w http.ResponseWriter, r *http.Request, currentUsername, userContext, docroot, phpContainer, website string) {
	debugOptions := map[string]string{
		"WP_DEBUG":         onOff(r.URL.Query().Get("WP_DEBUG")),
		"WP_DEBUG_LOG":     onOff(r.URL.Query().Get("WP_DEBUG_LOG")),
		"WP_DEBUG_DISPLAY": onOff(r.URL.Query().Get("WP_DEBUG_DISPLAY")),
		"SCRIPT_DEBUG":     onOff(r.URL.Query().Get("SCRIPT_DEBUG")),
		"SAVEQUERIES":      onOff(r.URL.Query().Get("SAVEQUERIES")),
	}
	wpBaseCmd := wpCLIPHPEnvFlags + " --raw --type=constant --path=" + docroot + " --skip-themes --allow-root"
	var commands []string
	for _, opt := range []string{"WP_DEBUG", "WP_DEBUG_LOG", "WP_DEBUG_DISPLAY", "SCRIPT_DEBUG", "SAVEQUERIES"} {
		commands = append(commands, wpBaseCmd+" "+opt+" "+debugOptions[opt])
	}
	fullCmd := strings.Join(commands, " && ")
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", fullCmd)
	if runErr := podmanmanager.Command(r.Context(), userContext, argv).Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "updated debug options for WordPress website "+website, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Debugging options updated successfully"})
}

func onOff(v string) string {
	if strings.ToLower(v) == "on" {
		return "true"
	}
	return "false"
}

func handleWPCLIUpdateSiteInfo(a *appctx.App, w http.ResponseWriter, r *http.Request, currentUsername, userContext, docroot string, wpBase []string, website string) {
	q := r.URL.Query()
	for _, opt := range []string{"siteurl", "home", "blogname", "blogdescription", "admin_email", "users_can_register", "default_ping_status", "blog_public"} {
		if !q.Has(opt) {
			continue
		}
		value := q.Get(opt)
		argv := append(append([]string{}, wpBase...), "option", "update", opt, value, "--path="+docroot, "--skip-themes", "--allow-root")
		_ = podmanmanager.Command(r.Context(), userContext, argv).Run()
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "updated site information for WordPress website "+website, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "General options edited successfully"})
}

func handleWPCLIUpdateNow(a *appctx.App, w http.ResponseWriter, r *http.Request, currentUsername, userContext, docroot string, wpBase []string) {
	argv := append(append([]string{}, wpBase...), "core", "update", "--path="+docroot, "--skip-themes", "--allow-root")
	_ = logger.RecordUserAction(a.Config, currentUsername, "started core update for WordPress website in "+docroot, reqip.ClientIP(r))
	if runErr := podmanmanager.Command(r.Context(), userContext, argv).Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "WordPress updated successfully"})
}

func handleWPCLIUpdatePreferences(a *appctx.App, w http.ResponseWriter, r *http.Request, currentUsername, userContext, docroot, phpContainer string) {
	q := r.URL.Query()
	get := func(key, def string) string {
		if v := q.Get(key); v != "" {
			return v
		}
		return def
	}
	updateOpts := map[string]string{
		"WP_AUTO_UPDATE_PLUGINS": get("WP_AUTO_UPDATE_PLUGINS", "false"),
		"WP_AUTO_UPDATE_THEMES":  get("WP_AUTO_UPDATE_THEMES", "false"),
		"WP_AUTO_UPDATE_CORE":    get("WP_AUTO_UPDATE_CORE", "false"),
	}
	allowedValues := map[string]map[string]bool{
		"WP_AUTO_UPDATE_PLUGINS": {"true": true, "false": true},
		"WP_AUTO_UPDATE_THEMES":  {"true": true, "false": true},
		"WP_AUTO_UPDATE_CORE":    {"true": true, "false": true, "minor": true},
	}
	for opt, value := range updateOpts {
		if !allowedValues[opt][value] {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid value for " + opt})
			return
		}
	}

	var commands []string
	for _, opt := range []string{"WP_AUTO_UPDATE_PLUGINS", "WP_AUTO_UPDATE_THEMES", "WP_AUTO_UPDATE_CORE"} {
		value := updateOpts[opt]
		rawFlag := "--raw"
		if opt == "WP_AUTO_UPDATE_CORE" && value == "minor" {
			rawFlag = ""
		}
		wpBaseCmd := wpCLIPHPEnvFlags + " " + rawFlag + " --type=constant --path=" + docroot + " --skip-themes --allow-root"
		commands = append(commands, strings.TrimSpace(wpBaseCmd)+" "+opt+" "+value)
	}
	fullCmd := strings.Join(commands, " && ")
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c", fullCmd)
	if runErr := podmanmanager.Command(r.Context(), userContext, argv).Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "edited auto-update preferences for WordPress website in "+docroot, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Update preferences saved successfully"})
}

var defineBoolOrStringRE = regexp.MustCompile(`(?i)define\(\s*['"](?P<name>[^'"]+)['"]\s*,\s*(?P<value>true|false|['"][^'"]+['"])\s*\)\s*;`)

func handleWPCLIDebugInfo(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext, wpConfigFile, docroot string, wpBase []string) {
	debugConstants := map[string]string{"WP_DEBUG": "false", "WP_DEBUG_LOG": "false", "WP_DEBUG_DISPLAY": "false", "SCRIPT_DEBUG": "false", "SAVEQUERIES": "false"}
	content, readErr := os.ReadFile(wpConfigFile)
	if readErr == nil {
		fillDefinesFromContent(string(content), debugConstants)
	} else {
		fillDefinesFromWPCLIConfigList(r, userContext, docroot, wpBase, debugConstants)
	}
	writeJSON(w, http.StatusOK, debugConstants)
}

func handleWPCLIUpdateInfo(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext, wpConfigFile, docroot string, wpBase []string) {
	updateConstants := map[string]string{"WP_AUTO_UPDATE_PLUGINS": "false", "WP_AUTO_UPDATE_THEMES": "false", "WP_AUTO_UPDATE_CORE": "minor"}
	content, readErr := os.ReadFile(wpConfigFile)
	if readErr == nil {
		fillDefinesFromContent(string(content), updateConstants)
	} else {
		fillDefinesFromWPCLIConfigList(r, userContext, docroot, wpBase, updateConstants)
	}
	writeJSON(w, http.StatusOK, updateConstants)
}

func fillDefinesFromContent(content string, constants map[string]string) {
	for _, m := range defineBoolOrStringRE.FindAllStringSubmatch(content, -1) {
		name := m[1]
		value := strings.ToLower(strings.Trim(m[2], `'"`))
		if _, ok := constants[name]; ok {
			constants[name] = value
		}
	}
}

func fillDefinesFromWPCLIConfigList(r *http.Request, userContext, docroot string, wpBase []string, constants map[string]string) {
	argv := append(append([]string{}, wpBase...), "config", "list", "--format=json", "--path="+docroot, "--skip-themes", "--allow-root")
	out, err := podmanmanager.Command(r.Context(), userContext, argv).Output()
	if err != nil {
		return
	}
	var configData []map[string]any
	if json.Unmarshal(out, &configData) != nil {
		return
	}
	for name := range constants {
		for _, cfg := range configData {
			if cfg["name"] == name {
				constants[name] = strings.ToLower(fmtValue(cfg["value"]))
				break
			}
		}
	}
}

func fmtValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return ""
	}
}
