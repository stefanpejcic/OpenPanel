// Package websites implements the /sites listing page, the /website
// CMS-type dispatcher, and the side JSON endpoints (safebrowsing,
// PageSpeed, WP vulnerability scan, pm2 package installs, WP info, and the
// distinct /wordpress/wp-cli/<action> passthrough used by the site-manager
// UI). Mautic is not a supported CMS type.
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

// ---------------------- CMS DB INFO (shared) ---------------------- //

// cmsDBField pairs an info map key with the precompiled regex that extracts
// it from a CMS config file's text.
type cmsDBField struct {
	key string
	re  *regexp.Regexp
}

// cmsMappedDir maps a site's docroot-relative directory (as stored in the
// "/var/www/html/..." form every handler in this package uses) to its
// backing path on the host filesystem.
func cmsMappedDir(userContext, directory string) (string, bool) {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(directory, wwwPrefix) {
		return "", false
	}
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(directory, wwwPrefix), true
}

// applyCMSDBFields runs each field's regex against text, first match wins.
func applyCMSDBFields(text string, fields []cmsDBField) map[string]string {
	info := map[string]string{}
	for _, f := range fields {
		if _, exists := info[f.key]; exists {
			continue
		}
		if m := f.re.FindStringSubmatch(text); m != nil {
			info[f.key] = m[1]
		}
	}
	return info
}

// readCMSConfig reads configPath and evaluates fields against its contents
// (after an optional preprocess step), producing the same
// {"error": "..."} / field-map shape every extract<CMS>DatabaseInfo
// function below returns.
func readCMSConfig(configPath, label string, fields []cmsDBField, preprocess func(string) string) map[string]string {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return map[string]string{"error": label + " not found"}
	}
	text := string(content)
	if preprocess != nil {
		text = preprocess(text)
	}
	info := applyCMSDBFields(text, fields)
	if len(info) == 0 {
		return map[string]string{"error": "No database information found in " + label}
	}
	return info
}

// ---------------------- WP INFO ---------------------- //

// wpDBFieldSets holds the DB_NAME/DB_USER/DB_PASSWORD/DB_HOST and
// $table_prefix patterns for wp-config.php, once for each quote style.
// Go's RE2 has no backreferences to tie the opening/closing quote character
// together (both single or both double), so each quote style is tried
// separately - single-quote fields win when both are present, matching the
// original per-call regexp.MustCompile behavior this replaces.
var wpDBFieldSets = [2][]cmsDBField{
	{
		{"database_name", regexp.MustCompile(`define\(\s*'DB_NAME'\s*,\s*'(.+?)'\s*\)`)},
		{"database_user", regexp.MustCompile(`define\(\s*'DB_USER'\s*,\s*'(.+?)'\s*\)`)},
		{"database_password", regexp.MustCompile(`define\(\s*'DB_PASSWORD'\s*,\s*'(.+?)'\s*\)`)},
		{"database_host", regexp.MustCompile(`define\(\s*'DB_HOST'\s*,\s*'(.+?)'\s*\)`)},
		{"database_table_prefix", regexp.MustCompile(`\$table_prefix\s*=\s*'(.+?)'`)},
	},
	{
		{"database_name", regexp.MustCompile(`define\(\s*"DB_NAME"\s*,\s*"(.+?)"\s*\)`)},
		{"database_user", regexp.MustCompile(`define\(\s*"DB_USER"\s*,\s*"(.+?)"\s*\)`)},
		{"database_password", regexp.MustCompile(`define\(\s*"DB_PASSWORD"\s*,\s*"(.+?)"\s*\)`)},
		{"database_host", regexp.MustCompile(`define\(\s*"DB_HOST"\s*,\s*"(.+?)"\s*\)`)},
		{"database_table_prefix", regexp.MustCompile(`\$table_prefix\s*=\s*"(.+?)"`)},
	},
}

