package prestashop

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cmsclone"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
)

// This file mirrors wordpress/manage.go's handleCloneWordPress in overall
// shape (site-limit check, file copy, DB create+dump-pipe, config
// rewrite, sites-table insert), sharing everything but the docroot copy
// and config-rewrite steps with every other CMS's clone.go via
// internal/core/cmsclone - see that package's doc comment for why those
// two steps stay local. Unlike Joomla/Drupal, PrestaShop DOES hardcode its
// domain in the DB - the `{prefix}shop_url` table's domain/domain_ssl/
// physical_uri columns (confirmed live against a real installed
// PrestaShop: `id_shop_url, id_shop, domain, domain_ssl, physical_uri,
// virtual_uri, main, active`) - so after the DB import this updates that
// row to point the clone at its own new domain/subdirectory, the
// PrestaShop equivalent of wp-cli's search-replace step.
//
// cmsclone.ValidDocroot accepts the real "/var/www/html/..." absolute-path
// form .Docroot actually uses everywhere else in this codebase -
// WordPress's own validateDocroot() rejects any leading "/", which would
// reject its own clone form's real source_folder value. Not replicating
// that bug here.

var (
	clonePrestaDBNameRE   = regexp.MustCompile(`'database_name'\s*=>\s*'.*?',`)
	clonePrestaDBUserRE   = regexp.MustCompile(`'database_user'\s*=>\s*'.*?',`)
	clonePrestaDBPasswdRE = regexp.MustCompile(`'database_password'\s*=>\s*'.*?',`)
)

// handlePrestashopClone mirrors wordpress/manage.go's handleCloneWordPress.
func handlePrestashopClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	dstDB := strings.ToLower(formOr(r, "target_db", "presta_clone_"+generateRandomString(6)))
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
	dstSubdirURI := "/"
	if dstFolder != "" {
		dstSubdirURI = "/" + dstFolder + "/"
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy PrestaShop files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy PrestaShop files: " + cpErr.Error()})
		return
	}
	cmsclone.ChownRecursive(ctx, userContext, dstPath)
	// PrestaShop's Symfony var/cache/{prod,dev} holds a compiled service
	// container with the SOURCE site's DB credentials baked in - a plain
	// file copy carries that stale cache along, which makes the clone 500
	// on its very first request even though parameters.php itself is
	// correct (confirmed live: clearing this directory was the fix).
	// PrestaShop regenerates it automatically on the next request.
	_ = os.RemoveAll(filepath.Join(dstPath, "var", "cache", "prod"))
	_ = os.RemoveAll(filepath.Join(dstPath, "var", "cache", "dev"))

	_, dbErr := cmsclone.CreateDatabaseAndDump(ctx, userContext, mysqlVersion, dumpCmd, srcDB, dstDB, dstDBUser, dstDBUserPassword)
	if dbErr != nil {
		if cmsclone.DumpStageFailed(dbErr) {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "command_failed"})
		} else {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": dbErr.Error()})
		}
		return
	}

	parametersFile := filepath.Join(dstPath, "app", "config", "parameters.php")
	content, readErr := os.ReadFile(parametersFile)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": readErr.Error()})
		return
	}
	strContent := string(content)
	strContent = clonePrestaDBNameRE.ReplaceAllString(strContent, "'database_name' => '"+escapePHPSingleQuoted(dstDB)+"',")
	strContent = clonePrestaDBUserRE.ReplaceAllString(strContent, "'database_user' => '"+escapePHPSingleQuoted(dstDBUser)+"',")
	strContent = clonePrestaDBPasswdRE.ReplaceAllString(strContent, "'database_password' => '"+escapePHPSingleQuoted(dstDBUserPassword)+"',")
	if writeErr := os.WriteFile(parametersFile, []byte(strContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "details": writeErr.Error()})
		return
	}
	if !strings.Contains(strContent, dstDB) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"status": "error", "step": "Failed to set database_name in parameters.php"})
		return
	}

	// Point the clone's own ps_shop_url row at its new domain/subdirectory -
	// the same fix install.go applies at install time via --domain/--base_uri,
	// see that file's comment on why domain and physical_uri must stay split.
	escapedDstDomain := escapeMySQLString(dstDomain)
	escapedURI := escapeMySQLString(dstSubdirURI)
	_, _ = mysqlmanager.Exec(ctx, userContext,
		"UPDATE `ps_shop_url` SET domain = '"+escapedDstDomain+"', domain_ssl = '"+escapedDstDomain+"', physical_uri = '"+escapedURI+"'",
		dstDB)

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	prestashopVersion := formOr(r, "prestashop_version", "latest")
	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "PrestaShop", CMSType: "prestashop",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: prestashopVersion,
		SrcPath: srcPath, DstPath: dstPath, DstDB: dstDB,
	})
}

func escapePHPSingleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}

func escapeMySQLString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return value
}
