package nextcloud

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

// This file mirrors wordpress/manage.go's handleCloneWordPress in
// structure (site-limit check, file copy, DB create+dump-pipe, config
// rewrite, sites-table insert). Nextcloud is the most config-heavy of the
// 5: config/config.php (confirmed live against a real installed
// Nextcloud's $CONFIG array) hardcodes 'datadirectory' as an ABSOLUTE
// container path that MUST match the new docroot after the file copy or
// Nextcloud refuses to serve, plus 'trusted_domains' (a numbered PHP
// array, index 1 holds the install-time domain per install.go), plus
// 'overwrite.cli.url', plus the usual dbname/dbuser/dbpassword. All of
// these are fixed via plain regex text-replace on config.php directly
// rather than shelling out to `occ config:system:set` - occ requires a
// working PHP runtime matching Nextcloud's own minimum version, and this
// session already hit PHP-version failures running occ against some test
// containers (see maintenance.go's handler failing the same way on a
// PHP 7.2 container); a text edit has no such runtime dependency and this
// codebase already edits config.php directly elsewhere (extractNextcloud-
// DatabaseInfoForLogin reads it the same way). 'instanceid' is
// deliberately left untouched: the file copy also copies
// data/appdata_<instanceid>/, so keeping the same instanceid keeps that
// directory's name consistent with what's now inside it.
//
// cloneValidateDocroot accepts the real "/var/www/html/..." absolute-path
// form .Docroot actually uses everywhere else in this codebase -
// WordPress's own validateDocroot() rejects any leading "/", which would
// reject its own clone form's real source_folder value. Not replicating
// that bug here.

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
	cloneNCDataDirRE   = regexp.MustCompile(`'datadirectory'\s*=>\s*'.*?',`)
	cloneNCOverwriteRE = regexp.MustCompile(`'overwrite\.cli\.url'\s*=>\s*'.*?',`)
	cloneNCDBNameRE    = regexp.MustCompile(`'dbname'\s*=>\s*'.*?',`)
	cloneNCDBUserRE    = regexp.MustCompile(`'dbuser'\s*=>\s*'.*?',`)
	cloneNCDBPasswdRE  = regexp.MustCompile(`'dbpassword'\s*=>\s*'.*?',`)
	// Matches the trusted_domains array's second entry specifically
	// (index 1, right after the always-present "0 => 'localhost'," line -
	// see install.go's trusted_domains handling for why index 1 is where
	// the site's own domain lives).
	cloneNCTrustedDomainRE = regexp.MustCompile(`(0 => 'localhost',\s*\n\s*1 => )'.*?'(,)`)
)

// handleNextcloudClone mirrors wordpress/manage.go's handleCloneWordPress.
func handleNextcloudClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dstDB := strings.ToLower(formOr(r, "target_db", "nc_clone_"+generateRandomString(6)))
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Nextcloud files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Nextcloud files: " + cpErr.Error()})
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

	configFile := filepath.Join(dstPath, "config", "config.php")
	content, readErr := os.ReadFile(configFile)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": readErr.Error()})
		return
	}
	strContent := string(content)
	strContent = cloneNCDataDirRE.ReplaceAllString(strContent, "'datadirectory' => '"+escapePHPSingleQuoted(strings.TrimSuffix(docroot, "/")+"/data")+"',")
	strContent = cloneNCOverwriteRE.ReplaceAllString(strContent, "'overwrite.cli.url' => 'https://"+escapePHPSingleQuoted(dstDomainWithSubdir)+"',")
	strContent = cloneNCDBNameRE.ReplaceAllString(strContent, "'dbname' => '"+escapePHPSingleQuoted(dstDB)+"',")
	strContent = cloneNCDBUserRE.ReplaceAllString(strContent, "'dbuser' => '"+escapePHPSingleQuoted(dstDBUser)+"',")
	strContent = cloneNCDBPasswdRE.ReplaceAllString(strContent, "'dbpassword' => '"+escapePHPSingleQuoted(dstDBUserPassword)+"',")
	strContent = cloneNCTrustedDomainRE.ReplaceAllString(strContent, "${1}'"+escapePHPSingleQuoted(dstDomain)+"'${2}")
	if writeErr := os.WriteFile(configFile, []byte(strContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": writeErr.Error()})
		return
	}
	if !strings.Contains(strContent, dstDB) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "Failed to set dbname in config.php"})
		return
	}

	_ = a.Cache.Delete(ctx, "get_user_websites:"+itoa(userID))

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	nextcloudVersion := formOr(r, "nextcloud_version", "latest")
	if _, insertErr := a.DB.ExecContext(ctx, `
		INSERT INTO sites (site_name, domain_id, admin_email, version, type)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE domain_id = VALUES(domain_id), admin_email = VALUES(admin_email), version = VALUES(version), type = VALUES(type)`,
		dstDomainWithSubdir, domainID, adminEmail, nextcloudVersion, "nextcloud"); insertErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cloned Nextcloud website from "+providedDomain+" to "+dstDomainWithSubdir, reqip.ClientIP(r))

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
