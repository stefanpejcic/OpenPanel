package serverinfo

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// RegisterInfo wires the server info page route onto mux.
func RegisterInfo(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "info")(h)
	}
	mux.Handle("GET /server/info", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleServerInfo(a, w, r) }))
}

// RegisterHostingJSON wires the /json/system/hosting/* routes onto mux.
// Unlike RegisterInfo, these are always registered rather than gated by
// enabled_modules, so callers should register this alongside the
// always-on modules, not behind the "info" feature flag.
func RegisterHostingJSON(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "info")(h)
	}
	mux.Handle("GET /json/system/hosting/info", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleSystemHostingInfo(a, w, r) }))
	mux.Handle("GET /json/system/hosting/plan", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleSystemHostingPlan(a, w, r) }))
	mux.Handle("GET /json/system/hosting/ports", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleSystemHostingPorts(a, w, r) }))
}

// RegisterUsage wires the resource-usage page routes onto mux.
func RegisterUsage(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "usage")(h)
	}
	mux.Handle("GET /server/usage", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleContainerUsage(a, w, r) }))
	mux.Handle("GET /server/usage/history", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleUsageHistory(a, w, r) }))
}
