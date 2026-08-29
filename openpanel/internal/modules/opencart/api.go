package opencart

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

// apiInstallOpenCart delegates straight to handleInstallPage (which itself
// calls handleInstallStream on POST): same site-limit check, same NDJSON
// progress stream written directly to the response - just fed from the
// API's JSON body instead of a UI form post.
func apiInstallOpenCart(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID        string `json:"domain_id"`
		Subdirectory    string `json:"subdirectory"`
		OpenCartVersion string `json:"opencart_version"`
		AdminUsername   string `json:"admin_username"`
		AdminPassword   string `json:"admin_password"`
		AdminEmail      string `json:"admin_email"`
		DBName          string `json:"db_name"`
		DBUser          string `json:"db_user"`
		DBPassword      string `json:"db_password"`
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
		"opencart_version": {body.OpenCartVersion},
		"admin_username":   {body.AdminUsername}, "admin_password": {body.AdminPassword}, "admin_email": {body.AdminEmail},
		"db_name": {body.DBName}, "db_user": {body.DBUser}, "db_password": {body.DBPassword},
	}
	handleInstallPage(a, w, withOpenCartForm(r, form))
}

// apiRemoveOpenCart delegates to handleRemoveOpenCart with the path's
// {site_id} translated into the "id" form field it expects, and
// output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemoveOpenCart(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withOpenCartForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveOpenCart(a, w, cloned)
}

// resolveOpenCartSiteID looks up the path's {site_id} the same way
// apiRemoveOpenCart (via handleRemoveOpenCart) and manage.go's own "id"
// lookup do, returning the site's full site_name (domain[/subdirectory])
// and docroot.
func resolveOpenCartSiteID(a *appctx.App, r *http.Request, siteID string) (siteName, docroot string, ok bool) {
	row := a.DB.QueryRowContext(r.Context(), `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'opencart'`, siteID)
	if err := row.Scan(&siteName, &docroot); err != nil {
		return "", "", false
	}
	return siteName, docroot, true
}

// apiCloneOpenCart resolves the path's {site_id} into the source domain/
// docroot handleOpenCartClone expects as source_domain/source_folder (same
// "id" lookup as apiRemoveOpenCart), derives source_db from config.php via
// extractOpenCartDatabaseInfoForLogin, and takes the destination-side
// fields from the JSON body.
func apiCloneOpenCart(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	siteName, docroot, ok := resolveOpenCartSiteID(a, r, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}

	_, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbInfo := extractOpenCartDatabaseInfoForLogin(userContext, docroot)
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
		OpenCartVersion      string `json:"opencart_version"`
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
		"admin_email": {body.AdminEmail}, "opencart_version": {body.OpenCartVersion},
	}
	handleOpenCartClone(a, w, withOpenCartForm(r, form))
}

// apiCacheOpenCart resolves the path's {site_id} the same way
// apiCloneOpenCart does, then delegates to handleOpenCartCacheClean with
// domain/docroot set as URL query params (openCartRequestParams reads
// r.URL.Query(), not form values).
func apiCacheOpenCart(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	siteName, docroot, ok := resolveOpenCartSiteID(a, r, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}

	cloned := r.Clone(r.Context())
	cloned.Method = http.MethodPost
	q := cloned.URL.Query()
	q.Set("domain", siteName)
	q.Set("docroot", docroot)
	cloned.URL.RawQuery = q.Encode()
	handleOpenCartCacheClean(a, w, cloned)
}
