package search

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterAPI wires the /api/search/{what} route onto mux - reuses HandleSearch as-is since it already reads the user via auth.UserID, which RequireAPI populates the same way RequireLogin does, gated per-what internally via gateFor
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "search", "GET /api/search/{what}", func(w http.ResponseWriter, r *http.Request) {
		HandleSearch(a, w, r)
	})
}
