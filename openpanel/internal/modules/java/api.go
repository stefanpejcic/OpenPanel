package java

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/appinstall"
)

// RegisterAPI wires POST /api/java/install onto mux, gated behind the
// "java" feature flag - same shape as nodejs/python/ruby's RegisterAPI.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "java", "POST /api/java/install", func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstall(appinstall.Java, a, w, r)
	})
}
