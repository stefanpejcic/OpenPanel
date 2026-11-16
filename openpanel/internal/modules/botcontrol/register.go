package botcontrol

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register registers the AI & Bot Control module's routes onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "bot_control")(h)
	}

	mux.Handle("/server/bot-control", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBotControlList(a, w, r) }))
	mux.Handle("/server/bot-control/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBotControlDomain(a, w, r) }))
}
