package sofawiki

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cmsclone"
)

// handleSofawikiClone copies a SofaWiki install's files to a new
// domain/subdirectory. There's no database to dump/restore, and (since a
// fresh install is left unconfigured until the site owner completes
// SofaWiki's own setup wizard - see sofawiki.go's package doc comment) no
// config file with a hardcoded base URL to rewrite either for a
// not-yet-configured source. If the source HAS since been configured
// through that wizard, whatever site title/URL it wrote to
// inc/configuration.php is copied as-is and not rewritten for the new
// domain - the same known, deliberate limitation drupal/clone.go
// documents for hardcoded URLs in content, for the same reason (no
// generic way to know what to replace it with).
func handleSofawikiClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
	srcFolder := r.FormValue("source_folder")
	dstFolder := r.FormValue("subdirectory")

	if providedDomain == "" || dstDomain == "" || srcFolder == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing required form fields"})
		return
	}

	domainID, docroot, _, dstDomainWithSubdir, ok := cmsclone.ResolveDestination(ctx, a, dstDomain, dstFolder)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Destination domain not found in database"})
		return
	}

	srcDomain := strings.Split(providedDomain, "/")[0]

	if !cmsclone.ValidDomain(srcDomain) || !cmsclone.ValidDomain(dstDomain) || !cmsclone.ValidDocroot(srcFolder) || !cmsclone.ValidDocroot(docroot) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid input or unsafe docroot"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, srcDomain) || !a.CheckDomainBelongsToUser(ctx, userID, dstDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy SofaWiki files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy SofaWiki files: " + cpErr.Error()})
		return
	}
	cmsclone.ChownRecursive(ctx, userContext, dstPath)

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "SofaWiki", CMSType: "sofawiki",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: "master",
		SrcPath: srcPath, DstPath: dstPath,
	})
}
