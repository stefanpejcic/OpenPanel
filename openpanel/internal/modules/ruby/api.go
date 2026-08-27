package ruby

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/appinstall"
)

// RegisterAPI wires POST /api/ruby/install onto mux, gated behind the
// "ruby" feature flag - same shape as nodejs/python's RegisterAPI.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "ruby", "POST /api/ruby/install", func(w http.ResponseWriter, r *http.Request) {
		appinstall.HandleInstall(appinstall.Ruby, a, w, r)
	})
}
