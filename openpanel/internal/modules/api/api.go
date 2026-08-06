// Package api ports modules/api_core.py's introspection endpoint. The
// actual per-feature endpoints (modules/api/*.py) live inside their
// existing feature packages (mysql, postgresql, domains, ...) as
// apiXxx.go files, each registering through
// internal/core/apiregistry.Handle - which is also what this package's
// /api/endpoints reads back from.
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
