package mediawiki

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
)

// This file mirrors drupal/clone.go in structure (site-limit check, file
// copy, DB create+dump-pipe, config rewrite, sites-table insert). Two
// MediaWiki-specific differences from Drupal:
//
//  1. LocalSettings.php hardcodes $wgServer/$wgScriptPath to the source
//     domain (install.go passes --server=/--scriptpath= to
//     maintenance/install.php) - unlike Drupal, which derives its base URL
//     at runtime - so both must be rewritten to the clone's own domain or
//     the clone keeps generating/accepting links only for the source.
//  2. MediaWiki needs its own per-site cron job (maintenance/runJobs.php)
//     registered for the clone, mirroring install.go's own crons.AddJob
//     call - without it the clone's job queue never runs.

var (
	cloneValidDomainRE = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	cloneValidDBRE     = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func cloneValidateDomain(name string) bool { return name != "" && cloneValidDomainRE.MatchString(name) }
func cloneValidateDB(name string) bool     { return name != "" && cloneValidDBRE.MatchString(name) }
func cloneValidateDocroot(path string) bool {
	return path != "" && !strings.Contains(path, "..") && strings.HasPrefix(path, "/var/www/html/")
}

var (
	cloneMediaWikiDBNameRE     = regexp.MustCompile(`\$wgDBname\s*=\s*"[^"]*"`)
	cloneMediaWikiDBUserRE     = regexp.MustCompile(`\$wgDBuser\s*=\s*"[^"]*"`)
	cloneMediaWikiDBPasswordRE = regexp.MustCompile(`\$wgDBpassword\s*=\s*"[^"]*"`)
	cloneMediaWikiServerRE     = regexp.MustCompile(`\$wgServer\s*=\s*"[^"]*"`)
	cloneMediaWikiScriptPathRE = regexp.MustCompile(`\$wgScriptPath\s*=\s*"[^"]*"`)
)

// handleMediaWikiClone mirrors drupal/clone.go's handleDrupalClone.
func handleMediaWikiClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "You have reached the maximum number of sites allowed"})
		return
	}

	providedDomain := r.FormValue("source_domain")
	dstDomain := r.FormValue("target_domain")
	srcDB := r.FormValue("source_db")
	srcFolder := r.FormValue("source_folder")
	dstFolder := r.FormValue("subdirectory")

	dstDB := strings.ToLower(formOr(r, "target_db", "mediawiki_clone_"+generateRandomString(6)))
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

	if !cloneValidateDomain(srcDomain) || !cloneValidateDomain(dstDomain) || !cloneValidateDB(srcDB) || !cloneValidateDB(dstDB) ||
		!cloneValidateDB(dstDBUser) || !cloneValidateDocroot(srcFolder) || !cloneValidateDocroot(docroot) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid input or unsafe docroot"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, srcDomain) || !a.CheckDomainBelongsToUser(ctx, userID, dstDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy MediaWiki files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy MediaWiki files: " + cpErr.Error()})
		return
	}
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		_ = exec.CommandContext(ctx, "chown", "-R", itoa(uid)+":"+itoa(uid), dstPath).Run()
	}
	settingsFile := filepath.Join(dstPath, "LocalSettings.php")
	_ = exec.CommandContext(ctx, "chmod", "644", settingsFile).Run()

	escapedPassword := strings.ReplaceAll(dstDBUserPassword, `"`, `\"`)
	cloneQueries := []string{
		"CREATE DATABASE IF NOT EXISTS `" + dstDB + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		"CREATE USER IF NOT EXISTS '" + dstDBUser + "'@'%' IDENTIFIED BY '" + strings.ReplaceAll(strings.ReplaceAll(dstDBUserPassword, `\`, `\\`), `'`, `\'`) + "'",
		"GRANT ALL PRIVILEGES ON `" + dstDB + "`.* TO '" + dstDBUser + "'@'%'",
		"FLUSH PRIVILEGES",
	}
	for _, q := range cloneQueries {
		if _, execErr := mysqlmanager.Exec(ctx, userContext, q, ""); execErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": execErr.Error()})
			return
		}
	}

	// srcDB/dstDB are already validated against ^[a-zA-Z0-9_]+$ - no
	// identifier quoting needed, and backticks would be actively wrong
	// (bash -c treats an unescaped backtick as command substitution) - see
	// drupal/clone.go's identical comment, confirmed live there.
	dumpTablesCmd := dumpCmd + " --single-transaction --quick " + srcDB + " | " + mysqlVersion + " " + dstDB
	fullDBArgv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, "bash", "-c", dumpTablesCmd)
	if runErr := podmanmanager.Command(ctx, userContext, fullDBArgv).Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "command_failed"})
		return
	}

	content, readErr := os.ReadFile(settingsFile)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": readErr.Error()})
		return
	}
	dstScriptPath := ""
	if dstFolder != "" {
		dstScriptPath = "/" + dstFolder
	}
	strContent := string(content)
	// ReplaceAllString (not ReplaceAllStringFunc) would silently swallow
	// the "$wg..." prefix here: Go's regexp package treats a bare "$name"
	// in the REPLACEMENT string as a submatch-expansion reference, not
	// literal text, and since these regexes have no such named group it
	// expands to "" - confirmed live: this produced ` = "dbname";` (no
	// variable on the left of "="), a PHP parse error on every request to
	// the clone. ReplaceAllStringFunc never does submatch expansion.
	strContent = cloneMediaWikiDBNameRE.ReplaceAllStringFunc(strContent, func(string) string { return `$wgDBname = "` + dstDB + `"` })
	strContent = cloneMediaWikiDBUserRE.ReplaceAllStringFunc(strContent, func(string) string { return `$wgDBuser = "` + dstDBUser + `"` })
	strContent = cloneMediaWikiDBPasswordRE.ReplaceAllStringFunc(strContent, func(string) string { return `$wgDBpassword = "` + escapedPassword + `"` })
	strContent = cloneMediaWikiServerRE.ReplaceAllStringFunc(strContent, func(string) string { return `$wgServer = "https://` + dstDomain + `"` })
	strContent = cloneMediaWikiScriptPathRE.ReplaceAllStringFunc(strContent, func(string) string { return `$wgScriptPath = "` + dstScriptPath + `"` })
	if writeErr := os.WriteFile(settingsFile, []byte(strContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": writeErr.Error()})
		return
	}
	if !strings.Contains(strContent, dstDB) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "Failed to set 'wgDBname' in LocalSettings.php"})
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + phpVersion
	}
	cronComment := mediawikiCronComment(dstDomainWithSubdir)
	cronCommand := "php " + docroot + "/maintenance/runJobs.php --maxjobs=50"
	_ = crons.AddJob(ctx, userContext, cronComment, "0 * * * * *", phpContainer, cronCommand, true)

	_ = a.Cache.Delete(ctx, "get_user_websites:"+itoa(userID))

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	mediawikiVersion := formOr(r, "mediawiki_version", "latest")
	if _, insertErr := a.DB.ExecContext(ctx, `
		INSERT INTO sites (site_name, domain_id, admin_email, version, type)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE domain_id = VALUES(domain_id), admin_email = VALUES(admin_email), version = VALUES(version), type = VALUES(type)`,
		dstDomainWithSubdir, domainID, adminEmail, mediawikiVersion, "mediawiki"); insertErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cloned MediaWiki website from "+providedDomain+" to "+dstDomainWithSubdir, reqip.ClientIP(r))

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "success", "source": providedDomain, "target": dstDomainWithSubdir,
		"source_path": srcPath, "target_path": dstPath, "target_db": dstDB,
	})
}
