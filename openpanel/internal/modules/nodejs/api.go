package nodejs

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/appinstall"
)

// RegisterAPI wires POST /api/nodejs/install onto mux, gated behind the "nodejs" feature flag - reuses appinstall.HandleInstall as-is since it's already pure form-in/NDJSON-out with no session or flash usage
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "nodejs", "POST /api/nodejs/install", func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstall(appinstall.NodeJS, a, w, r)
	})
}
