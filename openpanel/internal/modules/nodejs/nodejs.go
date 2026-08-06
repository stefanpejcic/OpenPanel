// Package nodejs installs a NodeJS application into a domain via a
// docker-compose service + reverse proxy. All the actual logic lives in
// internal/modules/appinstall, shared with the python package.
package nodejs

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/appinstall"
)

// Register wires the /nodejs/install route onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "nodejs")(h)
	}
	mux.Handle("GET /nodejs/install", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstallPage(appinstall.NodeJS, a, w, r)
	}))
	mux.Handle("POST /nodejs/install", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstallPage(appinstall.NodeJS, a, w, r)
	}))
}
