package dynamicdns

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var pageFiles = []string{
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"domains/_shared.html",
}

var dynamicDNSPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "domains/dynamic_dns.html")...)

// DomainEntries is one domain's group of dynamic DNS entries, kept as a
// slice in userDomains' insertion order rather than a map, since Go maps
// don't preserve iteration order.
type DomainEntries struct {
	DomainName string
	Entries    []DynDNSEntry
}

// DynamicDNSPageData is domains/dynamic_dns.html's template context.
type DynamicDNSPageData struct {
	web.LayoutData
	DomainEntries []DomainEntries
	Domains       []appctx.Domain
	// AllEntries is DomainEntries flattened in the same order - the JS
	// array openEdit(idx)/openDelete(idx) index into by Index.
	AllEntries []DynDNSEntry
	// BaseURL is the reconstructed scheme://host:port/, used to build the
	// absolute webcall update URL shown for each entry.
	BaseURL string
}

// publicBaseURL reconstructs the scheme+host+port the same way
// enforceAccessDomain() (internal/auth/loaduser.go) does for the
// canonical-domain redirect - by the time an authenticated page like this
// one renders, the request's host has already passed that check.
func publicBaseURL(ctx context.Context, a *appctx.App, r *http.Request) string {
	host := strings.TrimSpace(a.ForceDomain)
	if host == "" {
		host = r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
	}
	scheme := "http"
	if sysinfo.HasSSL(ctx, a.Cache, host) {
		scheme = "https"
	}
	portSuffix := ""
	if a.ForcePort != "" {
		portSuffix = ":" + a.ForcePort
	}
	return scheme + "://" + host + portSuffix + "/"
}

func renderDynamicDNSPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainEntries []DomainEntries, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Dynamic DNS")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var allEntries []DynDNSEntry
	for _, group := range domainEntries {
		allEntries = append(allEntries, group.Entries...)
	}

	data := DynamicDNSPageData{
		LayoutData: layout, DomainEntries: domainEntries, Domains: domains, AllEntries: allEntries,
		BaseURL: publicBaseURL(r.Context(), a, r),
	}
	if err := dynamicDNSPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DYNAMICDNS - template render error: %v", err)
	}
}
