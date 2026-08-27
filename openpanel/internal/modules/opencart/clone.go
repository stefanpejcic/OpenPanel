package opencart

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
// two steps stay local. OpenCart hardcodes its own URL and filesystem path
// in TWO files - confirmed live against a real installed OpenCart's
// config.php and admin/config.php: both only need HTTP_SERVER (+ admin's
// HTTP_CATALOG) and DIR_OPENCART rewritten - every other DIR_* constant
// derives from DIR_OPENCART via string concatenation in the file itself,
// so fixing that one constant fixes all of them. `oc_setting` has no
// url-related key stored (config_url/config_ssl are absent on a stock
// install; confirmed via a live SELECT), so no DB-side URL fix is needed
// the way PrestaShop's ps_shop_url needs one.
//
// cmsclone.ValidDocroot accepts the real "/var/www/html/..." absolute-path
// form .Docroot actually uses everywhere else in this codebase -
// WordPress's own validateDocroot() rejects any leading "/", which would
// reject its own clone form's real source_folder value. Not replicating
// that bug here.

var (
	cloneOCHTTPServerRE  = regexp.MustCompile(`define\('HTTP_SERVER',\s*'.*?'\);`)
	cloneOCHTTPCatalogRE = regexp.MustCompile(`define\('HTTP_CATALOG',\s*'.*?'\);`)
	cloneOCDirOpenCartRE = regexp.MustCompile(`define\('DIR_OPENCART',\s*'.*?'\);`)
	cloneOCDBHostnameRE  = regexp.MustCompile(`define\('DB_HOSTNAME',\s*'.*?'\);`)
	cloneOCDBUsernameRE  = regexp.MustCompile(`define\('DB_USERNAME',\s*'.*?'\);`)
	cloneOCDBPasswordRE  = regexp.MustCompile(`define\('DB_PASSWORD',\s*'.*?'\);`)
	cloneOCDBDatabaseRE  = regexp.MustCompile(`define\('DB_DATABASE',\s*'.*?'\);`)
)

// handleOpenCartClone mirrors wordpress/manage.go's handleCloneWordPress.
func handleOpenCartClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dstDB := strings.ToLower(formOr(r, "target_db", "oc_clone_"+generateRandomString(6)))
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
	dstBaseURLPath := ""
	if dstFolder != "" {
		dstBaseURLPath = dstFolder + "/"
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy OpenCart files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy OpenCart files: " + cpErr.Error()})
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

	newHTTPServer := "https://" + dstDomain + "/" + dstBaseURLPath
	newHTTPCatalog := "https://" + dstDomain + "/" + dstBaseURLPath
	newDirOpenCart := "/var/www/html/" + dstDomainWithSubdir + "/"

	rewriteConfig := func(relPath string, isAdmin bool) error {
		fp := filepath.Join(dstPath, relPath)
		content, readErr := os.ReadFile(fp)
		if readErr != nil {
			return readErr
		}
		strContent := string(content)
		if isAdmin {
			strContent = cloneOCHTTPServerRE.ReplaceAllString(strContent, "define('HTTP_SERVER', '"+escapePHPSingleQuoted(newHTTPServer+"admin/")+"');")
			strContent = cloneOCHTTPCatalogRE.ReplaceAllString(strContent, "define('HTTP_CATALOG', '"+escapePHPSingleQuoted(newHTTPCatalog)+"');")
		} else {
			strContent = cloneOCHTTPServerRE.ReplaceAllString(strContent, "define('HTTP_SERVER', '"+escapePHPSingleQuoted(newHTTPServer)+"');")
		}
		strContent = cloneOCDirOpenCartRE.ReplaceAllString(strContent, "define('DIR_OPENCART', '"+escapePHPSingleQuoted(newDirOpenCart)+"');")
		strContent = cloneOCDBHostnameRE.ReplaceAllString(strContent, "define('DB_HOSTNAME', 'mariadb');")
		strContent = cloneOCDBUsernameRE.ReplaceAllString(strContent, "define('DB_USERNAME', '"+escapePHPSingleQuoted(dstDBUser)+"');")
		strContent = cloneOCDBPasswordRE.ReplaceAllString(strContent, "define('DB_PASSWORD', '"+escapePHPSingleQuoted(dstDBUserPassword)+"');")
		strContent = cloneOCDBDatabaseRE.ReplaceAllString(strContent, "define('DB_DATABASE', '"+escapePHPSingleQuoted(dstDB)+"');")
		return os.WriteFile(fp, []byte(strContent), 0o644)
	}

	if rwErr := rewriteConfig("config.php", false); rwErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": "config.php: " + rwErr.Error()})
		return
	}
	if rwErr := rewriteConfig(filepath.Join("admin", "config.php"), true); rwErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": "admin/config.php: " + rwErr.Error()})
		return
	}

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	openCartVersion := formOr(r, "opencart_version", "latest")
	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "OpenCart", CMSType: "opencart",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: openCartVersion,
		SrcPath: srcPath, DstPath: dstPath, DstDB: dstDB,
	})
}

func escapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}
