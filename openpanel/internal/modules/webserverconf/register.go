package webserverconf

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires the webserver config editor route onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "webserver_conf")(h)
	}
	mux.Handle(redirectPath, requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWebserverConf(a, w, r) }))
}
