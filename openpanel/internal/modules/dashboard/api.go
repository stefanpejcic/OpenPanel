package dashboard

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterAPI wires /api/ equivalents of the two dashboard JSON widgets onto mux, reusing handleResourceUsage/handleDiskInodes as-is since both are already pure JSON handlers. Always available, matching Register's unconditional wiring (dashboard is always in EnabledModules).
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "dashboard", "GET /api/dashboard/resource-usage", func(w http.ResponseWriter, r *http.Request) { handleResourceUsage(a, w, r) })
	apiregistry.Handle(mux, a, "dashboard", "GET /api/dashboard/disk-inodes", func(w http.ResponseWriter, r *http.Request) { handleDiskInodes(a, w, r) })
	apiregistry.Handle(mux, a, "dashboard", "POST /api/dashboard/tour/complete", func(w http.ResponseWriter, r *http.Request) { handleTourComplete(a, w, r) })
}
