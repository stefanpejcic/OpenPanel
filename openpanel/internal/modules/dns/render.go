package dns

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
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

func loadPage(files ...string) *web.Page {
	return web.MustLoadPage(append(append([]string{}, pageFiles...), files...)...)
}

var dnsPage = loadPage("domains/dns.html")

// DNSPageData is domains/dns.html's template context, covering all three
// states the page can render: the domain-list landing page (Domain == ""),
// the table view (ViewMode == "table"), and the raw code view
// (ViewMode == "code").
type DNSPageData struct {
	web.LayoutData
	Domain       string
	ViewMode     string
	Domains      []appctx.Domain // zone-having domains only, for the selector
	Rows         []ZoneRow
	Serial       string
	ZoneContent  string
	Issues       []HealthIssue
	TotalRecords int
}

func renderDNSListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainRows []DomainZoneRow) {
	layout, _, err := web.BuildLayoutData(a, w, r, "DNS Zone Editor")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var withZone []appctx.Domain
	for _, d := range domainRows {
		if d.ZoneFileExists {
			withZone = append(withZone, appctx.Domain{DomainURL: d.DomainURL})
		}
	}
	data := DNSPageData{LayoutData: layout, Domains: withZone}
	if err := dnsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DNS - list template render error: %v", err)
	}
}

func renderDNSTablePage(a *appctx.App, w http.ResponseWriter, r *http.Request, domain string, rows []ZoneRow, serial string, issues []HealthIssue) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit DNS Zone")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DNSPageData{
		LayoutData: layout, Domain: domain, ViewMode: "table", Rows: rows,
		Serial: serial, Issues: issues, TotalRecords: len(rows),
	}
	if err := dnsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DNS - table template render error: %v", err)
	}
}

func renderDNSCodePage(a *appctx.App, w http.ResponseWriter, r *http.Request, domain, zoneContent string, issues []HealthIssue) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit DNS File")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DNSPageData{LayoutData: layout, Domain: domain, ViewMode: "code", ZoneContent: zoneContent, Issues: issues}
	if err := dnsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DNS - code template render error: %v", err)
	}
}
