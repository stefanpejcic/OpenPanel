package websites

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dashboard"
)

// RegisterWebsiteBuilderAPI wires the website builder API routes onto mux.
func RegisterWebsiteBuilderAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "website_builder", "GET /api/website-builder/{domain...}", func(w http.ResponseWriter, r *http.Request) { apiWebsiteBuilderGet(a, w, r) })
	apiregistry.Handle(mux, a, "website_builder", "PUT /api/website-builder/{domain...}", func(w http.ResponseWriter, r *http.Request) { apiWebsiteBuilderSave(a, w, r) })
	apiregistry.Handle(mux, a, "website_builder", "POST /api/website-builder", func(w http.ResponseWriter, r *http.Request) { apiWebsiteBuilderInstall(a, w, r) })
	apiregistry.Handle(mux, a, "website_builder", "DELETE /api/website-builder/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiWebsiteBuilderRemove(a, w, r) })
	apiregistry.Handle(mux, a, "website_builder", "POST /api/website-builder/sites/{site_id}/detach", func(w http.ResponseWriter, r *http.Request) { apiWebsiteBuilderDetach(a, w, r) })
}

// apiSiteDir resolves the on-disk site directory for a domain[/folder]
// path param: look up the domain's docroot, append the folder if any, and
// strip the '/var/www/html/' prefix to get the volume-relative path.
func apiSiteDir(a *appctx.App, r *http.Request, userContext, domain string) (siteDir, docroot string, ok bool, statusIfNotFound int) {
	domainRoot, _ := splitDomainAndFolder(domain)
	row := a.DB.QueryRowContext(r.Context(), "SELECT docroot FROM domains WHERE domain_url = ?", domainRoot)
	if scanErr := row.Scan(&docroot); scanErr != nil {
		return "", "", false, http.StatusNotFound
	}

	folderParam := strings.TrimPrefix(strings.TrimPrefix(domain, domainRoot), "/")
	if folderParam != "" {
		docroot = docroot + "/" + folderParam
	}

	stripped := strings.TrimPrefix(docroot, "/var/www/html/")
	siteDir = "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + stripped
	return siteDir, docroot, true, 0
}

// apiWebsiteBuilderGet returns the current HTML/CSS content for a site.
func apiWebsiteBuilderGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")
	domainRoot, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(r.Context(), userID, domainRoot) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	siteDir, docroot, ok, status := apiSiteDir(a, r, userContext, domain)
	if !ok {
		writeJSON(w, status, map[string]string{"error": "Domain not found"})
		return
	}

	html := ""
	if content, readErr := os.ReadFile(filepath.Join(siteDir, "index.html")); readErr == nil {
		html = string(content)
	}
	css := ""
	if content, readErr := os.ReadFile(filepath.Join(siteDir, "style.css")); readErr == nil {
		css = string(content)
	}

	writeJSON(w, http.StatusOK, map[string]string{"domain": domain, "docroot": docroot, "html": html, "css": css})
}

// apiWebsiteBuilderSave writes submitted HTML/CSS content to a site's
// directory.
func apiWebsiteBuilderSave(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")
	domainRoot, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	var body struct {
		HTML string `json:"html"`
		CSS  string `json:"css"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	siteDir, _, ok, status := apiSiteDir(a, r, userContext, domain)
	if !ok {
		writeJSON(w, status, map[string]string{"error": "Domain not found"})
		return
	}

	if mkErr := os.MkdirAll(siteDir, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
		return
	}
	if wErr := os.WriteFile(filepath.Join(siteDir, "index.html"), []byte(body.HTML), 0o644); wErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": wErr.Error()})
		return
	}
	if wErr := os.WriteFile(filepath.Join(siteDir, "style.css"), []byte(body.CSS), 0o644); wErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": wErr.Error()})
		return
	}

	if uid, uidErr := a.GetUID(ctx, userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = exec.CommandContext(ctx, "chown", "-R", uidStr+":"+uidStr, siteDir).Run()
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "saved website builder content for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Content saved successfully", "domain": domain})
}

// apiWebsiteBuilderInstall creates a new website-builder site: validates
// plan limits and the target path, then writes starter HTML/CSS files and
// inserts the site's database row.
func apiWebsiteBuilderInstall(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	injectedData, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	planID, _ := injectedData["hosting_plan"].(int)
	plan, _ := a.QueryPlanDetailsByID(ctx, planID)
	websitesLimit := atoiDefaultWB(plan.WebsitesLimit, 0)
	userWebsites, _ := dashboard.GetUserWebsites(a, ctx, userID)
	if websitesLimit != 0 && len(userWebsites) >= websitesLimit {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "You have reached the maximum number of sites allowed" + plan.UpgradeMessage()})
		return
	}

	var body struct {
		DomainID     any    `json:"domain_id"`
		Subdirectory string `json:"subdirectory"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	subdirectory := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(body.Subdirectory)), " ", "")

	var domainIDStr string
	switch v := body.DomainID.(type) {
	case string:
		domainIDStr = v
	case float64:
		domainIDStr = strconv.Itoa(int(v))
	}
	if domainIDStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain_id is required"})
		return
	}

	var selectedDomain, docroot sql.NullString
	row := a.DB.QueryRowContext(ctx, "SELECT domain_url, docroot FROM domains WHERE domain_id = ?", domainIDStr)
	if scanErr := row.Scan(&selectedDomain, &docroot); scanErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Domain not found"})
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, selectedDomain.String) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	installPath := docroot.String
	siteName := selectedDomain.String
	if subdirectory != "" {
		installPath = docroot.String + "/" + subdirectory
		siteName = selectedDomain.String + "/" + subdirectory
	}

	stripped := strings.TrimPrefix(installPath, "/var/www/html/")
	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + stripped

	for _, fname := range []string{"index.html", "index.php", ".htaccess"} {
		if _, statErr := os.Stat(path.Join(volume, fname)); statErr == nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": fname + " already exists — website creation cannot proceed"})
			return
		}
	}

	if mkErr := os.MkdirAll(volume, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
		return
	}

	const htmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