// extractDatabaseInfo parses DB_NAME/DB_USER/DB_PASSWORD/DB_HOST and the
// table prefix out of wp-config.php.
func extractDatabaseInfo(userContext, directory string) map[string]string {
	mappedDir, ok := cmsMappedDir(userContext, directory)
	if !ok {
		return nil
	}
	content, err := os.ReadFile(filepath.Join(mappedDir, "wp-config.php"))
	if err != nil {
		return map[string]string{"error": "wp-config.php not found"}
	}
	text := string(content)
	info := map[string]string{}
	for _, fieldSet := range wpDBFieldSets {
		for key, val := range applyCMSDBFields(text, fieldSet) {
			if _, exists := info[key]; !exists {
				info[key] = val
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

// ---------------------- DRUPAL INFO ---------------------- //

var drupalDBFields = []cmsDBField{
	{"database_name", regexp.MustCompile(`'database'\s*=>\s*'([^']*)'`)},
	{"database_user", regexp.MustCompile(`'username'\s*=>\s*'([^']*)'`)},
	{"database_password", regexp.MustCompile(`'password'\s*=>\s*'([^']*)'`)},
	{"database_host", regexp.MustCompile(`'host'\s*=>\s*'([^']*)'`)},
}

// stripPHPStarCommentLines drops every line whose trimmed form starts with
// '*' - Drupal's stock settings.php has a large documentation block near
// the top with placeholder 'database' => 'database_name' style example
// lines inside a /** ... */ comment (every line prefixed with '*'), which
// would otherwise match before the real $databases['default'] array drush
// appends near the end of the file.
func stripPHPStarCommentLines(text string) string {
	var codeLines []string
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "*") {
			continue
		}
		codeLines = append(codeLines, line)
	}
	return strings.Join(codeLines, "\n")
}

// extractDrupalDatabaseInfo parses the $databases['default']['default']
// array out of settings.php - the live-read-from-config approach, same as
// extractDatabaseInfo above does for wp-config.php, so no DB credentials
// need to be persisted anywhere else.
func extractDrupalDatabaseInfo(userContext, directory string) map[string]string {
	mappedDir, ok := cmsMappedDir(userContext, directory)
	if !ok {
		return nil
	}
	configPath := filepath.Join(mappedDir, "web", "sites", "default", "settings.php")
	return readCMSConfig(configPath, "settings.php", drupalDBFields, stripPHPStarCommentLines)
}

// getDrupalVersion reads drupal/core-recommended's resolved version out of
// composer.lock, mirroring getWPVersion's live-read-from-disk approach.
func getDrupalVersion(userContext, realPath string) string {
	relPath := strings.TrimPrefix(realPath, "/var/www/html/")
	filePath := filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", relPath, "composer.lock")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Unknown"
	}
	var lock struct {
		Packages []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"packages"`
	}
	if json.Unmarshal(content, &lock) != nil {
		return "Unknown"
	}
	for _, pkg := range lock.Packages {
		if pkg.Name == "drupal/core-recommended" || pkg.Name == "drupal/core" {
			return pkg.Version
		}
	}
	return "Unknown"
}

// ---------------------- FLARUM INFO ---------------------- //

// flarumDBFields matches config.php's 'database' => [...] array - unlike
// Drupal's settings.php, Flarum's config.php is a plain generated PHP
// array with no preceding documentation/placeholder block to strip first.
var flarumDBFields = []cmsDBField{
	{"database_name", regexp.MustCompile(`'database'\s*=>\s*'([^']*)'`)},
	{"database_user", regexp.MustCompile(`'username'\s*=>\s*'([^']*)'`)},
	{"database_password", regexp.MustCompile(`'password'\s*=>\s*'([^']*)'`)},
	{"database_host", regexp.MustCompile(`'host'\s*=>\s*'([^']*)'`)},
}

// extractFlarumDatabaseInfo parses the 'database' array out of config.php
// - config.php lives directly at the docroot's base (unlike Drupal's
// nested sites/default/settings.php), since Flarum's own Paths value
// object writes it to $paths->base, not $paths->public.
func extractFlarumDatabaseInfo(userContext, directory string) map[string]string {
	mappedDir, ok := cmsMappedDir(userContext, directory)
	if !ok {
		return nil
	}
	configPath := filepath.Join(mappedDir, "config.php")
	return readCMSConfig(configPath, "config.php", flarumDBFields, nil)
}

