package sofawiki

import (
	"encoding/json"
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// apiCloneForm builds the clone form from the API's JSON body and the site behind {site_id}
func apiCloneForm(a *appctx.App, w http.ResponseWriter, r *http.Request) (url.Values, bool) {
	siteID := r.PathValue("site_id")

	var siteName, docroot string
	row := a.DB.QueryRowContext(r.Context(), `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'sofawiki'`, siteID)
	if scanErr := row.Scan(&siteName, &docroot); scanErr != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return nil, false
	}

	var body struct {
		TargetDomain string `json:"target_domain"`
		Subdirectory string `json:"subdirectory"`
		AdminEmail   string `json:"admin_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return nil, false
	}
	if body.TargetDomain == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "target_domain is required"})
		return nil, false
	}

	form := url.Values{
		"source_domain": {siteName}, "source_folder": {docroot},
		"target_domain": {body.TargetDomain}, "subdirectory": {body.Subdirectory},
		"admin_email": {body.AdminEmail},
	}
	return form, true
}
