package waf

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// Register registers the WAF module's routes onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "waf")(h)
	}

	mux.Handle("GET /json/waf/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFJSONForDomain(a, w, r) }))
	mux.Handle("/server/waf/type/{type}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFType(a, w, r) }))
	mux.Handle("/server/waf/log", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFLog(a, w, r) }))
	// "/server/waf/log/{domain_name...}" also covers the bare "/server/waf/log/" case (domain_name resolves to ""), so a single route handles both the domain-scoped and the all-domains log view
	mux.Handle("/server/waf/log/{domain_name...}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFLog(a, w, r) }))
	mux.Handle("/server/waf/{domain}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFDomain(a, w, r) }))
	mux.Handle("/server/waf", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWAFList(a, w, r) }))
	mux.Handle("POST /server/waf/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		web.ServeBulkDispatch(a, mux, w, r, wafBulkActions(web.RequestTranslator(a, r)), func(action, _, domain string) (*web.BulkCall, web.BulkResult) {
			status := map[string]string{"enable": "On", "disable": "Off"}[action]
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/server/waf", Form: url.Values{"domain_name": {domain}, "modsec_action": {status}}})
		})
	}))
}

func wafBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "enable", Label: t.Get("Enable WAF"), Confirm: t.Get("Enable the firewall for the selected domains?")},
		{Key: "disable", Label: t.Get("Disable WAF"), Confirm: t.Get("Disable the firewall for the selected domains?"), Danger: true},
	}
}
