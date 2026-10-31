package dns

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
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
	Rows         []ZoneRow
	Serial       string
	ZoneContent  string
	Issues       []HealthIssue
	TotalRecords int
	Zones        []ZoneSummary // list page: every zone with its record counts
	ZoneRecords  int           // list page: records across all zones
	Nameservers  []string      // list page: the server's ns1-ns4, for the footer
}

// ZoneSummary is one row of the zone list: the domain, how many records its zone holds and how many of each type
type ZoneSummary struct {
	Domain  string
	Records int
	Types   []TypeCount
}

// TypeCount is how many records of one type a zone holds
type TypeCount struct {
	Type  string
	Count int
}

// summarizeZone counts a zone file's records by type, in the order the types first appear
func summarizeZone(domain, content string) ZoneSummary {
	rows := buildZoneRows(parseZoneWithLineNumbers(content))
	z := ZoneSummary{Domain: domain, Records: len(rows)}
	idx := map[string]int{}
	for _, r := range rows {
		if i, ok := idx[r.Type]; ok {
			z.Types[i].Count++
			continue
		}
		idx[r.Type] = len(z.Types)
		z.Types = append(z.Types, TypeCount{Type: r.Type, Count: 1})
	}
	return z
}

func renderDNSListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainRows []DomainZoneRow) {
	layout, _, err := web.BuildLayoutData(a, w, r, "DNS Zone Editor")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var withZone []appctx.Domain
	var zones []ZoneSummary
	total := 0
	for _, d := range domainRows {
		if d.ZoneFileExists {
			withZone = append(withZone, appctx.Domain{DomainURL: d.DomainURL})
			content, _ := tenantfs.ReadFile(zoneFilePath(d.DomainURL))
			z := summarizeZone(d.DomainURL, string(content))
			total += z.Records
			zones = append(zones, z)
		}
	}
	data := DNSPageData{LayoutData: layout, Domains: withZone, Zones: zones, ZoneRecords: total}
	for _, k := range []string{"ns1", "ns2", "ns3", "ns4"} {
		if v := a.Config.Get(k, ""); v != "" {
			data.Nameservers = append(data.Nameservers, v)
		}
	}
	if err := dnsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DNS - list template render error: %v", err)
	}
}

func renderDNSTablePage(a *appctx.App, w http.ResponseWriter, r *http.Request, domain string, rows []ZoneRow, serial string, issues []HealthIssue, zoneDomains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit DNS Zone")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.BulkActions = dnsBulkActions(layout.T)
	data := DNSPageData{
		LayoutData: layout, Domain: domain, ViewMode: "table", Rows: rows,
		Serial: serial, Issues: issues, TotalRecords: len(rows), Domains: zoneDomains,
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