// getFlarumVersion reads flarum/core's resolved version out of
// composer.lock, mirroring getDrupalVersion's live-read-from-disk approach.
func getFlarumVersion(userContext, realPath string) string {
	relPath := strings.TrimPrefix(realPath, "/var/www/html/")
	filePath := filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", relPath, "composer.lock")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Unknown"
	}
	var lock struct {
		Packages []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"packages"`
	}
	if json.Unmarshal(content, &lock) != nil {
		return "Unknown"
	}
	for _, pkg := range lock.Packages {
		if pkg.Name == "flarum/core" {
			return pkg.Version
		}
	}
	return "Unknown"
}

// extractJoomlaDatabaseInfo parses the $host/$user/$password/$db/$dbprefix
// properties out of configuration.php - much simpler than
// extractDrupalDatabaseInfo's settings.php scrape, since Joomla's installer
// writes a plain generated PHP class with no preceding documentation/
// placeholder block to accidentally match first.
var joomlaDBFields = []cmsDBField{
	{"database_name", regexp.MustCompile(`\$db\s*=\s*'([^']*)'`)},
	{"database_user", regexp.MustCompile(`\$user\s*=\s*'([^']*)'`)},
	{"database_password", regexp.MustCompile(`\$password\s*=\s*'([^']*)'`)},
	{"database_host", regexp.MustCompile(`\$host\s*=\s*'([^']*)'`)},
	{"database_prefix", regexp.MustCompile(`\$dbprefix\s*=\s*'([^']*)'`)},
}

func extractJoomlaDatabaseInfo(userContext, directory string) map[string]string {
	mappedDir, ok := cmsMappedDir(userContext, directory)
	if !ok {
		return nil
	}
	return readCMSConfig(filepath.Join(mappedDir, "configuration.php"), "configuration.php", joomlaDBFields, nil)
}

// getJoomlaVersion reads the MAJOR/MINOR/PATCH version constants out of
// libraries/src/Version.php, mirroring getDrupalVersion's live-read-from-
// disk approach (no persisted version tracking anywhere else).
func getJoomlaVersion(userContext, realPath string) string {
	relPath := strings.TrimPrefix(realPath, "/var/www/html/")
	filePath := filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", relPath, "libraries", "src", "Version.php")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Unknown"
	}
	text := string(content)
	major := regexp.MustCompile(`MAJOR_VERSION\s*=\s*(\d+)`).FindStringSubmatch(text)
	minor := regexp.MustCompile(`MINOR_VERSION\s*=\s*(\d+)`).FindStringSubmatch(text)
	patch := regexp.MustCompile(`PATCH_VERSION\s*=\s*(\d+)`).FindStringSubmatch(text)
	if major == nil || minor == nil || patch == nil {
		return "Unknown"
	}
	return major[1] + "." + minor[1] + "." + patch[1]
}

// extractOpenCartDatabaseInfo parses the DB_HOSTNAME/DB_USERNAME/
// DB_PASSWORD/DB_DATABASE/DB_PREFIX constants out of config.php - same
// plain-define()-list shape as Joomla's configuration.php, no comment-block
// gotcha to work around.
var openCartDBFields = []cmsDBField{
	{"database_name", regexp.MustCompile(`DB_DATABASE'\s*,\s*'([^']*)'`)},
	{"database_user", regexp.MustCompile(`DB_USERNAME'\s*,\s*'([^']*)'`)},
	{"database_password", regexp.MustCompile(`DB_PASSWORD'\s*,\s*'([^']*)'`)},
	{"database_host", regexp.MustCompile(`DB_HOSTNAME'\s*,\s*'([^']*)'`)},
	{"database_prefix", regexp.MustCompile(`DB_PREFIX'\s*,\s*'([^']*)'`)},
}

func extractOpenCartDatabaseInfo(userContext, directory string) map[string]string {
	mappedDir, ok := cmsMappedDir(userContext, directory)
	if !ok {
		return nil
	}
	return readCMSConfig(filepath.Join(mappedDir, "config.php"), "config.php", openCartDBFields, nil)
}

