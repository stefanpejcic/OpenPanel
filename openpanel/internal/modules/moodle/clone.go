package moodle

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

// This file mirrors drupal/clone.go's overall shape (site-limit check, DB
// create+dump-pipe, config rewrite, sites-table insert, cron
// registration), sharing the site-limit check, DB create+dump-pipe, and
// sites-table-insert/log/response tail with every other CMS's clone.go via
// internal/core/cmsclone - see that package's doc comment for why those
// steps are shared but the file-copy and config-rewrite steps aren't. The
// file-copy step here is NOT a plain "copy the docroot" like every other
// module's clone - see install.go's package doc comment and its own
// approotHostPath/datarootHostPath/symlink construction (read in full
// before touching this file). Moodle's docroot is a symlink, not a real
// directory: the actual code lives in a sibling "<slug>_moodleapp/"
// directory and user data lives in a separate sibling "<slug>_moodledata/"
// directory, neither of which install.go's own dispatch.go/websites.go
// plumbing exposes as "source_folder" the way it does for every
// flat-docroot module - so this clone resolves both sibling directories
// itself from the source domain name, the same siteSlug() call install.go
// used to create them. This is also why cmsclone.ValidDocroot is never
// called here: there's no source_folder/docroot form field to validate the
// way the other seven modules do.

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

	websiteCount, _ := countUserWebsites(a, userID)
	if !cmsclone.WithinSiteLimit(ctx, a, userID, websiteCount) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "You have reached the maximum number of sites allowed" + a.UpgradeMessageForUser(ctx, userID)})
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
	cmsclone.ChownRecursive(ctx, userContext, dstApprootHostPath, dstDatarootHostPath)
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

	escapedPassword, dbErr := cmsclone.CreateDatabaseAndDump(ctx, userContext, mysqlVersion, dumpCmd, srcDB, dstDB, dstDBUser, dstDBUserPassword)
	if dbErr != nil {
		if cmsclone.DumpStageFailed(dbErr) {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "command_failed"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": dbErr.Error()})
		}
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

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	moodleVersion := formOr(r, "moodle_version", "latest")
	// Rewrites hardcoded source-domain URLs left in page/content body text (the config-file rewrite above only fixes the DB connection settings, not application data) - the generic equivalent of wp-cli's search-replace, which this CMS's own CLI has no built-in version of.
	cmsclone.SearchReplaceDatabase(ctx, userContext, dstDB, "https://"+providedDomain, "https://"+dstDomainWithSubdir)

	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "Moodle", CMSType: "moodle",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: moodleVersion,
		SrcPath: srcApprootHostPath, DstPath: dstApprootHostPath, DstDB: dstDB,
	})
}
