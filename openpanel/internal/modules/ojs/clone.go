package ojs

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cmsclone"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
)

// This file mirrors moodle/clone.go's overall shape (site-limit check, DB
// create+dump-pipe via internal/core/cmsclone, sites-table insert, cron
// registration) - see that package's doc comment for why those steps are
// shared but the file-copy and config-rewrite steps aren't. Like Moodle,
// OJS's docroot is a symlink, not a real directory (see ojs.go's package
// doc comment), so this resolves the approot/files sibling directories
// itself from the source domain name, the same siteSlug() call install.go
// used to create them, and config.inc.php's INI syntax needs line-based
// regex replacement instead of Moodle's PHP $CFG-> assignment syntax (see
// config.go).

// handleOJSClone mirrors moodle/clone.go's handleMoodleClone, adapted for
// OJS's approot/files/symlink layout (see install.go) and config.inc.php's
// INI format.
func handleOJSClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
	dstFolder := r.FormValue("subdirectory")

	dstDB := strings.ToLower(formOr(r, "target_db", "ojs_clone_"+generateRandomString(6)))
	dstDBUser := strings.ToLower(formOr(r, "target_db_user", dstDB))
	dstDBUserPassword := formOr(r, "target_db_user_password", generateRandomString(16))

	if providedDomain == "" || dstDomain == "" || srcDB == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing required form fields"})
		return
	}

	domainID, docroot, phpVersion, dstDomainWithSubdir, ok := cmsclone.ResolveDestination(ctx, a, dstDomain, dstFolder)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Destination domain not found in database"})
		return
	}

	srcDomain := strings.Split(providedDomain, "/")[0]

	if !cmsclone.ValidDomain(srcDomain) || !cmsclone.ValidDomain(dstDomain) || !cmsclone.ValidDB(srcDB) || !cmsclone.ValidDB(dstDB) || !cmsclone.ValidDB(dstDBUser) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid input"})
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

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	srcSlug := siteSlug(providedDomain)
	dstSlug := siteSlug(dstDomainWithSubdir)

	srcApprootHostPath := filepath.Join(htmlVolume, srcSlug+"_ojsapp")
	srcFilesHostPath := filepath.Join(htmlVolume, srcSlug+"_ojsfiles")
	dstApprootHostPath := filepath.Join(htmlVolume, dstSlug+"_ojsapp")
	dstApprootContainerPath := "/var/www/html/" + dstSlug + "_ojsapp"
	dstFilesHostPath := filepath.Join(htmlVolume, dstSlug+"_ojsfiles")
	dstFilesContainerPath := "/var/www/html/" + dstSlug + "_ojsfiles"

	if info, statErr := os.Stat(srcApprootHostPath); statErr != nil || !info.IsDir() {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Source OJS app directory not found: " + srcApprootHostPath})
		return
	}

	if mkErr := os.MkdirAll(dstApprootHostPath, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy OJS files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcApprootHostPath+"/.", dstApprootHostPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy OJS app files: " + cpErr.Error()})
		return
	}
	if mkErr := os.MkdirAll(dstFilesHostPath, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create OJS files directory: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcFilesHostPath+"/.", dstFilesHostPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy OJS files directory: " + cpErr.Error()})
		return
	}
	cmsclone.ChownRecursive(ctx, userContext, dstApprootHostPath, dstFilesHostPath)
	_ = exec.Command("chmod", "-R", "777", dstFilesHostPath).Run()

	// Docroot destination host path, and the symlink target must be the
	// container-visible path (/var/www/html/...), not the host filesystem
	// path - same gotcha install.go's own symlink call documents.
	const wwwBaseDirectory = "/var/www/html/"
	dstHostOSPath := strings.Replace(filepath.Clean(docroot), wwwBaseDirectory, htmlVolume, 1)
	if _, statErr := os.Lstat(dstHostOSPath); statErr == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Destination path " + docroot + " already exists."})
		_ = os.RemoveAll(dstApprootHostPath)
		_ = os.RemoveAll(dstFilesHostPath)
		return
	}
	if symErr := os.Symlink(dstApprootContainerPath, dstHostOSPath); symErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error creating web root symlink: " + symErr.Error()})
		_ = os.RemoveAll(dstApprootHostPath)
		_ = os.RemoveAll(dstFilesHostPath)
		return
	}

	escapedPassword, dbErr := cmsclone.CreateDatabaseAndDump(ctx, userContext, mysqlVersion, dumpCmd, srcDB, dstDB, dstDBUser, dstDBUserPassword)
	if dbErr != nil {
		if cmsclone.DumpStageFailed(dbErr) {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "command_failed"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": dbErr.Error()})
		}
		return
	}

	configFile := filepath.Join(dstApprootHostPath, "config.inc.php")
	content, readErr := os.ReadFile(configFile)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": readErr.Error()})
		return
	}
	strContent := string(content)
	strContent = iniDatabaseNameRE.ReplaceAllString(strContent, iniBare("name", dstDB))
	strContent = iniDatabaseUsernameRE.ReplaceAllString(strContent, iniBare("username", dstDBUser))
	strContent = iniDatabasePasswordRE.ReplaceAllString(strContent, iniBare("password", escapedPassword))
	strContent = iniBaseURLRE.ReplaceAllString(strContent, iniQuoted("base_url", "https://"+dstDomainWithSubdir))
	strContent = iniFilesDirRE.ReplaceAllString(strContent, iniBare("files_dir", dstFilesContainerPath))
	if writeErr := os.WriteFile(configFile, []byte(strContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": writeErr.Error()})
		return
	}
	if !strings.Contains(strContent, dstDB) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "Failed to set 'name' in config.inc.php"})
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpContainer := webServer
	if !isLitespeed {
		phpContainer = "php-fpm-" + phpVersion
	}
	cronComment := ojsCronComment(dstDomainWithSubdir)
	cronCommand := "php " + dstApprootContainerPath + "/lib/pkp/tools/scheduler.php run"
	_ = crons.AddJob(ctx, userContext, cronComment, "0 * * * * *", phpContainer, cronCommand, true)

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	ojsVersion := formOr(r, "ojs_version", "latest")
	// Rewrites hardcoded source-domain URLs left in the database (the
	// config-file rewrite above only fixes the DB connection settings and
	// base_url, not application data) - the generic equivalent of wp-cli's
	// search-replace, which OJS has no built-in version of.
	cmsclone.SearchReplaceDatabase(ctx, userContext, dstDB, "https://"+providedDomain, "https://"+dstDomainWithSubdir)

	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "OJS", CMSType: "ojs",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: ojsVersion,
		SrcPath: srcApprootHostPath, DstPath: dstApprootHostPath, DstDB: dstDB,
	})
}
