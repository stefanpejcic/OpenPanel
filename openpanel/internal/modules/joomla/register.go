package joomla

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the Joomla install/remove/manage routes onto mux, gated
// behind the "joomla" feature flag. No list page (matches drupal's current
// scope: manage via the general Site Manager instead).
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "joomla")(h)
	}
	mux.Handle("/joomla/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /joomla/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveJoomla(a, w, r) }))
	mux.Handle("GET /joomla/login", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleJoomlaLogin(a, w, r) }))
	mux.Handle("POST /joomla/cache", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleJoomlaCacheClean(a, w, r) }))
	mux.Handle("GET /joomla/logs", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleJoomlaLogs(a, w, r) }))
	mux.Handle("/joomla/maintenance", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleJoomlaMaintenance(a, w, r) }))
	mux.Handle("GET /joomla/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleJoomlaGetBackupDates(a, w, r) }))
	mux.Handle("GET /joomla/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleJoomlaRestoreBackup(a, w, r) }))
	mux.Handle("GET /joomla/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleJoomlaRunBackup(a, w, r) }))
	mux.Handle("POST /joomla/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleJoomlaClone(a, w, r) }))
}

// withJoomlaForm clones r as a POST carrying the given values as both Form
// and PostForm, so a UI handler that reads r.FormValue(...) sees exactly
// the fields the API's JSON body supplied - same pattern used by
// drupal/register.go's withDrupalForm.
func withJoomlaForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/joomla/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "joomla", "POST /api/joomla/install", func(w http.ResponseWriter, r *http.Request) { apiInstallJoomla(a, w, r) })
	apiregistry.Handle(mux, a, "joomla", "DELETE /api/joomla/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveJoomla(a, w, r) })
	apiregistry.Handle(mux, a, "joomla", "POST /api/joomla/clone", func(w http.ResponseWriter, r *http.Request) { apiCloneJoomla(a, w, r) })
}
