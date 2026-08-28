package phpbb

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

// This file mirrors flarum/clone.go's shape (site-limit check, file copy,
// DB create+dump-pipe, config rewrite, sites-table insert). config.php's
// dbname/dbuser/dbpasswd keys need rewriting for the new DB; the board's
// base URL isn't stored in config.php at all (phpBB derives it from
// $_SERVER at runtime like Drupal, not hardcoded like Flarum), so only
// SearchReplaceDatabase (which fixes hardcoded URLs left in post/page
// content) handles that side, same limitation drupal/clone.go documents.

var (
	clonePhpbbDatabaseRE = regexp.MustCompile(`\$dbname\s*=\s*'.*?';`)
	clonePhpbbUsernameRE = regexp.MustCompile(`\$dbuser\s*=\s*'.*?';`)
	clonePhpbbPasswordRE = regexp.MustCompile(`\$dbpasswd\s*=\s*'.*?';`)
)

func handlePhpbbClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dstDB := strings.ToLower(formOr(r, "target_db", "phpbb_clone_"+generateRandomString(6)))
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy phpBB files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy phpBB files: " + cpErr.Error()})
		return
	}
	cmsclone.ChownRecursive(ctx, userContext, dstPath)
	configFile := filepath.Join(dstPath, "config.php")
	_ = exec.CommandContext(ctx, "chmod", "644", configFile).Run()

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
	strContent = clonePhpbbDatabaseRE.ReplaceAllString(strContent, "$$dbname = '"+escapePHPSingleQuoted(dstDB)+"';")
	strContent = clonePhpbbUsernameRE.ReplaceAllString(strContent, "$$dbuser = '"+escapePHPSingleQuoted(dstDBUser)+"';")
	strContent = clonePhpbbPasswordRE.ReplaceAllString(strContent, "$$dbpasswd = '"+escapePHPSingleQuoted(dstDBUserPassword)+"';")
	if writeErr := os.WriteFile(configFile, []byte(strContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": writeErr.Error()})
		return
	}
	if !strings.Contains(strContent, dstDB) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "Failed to set 'dbname' in config.php"})
		return
	}

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	version := formOr(r, "version", phpbbVersion)
	// Rewrites hardcoded source-domain URLs left in post/page content -
	// config.php has no base-URL field to rewrite (phpBB derives its URL
	// from $_SERVER at runtime, like Drupal), so this is the only URL
	// rewrite this clone can do.
	cmsclone.SearchReplaceDatabase(ctx, userContext, dstDB, "https://"+providedDomain, "https://"+dstDomainWithSubdir)

	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "phpBB", CMSType: "phpbb",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: version,
		SrcPath: srcPath, DstPath: dstPath, DstDB: dstDB,
	})
}

func escapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}
