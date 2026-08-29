package tinyfilemanager

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the TinyFileManager install/remove/backup routes onto
// mux, gated behind the "tinyfilemanager" feature flag. No maintenance,
// login, cache, or clone routes - see tinyfilemanager.go's package doc
// comment for why those aren't implemented.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "tinyfilemanager")(h)
	}
	mux.Handle("/tinyfilemanager/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /tinyfilemanager/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveTinyFileManager(a, w, r) }))
	mux.Handle("GET /tinyfilemanager/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleTinyFileManagerGetBackupDates(a, w, r) }))
	mux.Handle("GET /tinyfilemanager/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleTinyFileManagerRestoreBackup(a, w, r) }))
	mux.Handle("GET /tinyfilemanager/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleTinyFileManagerRunBackup(a, w, r) }))
}

// withTinyFileManagerForm clones r as a POST carrying the given values as
// both Form and PostForm, so a UI handler that reads r.FormValue(...) sees
// exactly the fields the API's JSON body supplied.
func withTinyFileManagerForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/tinyfilemanager/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "tinyfilemanager", "POST /api/tinyfilemanager/install", func(w http.ResponseWriter, r *http.Request) { apiInstallTinyFileManager(a, w, r) })
	apiregistry.Handle(mux, a, "tinyfilemanager", "DELETE /api/tinyfilemanager/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveTinyFileManager(a, w, r) })
}