// getOpenCartVersion reads the VERSION constant out of the top-level
// index.php, mirroring getJoomlaVersion/getDrupalVersion's live-read-from-
// disk approach (no persisted version tracking anywhere else).
func getOpenCartVersion(userContext, realPath string) string {
	relPath := strings.TrimPrefix(realPath, "/var/www/html/")
	filePath := filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", relPath, "index.php")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Unknown"
	}
	m := regexp.MustCompile(`VERSION'\s*,\s*'([^']*)'`).FindStringSubmatch(string(content))
	if m == nil {
		return "Unknown"
	}
	return m[1]
}

// extractNextcloudDatabaseInfo parses the dbname/dbuser/dbpassword/dbhost/
// dbtableprefix entries out of config/config.php's `$CONFIG = array(...)`
// literal.
var nextcloudDBFields = []cmsDBField{
	{"database_name", regexp.MustCompile(`'dbname'\s*=>\s*'([^']*)'`)},
	{"database_user", regexp.MustCompile(`'dbuser'\s*=>\s*'([^']*)'`)},
	{"database_password", regexp.MustCompile(`'dbpassword'\s*=>\s*'([^']*)'`)},
	{"database_host", regexp.MustCompile(`'dbhost'\s*=>\s*'([^']*)'`)},
	{"database_prefix", regexp.MustCompile(`'dbtableprefix'\s*=>\s*'([^']*)'`)},
}

func extractNextcloudDatabaseInfo(userContext, directory string) map[string]string {
	mappedDir, ok := cmsMappedDir(userContext, directory)
	if !ok {
		return nil
	}
	return readCMSConfig(filepath.Join(mappedDir, "config", "config.php"), "config/config.php", nextcloudDBFields, nil)
}

// getNextcloudVersion reads $OC_VersionString out of version.php,
// mirroring getOpenCartVersion/getJoomlaVersion's live-read-from-disk
// approach.
func getNextcloudVersion(userContext, realPath string) string {
	relPath := strings.TrimPrefix(realPath, "/var/www/html/")
	filePath := filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", relPath, "version.php")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Unknown"
	}
	m := regexp.MustCompile(`OC_VersionString\s*=\s*'([^']*)'`).FindStringSubmatch(string(content))
	if m == nil {
		return "Unknown"
	}
	return m[1]
}

// extractPrestashopDatabaseInfo parses the database_name/database_user/
// database_password/database_host/database_prefix entries out of
// app/config/parameters.php's `<?php return array('parameters' =>
// array(...))` literal.
var prestashopDBFields = []cmsDBField{
	{"database_name", regexp.MustCompile(`'database_name'\s*=>\s*'([^']*)'`)},
	{"database_user", regexp.MustCompile(`'database_user'\s*=>\s*'([^']*)'`)},
	{"database_password", regexp.MustCompile(`'database_password'\s*=>\s*'([^']*)'`)},
	{"database_host", regexp.MustCompile(`'database_host'\s*=>\s*'([^']*)'`)},
	{"database_prefix", regexp.MustCompile(`'database_prefix'\s*=>\s*'([^']*)'`)},
}

func extractPrestashopDatabaseInfo(userContext, directory string) map[string]string {
	mappedDir, ok := cmsMappedDir(userContext, directory)
	if !ok {
		return nil
	}
	return readCMSConfig(filepath.Join(mappedDir, "app", "config", "parameters.php"), "app/config/parameters.php", prestashopDBFields, nil)
}

// getPrestashopVersion reads the VERSION const out of src/Core/Version.php,
// mirroring getNextcloudVersion/getOpenCartVersion's live-read-from-disk
// approach.
func getPrestashopVersion(userContext, realPath string) string {
	relPath := strings.TrimPrefix(realPath, "/var/www/html/")
	filePath := filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", relPath, "src", "Core", "Version.php")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Unknown"
	}
	m := regexp.MustCompile(`const VERSION\s*=\s*'([^']*)'`).FindStringSubmatch(string(content))
	if m == nil {
		return "Unknown"
	}
	return m[1]
}

