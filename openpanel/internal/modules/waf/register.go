package waf

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register registers the WAF module's routes onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "waf")(h)
	}

	mux.Handle("GET /json/waf/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFJSONForDomain(a, w, r) }))
	mux.Handle("/server/waf/type/{type}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFType(a, w, r) }))
	mux.Handle("/server/waf/log", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFLog(a, w, r) }))
	// "/server/waf/log/{domain_name...}" also covers the bare
	// "/server/waf/log/" case (domain_name resolves to ""), so a single
	// route handles both the domain-scoped and the all-domains log view.
	mux.Handle("/server/waf/log/{domain_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFLog(a, w, r) }))
	mux.Handle("/server/waf/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFDomain(a, w, r) }))
	mux.Handle("/server/waf", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFList(a, w, r) }))
}
