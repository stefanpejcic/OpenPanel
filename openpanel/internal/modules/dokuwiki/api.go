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
