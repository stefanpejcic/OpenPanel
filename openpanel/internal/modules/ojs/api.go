package ojs

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

// apiInstallOJS delegates straight to handleInstallPage (which itself
// calls handleInstallStream on POST): same site-limit check, same NDJSON
// progress stream written directly to the response - just fed from the
// API's JSON body instead of a UI form post.
func apiInstallOJS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID      string `json:"domain_id"`
		Subdirectory  string `json:"subdirectory"`
		OJSVersion    string `json:"ojs_version"`
		AdminUsername string `json:"admin_username"`
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
		"ojs_version":    {body.OJSVersion},
		"admin_username": {body.AdminUsername}, "admin_password": {body.AdminPassword}, "admin_email": {body.AdminEmail},
		"db_name": {body.DBName}, "db_user": {body.DBUser}, "db_password": {body.DBPassword},
	}
	handleInstallPage(a, w, withOJSForm(r, form))
}

// apiRemoveOJS delegates to handleRemoveOJS with the path's {site_id}
// translated into the "id" form field it expects, and output=json forced
// so it returns JSON instead of a flash-and-redirect.
func apiRemoveOJS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withOJSForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveOJS(a, w, cloned)
}

// apiResolveOJSSite resolves {site_id} into the (domain, docroot) pair
// every handler in this file needs - mirrors moodle/api.go's
// apiResolveMoodleSite. docroot isn't read by handleOJSClone (OJS's docroot
// is a symlink derived from siteSlug(), not a stored source_folder form
// field - see clone.go's package doc comment) but is still needed by
// ojsRequestParams for the cache endpoint.
func apiResolveOJSSite(ctx context.Context, a *appctx.App, siteID string) (domain, docroot string, ok bool) {
	var siteName string
	var rootDocroot sql.NullString
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'ojs'`, siteID)
	if err := row.Scan(&siteName, &rootDocroot); err != nil || !rootDocroot.Valid || rootDocroot.String == "" {
		return "", "", false
	}
	docroot = rootDocroot.String
	if idx := strings.Index(siteName, "/"); idx != -1 {
		docroot = strings.TrimSuffix(rootDocroot.String, "/") + "/" + siteName[idx+1:]
	}
	return siteName, docroot, true
}

// apiOJSClone delegates to handleOJSClone, resolving {site_id} into the
// source_domain field it expects (no source_folder - see
// apiResolveOJSSite's comment) and taking every other clone field from the
// JSON body - mirrors apiMoodleClone.
func apiOJSClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	sourceDomain, _, ok := apiResolveOJSSite(r.Context(), a, siteID)
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
		OJSVersion           string `json:"ojs_version"`
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
		"source_db": {body.SourceDB}, "subdirectory": {body.Subdirectory},
		"target_db": {body.TargetDB}, "target_db_user": {body.TargetDBUser}, "target_db_user_password": {body.TargetDBUserPassword},
		"admin_email": {body.AdminEmail}, "ojs_version": {body.OJSVersion},
	}
	handleOJSClone(a, w, withOJSForm(r, form))
}

// apiOJSUpdate resolves {site_id} into the domain query param
// handleOJSUpdate reads directly, then delegates to it as-is.
func apiOJSUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	domain, docroot, ok := apiResolveOJSSite(r.Context(), a, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}
	q := r.URL.Query()
	q.Set("domain", domain)
	q.Set("docroot", docroot)
	r.URL.RawQuery = q.Encode()
	handleOJSUpdate(a, w, r)
}

// apiOJSCache resolves {site_id} into the domain/docroot query params
// handleOJSCacheClean reads (via ojsRequestParams), then delegates to it
// as-is.
func apiOJSCache(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	domain, docroot, ok := apiResolveOJSSite(r.Context(), a, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}
	q := r.URL.Query()
	q.Set("domain", domain)
	q.Set("docroot", docroot)
	r.URL.RawQuery = q.Encode()
	handleOJSCacheClean(a, w, r)
}
