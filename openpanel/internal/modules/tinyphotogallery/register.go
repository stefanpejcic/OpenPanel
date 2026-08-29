package tinyphotogallery

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires the TinyPhotoGallery install/remove routes onto mux,
// gated behind the "tinyphotogallery" feature flag. No maintenance,
// login, cache, backup, or clone routes - see tinyphotogallery.go's
// package doc comment for why those aren't implemented.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "tinyphotogallery")(h)
	}
	mux.Handle("/tinyphotogallery/install", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleInstallPage(a, w, r) }))
	mux.Handle("POST /tinyphotogallery/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveTinyPhotoGallery(a, w, r) }))
}

// withTinyPhotoGalleryForm clones r as a POST carrying the given values as
// both Form and PostForm, so a UI handler that reads r.FormValue(...) sees
// exactly the fields the API's JSON body supplied.
func withTinyPhotoGalleryForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

// RegisterAPI wires the /api/tinyphotogallery/* routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "tinyphotogallery", "POST /api/tinyphotogallery/install", func(w http.ResponseWriter, r *http.Request) { apiInstallTinyPhotoGallery(a, w, r) })
	apiregistry.Handle(mux, a, "tinyphotogallery", "DELETE /api/tinyphotogallery/sites/{site_id}", func(w http.ResponseWriter, r *http.Request) { apiRemoveTinyPhotoGallery(a, w, r) })
}
