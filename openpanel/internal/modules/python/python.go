// Package python installs a Python application into a domain via a
// docker-compose service + reverse proxy. All the actual logic lives in
// internal/modules/appinstall, shared with the nodejs package.
package python

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/appinstall"
)

// Register wires the /python/install route onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "python")(h)
	}
	mux.Handle("GET /python/install", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstallPage(appinstall.Python, a, w, r)
	}))
	mux.Handle("POST /python/install", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstallPage(appinstall.Python, a, w, r)
	}))
}
