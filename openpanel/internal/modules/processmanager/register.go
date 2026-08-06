package processmanager

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires the process manager page route onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "process_manager")(h)
	}
	mux.Handle("/process-manager", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleProcessManager(a, w, r) }))
}
