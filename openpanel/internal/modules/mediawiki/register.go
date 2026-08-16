package mediawiki

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the MediaWiki install/remove/manage routes onto mux,
// gated behind the "mediawiki" feature flag. No list page (matches
// joomla/moodle's scope: manage via the general Site Manager instead).
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "mediawiki")(h)
	}
	mux.Handle("/mediawiki/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("GET /mediawiki/versions", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMediaWikiVersions(a, w, r) }))
	mux.Handle("POST /mediawiki/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveMediaWiki(a, w, r) }))
	mux.Handle("POST /mediawiki/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMediaWikiClone(a, w, r) }))
	mux.Handle("GET /mediawiki/login", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMediaWikiLogin(a, w, r) }))
	mux.Handle("GET /mediawiki/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMediaWikiLogs(a, w, r) }))
	mux.Handle("GET /mediawiki/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMediaWikiGetBackupDates(a, w, r) }))
	mux.Handle("GET /mediawiki/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMediaWikiRestoreBackup(a, w, r) }))
	mux.Handle("GET /mediawiki/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMediaWikiRunBackup(a, w, r) }))
}

// withMediaWikiForm clones r as a POST carrying the given values as both
// Form and PostForm, so a UI handler that reads r.FormValue(...) sees
// exactly the fields the API's JSON body supplied - same pattern used by
// every other CMS module's with{CMS}Form.
func withMediaWikiForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/mediawiki/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "mediawiki", "POST /api/mediawiki/install", func(w http.ResponseWriter, r *http.Request) { apiInstallMediaWiki(a, w, r) })
	apiregistry.Handle(mux, a, "mediawiki", "DELETE /api/mediawiki/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveMediaWiki(a, w, r) })
}
