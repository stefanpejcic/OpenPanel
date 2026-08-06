// Package plugins implements the `/plugins` route: the JSON listing
// base.html's client-side JS fetches to inject plugin entries into the
// sidebar and dashboard. See internal/core/plugins for the underlying
// plugin system.
package plugins

import (
	"encoding/json"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	corePlugins "gist.github.com/stefanpejcic/openpanel/internal/core/plugins"
)

// Register wires the /plugins route onto mux - the route only exists at
// all when at least one plugin was found at startup.
func Register(mux *http.ServeMux, a *appctx.App) {
	if len(a.PluginNames) == 0 {
		return
	}
	// The feature name for a route that isn't tied to a specific module is
	// "app" - always granted, see baselineFeatures.
	requireLogin := auth.RequireLogin(a, "app")
	mux.Handle("GET /plugins", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlePluginsPage(a, w, r)
	})))
}

// handlePluginsPage writes the discovered plugin list as JSON.
func handlePluginsPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	list := corePlugins.List(corePlugins.BaseDir)
	if list == nil {
		list = []corePlugins.Plugin{}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}
