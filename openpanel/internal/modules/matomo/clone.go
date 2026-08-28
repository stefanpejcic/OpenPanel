package matomo

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

// This file mirrors drupal/clone.go in structure (site-limit check, file
// copy, DB create+dump-pipe, config rewrite, sites-table insert), sharing
// everything but the docroot copy and config-rewrite steps with every
// other CMS's clone.go via internal/core/cmsclone - see that package's doc
// comment for why those two steps stay local. Matomo's own difference:
// config/config.ini.php is INI, not a PHP array/define() list, and it
// hardcodes a [General] trusted_hosts[] array of allowed HTTP Host headers
// - confirmed live against a real install (php.tests.openpanel.org/analytics)
// - so the clone's own domain must be appended there or Matomo rejects
// every request to the clone with "untrusted host".

var (
	cloneMatomoDBNameRE       = regexp.MustCompile(`(?m)^dbname\s*=\s*"[^"]*"`)
	cloneMatomoDBUserRE       = regexp.MustCompile(`(?m)^username\s*=\s*"[^"]*"`)
	cloneMatomoDBPasswordRE   = regexp.MustCompile(`(?m)^password\s*=\s*"[^"]*"`)
	cloneMatomoTrustedHostsRE = regexp.MustCompile(`(?m)^trusted_hosts\[\]\s*=\s*"[^"]*"$`)
)

// handleMatomoClone mirrors drupal/clone.go's handleDrupalClone.
func handleMatomoClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dstDB := strings.ToLower(formOr(r, "target_db", "matomo_clone_"+generateRandomString(6)))
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Matomo files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Matomo files: " + cpErr.Error()})
		return
	}
	cmsclone.ChownRecursive(ctx, userContext, dstPath)
	configFile := filepath.Join(dstPath, "config", "config.ini.php")
	_ = exec.CommandContext(ctx, "chmod", "644", configFile).Run()

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

	content, readErr := os.ReadFile(configFile)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": readErr.Error()})
		return
	}
	strContent := string(content)
	strContent = cloneMatomoDBNameRE.ReplaceAllString(strContent, `dbname = "`+dstDB+`"`)
	strContent = cloneMatomoDBUserRE.ReplaceAllString(strContent, `username = "`+dstDBUser+`"`)
	strContent = cloneMatomoDBPasswordRE.ReplaceAllString(strContent, `password = "`+escapedPassword+`"`)
	if cloneMatomoTrustedHostsRE.MatchString(strContent) {
		if !strings.Contains(strContent, `trusted_hosts[] = "`+dstDomain+`"`) {
			lastMatch := cloneMatomoTrustedHostsRE.FindAllStringIndex(strContent, -1)
			insertAt := lastMatch[len(lastMatch)-1][1]
			strContent = strContent[:insertAt] + "\ntrusted_hosts[] = \"" + dstDomain + "\"" + strContent[insertAt:]
		}
	} else {
		strContent = strings.Replace(strContent, "[General]", "[General]\ntrusted_hosts[] = \""+dstDomain+"\"", 1)
	}
	if writeErr := os.WriteFile(configFile, []byte(strContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": writeErr.Error()})
		return
	}
	if !strings.Contains(strContent, dstDB) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "Failed to set 'dbname' in config.ini.php"})
		return
	}

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	matomoVersion := formOr(r, "matomo_version", "latest")
	// Rewrites hardcoded source-domain URLs left in page/content body text (the config-file rewrite above only fixes the DB connection settings, not application data) - the generic equivalent of wp-cli's search-replace, which this CMS's own CLI has no built-in version of.
	cmsclone.SearchReplaceDatabase(ctx, userContext, dstDB, "https://"+providedDomain, "https://"+dstDomainWithSubdir)

	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "Matomo", CMSType: "matomo",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: matomoVersion,
		SrcPath: srcPath, DstPath: dstPath, DstDB: dstDB,
	})
}
