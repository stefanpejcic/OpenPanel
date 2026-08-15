package nextcloud

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

// handleNextcloudVersions backs nextcloud_install.html's version dropdown -
// there's no GitHub releases API to hit client-side the way
// joomla/opencart's install forms do, so this exposes the server-side
// HTML-scrape result (version.go) as JSON instead.
func handleNextcloudVersions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	versions, err := listNextcloudVersions(r.Context())
	if err != nil {
		writeAPIJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeAPIJSON(w, http.StatusOK, map[string]any{"versions": versions})
}

// apiInstallNextcloud delegates straight to handleInstallPage (which itself
// calls handleInstallStream on POST): same site-limit check, same NDJSON
// progress stream written directly to the response - just fed from the
// API's JSON body instead of a UI form post.
func apiInstallNextcloud(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID         string `json:"domain_id"`
		Subdirectory     string `json:"subdirectory"`
		NextcloudVersion string `json:"nextcloud_version"`
		AdminUsername    string `json:"admin_username"`
		AdminPassword    string `json:"admin_password"`
		AdminEmail       string `json:"admin_email"`
		DBName           string `json:"db_name"`
		DBUser           string `json:"db_user"`
		DBPassword       string `json:"db_password"`
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
		"nextcloud_version": {body.NextcloudVersion},
		"admin_username":    {body.AdminUsername}, "admin_password": {body.AdminPassword}, "admin_email": {body.AdminEmail},
		"db_name": {body.DBName}, "db_user": {body.DBUser}, "db_password": {body.DBPassword},
	}
	handleInstallPage(a, w, withNextcloudForm(r, form))
}

// apiRemoveNextcloud delegates to handleRemoveNextcloud with the path's
// {site_id} translated into the "id" form field it expects, and
// output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemoveNextcloud(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withNextcloudForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveNextcloud(a, w, cloned)
}
