package sofawiki

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the SofaWiki install/remove/manage routes onto mux,
// gated behind the "sofawiki" feature flag. No maintenance, login, or
// cache routes - see sofawiki.go's package doc comment for why those
// aren't implemented.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "sofawiki")(h)
	}
	mux.Handle("/sofawiki/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /sofawiki/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveSofawiki(a, w, r) }))
	mux.Handle("GET /sofawiki/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleSofawikiGetBackupDates(a, w, r) }))
	mux.Handle("GET /sofawiki/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleSofawikiRestoreBackup(a, w, r) }))
	mux.Handle("GET /sofawiki/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleSofawikiRunBackup(a, w, r) }))
	mux.Handle("POST /sofawiki/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleSofawikiClone(a, w, r) }))
}

// withSofawikiForm clones r as a POST carrying the given values as both
// Form and PostForm, so a UI handler that reads r.FormValue(...) sees
// exactly the fields the API's JSON body supplied.
func withSofawikiForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/sofawiki/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "sofawiki", "POST /api/sofawiki/install", func(w http.ResponseWriter, r *http.Request) { apiInstallSofawiki(a, w, r) })
	apiregistry.Handle(mux, a, "sofawiki", "DELETE /api/sofawiki/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveSofawiki(a, w, r) })
}
