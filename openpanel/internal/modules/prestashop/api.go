package prestashop

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

// handlePrestashopVersions backs prestashop_install.html's version dropdown -
// GitHub's releases API is CORS-open and could be hit client-side, but
// filtering out versions with no downloadable asset (see version.go) needs
// server-side logic, so this exposes the already-filtered list as JSON
// instead, matching nextcloud's identical approach.
func handlePrestashopVersions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	versions, err := listPrestashopVersions(r.Context())
	if err != nil {
		writeAPIJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeAPIJSON(w, http.StatusOK, map[string]any{"versions": versions})
}

// apiInstallPrestashop delegates straight to handleInstallPage (which
// itself calls handleInstallStream on POST): same site-limit check, same
// NDJSON progress stream written directly to the response - just fed from
// the API's JSON body instead of a UI form post.
func apiInstallPrestashop(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID          string `json:"domain_id"`
		Subdirectory      string `json:"subdirectory"`
		PrestashopVersion string `json:"prestashop_version"`
		AdminFirstname    string `json:"admin_firstname"`
		AdminLastname     string `json:"admin_lastname"`
		AdminPassword     string `json:"admin_password"`
		AdminEmail        string `json:"admin_email"`
		DBName            string `json:"db_name"`
		DBUser            string `json:"db_user"`
		DBPassword        string `json:"db_password"`
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
		"prestashop_version": {body.PrestashopVersion},
		"admin_firstname":    {body.AdminFirstname}, "admin_lastname": {body.AdminLastname},
		"admin_password": {body.AdminPassword}, "admin_email": {body.AdminEmail},
		"db_name": {body.DBName}, "db_user": {body.DBUser}, "db_password": {body.DBPassword},
	}
	handleInstallPage(a, w, withPrestashopForm(r, form))
}

// apiRemovePrestashop delegates to handleRemovePrestashop with the path's
// {site_id} translated into the "id" form field it expects, and
// output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemovePrestashop(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withPrestashopForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemovePrestashop(a, w, cloned)
}

// resolvePrestashopSiteID looks up the path's {site_id} the same way
// apiRemovePrestashop (via handleRemovePrestashop) and manage.go's own "id"
// lookup do, returning the site's full site_name (domain[/subdirectory])
// and docroot.
func resolvePrestashopSiteID(a *appctx.App, r *http.Request, siteID string) (siteName, docroot string, ok bool) {
	row := a.DB.QueryRowContext(r.Context(), `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'prestashop'`, siteID)
	if err := row.Scan(&siteName, &docroot); err != nil {
		return "", "", false
	}
	return siteName, docroot, true
}

// apiClonePrestashop resolves the path's {site_id} into the source domain/
// docroot handlePrestashopClone expects as source_domain/source_folder
// (same "id" lookup as apiRemovePrestashop), derives source_db from
// app/config/parameters.php via extractPrestashopDatabaseInfoForLogin, and
// takes the destination-side fields from the JSON body.
func apiClonePrestashop(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	siteName, docroot, ok := resolvePrestashopSiteID(a, r, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}

	_, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbInfo := extractPrestashopDatabaseInfoForLogin(userContext, docroot)
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
		PrestashopVersion    string `json:"prestashop_version"`
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
		"admin_email": {body.AdminEmail}, "prestashop_version": {body.PrestashopVersion},
	}
	handlePrestashopClone(a, w, withPrestashopForm(r, form))
}

// apiCachePrestashop resolves the path's {site_id} the same way
// apiClonePrestashop does, then delegates to handlePrestashopCacheClean
// with domain/docroot set as URL query params (prestashopRequestParams
// reads r.URL.Query(), not form values).
func apiCachePrestashop(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	siteName, docroot, ok := resolvePrestashopSiteID(a, r, siteID)
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
	handlePrestashopCacheClean(a, w, cloned)
}
