package ipblocker

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires modules/ip_blocker.py's route onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "ip_blocker")(h)
	}
	mux.Handle("/security/ip-blocker", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleIPBlocker(a, w, r) }))
}
