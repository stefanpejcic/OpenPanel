package domains

import (
	"net/http"
	"net/url"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// domainsBulkActions lists every /domains bulk action, AllowedBulkActions then drops the ones the user's plan doesn't have
func domainsBulkActions(t i18n.Translator, phpVersions []string) []web.BulkAction {
	return []web.BulkAction{
		{Key: "suspend", Label: t.Get("Suspend"), Confirm: t.Get("Suspend the selected domains? Visitors will see a suspended page."), Feature: "domain_suspend"},
		{Key: "unsuspend", Label: t.Get("Unsuspend"), Confirm: t.Get("Unsuspend the selected domains?"), Feature: "domain_suspend"},
		{Key: "php", Label: t.Get("Change PHP version"), Confirm: t.Get("Switch the selected domains to:"), Feature: "php",
			Input: &web.BulkInput{Type: "select", Options: php.VersionOptions(phpVersions)}},
		{Key: "redirect", Label: t.Get("Redirect to"), Confirm: t.Get("Redirect the selected domains to:"), Feature: "redirects",
			Input: &web.BulkInput{Type: "url", Placeholder: "https://example.com"}},
		{Key: "unredirect", Label: t.Get("Remove redirect"), Confirm: t.Get("Remove the redirect from the selected domains?"), Feature: "redirects"},
		{Key: "waf", Label: t.Get("WAF"), Confirm: t.Get("Set the firewall for the selected domains to:"), Feature: "waf",
			Input: &web.BulkInput{Type: "select", Options: []web.BulkOption{{Value: "On", Label: t.Get("On")}, {Value: "Off", Label: t.Get("Off")}}}},
		{Key: "cloudflare", Label: t.Get("Cloudflare"), Confirm: t.Get("Cloudflare-only access for the selected domains, restricted domains not proxied through Cloudflare return 403 to visitors:"),
			Input: &web.BulkInput{Type: "select", Options: []web.BulkOption{{Value: "enable", Label: t.Get("Restrict to Cloudflare")}, {Value: "disable", Label: t.Get("Unrestrict")}}}},
		{Key: "delete", Label: t.Get("Delete"), Confirm: t.Get("Permanently delete the selected domains, including their websites, files and DNS zones? This cannot be undone."), Danger: true},
	}
}

func handleDomainsBulk(a *appctx.App, mux http.Handler, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	_, userContext, err := injected(a, r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	actions := web.AllowedBulkActions(a, r, domainsBulkActions(web.RequestTranslator(a, r), php.InstalledVersionsSorted(r, a, userContext)))

	var switcher *php.VersionSwitcher
	web.ServeBulkDispatch(a, mux, w, r, actions, func(action, value, domain string) (*web.BulkCall, web.BulkResult) {
		form := url.Values{"domain_name": {domain}}
		switch action {
		case "suspend":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/domains/suspend", Form: form})
		case "unsuspend":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/domains/unsuspend", Form: form})
		case "redirect":
			form.Set("redirect_url", value)
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/domains/redirect", Form: form})
		case "unredirect":
			// the delete route removes exactly this redir line, like the row's hidden field
			current := getRedirectURL(domain)
			if current == "" {
				return web.Skip("No redirect set.")
			}
			form.Set("redirect_url", current)
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/domains/redirect/delete", Form: form})
		case "waf":
			form.Set("modsec_action", value)
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/server/waf", Form: form})
		case "cloudflare":
			if strings.HasSuffix(domain, ".onion") {
				return web.Skip("Not available for .onion domains.")
			}
			form.Set("cloudflare_action", value)
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/domains/cloudflare", Form: form})
		case "php":
			if switcher == nil {
				switcher = php.NewVersionSwitcher(a, r, userContext)
			}
			return switcher.Route(domain, value)
		case "delete":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/domains/delete", Form: url.Values{"domain_url": {domain}}})
		}
		return web.Skip("Unknown bulk action.")
	})
}
