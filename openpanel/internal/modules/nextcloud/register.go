package nextcloud

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the Nextcloud install/remove/manage routes onto mux,
// gated behind the "nextcloud" feature flag. No list page (matches
// drupal/joomla/opencart's scope: manage via the general Site Manager
// instead).
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "nextcloud")(h)
	}
	mux.Handle("/nextcloud/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /nextcloud/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveNextcloud(a, w, r) }))
	mux.Handle("GET /nextcloud/login", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleNextcloudLogin(a, w, r) }))
	mux.Handle("POST /nextcloud/cache", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleNextcloudCacheClean(a, w, r) }))
	mux.Handle("GET /nextcloud/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleNextcloudLogs(a, w, r) }))
	mux.Handle("GET /nextcloud/versions", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleNextcloudVersions(a, w, r) }))
	mux.Handle("/nextcloud/maintenance", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleNextcloudMaintenance(a, w, r) }))
	mux.Handle("GET /nextcloud/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleNextcloudGetBackupDates(a, w, r) }))
	mux.Handle("GET /nextcloud/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleNextcloudRestoreBackup(a, w, r) }))
	mux.Handle("GET /nextcloud/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleNextcloudRunBackup(a, w, r) }))
	mux.Handle("POST /nextcloud/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleNextcloudClone(a, w, r) }))
	mux.Handle("POST /nextcloud/update", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleNextcloudUpdate(a, w, r) }))
}

// withNextcloudForm clones r as a POST carrying the given values as both
// Form and PostForm, so a UI handler that reads r.FormValue(...) sees
// exactly the fields the API's JSON body supplied - same pattern used by
// opencart/register.go's withOpenCartForm.
func withNextcloudForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/nextcloud/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "nextcloud", "POST /api/nextcloud/install", func(w http.ResponseWriter, r *http.Request) { apiInstallNextcloud(a, w, r) })
	apiregistry.Handle(mux, a, "nextcloud", "DELETE /api/nextcloud/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveNextcloud(a, w, r) })
	apiregistry.Handle(mux, a, "nextcloud", "POST /api/nextcloud/sites/{site_id}/clone", func(w http.ResponseWriter, r *http.Request) { apiNextcloudClone(a, w, r) })
	apiregistry.Handle(mux, a, "nextcloud", "POST /api/nextcloud/sites/{site_id}/update", func(w http.ResponseWriter, r *http.Request) { apiNextcloudUpdate(a, w, r) })
	apiregistry.Handle(mux, a, "nextcloud", "POST /api/nextcloud/sites/{site_id}/cache", func(w http.ResponseWriter, r *http.Request) { apiNextcloudCache(a, w, r) })
}
