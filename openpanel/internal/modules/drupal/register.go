package drupal

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the Drupal list/install/remove routes onto mux, gated
// behind the "drupal" feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "drupal")(h)
	}
	mux.Handle("/drupal/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /drupal/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveDrupal(a, w, r) }))
	mux.Handle("GET /drupal/login", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDrupalLogin(a, w, r) }))
	mux.Handle("POST /drupal/cache", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDrupalCacheRebuild(a, w, r) }))
	mux.Handle("GET /drupal/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDrupalLogs(a, w, r) }))
	mux.Handle("/drupal/maintenance", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDrupalMaintenance(a, w, r) }))
	mux.Handle("GET /drupal/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDrupalGetBackupDates(a, w, r) }))
	mux.Handle("GET /drupal/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDrupalRestoreBackup(a, w, r) }))
	mux.Handle("GET /drupal/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDrupalRunBackup(a, w, r) }))
	mux.Handle("POST /drupal/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDrupalClone(a, w, r) }))
	mux.Handle("POST /drupal/update", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDrupalUpdate(a, w, r) }))
}

// withDrupalForm clones r as a POST carrying the given values as both Form
// and PostForm, so a UI handler that reads r.FormValue(...) sees exactly
// the fields the API's JSON body supplied - same pattern used by
// wordpress/api.go's withWPForm and backups/api.go's withForm.
func withDrupalForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/drupal/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "drupal", "POST /api/drupal/install", func(w http.ResponseWriter, r *http.Request) { apiInstallDrupal(a, w, r) })
	apiregistry.Handle(mux, a, "drupal", "DELETE /api/drupal/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveDrupal(a, w, r) })
	apiregistry.Handle(mux, a, "drupal", "POST /api/drupal/sites/{site_id}/clone", func(w http.ResponseWriter, r *http.Request) { apiDrupalClone(a, w, r) })
	apiregistry.Handle(mux, a, "drupal", "POST /api/drupal/sites/{site_id}/update", func(w http.ResponseWriter, r *http.Request) { apiDrupalUpdate(a, w, r) })
	apiregistry.Handle(mux, a, "drupal", "POST /api/drupal/sites/{site_id}/cache", func(w http.ResponseWriter, r *http.Request) { apiDrupalCache(a, w, r) })
}
