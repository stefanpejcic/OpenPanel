package python

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/appinstall"
)

// RegisterAPI wires POST /api/python/install onto mux, gated behind the
// "python" feature flag. See nodejs.RegisterAPI's comment - same shared
// appinstall.HandleInstall, no API-specific variant needed.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "python", "POST /api/python/install", func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstall(appinstall.Python, a, w, r)
	})
}
