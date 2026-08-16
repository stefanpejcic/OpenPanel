package joomla

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
)

// This file mirrors wordpress/manage.go's handleCloneWordPress exactly in
// structure (site-limit check, file copy, DB create+dump-pipe, config
// rewrite, sites-table insert) - only two things differ from WordPress:
//
//  1. Joomla's configuration.php has no hardcoded site URL (Joomla derives
//     it from the request at runtime), so there is no wp-cli
//     `search-replace`-equivalent step here. This is a known, deliberate
//     limitation: any URL a user has hardcoded into article/module content
//     will still point at the source domain after cloning - out of scope,
//     same as a stock Joomla site itself doesn't rewrite such content on a
//     manual domain move either.
//  2. validateDocroot below intentionally accepts the "/var/www/html/..."
//     absolute-path form every other handler in this package already uses
//     for .Docroot (see joomlaRequestParams) - WordPress's own
//     validateDocroot() rejects any leading "/", which would reject the
//     real .Docroot value WordPress's own clone form submits as
//     source_folder. Not replicating that bug here.

var (
	cloneValidDomainRE = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	cloneValidDBRE      = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func cloneValidateDomain(name string) bool { return name != "" && cloneValidDomainRE.MatchString(name) }
func cloneValidateDB(name string) bool     { return name != "" && cloneValidDBRE.MatchString(name) }
func cloneValidateDocroot(path string) bool {
	return path != "" && !strings.Contains(path, "..") && strings.HasPrefix(path, "/var/www/html/")
}

var (
	cloneJoomlaUserRE     = regexp.MustCompile(`\$user\s*=\s*'.*?';`)
	cloneJoomlaPasswordRE = regexp.MustCompile(`\$password\s*=\s*'.*?';`)
	cloneJoomlaDBRE       = regexp.MustCompile(`\$db\s*=\s*'.*?';`)
)

// handleJoomlaClone mirrors wordpress/manage.go's handleCloneWordPress.
func handleJoomlaClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dstDB := strings.ToLower(formOr(r, "target_db", "joomla_clone_"+generateRandomString(6)))
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Joomla files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Joomla files: " + cpErr.Error()})
		return
	}
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		_ = exec.CommandContext(ctx, "chown", "-R", itoa(uid)+":"+itoa(uid), dstPath).Run()
	}

	escapedPassword := strings.ReplaceAll(strings.ReplaceAll(dstDBUserPassword, `\`, `\\`), `'`, `\'`)
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

	// srcDB/dstDB are already validated against ^[a-zA-Z0-9_]+$ by
	// cloneValidateDB, so no identifier quoting is needed here - backticks
	// would be actively wrong: bash -c interprets an unescaped backtick as
	// command substitution, not as passed-through SQL quoting, which
	// silently truncated this exact command when copied from WordPress's
	// clone (verified live: bash treated `srcDB` as "run srcDB as a command").
	dumpTablesCmd := dumpCmd + " --single-transaction --quick " + srcDB + " | " + mysqlVersion + " " + dstDB
	fullDBArgv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, "bash", "-c", dumpTablesCmd)
	if runErr := podmanmanager.Command(ctx, userContext, fullDBArgv).Run(); runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "command_failed"})
		return
	}

	configFile := filepath.Join(dstPath, "configuration.php")
	content, readErr := os.ReadFile(configFile)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": readErr.Error()})
		return
	}
	strContent := string(content)
	// Go's regexp.ReplaceAllString treats a literal '$' in the replacement
	// string as a capture-group reference (e.g. $user would look for a
	// group named "user" and silently substitute empty string when none
	// exists) - '$$' is the escape for a literal '$', which every
	// replacement below needs since Joomla's configuration.php uses
	// "public $propertyName = ...;" syntax.
	strContent = cloneJoomlaUserRE.ReplaceAllString(strContent, "$$user = '"+escapePHPSingleQuoted(dstDBUser)+"';")
	strContent = cloneJoomlaPasswordRE.ReplaceAllString(strContent, "$$password = '"+escapePHPSingleQuoted(dstDBUserPassword)+"';")
	strContent = cloneJoomlaDBRE.ReplaceAllString(strContent, "$$db = '"+escapePHPSingleQuoted(dstDB)+"';")
	if writeErr := os.WriteFile(configFile, []byte(strContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": writeErr.Error()})
		return
	}
	if !strings.Contains(strContent, dstDB) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "Failed to set $db in configuration.php"})
		return
	}

	_ = a.Cache.Delete(ctx, "get_user_websites:"+itoa(userID))

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	joomlaVersion := formOr(r, "joomla_version", "latest")
	if _, insertErr := a.DB.ExecContext(ctx, `
		INSERT INTO sites (site_name, domain_id, admin_email, version, type)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE domain_id = VALUES(domain_id), admin_email = VALUES(admin_email), version = VALUES(version), type = VALUES(type)`,
		dstDomainWithSubdir, domainID, adminEmail, joomlaVersion, "joomla"); insertErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cloned Joomla website from "+providedDomain+" to "+dstDomainWithSubdir, reqip.ClientIP(r))

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "success", "source": providedDomain, "target": dstDomainWithSubdir,
		"source_path": srcPath, "target_path": dstPath, "target_db": dstDB,
	})
}

func escapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}
