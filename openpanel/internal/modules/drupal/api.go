package drupal

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

func writeAPIJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiListDrupal mirrors handleListDrupal but returns JSON instead of
// rendering the HTML list page.
func apiListDrupal(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	rows, execErr := a.DB.QueryContext(ctx, `
		SELECT site_name, domain_id, admin_email, version, created_date, type, id
		FROM sites
		WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?)
		AND type = 'drupal'`, userID)
	if execErr != nil {
		writeAPIJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}
	defer rows.Close()

	type siteEntry struct {
		SiteName    *string `json:"site_name"`
		DomainID    *int64  `json:"domain_id"`
		AdminEmail  *string `json:"admin_email"`
		Version     *string `json:"version"`
		CreatedDate *string `json:"created_date"`
		Type        *string `json:"type"`
		ID          *int64  `json:"id"`
	}
	sites := []siteEntry{}
	for rows.Next() {
		var siteName, adminEmail, version, createdDate, typ sql.NullString
		var domainID, id sql.NullInt64
		if scanErr := rows.Scan(&siteName, &domainID, &adminEmail, &version, &createdDate, &typ, &id); scanErr != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		entry := siteEntry{}
		if siteName.Valid {
			entry.SiteName = &siteName.String
		}
		if domainID.Valid {
			entry.DomainID = &domainID.Int64
		}
		if adminEmail.Valid {
			entry.AdminEmail = &adminEmail.String
		}
		if version.Valid {
			entry.Version = &version.String
		}
		if createdDate.Valid {
			entry.CreatedDate = &createdDate.String
		}
		if typ.Valid {
			entry.Type = &typ.String
		}
		if id.Valid {
			entry.ID = &id.Int64
		}
		sites = append(sites, entry)
	}
	writeAPIJSON(w, http.StatusOK, map[string]any{"sites": sites, "count": len(sites)})
}

// apiInstallDrupal delegates straight to handleInstallPage (which itself
// calls handleInstallStream on POST): same site-limit check, same NDJSON
// progress stream written directly to the response - just fed from the
// API's JSON body instead of a UI form post.
func apiInstallDrupal(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID      string `json:"domain_id"`
		Subdirectory  string `json:"subdirectory"`
		SiteName      string `json:"site_name"`
		DrupalVersion string `json:"drupal_version"`
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
		"site_name": {body.SiteName}, "drupal_version": {body.DrupalVersion},
		"admin_username": {body.AdminUsername}, "admin_password": {body.AdminPassword}, "admin_email": {body.AdminEmail},
		"db_name": {body.DBName}, "db_user": {body.DBUser}, "db_password": {body.DBPassword},
	}
	handleInstallPage(a, w, withDrupalForm(r, form))
}

// apiRemoveDrupal delegates to handleRemoveDrupal with the path's
// {site_id} translated into the "id" form field it expects, and
// output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemoveDrupal(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withDrupalForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveDrupal(a, w, cloned)
}
