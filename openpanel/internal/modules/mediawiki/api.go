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

// resolveMediaWikiSiteID looks up the path's {site_id} the same way
// apiRemoveMediaWiki (via handleRemoveMediaWiki) and manage.go's own "id"
// lookup do, returning the site's full site_name (domain[/subdirectory])
// and docroot.
func resolveMediaWikiSiteID(a *appctx.App, r *http.Request, siteID string) (siteName, docroot string, ok bool) {
	row := a.DB.QueryRowContext(r.Context(), `
		SELECT sites.site_name, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_url = SUBSTRING_INDEX(sites.site_name, '/', 1)
		WHERE sites.id = ? AND sites.type = 'mediawiki'`, siteID)
	if err := row.Scan(&siteName, &docroot); err != nil {
		return "", "", false
	}
	return siteName, docroot, true
}

// apiCloneMediaWiki resolves the path's {site_id} into the source domain/
// docroot handleMediaWikiClone expects as source_domain/source_folder
// (same "id" lookup as apiRemoveMediaWiki), derives source_db from
// LocalSettings.php via extractMediaWikiDatabaseInfoForLogin (the same
// helper mediawiki_app.html's clone form's server-rendered source_db value
// comes from), and takes the destination-side fields from the JSON body.
func apiCloneMediaWiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	siteName, docroot, ok := resolveMediaWikiSiteID(a, r, siteID)
	if !ok {
		writeAPIJSON(w, http.StatusNotFound, map[string]string{"error": "Site not found"})
		return
	}

	_, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbInfo := extractMediaWikiDatabaseInfoForLogin(userContext, docroot)
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
		MediaWikiVersion     string `json:"mediawiki_version"`
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
		"admin_email": {body.AdminEmail}, "mediawiki_version": {body.MediaWikiVersion},
	}
	handleMediaWikiClone(a, w, withMediaWikiForm(r, form))
}

// apiUpdateMediaWiki resolves the path's {site_id} the same way
// apiCloneMediaWiki does, then delegates to handleMediaWikiUpdate with
// domain/docroot set as URL query params (that handler reads
// r.URL.Query(), not form values) - the NDJSON progress stream is written
// directly to the response as-is.
func apiUpdateMediaWiki(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	siteName, docroot, ok := resolveMediaWikiSiteID(a, r, siteID)
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
	handleMediaWikiUpdate(a, w, cloned)
}
