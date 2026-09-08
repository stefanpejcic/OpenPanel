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

// mirrors wordpress/manage.go's handleCloneWordPress in shape (file copy, DB create+dump, config rewrite, sites insert), sharing everything but docroot copy and config rewrite with every other CMS via internal/core/cmsclone
// config.php's datadirectory, trusted_domains (index 1), overwrite.cli.url and db creds are fixed via regex text-replace rather than occ, since occ needs a matching PHP runtime and this hit PHP-version failures elsewhere (see maintenance.go)
// instanceid is left untouched so the copied data/appdata_<instanceid>/ dir keeps matching it
// cmsclone.ValidDocroot accepts the real absolute "/var/www/html/..." form used everywhere here, unlike wordpress's own validateDocroot which would reject it

var (
	cloneNCDataDirRE   = regexp.MustCompile(`'datadirectory'\s*=>\s*'.*?',`)
	cloneNCOverwriteRE = regexp.MustCompile(`'overwrite\.cli\.url'\s*=>\s*'.*?',`)
	cloneNCDBNameRE    = regexp.MustCompile(`'dbname'\s*=>\s*'.*?',`)
	cloneNCDBUserRE    = regexp.MustCompile(`'dbuser'\s*=>\s*'.*?',`)
	cloneNCDBPasswdRE  = regexp.MustCompile(`'dbpassword'\s*=>\s*'.*?',`)
	// matches trusted_domains index 1, right after the always-present "0 => 'localhost'," line - see install.go for why index 1 holds the site's domain
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
	// rewrites hardcoded source-domain URLs left in content body text, the generic equivalent of wp-cli's search-replace which Nextcloud's CLI lacks
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
