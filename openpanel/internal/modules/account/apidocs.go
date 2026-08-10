package account

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// handleAPIReference renders a static reference page listing every REST
// endpoint, driven entirely client-side by the JSON in internal/core/apidocs.
func handleAPIReference(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	renderAPIDocsPage(a, w, r)
}

// handleAPISwagger renders the interactive Swagger UI page, pre-authorized
// with a freshly minted Bearer token for the logged-in user - so "Try it
// out" works immediately against this same origin without another login.
func handleAPISwagger(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	token, err := mintAPIToken(a, userID)
	if err != nil {
		token = ""
	}
	renderAPISwaggerPage(a, w, r, token)
}

// RegisterAPIDocs wires the /account/api routes onto mux, gated behind the
// "api" feature flag: /account/api is the interactive Swagger UI,
// /account/api/reference the plain searchable list it replaced.
func RegisterAPIDocs(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "api")(h)
	}
	mux.Handle("GET /account/api", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAPISwagger(a, w, r) }))
	mux.Handle("GET /account/api/reference", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAPIReference(a, w, r) }))
}
