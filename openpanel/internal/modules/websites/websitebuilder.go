// Package websites (this file) implements the GrapesJS drag-and-drop HTML
// site builder's install/edit/detach/remove routes. Kept in this package
// (rather than its own) so it can reuse
// getContainerFromDatabase/ContainerInfo/splitDomainAndFolder without
// exporting them.
package websites

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// getAndValidateSite looks up a site by ID and confirms it belongs to
// userID before returning its domain and install path.
func getAndValidateSite(ctx context.Context, a *appctx.App, userID int, siteID string) (selectedDomain, installPath string, ok bool) {
	var domocroot sql.NullString
	var domain sql.NullString
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_id = sites.domain_id
		WHERE sites.id = ?`, siteID)
	if err := row.Scan(&domain, &domocroot); err != nil {
		return "", "", false
	}
	selectedDomain = domain.String
	installPath = domocroot.String

	apexDomain, _ := splitDomainAndFolder(selectedDomain)
	if !a.CheckDomainBelongsToUser(ctx, userID, apexDomain) {
		return "", "", false
	}
	return selectedDomain, installPath, true
}

func deleteSiteFromDB(ctx context.Context, a *appctx.App, userID int, siteID string) error {
	_, err := a.DB.ExecContext(ctx, "DELETE FROM sites WHERE id = ?", siteID)
	if err != nil {
		return err
	}
	_ = a.Cache.Delete(ctx, cacheKeyUserWebsites(userID))
	return nil
}

// handleWebsiteBuilderRemove deletes the generated site files and the
// site's database row.
func handleWebsiteBuilderRemove(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	siteID := r.Form.Get("id")

	selectedDomain, installPath, ok := getAndValidateSite(ctx, a, userID, siteID)
	if !ok {
		flashAndRedirect(a, w, r, "error", "No data found for the provided site ID", "/sites")
		return
	}
	if installPath == "" {
		flashAndRedirect(a, w, r, "error", "Website not found in the database", "/sites")
		return
	}

	realInstallPath := strings.Replace(installPath, "/var/www/html/", "", 1)
	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + realInstallPath
	_ = os.Remove(filepath.Join(volume, "index.html"))
	_ = os.Remove(filepath.Join(volume, "style.css"))

	if delErr := deleteSiteFromDB(ctx, a, userID, siteID); delErr != nil {
		message := "An error occurred during website deletion."
		flashSess(a, w, r, "error", message)
		_, _ = w.Write([]byte(message))
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "removed Website Builder for "+selectedDomain, reqip.ClientIP(r))
	message := "Website deleted successfully!"
	flashSess(a, w, r, "success", message)
	_, _ = w.Write([]byte(message))
}

// handleWebsiteBuilderDetach removes the site's database row without
// touching its files on disk.
func handleWebsiteBuilderDetach(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	siteID := r.Form.Get("id")

	selectedDomain, _, ok := getAndValidateSite(ctx, a, userID, siteID)
	if !ok {
		flashAndRedirect(a, w, r, "error", "No data found for the provided site ID", "/sites")
		return
	}

	if delErr := deleteSiteFromDB(ctx, a, userID, siteID); delErr != nil {
		message := "An error occurred during website detachment."
		flashSess(a, w, r, "error", message)
		_, _ = w.Write([]byte(message))
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "detached Website Builder for "+selectedDomain, reqip.ClientIP(r))
	message := "Website detached successfully!"
	flashSess(a, w, r, "success", message)
	_, _ = w.Write([]byte(message))
}

// handleWebsiteBuilderEdit serves the GrapesJS editor on GET and saves the
// submitted HTML/CSS to disk on POST.
func handleWebsiteBuilderEdit(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domainParam := r.URL.Query().Get("domain")
	if domainParam == "" {
		http.Redirect(w, r, "/sites", http.StatusFound)
		return
	}

	websiteParam := domainParam
	domain, folderParam := splitDomainAndFolder(domainParam)

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	lookupKey := domainParam
	if folderParam != "" {
		lookupKey = websiteParam
	}
	container, found := getContainerFromDatabase(a, r, lookupKey)
	if !found {
		http.NotFound(w, r)
		return
	}

	var docroot string
	row := a.DB.QueryRowContext(ctx, "SELECT docroot FROM domains WHERE domain_url = ?", domain)
	if scanErr := row.Scan(&docroot); scanErr != nil {
		flashAndRedirect(a, w, r, "error", "Unable to detect docroot for the domain.", "/sites")
		return
	}
	if folderParam != "" {
		docroot = docroot + "/" + folderParam
	}

	cmsType := strings.ToLower(container.Type)
	if cmsType != "websitebuilder" && cmsType != "sitebuilder" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid action"})
		return
	}

	strippedPath := strings.TrimPrefix(docroot, "/var/www/html/")
	htmlSiteDocroot := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strippedPath
	htmlFile := filepath.Join(htmlSiteDocroot, "index.html")
	cssFile := filepath.Join(htmlSiteDocroot, "style.css")

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		htmlContent := r.Form.Get("html")
		cssContent := r.Form.Get("css")

		if mkErr := os.MkdirAll(htmlSiteDocroot, 0o755); mkErr != nil {
			http.Error(w, mkErr.Error(), http.StatusInternalServerError)
			return
		}
		if writeErr := os.WriteFile(htmlFile, []byte(htmlContent), 0o644); writeErr != nil {
			http.Error(w, writeErr.Error(), http.StatusInternalServerError)
			return
		}
		if writeErr := os.WriteFile(cssFile, []byte(cssContent), 0o644); writeErr != nil {
			http.Error(w, writeErr.Error(), http.StatusInternalServerError)
			return
		}

		if uid, uidErr := a.GetUID(ctx, userContext); uidErr == nil {
			uidStr := strconv.Itoa(uid)
			chownArgv := []string{"chown", "-R", uidStr + ":" + uidStr, htmlSiteDocroot}
			_ = exec.CommandContext(ctx, chownArgv[0], chownArgv[1:]...).Run()
		}

		_, _ = w.Write([]byte("Saved successfully"))
		return
	}

	htmlContent := ""
	if data, readErr := os.ReadFile(htmlFile); readErr == nil {
		htmlContent = string(data)
	}
	cssContent := ""
	if data, readErr := os.ReadFile(cssFile); readErr == nil {
		cssContent = string(data)
	}

	renderGrapesJSEditor(a, w, r, websiteParam, htmlContent, cssContent)
}

// RegisterWebsiteBuilder wires the website builder's routes onto mux.
func RegisterWebsiteBuilder(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "website_builder")(h)
	}

	mux.Handle("POST /website-builder/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWebsiteBuilderRemove(a, w, r) }))
	mux.Handle("POST /website-builder/detach", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWebsiteBuilderDetach(a, w, r) }))
	mux.Handle("/website-builder/edit", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWebsiteBuilderEdit(a, w, r) }))
	mux.Handle("/website-builder/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWebsiteBuilderInstall(a, w, r) }))
}
