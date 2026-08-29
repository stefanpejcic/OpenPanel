package phpbb

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the phpBB install/remove/clone/backup routes onto mux,
// gated behind the "phpbb" feature flag. No update route - see phpbb.go's
// package doc comment for why (browser-link-only, like Joomla/OpenCart/
// PrestaShop).
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "phpbb")(h)
	}
	mux.Handle("/phpbb/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /phpbb/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemovePhpbb(a, w, r) }))
	mux.Handle("GET /phpbb/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePhpbbGetBackupDates(a, w, r) }))
	mux.Handle("GET /phpbb/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePhpbbRestoreBackup(a, w, r) }))
	mux.Handle("GET /phpbb/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePhpbbRunBackup(a, w, r) }))
	mux.Handle("POST /phpbb/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePhpbbClone(a, w, r) }))
}

// withPhpbbForm clones r as a POST carrying the given values as both Form
// and PostForm, so a UI handler that reads r.FormValue(...) sees exactly
// the fields the API's JSON body supplied.
func withPhpbbForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/phpbb/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "phpbb", "POST /api/phpbb/install", func(w http.ResponseWriter, r *http.Request) { apiInstallPhpbb(a, w, r) })
	apiregistry.Handle(mux, a, "phpbb", "DELETE /api/phpbb/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemovePhpbb(a, w, r) })
	apiregistry.Handle(mux, a, "phpbb", "POST /api/phpbb/sites/{site_id}/clone", func(w http.ResponseWriter, r *http.Request) { apiClonePhpbb(a, w, r) })
}