</head>
<body>
<body></body>
<style>* { box-sizing: border-box; } body {margin: 0;}</style>
</body>
</html>`
	if wErr := os.WriteFile(path.Join(volume, "index.html"), []byte(htmlContent), 0o644); wErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": wErr.Error()})
		return
	}
	if wErr := os.WriteFile(path.Join(volume, "style.css"), []byte("* { box-sizing: border-box; } body {margin: 0;}"), 0o644); wErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": wErr.Error()})
		return
	}

	_ = exec.CommandContext(ctx, "chmod", "-R", "755", volume).Run()
	if uid, uidErr := a.GetUID(ctx, userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = exec.CommandContext(ctx, "chown", "-R", uidStr+":"+uidStr, volume).Run()
	}

	if _, insertErr := a.DB.ExecContext(ctx, "INSERT INTO sites (site_name, domain_id, type) VALUES (?, ?, ?)", siteName, domainIDStr, "websitebuilder"); insertErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": insertErr.Error()})
		return
	}

	_ = a.Cache.Delete(ctx, cacheKeyUserWebsites(userID))
	_ = logger.RecordUserAction(a.Config, currentUsername, "installed Website Builder on "+siteName, reqip.ClientIP(r))
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Website created successfully on " + siteName, "site_name": siteName})
}

// apiGetSite looks up a site by ID and reports ownership separately from
// existence, so callers can distinguish 403 from 404.
func apiGetSite(ctx context.Context, a *appctx.App, userID int, siteID string) (siteName, docroot string, forbidden, found bool) {
	var name, root sql.NullString
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_id = sites.domain_id
		WHERE sites.id = ?`, siteID)
	if scanErr := row.Scan(&name, &root); scanErr != nil {
		return "", "", false, false
	}

	domainRoot, _ := splitDomainAndFolder(name.String)
	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		return "", "", true, true
	}
	return name.String, root.String, false, true
}

// apiWebsiteBuilderRemove deletes a site's generated files and its
// database row.
func apiWebsiteBuilderRemove(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	siteID := r.PathValue("site_id")

	siteName, docroot, forbidden, found := apiGetSite(ctx, a, userID, siteID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}
	if forbidden {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}
	if docroot == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found in database"})
		return
	}

	realPath := strings.TrimPrefix(docroot, "/var/www/html/")
	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + realPath
	_ = os.Remove(filepath.Join(volume, "index.html"))
	_ = os.Remove(filepath.Join(volume, "style.css"))

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", siteID); delErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": delErr.Error()})
		return
	}
	_ = a.Cache.Delete(ctx, cacheKeyUserWebsites(userID))
	_ = logger.RecordUserAction(a.Config, currentUsername, "removed Website Builder for "+siteName, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Website deleted successfully"})
}

// apiWebsiteBuilderDetach removes a site's database row without touching
// its files on disk.
func apiWebsiteBuilderDetach(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	siteID := r.PathValue("site_id")

	siteName, _, forbidden, found := apiGetSite(ctx, a, userID, siteID)
	if !found {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}
	if forbidden {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	if _, delErr := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", siteID); delErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": delErr.Error()})
		return
	}
	_ = a.Cache.Delete(ctx, cacheKeyUserWebsites(userID))
	_ = logger.RecordUserAction(a.Config, currentUsername, "detached Website Builder for "+siteName, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Website detached successfully"})
}
