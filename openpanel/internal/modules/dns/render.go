package dns

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func loadPage(files ...string) *web.Page {
	return web.LoadPage(append([]string{"domains/_shared.html"}, files...)...)
}

var dnsPage = loadPage("domains/dns.html")

// DNSPageData is domains/dns.html's template context, covering all three states the page can render: the domain-list landing page (Domain == ""), the table view (ViewMode == "table"), and the raw code view (ViewMode == "code")
type DNSPageData struct {
	web.LayoutData
	Domain       string
	ViewMode     string
	Domains      []appctx.Domain // zone-having domains only, for the selector
	Picks        []web.DomainPick
	Rows         []ZoneRow
	Serial       string
	ZoneContent  string
	Issues       []HealthIssue
	TotalRecords int
}

func renderDNSListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainRows []DomainZoneRow) {
	layout, _, err := web.BuildLayoutData(a, w, r, "DNS zone editor")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var withZone []appctx.Domain
	var picks []web.DomainPick
	for _, d := range domainRows {
		if d.ZoneFileExists {
			withZone = append(withZone, appctx.Domain{DomainURL: d.DomainURL})
			picks = append(picks, web.DomainPick{Domain: d.DomainURL, Detail: zoneRecordCount(d.DomainURL)})
		} else {
			note := "No zone file"
			if hasSubdomainLabel(d.DomainURL) {
				note = "Uses the main domain's zone"
			}
			picks = append(picks, web.DomainPick{Domain: d.DomainURL, Disabled: true, Note: note})
		}
	}
	data := DNSPageData{LayoutData: layout, Domains: withZone, Picks: picks}
	if err := dnsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DNS - list template render error: %v", err)
	}
}

func renderDNSTablePage(a *appctx.App, w http.ResponseWriter, r *http.Request, domain string, rows []ZoneRow, serial string, issues []HealthIssue) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit DNS zone")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.BulkActions = dnsBulkActions(layout.T)
	data := DNSPageData{
		LayoutData: layout, Domain: domain, ViewMode: "table", Rows: rows,
		Serial: serial, Issues: issues, TotalRecords: len(rows),
	}
	if err := dnsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DNS - table template render error: %v", err)
	}
}

func renderDNSCodePage(a *appctx.App, w http.ResponseWriter, r *http.Request, domain, zoneContent string, issues []HealthIssue) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit DNS file")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DNSPageData{LayoutData: layout, Domain: domain, ViewMode: "code", ZoneContent: zoneContent, Issues: issues}
	if err := dnsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DNS - code template render error: %v", err)
	}
}
