package nextcloud

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cmsclone"
)

// This file mirrors wordpress/manage.go's handleCloneWordPress in overall
// shape (site-limit check, file copy, DB create+dump-pipe, config
// rewrite, sites-table insert), sharing everything but the docroot copy
// and config-rewrite steps with every other CMS's clone.go via
// internal/core/cmsclone - see that package's doc comment for why those
// two steps stay local. Nextcloud is the most config-heavy of the 8:
// config/config.php (confirmed live against a real installed Nextcloud's
// $CONFIG array) hardcodes 'datadirectory' as an ABSOLUTE container path
// that MUST match the new docroot after the file copy or Nextcloud refuses
// to serve, plus 'trusted_domains' (a numbered PHP array, index 1 holds
// the install-time domain per install.go), plus 'overwrite.cli.url', plus
// the usual dbname/dbuser/dbpassword. All of these are fixed via plain
// regex text-replace on config.php directly rather than shelling out to
// `occ config:system:set` - occ requires a working PHP runtime matching
// Nextcloud's own minimum version, and this session already hit
// PHP-version failures running occ against some test containers (see
// maintenance.go's handler failing the same way on a PHP 7.2 container);
// a text edit has no such runtime dependency and this codebase already
// edits config.php directly elsewhere (extractNextcloudDatabaseInfoForLogin
// reads it the same way). 'instanceid' is deliberately left untouched: the
// file copy also copies data/appdata_<instanceid>/, so keeping the same
// instanceid keeps that directory's name consistent with what's now inside
// it.
//
// cmsclone.ValidDocroot accepts the real "/var/www/html/..." absolute-path
// form .Docroot actually uses everywhere else in this codebase -
// WordPress's own validateDocroot() rejects any leading "/", which would
// reject its own clone form's real source_folder value. Not replicating
// that bug here.

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

	websiteCount, _ := countUserWebsites(a, userID)
	if !cmsclone.WithinSiteLimit(ctx, a, userID, websiteCount) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "You have reached the maximum number of sites allowed" + a.UpgradeMessageForUser(ctx, userID)})
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

	domainID, docroot, _, dstDomainWithSubdir, ok := cmsclone.ResolveDestination(ctx, a, dstDomain, dstFolder)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Destination domain not found in database"})
		return
	}

	srcDomain := strings.Split(providedDomain, "/")[0]

	if !cmsclone.ValidDomain(srcDomain) || !cmsclone.ValidDomain(dstDomain) || !cmsclone.ValidDB(srcDB) || !cmsclone.ValidDB(dstDB) ||
		!cmsclone.ValidDB(dstDBUser) || !cmsclone.ValidDocroot(srcFolder) || !cmsclone.ValidDocroot(docroot) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid input or unsafe docroot"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, srcDomain) || !a.CheckDomainBelongsToUser(ctx, userID, dstDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	dumpCmd, mysqlVersion, dumpCmdErr := cmsclone.SelectDumpCommand(userContext)
	if dumpCmdErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": dumpCmdErr.Error()})
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
	cmsclone.ChownRecursive(ctx, userContext, dstPath)

	_, dbErr := cmsclone.CreateDatabaseAndDump(ctx, userContext, mysqlVersion, dumpCmd, srcDB, dstDB, dstDBUser, dstDBUserPassword)
	if dbErr != nil {
		if cmsclone.DumpStageFailed(dbErr) {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "command_failed"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": dbErr.Error()})
		}
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

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	nextcloudVersion := formOr(r, "nextcloud_version", "latest")
	// Rewrites hardcoded source-domain URLs left in page/content body text (the config-file rewrite above only fixes the DB connection settings, not application data) - the generic equivalent of wp-cli's search-replace, which this CMS's own CLI has no built-in version of.
	cmsclone.SearchReplaceDatabase(ctx, userContext, dstDB, "https://"+providedDomain, "https://"+dstDomainWithSubdir)

	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "Nextcloud", CMSType: "nextcloud",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: nextcloudVersion,
		SrcPath: srcPath, DstPath: dstPath, DstDB: dstDB,
	})
}

func escapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}
