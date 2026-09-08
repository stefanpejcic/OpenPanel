package account

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// handleAPISwagger renders the Swagger UI page, pre-authorized with a freshly minted Bearer token so "Try it out" works immediately without another login
func handleAPISwagger(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	token, err := mintAPIToken(a, userID)
	if err != nil {
		token = ""
	}
	renderAPISwaggerPage(a, w, r, token)
}

// RegisterAPIDocs wires the /account/api route onto mux, gated behind the "api" feature flag, to the interactive Swagger UI
func RegisterAPIDocs(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "api")(h)
	}
	mux.Handle("GET /account/api", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAPISwagger(a, w, r) }))
}
