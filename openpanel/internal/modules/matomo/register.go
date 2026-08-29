package matomo

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the Matomo install/remove/manage routes onto mux, gated
// behind the "matomo" feature flag. No list page (matches
// drupal/joomla/opencart/nextcloud/prestashop's scope: manage via the
// general Site Manager instead).
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "matomo")(h)
	}
	mux.Handle("/matomo/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /matomo/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveMatomo(a, w, r) }))
	mux.Handle("POST /matomo/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMatomoClone(a, w, r) }))
	mux.Handle("GET /matomo/login", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMatomoLogin(a, w, r) }))
	mux.Handle("POST /matomo/cache", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMatomoCacheClean(a, w, r) }))
	mux.Handle("GET /matomo/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMatomoLogs(a, w, r) }))
	mux.Handle("GET /matomo/versions", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMatomoVersions(a, w, r) }))
	mux.Handle("GET /matomo/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMatomoGetBackupDates(a, w, r) }))
	mux.Handle("GET /matomo/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMatomoRestoreBackup(a, w, r) }))
	mux.Handle("GET /matomo/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMatomoRunBackup(a, w, r) }))
	mux.Handle("POST /matomo/update", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMatomoUpdate(a, w, r) }))
}

// withMatomoForm clones r as a POST carrying the given values as both Form
// and PostForm, so a UI handler that reads r.FormValue(...) sees exactly
// the fields the API's JSON body supplied - same pattern used by every
// other CMS module's with{CMS}Form.
func withMatomoForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/matomo/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "matomo", "POST /api/matomo/install", func(w http.ResponseWriter, r *http.Request) { apiInstallMatomo(a, w, r) })
	apiregistry.Handle(mux, a, "matomo", "DELETE /api/matomo/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveMatomo(a, w, r) })
	apiregistry.Handle(mux, a, "matomo", "POST /api/matomo/sites/{site_id}/clone", func(w http.ResponseWriter, r *http.Request) { apiMatomoClone(a, w, r) })
	apiregistry.Handle(mux, a, "matomo", "POST /api/matomo/sites/{site_id}/update", func(w http.ResponseWriter, r *http.Request) { apiMatomoUpdate(a, w, r) })
	apiregistry.Handle(mux, a, "matomo", "POST /api/matomo/sites/{site_id}/cache", func(w http.ResponseWriter, r *http.Request) { apiMatomoCache(a, w, r) })
}
