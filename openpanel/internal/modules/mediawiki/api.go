package mediawiki

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

// handleMediaWikiVersions backs mediawiki_install.html's version dropdown
// and mediawiki_app.html's Update tab - there's no GitHub releases API to
// hit client-side (releases.wikimedia.org sends no CORS headers either, so
// even a direct scrape can't be fetched from the browser), so this exposes
// the server-side HTML-scrape result (version.go) as JSON instead.
func handleMediaWikiVersions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	versions, err := listMediaWikiVersions(r.Context())
	if err != nil {
		writeAPIJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeAPIJSON(w, http.StatusOK, map[string]any{"versions": versions})
}

// apiInstallMediaWiki delegates straight to handleInstallPage (which itself
// calls handleInstallStream on POST): same site-limit check, same NDJSON
// progress stream written directly to the response - just fed from the
// API's JSON body instead of a UI form post.
func apiInstallMediaWiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID         string `json:"domain_id"`
		Subdirectory     string `json:"subdirectory"`
		MediaWikiVersion string `json:"mediawiki_version"`
		SiteName         string `json:"site_name"`
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
		"mediawiki_version": {body.MediaWikiVersion},
		"site_name":         {body.SiteName},
		"admin_username":    {body.AdminUsername}, "admin_password": {body.AdminPassword}, "admin_email": {body.AdminEmail},
		"db_name": {body.DBName}, "db_user": {body.DBUser}, "db_password": {body.DBPassword},
	}
	handleInstallPage(a, w, withMediaWikiForm(r, form))
}

// apiRemoveMediaWiki delegates to handleRemoveMediaWiki with the path's
// {site_id} translated into the "id" form field it expects, and
// output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemoveMediaWiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withMediaWikiForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveMediaWiki(a, w, cloned)
}
