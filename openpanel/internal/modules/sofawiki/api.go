package sofawiki

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

// apiInstallSofawiki delegates straight to handleInstallPage (which
// itself calls handleInstallStream on POST): same site-limit check, same
// NDJSON progress stream written directly to the response - just fed from
// the API's JSON body instead of a UI form post.
func apiInstallSofawiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID     string `json:"domain_id"`
		Subdirectory string `json:"subdirectory"`
		AdminEmail   string `json:"admin_email"`
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
	}
	handleInstallPage(a, w, withSofawikiForm(r, form))
}

// apiRemoveSofawiki delegates to handleRemoveSofawiki with the path's
// {site_id} translated into the "id" form field it expects, and
// output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemoveSofawiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withSofawikiForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveSofawiki(a, w, cloned)
}

// apiCloneSofawiki resolves the path's {site_id} into the source domain/
// docroot handleSofawikiClone expects as source_domain/source_folder,
// using the same "id" lookup query apiRemoveSofawiki (via
// handleRemoveSofawiki) and manage.go use - there's no database to derive
// a source_db from (SofaWiki is flat-file) - and takes the
// destination-side fields from the JSON body.
func apiCloneSofawiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")

	var siteName, docroot string
	row := a.DB.QueryRowContext(r.Context(), `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'sofawiki'`, siteID)
	if scanErr := row.Scan(&siteName, &docroot); scanErr != nil {
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
	handleSofawikiClone(a, w, withSofawikiForm(r, form))
}
