package search

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires the /json/search/{what} route onto mux, gated by the
// "search" feature - unconditionally granted to every user (see
// baselineFeatures). Each `what` sub-type then has its own finer-grained
// permission gate inside HandleSearch (see gateFor).
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := auth.RequireLogin(a, "search")
	mux.Handle("GET /json/search/{what}", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleSearch(a, w, r)
	})))
}
