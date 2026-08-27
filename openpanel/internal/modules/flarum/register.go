package flarum

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the Flarum install/remove/manage routes onto mux, gated
// behind the "flarum" feature flag. No maintenance or login routes - see
// flarum.go's package doc comment for why those two aren't implemented.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "flarum")(h)
	}
	mux.Handle("/flarum/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /flarum/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveFlarum(a, w, r) }))
	mux.Handle("POST /flarum/cache", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFlarumCacheClear(a, w, r) }))
	mux.Handle("GET /flarum/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFlarumLogs(a, w, r) }))
	mux.Handle("GET /flarum/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFlarumGetBackupDates(a, w, r) }))
	mux.Handle("GET /flarum/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFlarumRestoreBackup(a, w, r) }))
	mux.Handle("GET /flarum/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFlarumRunBackup(a, w, r) }))
	mux.Handle("POST /flarum/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFlarumClone(a, w, r) }))
	mux.Handle("POST /flarum/update", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFlarumUpdate(a, w, r) }))
}

// withFlarumForm clones r as a POST carrying the given values as both Form
// and PostForm, so a UI handler that reads r.FormValue(...) sees exactly
// the fields the API's JSON body supplied.
func withFlarumForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/flarum/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "flarum", "POST /api/flarum/install", func(w http.ResponseWriter, r *http.Request) { apiInstallFlarum(a, w, r) })
	apiregistry.Handle(mux, a, "flarum", "DELETE /api/flarum/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveFlarum(a, w, r) })
}
