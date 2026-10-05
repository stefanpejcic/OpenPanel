package nextcloud

import (
	"encoding/json"
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleNextcloudVersions backs nextcloud_install.html's version dropdown - there's no GitHub releases API to hit client-side the way joomla/opencart's install forms do, so this exposes the server-side HTML-scrape result (version.go) as JSON instead
func handleNextcloudVersions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	versions, err := listNextcloudVersions(r.Context())
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"versions": versions})
}

// apiCloneForm builds the clone form from the API's JSON body and the site behind {site_id}
func apiCloneForm(a *appctx.App, w http.ResponseWriter, r *http.Request) (url.Values, bool) {
	siteID := r.PathValue("site_id")
	sourceDomain, sourceFolder, ok := cmsapp.ResolveSite(r.Context(), a, "nextcloud", siteID)
	if !ok {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return nil, false
	}

	var body struct {
		TargetDomain         string `json:"target_domain"`
		Subdirectory         string `json:"subdirectory"`
		SourceDB             string `json:"source_db"`
		TargetDB             string `json:"target_db"`
		TargetDBUser         string `json:"target_db_user"`
		TargetDBUserPassword string `json:"target_db_user_password"`
		AdminEmail           string `json:"admin_email"`
		NextcloudVersion     string `json:"nextcloud_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return nil, false
	}
	if body.TargetDomain == "" || body.SourceDB == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "target_domain and source_db are required"})
		return nil, false
	}

	form := url.Values{
		"source_domain": {sourceDomain}, "target_domain": {body.TargetDomain},
		"source_db": {body.SourceDB}, "source_folder": {sourceFolder}, "subdirectory": {body.Subdirectory},
		"target_db": {body.TargetDB}, "target_db_user": {body.TargetDBUser}, "target_db_user_password": {body.TargetDBUserPassword},
		"admin_email": {body.AdminEmail}, "nextcloud_version": {body.NextcloudVersion},
	}
	return form, true
}
