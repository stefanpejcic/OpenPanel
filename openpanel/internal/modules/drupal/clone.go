package drupal

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
// two steps stay local. Two things differ from WordPress specifically:
//
//  1. Drupal's base URL is request-derived at runtime and this install
//     flow never sets $settings['trusted_host_patterns'] (confirmed via
//     grep of install.go - absent), so there's no host-allowlist to update
//     and no wp-cli-style DB search-replace step: any URL a user has
//     hardcoded into node/content body text will still point at the source
//     domain after cloning. Same known, deliberate limitation as
//     joomla/clone.go documents for the same reason.
//  2. cmsclone.ValidDocroot accepts the real "/var/www/html/..."
//     absolute-path form .Docroot actually uses everywhere else in this
//     codebase - WordPress's own validateDocroot() rejects any leading
//     "/", which would reject its own clone form's real source_folder
//     value. Not replicating that bug here.

var (
	cloneDrupalDatabaseRE = regexp.MustCompile(`'database'\s*=>\s*'.*?',`)
	cloneDrupalUsernameRE = regexp.MustCompile(`'username'\s*=>\s*'.*?',`)
	cloneDrupalPasswordRE = regexp.MustCompile(`'password'\s*=>\s*'.*?',`)
)

// handleDrupalClone mirrors wordpress/manage.go's handleCloneWordPress.
func handleDrupalClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dstDB := strings.ToLower(formOr(r, "target_db", "drupal_clone_"+generateRandomString(6)))
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Drupal files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy Drupal files: " + cpErr.Error()})
		return
	}
	cmsclone.ChownRecursive(ctx, userContext, dstPath)
	settingsFile := filepath.Join(dstPath, "sites", "default", "settings.php")
	_ = exec.CommandContext(ctx, "chmod", "644", settingsFile).Run()

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
	strContent := string(content)
	strContent = cloneDrupalDatabaseRE.ReplaceAllString(strContent, "'database' => '"+escapePHPSingleQuoted(dstDB)+"',")
	strContent = cloneDrupalUsernameRE.ReplaceAllString(strContent, "'username' => '"+escapePHPSingleQuoted(dstDBUser)+"',")
	strContent = cloneDrupalPasswordRE.ReplaceAllString(strContent, "'password' => '"+escapePHPSingleQuoted(dstDBUserPassword)+"',")
	if writeErr := os.WriteFile(settingsFile, []byte(strContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": writeErr.Error()})
		return
	}
	if !strings.Contains(strContent, dstDB) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "Failed to set 'database' in settings.php"})
		return
	}

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	drupalVersion := formOr(r, "drupal_version", "latest")
	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "Drupal", CMSType: "drupal",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: drupalVersion,
		SrcPath: srcPath, DstPath: dstPath, DstDB: dstDB,
	})
}

func escapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}
