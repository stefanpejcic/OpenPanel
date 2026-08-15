package prestashop

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the PrestaShop install/remove/manage routes onto mux,
// gated behind the "prestashop" feature flag. No list page (matches
// drupal/joomla/opencart/nextcloud's scope: manage via the general Site
// Manager instead).
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "prestashop")(h)
	}
	mux.Handle("/prestashop/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /prestashop/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemovePrestashop(a, w, r) }))
	mux.Handle("GET /prestashop/login", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePrestashopLogin(a, w, r) }))
	mux.Handle("POST /prestashop/cache", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePrestashopCacheClean(a, w, r) }))
	mux.Handle("GET /prestashop/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePrestashopLogs(a, w, r) }))
	mux.Handle("GET /prestashop/versions", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePrestashopVersions(a, w, r) }))
}

// withPrestashopForm clones r as a POST carrying the given values as both
// Form and PostForm, so a UI handler that reads r.FormValue(...) sees
// exactly the fields the API's JSON body supplied - same pattern used by
// opencart/nextcloud's with{OpenCart,Nextcloud}Form.
func withPrestashopForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/prestashop/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "prestashop", "POST /api/prestashop/install", func(w http.ResponseWriter, r *http.Request) { apiInstallPrestashop(a, w, r) })
	apiregistry.Handle(mux, a, "prestashop", "DELETE /api/prestashop/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemovePrestashop(a, w, r) })
}
