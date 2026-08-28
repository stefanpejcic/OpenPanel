// Package java installs a Java application into a domain via a
// docker-compose service + reverse proxy. All the actual logic lives in
// internal/modules/appinstall, shared with the nodejs, python, and ruby
// packages.
package java

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/appinstall"
)

// Register wires the /java/install route onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "java")(h)
	}
	mux.Handle("GET /java/install", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstallPage(appinstall.Java, a, w, r)
	}))
	mux.Handle("POST /java/install", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstallPage(appinstall.Java, a, w, r)
	}))
}
