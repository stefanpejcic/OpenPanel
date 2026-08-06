package account

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// handleAPIDocs renders a static reference page listing every REST
// endpoint, driven entirely client-side by the JSON in internal/core/apidocs.
func handleAPIDocs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	renderAPIDocsPage(a, w, r)
}

// RegisterAPIDocs wires the /account/api docs route onto mux, gated behind
// the "api" feature flag.
func RegisterAPIDocs(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "api")(h)
	}
	mux.Handle("GET /account/api", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAPIDocs(a, w, r) }))
}
