package joomla

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

// apiInstallJoomla delegates straight to handleInstallPage (which itself
// calls handleInstallStream on POST): same site-limit check, same NDJSON
// progress stream written directly to the response - just fed from the
// API's JSON body instead of a UI form post.
func apiInstallJoomla(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID      string `json:"domain_id"`
		Subdirectory  string `json:"subdirectory"`
		SiteName      string `json:"site_name"`
		JoomlaVersion string `json:"joomla_version"`
		AdminName     string `json:"admin_name"`
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
		"site_name": {body.SiteName}, "joomla_version": {body.JoomlaVersion},
		"admin_name": {body.AdminName}, "admin_username": {body.AdminUsername},
		"admin_password": {body.AdminPassword}, "admin_email": {body.AdminEmail},
		"db_name": {body.DBName}, "db_user": {body.DBUser}, "db_password": {body.DBPassword},
	}
	handleInstallPage(a, w, withJoomlaForm(r, form))
}

// apiRemoveJoomla delegates to handleRemoveJoomla with the path's
// {site_id} translated into the "id" form field it expects, and
// output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemoveJoomla(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withJoomlaForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveJoomla(a, w, cloned)
}

// apiCloneJoomla delegates straight to handleJoomlaClone, which already
// writes a JSON response as-is.
func apiCloneJoomla(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		SourceDomain         string `json:"source_domain"`
		TargetDomain         string `json:"target_domain"`
		SourceDB             string `json:"source_db"`
		SourceFolder         string `json:"source_folder"`
		Subdirectory         string `json:"subdirectory"`
		TargetDB             string `json:"target_db"`
		TargetDBUser         string `json:"target_db_user"`
		TargetDBUserPassword string `json:"target_db_user_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	form := url.Values{
		"source_domain": {body.SourceDomain}, "target_domain": {body.TargetDomain},
		"source_db": {body.SourceDB}, "source_folder": {body.SourceFolder}, "subdirectory": {body.Subdirectory},
		"target_db": {body.TargetDB}, "target_db_user": {body.TargetDBUser}, "target_db_user_password": {body.TargetDBUserPassword},
	}
	handleJoomlaClone(a, w, withJoomlaForm(r, form))
}
