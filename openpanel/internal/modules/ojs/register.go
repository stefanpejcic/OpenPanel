package ojs

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the OJS install/remove/manage routes onto mux, gated
// behind the "ojs" feature flag. No list page (matches
// drupal/joomla/moodle's scope: manage via the general Site Manager
// instead).
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "ojs")(h)
	}
	mux.Handle("/ojs/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /ojs/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveOJS(a, w, r) }))
	mux.Handle("GET /ojs/login", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOJSLogin(a, w, r) }))
	mux.Handle("POST /ojs/cache", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOJSCacheClean(a, w, r) }))
	mux.Handle("GET /ojs/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOJSLogs(a, w, r) }))
	mux.Handle("GET /ojs/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOJSGetBackupDates(a, w, r) }))
	mux.Handle("GET /ojs/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOJSRestoreBackup(a, w, r) }))
	mux.Handle("GET /ojs/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOJSRunBackup(a, w, r) }))
	mux.Handle("POST /ojs/update", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOJSUpdate(a, w, r) }))
	mux.Handle("POST /ojs/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOJSClone(a, w, r) }))
}

// withOJSForm clones r as a POST carrying the given values as both Form and
// PostForm, so a UI handler that reads r.FormValue(...) sees exactly the
// fields the API's JSON body supplied - same pattern used by every other
// CMS module's with{CMS}Form.
func withOJSForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/ojs/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "ojs", "POST /api/ojs/install", func(w http.ResponseWriter, r *http.Request) { apiInstallOJS(a, w, r) })
	apiregistry.Handle(mux, a, "ojs", "DELETE /api/ojs/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveOJS(a, w, r) })
	apiregistry.Handle(mux, a, "ojs", "POST /api/ojs/sites/{site_id}/clone", func(w http.ResponseWriter, r *http.Request) { apiOJSClone(a, w, r) })
	apiregistry.Handle(mux, a, "ojs", "POST /api/ojs/sites/{site_id}/update", func(w http.ResponseWriter, r *http.Request) { apiOJSUpdate(a, w, r) })
	apiregistry.Handle(mux, a, "ojs", "POST /api/ojs/sites/{site_id}/cache", func(w http.ResponseWriter, r *http.Request) { apiOJSCache(a, w, r) })
}