// extractMatomoDatabaseInfo parses the dbname/username/password/host/
// tables_prefix entries out of config/config.ini.php's [database] section -
// Matomo's own generated config, an INI file (not a PHP array/define()
// list the way every other CMS here writes its config), so this matches on
// `key = "value"` lines instead.
var matomoDBFields = []cmsDBField{
	{"database_name", regexp.MustCompile(`(?m)^dbname\s*=\s*"([^"]*)"`)},
	{"database_user", regexp.MustCompile(`(?m)^username\s*=\s*"([^"]*)"`)},
	{"database_password", regexp.MustCompile(`(?m)^password\s*=\s*"([^"]*)"`)},
	{"database_host", regexp.MustCompile(`(?m)^host\s*=\s*"([^"]*)"`)},
	{"database_prefix", regexp.MustCompile(`(?m)^tables_prefix\s*=\s*"([^"]*)"`)},
}

func extractMatomoDatabaseInfo(userContext, directory string) map[string]string {
	mappedDir, ok := cmsMappedDir(userContext, directory)
	if !ok {
		return nil
	}
	return readCMSConfig(filepath.Join(mappedDir, "config", "config.ini.php"), "config/config.ini.php", matomoDBFields, nil)
}

// getMatomoVersion reads the VERSION const out of core/Version.php,
// mirroring getPrestashopVersion/getNextcloudVersion's live-read-from-disk
// approach.
func getMatomoVersion(userContext, realPath string) string {
	relPath := strings.TrimPrefix(realPath, "/var/www/html/")
	filePath := filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", relPath, "core", "Version.php")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Unknown"
	}
	m := regexp.MustCompile(`VERSION\s*=\s*'([^']*)'`).FindStringSubmatch(string(content))
	if m == nil {
		return "Unknown"
	}
	return m[1]
}

// moodleApprootDir maps a Moodle site's docroot (a symlink to
// <approot>/public - see internal/modules/moodle's package doc comment for
// why) to its backing "app root" directory, where config.php/admin/lib/
// actually live. directory's relative-to-docroot part IS the site's slash-
// joined domain+subdirectory site name (docroot is literally
// "/var/www/html/"+that), so this reapplies moodle.siteSlug's exact
// domain/dot -> underscore substitution to find the sibling
// "<slug>_moodleapp" directory install.go created, without needing to
// resolve the symlink.
func moodleApprootDir(userContext, directory string) string {
	const wwwPrefix = "/var/www/html/"
	relPath := strings.TrimPrefix(directory, wwwPrefix)
	slug := strings.ReplaceAll(strings.ReplaceAll(relPath, "/", "_"), ".", "_")
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + slug + "_moodleapp"
}

// extractMoodleDatabaseInfo parses the $CFG->dbname/dbuser/dbpass/dbhost/
// prefix plain-variable assignments out of the approot's config.php -
// Moodle's own generated config, not reachable via the docroot symlink
// (that only leads to public/config.php, a thin shim, not the real one).
var moodleDBFields = []cmsDBField{
	{"database_name", regexp.MustCompile(`CFG->dbname\s*=\s*'([^']*)'`)},
	{"database_user", regexp.MustCompile(`CFG->dbuser\s*=\s*'([^']*)'`)},
	{"database_password", regexp.MustCompile(`CFG->dbpass\s*=\s*'([^']*)'`)},
	{"database_host", regexp.MustCompile(`CFG->dbhost\s*=\s*'([^']*)'`)},
	{"database_prefix", regexp.MustCompile(`CFG->prefix\s*=\s*'([^']*)'`)},
}

func extractMoodleDatabaseInfo(userContext, directory string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(directory, wwwPrefix) {
		return nil
	}
	configPath := filepath.Join(moodleApprootDir(userContext, directory), "config.php")
	return readCMSConfig(configPath, "config.php", moodleDBFields, nil)
}

