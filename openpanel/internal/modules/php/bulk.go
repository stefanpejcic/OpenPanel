package php

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// VersionOptions turns installed versions into the bulk bar's select options
func VersionOptions(installed []string) []web.BulkOption {
	options := make([]web.BulkOption, 0, len(installed))
	for _, v := range installed {
		options = append(options, web.BulkOption{Value: v, Label: "PHP " + v})
	}
	return options
}

// InstalledVersionsSorted is FetchPHPVersions newest first, copied so the cached slice isn't reordered
func InstalledVersionsSorted(r *http.Request, a *appctx.App, userContext string) []string {
	installed := append([]string(nil), FetchPHPVersions(r.Context(), a, userContext)...)
	sortVersionsDesc(installed)
	return installed
}

func phpDomainsBulkActions(t i18n.Translator, installed []string) []web.BulkAction {
	return []web.BulkAction{
		{Key: "version", Label: t.Get("Change version"), Confirm: t.Get("Switch the selected domains to:"), Input: &web.BulkInput{Type: "select", Options: VersionOptions(installed)}},
		{Key: "default", Label: t.Get("Reset to default"), Confirm: t.Get("Switch the selected domains to the default PHP version?")},
	}
}

// VersionSwitcher reads every domain's current PHP version once, so bulk switches from /php/domains and /domains send the right old_php_version
type VersionSwitcher struct {
	Default string
	current map[string]string
	a       *appctx.App
	r       *http.Request
}

func NewVersionSwitcher(a *appctx.App, r *http.Request, userContext string) *VersionSwitcher {
	userID, _ := auth.UserID(r)
	domainsList, _ := a.AllDomainsForUser(r.Context(), userID)
	rows, _, _ := buildPHPDomainRows(userContext, domainsList, nil)
	current := make(map[string]string, len(rows))
	for _, row := range rows {
		current[row.DomainURL] = row.PHPVersion
	}
	return &VersionSwitcher{Default: webserver.GetEnvFileValue(userContext, "DEFAULT_PHP_VERSION"), current: current, a: a, r: r}
}

// Route switches domain to newVersion through the per-row POST /php/domains
func (v *VersionSwitcher) Route(domain, newVersion string) (*web.BulkCall, web.BulkResult) {
	old, ok := v.current[domain]
	switch {
	case !ok:
		return web.Skip("Domain not found.")
	case old == "/":
		return web.Skip("This domain doesn't use PHP.")
	case newVersion == old:
		return nil, web.BulkResult{OK: true, Message: web.Tr(v.a, v.r, "Already on PHP %(version)s.", "version", old)}
	}
	return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/php/domains", Form: url.Values{
		"domain_url": {domain}, "old_php_version": {old}, "new_php_version": {newVersion}, "redirect_to": {"/php/domains"},
	}})
}

func handlePHPDomainsBulk(a *appctx.App, mux http.Handler, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	switcher := NewVersionSwitcher(a, r, userContext)
	actions := phpDomainsBulkActions(web.RequestTranslator(a, r), InstalledVersionsSorted(r, a, userContext))
	web.ServeBulkDispatch(a, mux, w, r, actions, func(action, value, domain string) (*web.BulkCall, web.BulkResult) {
		if action == "default" {
			value = switcher.Default
		}
		return switcher.Route(domain, value)
	})
}
