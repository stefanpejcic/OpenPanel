package matomo

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
