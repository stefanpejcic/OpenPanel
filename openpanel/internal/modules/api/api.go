// Package api provides the /api/endpoints introspection endpoint - the actual per-feature endpoints live in their own packages as apiXxx.go files, registered through apiregistry.Handle, which this package reads back from.
package api

import (
	"encoding/json"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// Register wires /api/endpoints onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "api", "GET /api/endpoints", func(w http.ResponseWriter, r *http.Request) {
		handleAPIEndpoints(w, r)
	})
}

// handleAPIEndpoints mirrors api_endpoints().
func handleAPIEndpoints(w http.ResponseWriter, r *http.Request) {
	endpoints := apiregistry.All()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": endpoints, "total": len(endpoints)})
}
