package goaccess

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires the GoAccess stats routes onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "goaccess")(h)
	}

	mux.Handle("GET /domains/stats", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDomainStats(a, w, r) }))
	mux.Handle("GET /domains/stats/", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDomainStats(a, w, r) }))
	mux.Handle("GET /domains/stats/{domain_name}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDomainStats(a, w, r) }))
}
