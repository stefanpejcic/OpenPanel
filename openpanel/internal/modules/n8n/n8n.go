// Package n8n installs an n8n workflow-automation instance into a domain via a docker-compose service + reverse proxy - the actual logic lives in internal/modules/appinstall, shared with the nodejs/python/ruby/java packages
package n8n

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/appinstall"
)

// Register wires the /n8n/install route onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "n8n")(h)
	}
	mux.Handle("GET /n8n/install", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstallPage(appinstall.N8N, a, w, r)
	}))
	mux.Handle("POST /n8n/install", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstallPage(appinstall.N8N, a, w, r)
	}))
}
