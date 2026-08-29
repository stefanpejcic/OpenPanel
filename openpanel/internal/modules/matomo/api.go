package matomo

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

func writeAPIJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// handleMatomoVersions backs matomo_install.html's version dropdown -
// GitHub's releases API is CORS-open and could be hit client-side, but
// filtering out versions with no downloadable asset (see version.go) needs
// server-side logic, so this exposes the already-filtered list as JSON
// instead, matching nextcloud/prestashop's identical approach.
func handleMatomoVersions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	versions, err := listMatomoVersions(r.Context())
	if err != nil {
		writeAPIJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeAPIJSON(w, http.StatusOK, map[string]any{"versions": versions})
}

// apiInstallMatomo delegates straight to handleInstallPage (which itself
// calls handleInstallStream on POST): same site-limit check, same NDJSON
// progress stream written directly to the response - just fed from the
// API's JSON body instead of a UI form post.
func apiInstallMatomo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID      string `json:"domain_id"`
		Subdirectory  string `json:"subdirectory"`
		MatomoVersion string `json:"matomo_version"`
		AdminLogin    string `json:"admin_login"`
		AdminPassword string `json:"admin_password"`
		AdminEmail    string `json:"admin_email"`
		DBName        string `json:"db_name"`
		DBUser        string `json:"db_user"`
		DBPassword    string `json:"db_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if body.DomainID == "" {
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "domain_id is required"})
		return
	}

	form := url.Values{
		"domain_id": {body.DomainID}, "subdirectory": {body.Subdirectory},
		"matomo_version": {body.MatomoVersion},
		"admin_login":    {body.AdminLogin}, "admin_password": {body.AdminPassword},
		"admin_email": {body.AdminEmail},
		"db_name":     {body.DBName}, "db_user": {body.DBUser}, "db_password": {body.DBPassword},
	}
	handleInstallPage(a, w, withMatomoForm(r, form))
}

// apiRemoveMatomo delegates to handleRemoveMatomo with the path's
// {site_id} translated into the "id" form field it expects, and
// output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemoveMatomo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withMatomoForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveMatomo(a, w, cloned)
}

// apiResolveMatomoSite resolves {site_id} into the (domain, docroot) pair
// every handler in this file needs - mirrors drupal/api.go's
// apiResolveDrupalSite.
func apiResolveMatomoSite(ctx context.Context, a *appctx.App, siteID string) (domain, docroot string, ok bool) {
	var siteName string
	var rootDocroot sql.NullString
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'matomo'`, siteID)
	if err := row.Scan(&siteName, &rootDocroot); err != nil || !rootDocroot.Valid || rootDocroot.String == "" {
		return "", "", false
	}
	docroot = rootDocroot.String
	if idx := strings.Index(siteName, "/"); idx != -1 {
		docroot = strings.TrimSuffix(rootDocroot.String, "/") + "/" + siteName[idx+1:]
	}
	return siteName, docroot, true
}

// apiMatomoClone delegates to handleMatomoClone, resolving {site_id} into
// the source_domain/source_folder fields it expects and taking every other
// clone field from the JSON body - mirrors apiDrupalClone.
func apiMatomoClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	sourceDomain, sourceFolder, ok := apiResolveMatomoSite(r.Context(), a, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}

	var body struct {
		TargetDomain         string `json:"target_domain"`
		Subdirectory         string `json:"subdirectory"`
		SourceDB             string `json:"source_db"`
		TargetDB             string `json:"target_db"`
		TargetDBUser         string `json:"target_db_user"`
		TargetDBUserPassword string `json:"target_db_user_password"`
		AdminEmail           string `json:"admin_email"`
		MatomoVersion        string `json:"matomo_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if body.TargetDomain == "" || body.SourceDB == "" {
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "target_domain and source_db are required"})
		return
	}

	form := url.Values{
		"source_domain": {sourceDomain}, "target_domain": {body.TargetDomain},
		"source_db": {body.SourceDB}, "source_folder": {sourceFolder}, "subdirectory": {body.Subdirectory},
		"target_db": {body.TargetDB}, "target_db_user": {body.TargetDBUser}, "target_db_user_password": {body.TargetDBUserPassword},
		"admin_email": {body.AdminEmail}, "matomo_version": {body.MatomoVersion},
	}
	handleMatomoClone(a, w, withMatomoForm(r, form))
}

// apiMatomoUpdate resolves {site_id} into the domain/docroot query params
// handleMatomoUpdate reads directly, then delegates to it as-is.
func apiMatomoUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	domain, docroot, ok := apiResolveMatomoSite(r.Context(), a, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}
	q := r.URL.Query()
	q.Set("domain", domain)
	q.Set("docroot", docroot)
	r.URL.RawQuery = q.Encode()
	handleMatomoUpdate(a, w, r)
}

// apiMatomoCache resolves {site_id} into the domain/docroot query params
// handleMatomoCacheClean reads (via matomoRequestParams), then delegates to
// it as-is.
func apiMatomoCache(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	domain, docroot, ok := apiResolveMatomoSite(r.Context(), a, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}
	q := r.URL.Query()
	q.Set("domain", domain)
	q.Set("docroot", docroot)
	r.URL.RawQuery = q.Encode()
	handleMatomoCacheClean(a, w, r)
}
