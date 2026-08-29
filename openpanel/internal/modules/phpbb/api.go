package phpbb

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

func apiInstallPhpbb(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID         string `json:"domain_id"`
		Subdirectory     string `json:"subdirectory"`
		BoardName        string `json:"board_name"`
		BoardDescription string `json:"board_description"`
		AdminUsername    string `json:"admin_username"`
		AdminPassword    string `json:"admin_password"`
		AdminEmail       string `json:"admin_email"`
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
		"board_name": {body.BoardName}, "board_description": {body.BoardDescription},
		"admin_username": {body.AdminUsername}, "admin_password": {body.AdminPassword}, "admin_email": {body.AdminEmail},
	}
	handleInstallPage(a, w, withPhpbbForm(r, form))
}

func apiRemovePhpbb(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withPhpbbForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemovePhpbb(a, w, cloned)
}

// apiClonePhpbb resolves the path's {site_id} into the source domain/
// docroot handlePhpbbClone expects as source_domain/source_folder, using
// the same "id" lookup query apiRemovePhpbb (via handleRemovePhpbb) and
// manage.go use, derives source_db from config.php via
// extractPhpbbDatabaseInfoForBackup, and takes the destination-side
// fields from the JSON body.
func apiClonePhpbb(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")

	var siteName, docroot string
	row := a.DB.QueryRowContext(r.Context(), `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'phpbb'`, siteID)
	if scanErr := row.Scan(&siteName, &docroot); scanErr != nil {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}

	_, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbInfo := extractPhpbbDatabaseInfoForBackup(userContext, docroot)
	sourceDB := dbInfo["database_name"]
	if sourceDB == "" {
		writeAPIJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not determine source database: " + dbInfo["error"]})
		return
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
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if body.TargetDomain == "" {
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "target_domain is required"})
		return
	}

	form := url.Values{
		"source_domain": {siteName}, "source_folder": {docroot}, "source_db": {sourceDB},
		"target_domain": {body.TargetDomain}, "subdirectory": {body.Subdirectory},
		"target_db": {body.TargetDB}, "target_db_user": {body.TargetDBUser}, "target_db_user_password": {body.TargetDBUserPassword},
		"admin_email": {body.AdminEmail}, "version": {body.Version},
	}
	handlePhpbbClone(a, w, withPhpbbForm(r, form))
}
