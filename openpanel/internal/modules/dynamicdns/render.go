package dynamicdns

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var dynamicDNSPage = web.LoadPage("domains/_shared.html", "domains/dynamic_dns.html")

// DomainEntries is one domain's group of dynamic DNS entries, kept as a slice in userDomains' insertion order rather than a map since Go maps don't preserve order
type DomainEntries struct {
	DomainName string
	Entries    []DynDNSEntry
}

// DynamicDNSPageData is domains/dynamic_dns.html's template context.
type DynamicDNSPageData struct {
	web.LayoutData
	DomainEntries []DomainEntries
	Domains       []appctx.Domain
	// AllEntries is DomainEntries flattened in the same order, the JS array openEdit(idx)/openDelete(idx) index into by Index
	AllEntries []DynDNSEntry
	// BaseURL is the reconstructed scheme://host:port/, used to build the absolute webcall update URL shown for each entry
	BaseURL string
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
		BaseURL: web.PublicBaseURL(r.Context(), a, r),
	}
	if err := dynamicDNSPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DYNAMICDNS - template render error: %v", err)
	}
}
