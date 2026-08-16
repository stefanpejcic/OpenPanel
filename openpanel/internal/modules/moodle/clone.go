package moodle

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

// This file mirrors drupal/clone.go's overall shape (site-limit check, DB
// create+dump-pipe, config rewrite, sites-table insert, cron
// registration), but the file-copy step is NOT a plain "copy the docroot"
// like every other module's clone - see install.go's package doc comment
// and its own approotHostPath/datarootHostPath/symlink construction
// (read in full before touching this file). Moodle's docroot is a
// symlink, not a real directory: the actual code lives in a sibling
// "<slug>_moodleapp/" directory and user data lives in a separate sibling
// "<slug>_moodledata/" directory, neither of which install.go's own
// dispatch.go/websites.go plumbing exposes as "source_folder" the way it
// does for every flat-docroot module - so this clone resolves both
// sibling directories itself from the source domain name, the same
// siteSlug() call install.go used to create them.

var (
	cloneValidDomainRE = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	cloneValidDBRE     = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

func cloneValidateDomain(name string) bool { return name != "" && cloneValidDomainRE.MatchString(name) }
func cloneValidateDB(name string) bool     { return name != "" && cloneValidDBRE.MatchString(name) }

var (
	cloneMoodleDBNameRE     = regexp.MustCompile(`CFG->dbname\s*=\s*'[^']*'`)
	cloneMoodleDBUserRE     = regexp.MustCompile(`CFG->dbuser\s*=\s*'[^']*'`)
	cloneMoodleDBPasswordRE = regexp.MustCompile(`CFG->dbpass\s*=\s*'[^']*'`)
	cloneMoodleWwwrootRE    = regexp.MustCompile(`CFG->wwwroot\s*=\s*'[^']*'`)
	cloneMoodleDatarootRE   = regexp.MustCompile(`CFG->dataroot\s*=\s*'[^']*'`)
)

// handleMoodleClone mirrors drupal/clone.go's handleDrupalClone, adapted
// for Moodle's approot/dataroot/symlink layout (see install.go).
func handleMoodleClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
	dstFolder := r.FormValue("subdirectory")

	dstDB := strings.ToLower(formOr(r, "target_db", "moodle_clone_"+generateRandomString(6)))
	dstDBUser := strings.ToLower(formOr(r, "target_db_user", dstDB))
	dstDBUserPassword := formOr(r, "target_db_user_password", generateRandomString(16))

	if providedDomain == "" || dstDomain == "" || srcDB == "" {
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

	if !cloneValidateDomain(srcDomain) || !cloneValidateDomain(dstDomain) || !cloneValidateDB(srcDB) || !cloneValidateDB(dstDB) || !cloneValidateDB(dstDBUser) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid input"})
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

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	srcSlug := siteSlug(providedDomain)
	dstSlug := siteSlug(dstDomainWithSubdir)

	srcApprootHostPath := filepath.Join(htmlVolume, srcSlug+"_moodleapp")
	srcDatarootHostPath := filepath.Join(htmlVolume, srcSlug+"_moodledata")
	dstApprootHostPath := filepath.Join(htmlVolume, dstSlug+"_moodleapp")
	dstApprootContainerPath := "/var/www/html/" + dstSlug + "_moodleapp"
	dstDatarootHostPath := filepath.Join(htmlVolume, dstSlug+"_moodledata")
	dstDatarootContainerPath := "/var/www/html/" + dstSlug + "_moodledata"

	if info, statErr := os.Stat(srcApprootHostPath); statErr != nil || !info.IsDir() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Source Moodle app directory not found: " + srcApprootHostPath})
		return
	}

	if mkErr := os.MkdirAll(dstApprootHostPath, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Moodle files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcApprootHostPath+"/.", dstApprootHostPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Moodle app files: " + cpErr.Error()})
		return
	}
	if mkErr := os.MkdirAll(dstDatarootHostPath, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create Moodle data directory: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcDatarootHostPath+"/.", dstDatarootHostPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Moodle data files: " + cpErr.Error()})
		return
	}
	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		_ = exec.CommandContext(ctx, "chown", "-R", itoa(uid)+":"+itoa(uid), dstApprootHostPath).Run()
		_ = exec.CommandContext(ctx, "chown", "-R", itoa(uid)+":"+itoa(uid), dstDatarootHostPath).Run()
	}
	_ = exec.Command("chmod", "-R", "777", dstDatarootHostPath).Run()

	// Docroot destination host path, and the symlink target must be the
	// container-visible path (/var/www/html/...), not the host filesystem
	// path - same gotcha install.go's own symlink call documents and was
	// fixed live for it this session (a host-path symlink resolves fine
	// via `ls` on the host but is broken from inside the php-fpm
	// container, which only sees its own /var/www/html/ bind mount).
	const wwwBaseDirectory = "/var/www/html/"
	dstHostOSPath := strings.Replace(filepath.Clean(docroot), wwwBaseDirectory, htmlVolume, 1)
	if _, statErr := os.Lstat(dstHostOSPath); statErr == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Destination path " + docroot + " already exists."})
		_ = os.RemoveAll(dstApprootHostPath)
		_ = os.RemoveAll(dstDatarootHostPath)
		return
	}
	if symErr := os.Symlink(dstApprootContainerPath+"/public", dstHostOSPath); symErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error creating web root symlink: " + symErr.Error()})
		_ = os.RemoveAll(dstApprootHostPath)
		_ = os.RemoveAll(dstDatarootHostPath)
		return
	}

	escapedPassword := strings.ReplaceAll(dstDBUserPassword, `\`, `\\`)
	escapedPassword = strings.ReplaceAll(escapedPassword, `'`, `\'`)
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

	configFile := filepath.Join(dstApprootHostPath, "config.php")
	content, readErr := os.ReadFile(configFile)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": readErr.Error()})
		return
	}
	strContent := string(content)
	strContent = cloneMoodleDBNameRE.ReplaceAllString(strContent, `CFG->dbname   = '`+dstDB+`'`)
	strContent = cloneMoodleDBUserRE.ReplaceAllString(strContent, `CFG->dbuser   = '`+dstDBUser+`'`)
	strContent = cloneMoodleDBPasswordRE.ReplaceAllString(strContent, `CFG->dbpass   = '`+escapedPassword+`'`)
	strContent = cloneMoodleWwwrootRE.ReplaceAllString(strContent, `CFG->wwwroot   = 'https://`+dstDomainWithSubdir+`'`)
	strContent = cloneMoodleDatarootRE.ReplaceAllString(strContent, `CFG->dataroot  = '`+dstDatarootContainerPath+`'`)
	if writeErr := os.WriteFile(configFile, []byte(strContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": writeErr.Error()})
		return
	}
	if !strings.Contains(strContent, dstDB) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "Failed to set 'dbname' in config.php"})
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + phpVersion
	}
	cronComment := moodleCronComment(dstDomainWithSubdir)
	cronCommand := "php " + dstApprootContainerPath + "/admin/cli/cron.php"
	_ = crons.AddJob(ctx, userContext, cronComment, "0 * * * * *", phpContainer, cronCommand, true)

	_ = a.Cache.Delete(ctx, "get_user_websites:"+itoa(userID))

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	moodleVersion := formOr(r, "moodle_version", "latest")
	if _, insertErr := a.DB.ExecContext(ctx, `
		INSERT INTO sites (site_name, domain_id, admin_email, version, type)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE domain_id = VALUES(domain_id), admin_email = VALUES(admin_email), version = VALUES(version), type = VALUES(type)`,
		dstDomainWithSubdir, domainID, adminEmail, moodleVersion, "moodle"); insertErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": insertErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cloned Moodle website from "+providedDomain+" to "+dstDomainWithSubdir, reqip.ClientIP(r))

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "success", "source": providedDomain, "target": dstDomainWithSubdir,
		"source_path": srcApprootHostPath, "target_path": dstApprootHostPath, "target_db": dstDB,
	})
}
