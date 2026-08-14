package opencart

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the OpenCart install/remove/manage routes onto mux, gated
// behind the "opencart" feature flag. No list page (matches drupal/joomla's
// scope: manage via the general Site Manager instead).
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "opencart")(h)
	}
	mux.Handle("/opencart/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /opencart/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveOpenCart(a, w, r) }))
	mux.Handle("GET /opencart/login", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOpenCartLogin(a, w, r) }))
	mux.Handle("POST /opencart/cache", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOpenCartCacheClean(a, w, r) }))
	mux.Handle("GET /opencart/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleOpenCartLogs(a, w, r) }))
}

// withOpenCartForm clones r as a POST carrying the given values as both
// Form and PostForm, so a UI handler that reads r.FormValue(...) sees
// exactly the fields the API's JSON body supplied - same pattern used by
// joomla/register.go's withJoomlaForm.
func withOpenCartForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/opencart/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "opencart", "POST /api/opencart/install", func(w http.ResponseWriter, r *http.Request) { apiInstallOpenCart(a, w, r) })
	apiregistry.Handle(mux, a, "opencart", "DELETE /api/opencart/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveOpenCart(a, w, r) })
}
