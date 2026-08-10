package serverinfo

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterHostingAPI wires the /api/hosting/* JSON API routes onto mux.
// These mirror the /json/system/hosting/* handlers registered by
// RegisterHostingJSON exactly - same handlers, same "info" feature gate -
// just exposed under the API's flat /api/hosting/... naming convention.
func RegisterHostingAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "info", "GET /api/hosting/info", func(w http.ResponseWriter, r *http.Request) { handleSystemHostingInfo(a, w, r) })
	apiregistry.Handle(mux, a, "info", "GET /api/hosting/plan", func(w http.ResponseWriter, r *http.Request) { handleSystemHostingPlan(a, w, r) })
	apiregistry.Handle(mux, a, "info", "GET /api/hosting/ports", func(w http.ResponseWriter, r *http.Request) { handleSystemHostingPorts(a, w, r) })
}
