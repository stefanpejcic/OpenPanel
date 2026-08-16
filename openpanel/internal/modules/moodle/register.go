package moodle

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the Moodle install/remove/manage routes onto mux, gated
// behind the "moodle" feature flag. No list page (matches
// drupal/joomla/opencart/nextcloud/prestashop's scope: manage via the
// general Site Manager instead).
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "moodle")(h)
	}
	mux.Handle("/moodle/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /moodle/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveMoodle(a, w, r) }))
	mux.Handle("POST /moodle/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMoodleClone(a, w, r) }))
	mux.Handle("POST /moodle/cache", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMoodleCacheClean(a, w, r) }))
	mux.Handle("GET /moodle/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMoodleLogs(a, w, r) }))
	mux.Handle("GET /moodle/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMoodleGetBackupDates(a, w, r) }))
	mux.Handle("GET /moodle/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMoodleRestoreBackup(a, w, r) }))
	mux.Handle("GET /moodle/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMoodleRunBackup(a, w, r) }))
}

// withMoodleForm clones r as a POST carrying the given values as both Form
// and PostForm, so a UI handler that reads r.FormValue(...) sees exactly
// the fields the API's JSON body supplied - same pattern used by every
// other CMS module's with{CMS}Form.
func withMoodleForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/moodle/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "moodle", "POST /api/moodle/install", func(w http.ResponseWriter, r *http.Request) { apiInstallMoodle(a, w, r) })
	apiregistry.Handle(mux, a, "moodle", "DELETE /api/moodle/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveMoodle(a, w, r) })
}