// getMoodleVersion reads $release out of public/version.php (the
// human-readable "5.2.1+ (Build: 20260807)"-style string every Moodle
// release ships, confirmed live against a real 5.2 release tarball -
// $version is the internal numeric build timestamp, not a human version;
// version.php itself lives under public/, unlike config.php).
func getMoodleVersion(userContext, directory string) string {
	content, err := os.ReadFile(filepath.Join(moodleApprootDir(userContext, directory), "public", "version.php"))
	if err != nil {
		return "Unknown"
	}
	m := regexp.MustCompile(`\$release\s*=\s*'([^']*)'`).FindStringSubmatch(string(content))
	if m == nil {
		return "Unknown"
	}
	return m[1]
}

// extractMediaWikiDatabaseInfo parses the $wgDBname/$wgDBuser/$wgDBpassword/
// $wgDBserver/$wgDBprefix plain-variable assignments out of the docroot's
// LocalSettings.php - MediaWiki is installed flat (no approot/public split
// like Moodle), so this reads directly from directory, mirroring
// extractJoomlaDatabaseInfo's shape.
var mediaWikiDBFields = []cmsDBField{
	{"database_name", regexp.MustCompile(`\$wgDBname\s*=\s*"([^"]*)"`)},
	{"database_user", regexp.MustCompile(`\$wgDBuser\s*=\s*"([^"]*)"`)},
	{"database_password", regexp.MustCompile(`\$wgDBpassword\s*=\s*"([^"]*)"`)},
	{"database_host", regexp.MustCompile(`\$wgDBserver\s*=\s*"([^"]*)"`)},
	{"database_prefix", regexp.MustCompile(`\$wgDBprefix\s*=\s*"([^"]*)"`)},
}

func extractMediaWikiDatabaseInfo(userContext, directory string) map[string]string {
	mappedDir, ok := cmsMappedDir(userContext, directory)
	if !ok {
		return nil
	}
	return readCMSConfig(filepath.Join(mappedDir, "LocalSettings.php"), "LocalSettings.php", mediaWikiDBFields, nil)
}

// getMediaWikiVersion reads the $wgVersion constant out of
// includes/Defines.php (every MediaWiki release ships this - confirmed
// live against a real 1.42 release tarball).
func getMediaWikiVersion(userContext, directory string) string {
	const wwwPrefix = "/var/www/html/"
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(directory, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "includes", "Defines.php"))
	if err != nil {
		return "Unknown"
	}
	m := regexp.MustCompile(`define\(\s*'MW_VERSION',\s*'([^']*)'\s*\)`).FindStringSubmatch(string(content))
	if m == nil {
		return "Unknown"
	}
	return m[1]
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

// wpInfoForSite resolves a WordPress site's database credentials, WP
// version, PHP version, and MySQL version. Shared by the UI's
// handleWebsiteWPInfo and the API's apiWPInfo. Returns ok=false if the
// domain's docroot couldn't be found.
func wpInfoForSite(a *appctx.App, r *http.Request, userContext, siteName string) (map[string]any, bool) {
	ctx := r.Context()
	domainNameUsed, folderParam := splitDomainAndFolder(siteName)

	var docroot string
	row := a.DB.QueryRowContext(ctx, "SELECT docroot FROM domains WHERE domain_url = ?", domainNameUsed)
	if scanErr := row.Scan(&docroot); scanErr != nil {
		return nil, false
	}
	if folderParam != "" {
		docroot = docroot + "/" + folderParam
	}

	databaseInfo := extractDatabaseInfo(userContext, docroot)
	wpVersionFile := filepath.Join(docroot, "wp-includes", "version.php")
	wpVersion := getWPVersion(userContext, wpVersionFile)
	phpVersion := php.GetPHPVForDomain(ctx, a, userContext, domainNameUsed)
	mysqlVersion := getMySQLVersion(a, r, userContext)

	return map[string]any{
		"database_info": databaseInfo, "wp_version": wpVersion,
		"php_version": phpVersion, "mysql_version": mysqlVersion,
	}, true
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
	domainNameUsed, _ := splitDomainAndFolder(siteName)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainNameUsed) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	info, ok := wpInfoForSite(a, r, userContext, siteName)
	if !ok {
		flashAndRedirect(a, w, r, "error", "Unable to detect docroot for the domain.", "/sites")
		return
	}

	writeJSON(w, http.StatusOK, info)
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
