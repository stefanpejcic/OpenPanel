package phpbb

import (
	"encoding/json"
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
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
		WHERE sites.id = ? AND sites.type = 'phpbb'`, siteID)
	if scanErr := row.Scan(&siteName, &docroot); scanErr != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return nil, false
	}

	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return nil, false
	}
	dbInfo := extractPhpbbDatabaseInfoForBackup(userContext, docroot)
	sourceDB := dbInfo["database_name"]
	if sourceDB == "" {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not determine source database: " + dbInfo["error"]})
		return nil, false
	}

	var body struct {
		TargetDomain         string `json:"target_domain"`
		Subdirectory         string `json:"subdirectory"`
		TargetDB             string `json:"target_db"`
		TargetDBUser         string `json:"target_db_user"`
		TargetDBUserPassword string `json:"target_db_user_password"`
		AdminEmail           string `json:"admin_email"`
		Version              string `json:"version"`
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
		"source_domain": {siteName}, "source_folder": {docroot}, "source_db": {sourceDB},
		"target_domain": {body.TargetDomain}, "subdirectory": {body.Subdirectory},
		"target_db": {body.TargetDB}, "target_db_user": {body.TargetDBUser}, "target_db_user_password": {body.TargetDBUserPassword},
		"admin_email": {body.AdminEmail}, "version": {body.Version},
	}
	return form, true
}
