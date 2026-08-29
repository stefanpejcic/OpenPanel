package dokuwiki

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the DokuWiki install/remove/manage/update routes onto
// mux, gated behind the "dokuwiki" feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "dokuwiki")(h)
	}
	mux.Handle("/dokuwiki/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /dokuwiki/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveDokuwiki(a, w, r) }))
	mux.Handle("POST /dokuwiki/update", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDokuwikiUpdate(a, w, r) }))
	mux.Handle("GET /dokuwiki/backup/get_dates/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDokuwikiGetBackupDates(a, w, r) }))
	mux.Handle("GET /dokuwiki/backup/restore/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDokuwikiRestoreBackup(a, w, r) }))
	mux.Handle("GET /dokuwiki/backup/run/{selected_domain...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDokuwikiRunBackup(a, w, r) }))
	mux.Handle("POST /dokuwiki/clone", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDokuwikiClone(a, w, r) }))
}

// withDokuwikiForm clones r as a POST carrying the given values as both
// Form and PostForm, so a UI handler that reads r.FormValue(...) sees
// exactly the fields the API's JSON body supplied.
func withDokuwikiForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/dokuwiki/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "dokuwiki", "POST /api/dokuwiki/install", func(w http.ResponseWriter, r *http.Request) { apiInstallDokuwiki(a, w, r) })
	apiregistry.Handle(mux, a, "dokuwiki", "DELETE /api/dokuwiki/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveDokuwiki(a, w, r) })
	apiregistry.Handle(mux, a, "dokuwiki", "POST /api/dokuwiki/sites/{site_id}/clone", func(w http.ResponseWriter, r *http.Request) { apiCloneDokuwiki(a, w, r) })
	apiregistry.Handle(mux, a, "dokuwiki", "POST /api/dokuwiki/sites/{site_id}/update", func(w http.ResponseWriter, r *http.Request) { apiUpdateDokuwiki(a, w, r) })
}
