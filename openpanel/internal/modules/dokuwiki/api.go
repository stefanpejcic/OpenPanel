package dokuwiki

import (
	"encoding/json"
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

func writeAPIJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiInstallDokuwiki delegates straight to handleInstallPage (which
// itself calls handleInstallStream on POST): same site-limit check, same
// NDJSON progress stream written directly to the response.
func apiInstallDokuwiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID      string `json:"domain_id"`
		Subdirectory  string `json:"subdirectory"`
		AdminEmail    string `json:"admin_email"`
		AdminUser     string `json:"admin_user"`
		AdminPassword string `json:"admin_password"`
		AdminFullName string `json:"admin_full_name"`
		SiteTitle     string `json:"site_title"`
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
		"domain_id": {body.DomainID}, "subdirectory": {body.Subdirectory}, "admin_email": {body.AdminEmail},
		"admin_user": {body.AdminUser}, "admin_password": {body.AdminPassword},
		"admin_full_name": {body.AdminFullName}, "site_title": {body.SiteTitle},
	}
	handleInstallPage(a, w, withDokuwikiForm(r, form))
}

// apiRemoveDokuwiki delegates to handleRemoveDokuwiki with the path's
// {site_id} translated into the "id" form field it expects, and
// output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemoveDokuwiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withDokuwikiForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveDokuwiki(a, w, cloned)
}

// resolveDokuwikiSiteID looks up the path's {site_id} the same way
// apiRemoveDokuwiki (via handleRemoveDokuwiki) and manage.go's own "id"
// lookup do, returning the site's full site_name (domain[/subdirectory])
// and docroot.
func resolveDokuwikiSiteID(a *appctx.App, r *http.Request, siteID string) (siteName, docroot string, ok bool) {
	row := a.DB.QueryRowContext(r.Context(), `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'dokuwiki'`, siteID)
	if err := row.Scan(&siteName, &docroot); err != nil {
		return "", "", false
	}
	return siteName, docroot, true
}

// apiCloneDokuwiki resolves the path's {site_id} into the source domain/
// docroot handleDokuwikiClone expects as source_domain/source_folder (same
// "id" lookup as apiRemoveDokuwiki) - there's no database to derive a
// source_db from (DokuWiki is flat-file) - and takes the destination-side
// fields from the JSON body.
func apiCloneDokuwiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	siteName, docroot, ok := resolveDokuwikiSiteID(a, r, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}

	var body struct {
		TargetDomain string `json:"target_domain"`
		Subdirectory string `json:"subdirectory"`
		AdminEmail   string `json:"admin_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if body.TargetDomain == "" {
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "target_domain is required"})
		return
	}

	form := url.Values{
		"source_domain": {siteName}, "source_folder": {docroot},
		"target_domain": {body.TargetDomain}, "subdirectory": {body.Subdirectory},
		"admin_email": {body.AdminEmail},
	}
	handleDokuwikiClone(a, w, withDokuwikiForm(r, form))
}

// apiUpdateDokuwiki resolves the path's {site_id} the same way
// apiCloneDokuwiki does, then delegates to handleDokuwikiUpdate with
// domain/docroot set as URL query params (that handler reads
// r.URL.Query(), not form values) - the NDJSON progress stream is written
// directly to the response as-is.
func apiUpdateDokuwiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	siteName, docroot, ok := resolveDokuwikiSiteID(a, r, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}

	cloned := r.Clone(r.Context())
	cloned.Method = http.MethodPost
	q := cloned.URL.Query()
	q.Set("domain", siteName)
	q.Set("docroot", docroot)
	cloned.URL.RawQuery = q.Encode()
	handleDokuwikiUpdate(a, w, cloned)
}
