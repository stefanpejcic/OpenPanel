package joomla

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
// rewrite, sites-table insert), sharing everything but the docroot
// copy and config-rewrite steps with every other CMS's clone.go via
// internal/core/cmsclone - see that package's doc comment for why those
// two steps stay local. Two things differ from WordPress specifically:
//
//  1. Joomla's configuration.php has no hardcoded site URL (Joomla derives
//     it from the request at runtime), so there is no wp-cli
//     `search-replace`-equivalent step here. This is a known, deliberate
//     limitation: any URL a user has hardcoded into article/module content
//     will still point at the source domain after cloning - out of scope,
//     same as a stock Joomla site itself doesn't rewrite such content on a
//     manual domain move either.
//  2. cmsclone.ValidDocroot intentionally accepts the "/var/www/html/..."
//     absolute-path form every other handler in this package already uses
//     for .Docroot (see joomlaRequestParams) - WordPress's own
//     validateDocroot() rejects any leading "/", which would reject the
//     real .Docroot value WordPress's own clone form submits as
//     source_folder. Not replicating that bug here.

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

	dstDB := strings.ToLower(formOr(r, "target_db", "joomla_clone_"+generateRandomString(6)))
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Joomla files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Joomla files: " + cpErr.Error()})
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

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	joomlaVersion := formOr(r, "joomla_version", "latest")
	// Rewrites hardcoded source-domain URLs left in page/content body text (the config-file rewrite above only fixes the DB connection settings, not application data) - the generic equivalent of wp-cli's search-replace, which this CMS's own CLI has no built-in version of.
	cmsclone.SearchReplaceDatabase(ctx, userContext, dstDB, "https://"+providedDomain, "https://"+dstDomainWithSubdir)

	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "Joomla", CMSType: "joomla",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: joomlaVersion,
		SrcPath: srcPath, DstPath: dstPath, DstDB: dstDB,
	})
}

func escapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}
