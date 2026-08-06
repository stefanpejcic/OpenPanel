package dynamicdns

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// Register wires the dynamic DNS routes onto mux. The webcall update
// endpoint is deliberately NOT behind auth.RequireLogin - it's a public,
// token-authenticated URL a router/IoT device calls.
func Register(mux *http.ServeMux, a *appctx.App) {
	mux.Handle("/domains/dynamic-dns", auth.RequireLogin(a, "dynamic_dns")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDynamicDNS(a, w, r)
	})))
	mux.HandleFunc("GET /dynamic-dns/update", handleDynamicDNSUpdate)
}
