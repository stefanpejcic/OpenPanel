package drupal

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

// apiInstallDrupal delegates straight to handleInstallPage (which itself
// calls handleInstallStream on POST): same site-limit check, same NDJSON
// progress stream written directly to the response - just fed from the
// API's JSON body instead of a UI form post.
func apiInstallDrupal(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID      string `json:"domain_id"`
		Subdirectory  string `json:"subdirectory"`
		SiteName      string `json:"site_name"`
		DrupalVersion string `json:"drupal_version"`
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
		"site_name": {body.SiteName}, "drupal_version": {body.DrupalVersion},
		"admin_username": {body.AdminUsername}, "admin_password": {body.AdminPassword}, "admin_email": {body.AdminEmail},
		"db_name": {body.DBName}, "db_user": {body.DBUser}, "db_password": {body.DBPassword},
	}
	handleInstallPage(a, w, withDrupalForm(r, form))
}

// apiRemoveDrupal delegates to handleRemoveDrupal with the path's
// {site_id} translated into the "id" form field it expects, and
// output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemoveDrupal(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withDrupalForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveDrupal(a, w, cloned)
}

// apiResolveDrupalSite resolves {site_id} into the (domain, docroot) pair
// every drush-backed handler in this file needs - domain includes any
// subdirectory suffix (e.g. "example.com/blog") exactly like sites.site_name
// stores it, and docroot is the real install path (domains.docroot with
// that subdirectory appended), matching how install.go's own installPath
// and handleRemoveDrupal's own realInstallPath are both computed from the
// same two columns.
func apiResolveDrupalSite(ctx context.Context, a *appctx.App, siteID string) (domain, docroot string, ok bool) {
	var siteName string
	var rootDocroot sql.NullString
	row := a.DB.QueryRowContext(ctx, `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'drupal'`, siteID)
	if err := row.Scan(&siteName, &rootDocroot); err != nil || !rootDocroot.Valid || rootDocroot.String == "" {
		return "", "", false
	}
	docroot = rootDocroot.String
	if idx := strings.Index(siteName, "/"); idx != -1 {
		docroot = strings.TrimSuffix(rootDocroot.String, "/") + "/" + siteName[idx+1:]
	}
	return siteName, docroot, true
}

// apiDrupalClone delegates to handleDrupalClone (which already writes a
// JSON response as-is), resolving {site_id} into the source_domain/
// source_folder fields it expects and taking every other clone field from
// the JSON body - same shape as wordpress/api.go's apiWordPressClone, minus
// source_domain/source_folder since those come from the path here instead.
func apiDrupalClone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	sourceDomain, sourceFolder, ok := apiResolveDrupalSite(r.Context(), a, siteID)
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
		DrupalVersion        string `json:"drupal_version"`
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
		"admin_email": {body.AdminEmail}, "drupal_version": {body.DrupalVersion},
	}
	handleDrupalClone(a, w, withDrupalForm(r, form))
}

// apiDrupalUpdate resolves {site_id} into the domain/docroot query params
// handleDrupalUpdate reads directly (r.URL.Query(), not r.FormValue), then
// delegates to it as-is - same NDJSON progress stream the UI's Update
// button gets.
func apiDrupalUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	domain, docroot, ok := apiResolveDrupalSite(r.Context(), a, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}
	q := r.URL.Query()
	q.Set("domain", domain)
	q.Set("docroot", docroot)
	r.URL.RawQuery = q.Encode()
	handleDrupalUpdate(a, w, r)
}

// apiDrupalCache resolves {site_id} into the domain/docroot query params
// handleDrupalCacheRebuild reads (via drushRequestParams), then delegates
// to it as-is.
func apiDrupalCache(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	domain, docroot, ok := apiResolveDrupalSite(r.Context(), a, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}
	q := r.URL.Query()
	q.Set("domain", domain)
	q.Set("docroot", docroot)
	r.URL.RawQuery = q.Encode()
	handleDrupalCacheRebuild(a, w, r)
}
