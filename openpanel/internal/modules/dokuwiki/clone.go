package dokuwiki

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

// cloneDokuwikiTitleRE matches the $conf['title'] line writeDokuwikiConfig
// writes - the only place a source-domain-derived value ends up in
// conf/local.php by default (DokuWiki itself is otherwise request-derived
// at runtime, like Drupal, not URL-in-config like Flarum).
var cloneDokuwikiTitleRE = regexp.MustCompile(`\$conf\['title'\]\s*=\s*'.*?';`)

// handleDokuwikiClone copies a DokuWiki install's files to a new
// domain/subdirectory. There's no database to dump/restore. Unlike
// SofaWiki's clone (which skips config rewriting entirely since a fresh
// SofaWiki install is deliberately left unconfigured), DokuWiki IS
// configured at install time here, so conf/local.php's title is rewritten
// to the destination domain. Any hardcoded links left inside
// data/pages/*.txt page content are NOT rewritten - the same known,
// deliberate limitation drupal/clone.go and sofawiki/clone.go document for
// hardcoded URLs in content, for the same reason (no database to run a
// generic search-replace against).
func handleDokuwikiClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy DokuWiki files: " + mkErr.Error()})
		return
	}
	if cpErr := exec.CommandContext(ctx, "cp", "-a", srcPath+"/.", dstPath+"/").Run(); cpErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to copy DokuWiki files: " + cpErr.Error()})
		return
	}
	cmsclone.ChownRecursive(ctx, userContext, dstPath)

	localConfigFile := filepath.Join(dstPath, "conf", "local.php")
	if content, readErr := os.ReadFile(localConfigFile); readErr == nil {
		// "$$" in the replacement is a literal "$" - ReplaceAllString
		// otherwise treats a bare "$conf" in the replacement as a
		// (nonexistent) submatch reference and silently drops it,
		// corrupting local.php's PHP syntax. Confirmed live: without the
		// escape, the written line came out as "['title'] = '...';"
		// (missing "$conf"), a fatal parse error that blanked the whole
		// cloned site.
		newContent := cloneDokuwikiTitleRE.ReplaceAllString(string(content), "$$conf['title'] = '"+escapePHPSingleQuoted(dstDomainWithSubdir)+"';")
		_ = os.WriteFile(localConfigFile, []byte(newContent), 0o644)
	}

	version := "unknown"
	if versionBytes, readErr := os.ReadFile(filepath.Join(dstPath, "VERSION")); readErr == nil {
		version = strings.TrimSpace(string(versionBytes))
	}

	adminEmail := formOr(r, "admin_email", "admin@"+dstDomain)
	cmsclone.FinalizeSite(ctx, w, r, cmsclone.FinalizeParams{
		App: a, WriteJSON: writeJSON, UserID: userID, Username: currentUsername,
		CMSDisplayName: "DokuWiki", CMSType: "dokuwiki",
		ProvidedDomain: providedDomain, DstDomainWithSubdir: dstDomainWithSubdir, DomainID: domainID,
		AdminEmail: adminEmail, Version: version,
		SrcPath: srcPath, DstPath: dstPath,
	})
}
