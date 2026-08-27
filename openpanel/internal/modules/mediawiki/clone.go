package mediawiki

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cmsclone"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
)

// This file mirrors drupal/clone.go in structure (site-limit check, file
// copy, DB create+dump-pipe, config rewrite, sites-table insert), sharing
// everything but the docroot copy and config-rewrite steps with every
// other CMS's clone.go via internal/core/cmsclone - see that package's doc
// comment for why those two steps stay local. Two MediaWiki-specific
// differences from Drupal:
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

	dstDB := strings.ToLower(formOr(r, "target_db", "mediawiki_clone_"+generateRandomString(6)))
	dstDBUser := strings.ToLower(formOr(r, "target_db_user", dstDB))
	dstDBUserPassword := formOr(r, "target_db_user_password", generateRandomString(16))

	if providedDomain == "" || dstDomain == "" || srcDB == "" || srcFolder == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing required form fields"})
		return
	}

	domainID, docroot, phpVersion, dstDomainWithSubdir, ok := cmsclone.ResolveDestination(ctx, a, dstDomain, dstFolder)
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy MediaWiki files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy MediaWiki files: " + cpErr.Error()})
		return
	}
	cmsclone.ChownRecursive(ctx, userContext, dstPath)
	settingsFile := filepath.Join(dstPath, "LocalSettings.php")
	_ = exec.CommandContext(ctx, "chmod", "644", settingsFile).Run()

	escapedPassword := strings.ReplaceAll(dstDBUserPassword, `"`, `\"`)
	_, dbErr := cmsclone.CreateDatabaseAndDump(ctx, userContext, mysqlVersion, dumpCmd, srcDB, dstDB, dstDBUser, dstDBUserPassword)
	if dbErr != nil {
		if cmsclone.DumpStageFailed(dbErr) {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "command_failed"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": dbErr.Error()})
		}
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

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	mediawikiVersion := formOr(r, "mediawiki_version", "latest")
	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "MediaWiki", CMSType: "mediawiki",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: mediawikiVersion,
		SrcPath: srcPath, DstPath: dstPath, DstDB: dstDB,
	})
}
