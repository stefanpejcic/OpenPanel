package plugins

import (
	"encoding/json"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	corePlugins "gist.github.com/stefanpejcic/openpanel/internal/core/plugins"
)

// RegisterAPI wires GET /api/plugins onto mux - only when at least one
// plugin was found at startup, matching Register's own guard. Gated on
// the "app" feature, same as the web route (always granted, see
// baselineFeatures).
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	if len(a.PluginNames) == 0 {
		return
	}
	apiregistry.Handle(mux, a, "app", "GET /api/plugins", func(w http.ResponseWriter, r *http.Request) {
		list := corePlugins.List(corePlugins.BaseDir)
		if list == nil {
			list = []corePlugins.Plugin{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"plugins": list})
	})
}
