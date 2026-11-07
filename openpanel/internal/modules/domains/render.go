package domains

import (
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func loadPage(files ...string) *web.Page {
	return web.LoadPage(append([]string{"domains/_shared.html"}, files...)...)
}

var domainsPage = loadPage("domains/domains.html")
var newDomainPage = loadPage("domains/new.html")
var deleteDomainPage = loadPage("domains/delete.html")
var suspendPage = loadPage("domains/suspend.html")
var unsuspendPage = loadPage("domains/unsuspend.html")
var docrootPage = loadPage("domains/docroot.html")
var domainLogsPage = loadPage("domains/logs.html")
var capitalizePage = loadPage("domains/capitalize.html")
var vhostPage = loadPage("partials/_conf_editor.html", "domains/vhost.html")
var redirectPage = loadPage("domains/redirect.html")
var sslPage = loadPage("domains/ssl.html")

// DomainsPageData is domains.html's template context.
type DomainsPageData struct {
	web.LayoutData
	Domains         []DomainRow
	TotalPages      int
	CurrentPage     int
	StartLineNumber int
	EndLineNumber   int
	TotalDomains    int
	// MainDomains is what counts toward DomainsLimit (subdomains don't), a limit of 0 is unlimited
	MainDomains, DomainsLimit int
	PrevPage                  int
	NextPage                  int
	PageEntries               []PageEntry
	Filters                   []DomainFilter
}

// DomainsSummary feeds the tiles above the domains table, counted over the rows on this page
type DomainsSummary struct {
	Subdomains, SSLChecked, SSLValid, SSLExpiring, SSLBroken, WAFOn, Suspended int
	FirstSuspended                                                             string
	FirstSSLIssue                                                              string // the SSL note links here
}

func (d DomainsPageData) Summary() DomainsSummary {
	var s DomainsSummary
	for _, r := range d.Domains {
		if r.IsSubdomain {
			s.Subdomains++
		}
		if r.WAF == "On" || r.WAF == "DetectionOnly" {
			s.WAFOn++
		}
		if r.Status == "Suspended" {
			if s.Suspended == 0 {
				s.FirstSuspended = r.DomainURL
			}
			s.Suspended++
		}
		if strings.HasSuffix(r.DomainURL, ".onion") {
			continue
		}
		s.SSLChecked++
		switch {
		case r.HTTPS != "Automatic" && r.HTTPS != "Custom":
			s.SSLBroken++
		case r.SSL.Problem == "expiring":
			s.SSLExpiring++
		case r.SSL.Problem != "":
			s.SSLBroken++
		default:
			s.SSLValid++
			continue
		}
		if s.FirstSSLIssue == "" {
			s.FirstSSLIssue = r.DomainURL
		}
	}
	return s
}

// DomainFilter is one filter dropdown next to Show Columns on the domains list
type DomainFilter struct {
	Key     string
	Label   string
	Options []FilterOption
}

type FilterOption struct {
	Value string
	Label string
}

func buildDomainFilters(t i18n.Translator, allowed map[string]bool, rows []DomainRow) []DomainFilter {
	all := FilterOption{Value: "all", Label: t.Get("All")}
	filters := []DomainFilter{
		{Key: "status", Label: t.Get("Status"), Options: []FilterOption{all, {"active", t.Get("Active")}, {"suspended", t.Get("Suspended")}}},
	}

	php := DomainFilter{Key: "php", Label: t.Get("PHP Version"), Options: []FilterOption{all}}
	seen := map[string]bool{}
	for _, r := range rows {
		if r.PHPVersion != "" && !seen[r.PHPVersion] {
			seen[r.PHPVersion] = true
			php.Options = append(php.Options, FilterOption{r.PHPVersion, r.PHPVersion})
		}
	}
	sort.Slice(php.Options[1:], func(i, j int) bool { return php.Options[i+1].Value < php.Options[j+1].Value })
	filters = append(filters, php)

	if allowed["ssl"] {
		filters = append(filters, DomainFilter{Key: "ssl", Label: t.Get("SSL"), Options: []FilterOption{all, {"auto", t.Get("AutoSSL")}, {"custom", t.Get("Custom SSL")}}})
	}
	filters = append(filters, DomainFilter{Key: "type", Label: t.Get("Type"), Options: []FilterOption{all, {"domain", t.Get("Domain")}, {"subdomain", t.Get("Subdomain")}}})
	if allowed["redirects"] {
		filters = append(filters, DomainFilter{Key: "redirect", Label: t.Get("Redirect"), Options: []FilterOption{all, {"yes", t.Get("With redirect")}, {"no", t.Get("No redirect")}}})
	}
	return filters
}

// PageEntry is one rendered pagination control: either a page number link or an ellipsis
type PageEntry struct {
	Number     int
	IsEllipsis bool
}

// buildPageEntries mirrors domains.html's pagination loop: current page, first/last page, and current±2 render as links, page 2 and total_pages-1 render as an ellipsis unless they already qualify above, everything else renders nothing
func buildPageEntries(current, total int) []PageEntry {
	var entries []PageEntry
	for p := 1; p <= total; p++ {
		switch {
		case p == current:
			entries = append(entries, PageEntry{Number: p})
		case p == 1 || p == total || (p >= current-2 && p <= current+2):
			entries = append(entries, PageEntry{Number: p})
		case p == 2 || p == total-1:
			entries = append(entries, PageEntry{IsEllipsis: true})
		}
	}
	return entries
}

func renderDomainsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, rows []DomainRow, totalPages, currentPage, startLine, endLine, totalDomains, mainDomains int) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "Domains")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injectedData["context"].(string)
	layout.BulkActions = web.AllowedBulkActions(a, r, domainsBulkActions(layout.T, php.InstalledVersionsSorted(r, a, userContext)))
	data := DomainsPageData{
		LayoutData: layout, Domains: rows, TotalPages: totalPages, CurrentPage: currentPage,
		StartLineNumber: startLine, EndLineNumber: endLine, TotalDomains: totalDomains,
		PrevPage: currentPage - 1, NextPage: currentPage + 1,
		PageEntries: buildPageEntries(currentPage, totalPages),
		Filters:     buildDomainFilters(layout.T, layout.UserAllowed, rows),
		MainDomains: mainDomains,
	}
	if planID, _ := injectedData["hosting_plan"].(int); planID > 0 {
		if plan, err := a.QueryPlanDetailsByID(r.Context(), planID); err == nil {
			data.DomainsLimit = web.AtoiDefault(plan.DomainsLimit, 0)
		}
	}
	if err := domainsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - domains template render error: %v", err)
	}
}

// NewDomainPageData is new.html's template context (stateless besides layout, the form posts via fetch()/SSE, not a normal redirect-driven flow)
type NewDomainPageData struct {
	web.LayoutData
	Sites []appctx.Domain // existing domains, so a new one can share a site's folder
}

func renderNewDomainPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Add Domain")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userID, _ := auth.UserID(r)
	sites, _ := a.AllDomainsForUser(r.Context(), userID)
	if err := newDomainPage.Render(w, http.StatusOK, NewDomainPageData{LayoutData: layout, Sites: sites}); err != nil {
		log.Printf("DOMAINS - new domain template render error: %v", err)
	}
}

// DeleteDomainPageData is delete.html's template context.
type DeleteDomainPageData struct {
	web.LayoutData
	DomainURL      string
	Docroot        string
	SiteCount      int
	SubdomainCount int
	Sites          []SiteRow
}

func renderDeleteDomainPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainURL, docroot string, siteCount, subdomainCount int, sites []SiteRow) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Delete "+domainURL)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DeleteDomainPageData{
		LayoutData: layout, DomainURL: domainURL, Docroot: docroot,
		SiteCount: siteCount, SubdomainCount: subdomainCount, Sites: sites,
	}
	if err := deleteDomainPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - delete domain template render error: %v", err)
	}
}

// DomainSelectorPageData is the shared shape for every "pick a domain, then act on it" page (suspend/unsuspend/docroot/logs/vhost/redirect/ssl) in their no-domain-selected state
type DomainSelectorPageData struct {
	web.LayoutData
	DomainName string
	Domains    []appctx.Domain
}

func renderSuspendPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName string, domainsList []appctx.Domain) {
	title := "Suspend"
	if domainName != "" {
		title = "Suspend " + domainName
	}
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DomainSelectorPageData{LayoutData: layout, DomainName: domainName, Domains: domainsList}
	if err := suspendPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - suspend template render error: %v", err)
	}
}

func renderUnsuspendPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName string, domainsList []appctx.Domain) {
	title := "Unsuspend"
	if domainName != "" {
		title = "Unsuspend " + domainName
	}
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DomainSelectorPageData{LayoutData: layout, DomainName: domainName, Domains: domainsList}
	if err := unsuspendPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - unsuspend template render error: %v", err)
	}
}

// DocrootPageData is docroot.html's template context.
type DocrootPageData struct {
	web.LayoutData
	DomainName string
	Docroot    string
	Domains    []appctx.Domain
}

func renderDocrootPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName, docroot string, domainsList []appctx.Domain) {
	title := "Change docroot"
	if domainName != "" {
		title = "Change docroot for " + domainName
	}
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DocrootPageData{LayoutData: layout, DomainName: domainName, Docroot: docroot, Domains: domainsList}
	if err := docrootPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - docroot template render error: %v", err)
	}
}

// DomainLogsPageData is logs.html's template context.
type DomainLogsPageData struct {
	web.LayoutData
	DomainName string
	// Missing is a domain whose access.log doesn't exist, the picker shows it as an alert
	Missing                     string
	Domains                     []DomainLogFile
	Rows                        []AccessLogRow
	Details                     []AccessLogDetail
	Counts                      map[string]int
	Status, Search, View        string
	Matched                     int
	ShowAll                     bool
	CurrentPage, ItemsPerPage   int
	TotalPages, TotalLines      int
	TotalAllowedLinesForShowAll int
	LogPageEntries              []PageEntry
}

// URL links to another page or status chip of the same log, keeping the search and view
func (d DomainLogsPageData) URL(page int, status string) string {
	v := url.Values{}
	if d.Search != "" {
		v.Set("q", d.Search)
	}
	if status != "" {
		v.Set("status", status)
	}
	if d.View != "" {
		v.Set("view", d.View)
	}
	if page > 1 {
		v.Set("page", strconv.Itoa(page))
	}
	if d.ShowAll {
		v.Set("show_all", "true")
	}
	u := "/domains/log/" + d.DomainName
	if len(v) > 0 {
		u += "?" + v.Encode()
	}
	return u
}

// ShowAllURL is the same filters on one page
func (d DomainLogsPageData) ShowAllURL() string {
	d.ShowAll = true
	return d.URL(1, d.Status)
}

// buildLogPageEntries mirrors logs.html's own pagination window (distinct from domains.html's current±2 scheme): a fixed window of 2 pages around current, with page 1 and total_pages always shown, bridged by a single ellipsis each side when there's a gap
func buildLogPageEntries(current, total int) []PageEntry {
	const window = 2
	start := current - window
	if start < 1 {
		start = 1
	}
	end := current + window
	if end > total {
		end = total
	}

	var entries []PageEntry
	if start > 1 {
		entries = append(entries, PageEntry{Number: 1})
		if start > 2 {
			entries = append(entries, PageEntry{IsEllipsis: true})
		}
	}
	for p := start; p <= end; p++ {
		entries = append(entries, PageEntry{Number: p})
	}
	if end < total {
		if end < total-1 {
			entries = append(entries, PageEntry{IsEllipsis: true})
		}
		entries = append(entries, PageEntry{Number: total})
	}
	return entries
}

func renderDomainLogsSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, files []DomainLogFile, missing string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Access Logs")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	status := http.StatusOK
	if missing != "" {
		status = http.StatusNotFound
	}
	data := DomainLogsPageData{LayoutData: layout, Domains: files, Missing: missing}
	if err := domainLogsPage.Render(w, status, data); err != nil {
		log.Printf("DOMAINS - logs template render error: %v", err)
	}
}

func renderDomainLogsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data DomainLogsPageData, files []DomainLogFile) {
	layout, _, err := web.BuildLayoutData(a, w, r, data.DomainName+" Access Log")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData, data.Domains = layout, files
	if err := domainLogsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - logs template render error: %v", err)
	}
}

// CapitalizePageData is capitalize.html's template context.
type CapitalizePageData struct {
	web.LayoutData
	DomainURL         string
	CapitalizedDomain string
}

func renderCapitalizePage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainURL, capitalizedDomain string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Capitalize "+domainURL)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := CapitalizePageData{LayoutData: layout, DomainURL: domainURL, CapitalizedDomain: capitalizedDomain}
	if err := capitalizePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - capitalize template render error: %v", err)
	}
}

// VhostPageData is vhost.html's template context.
type VhostPageData struct {
	web.LayoutData
	DomainName          string
	WebServerPreference string
	VhostContent        string
	Domains             []appctx.Domain
	VhostFiles          map[string]VhostFileInfo // keyed by domain, missing when the file can't be read
	Domain              appctx.Domain            // edit page: docroot and PHP version for the summary tiles
	File                VhostFileInfo            // edit page: ports read from the file
}

func renderVhostSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, webServerPreference string, domainsList []appctx.Domain, vhostFiles map[string]VhostFileInfo) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit VirtualHosts File")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := VhostPageData{LayoutData: layout, WebServerPreference: webServerPreference, Domains: domainsList, VhostFiles: vhostFiles}
	if err := vhostPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - vhost template render error: %v", err)
	}
}

func renderVhostEditPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName, webServerPreference, vhostContent string, domain appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit VirtualHosts for "+domainName)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := VhostPageData{LayoutData: layout, DomainName: domainName, WebServerPreference: webServerPreference, VhostContent: vhostContent, Domain: domain, File: parseVhostFile(vhostContent)}
	if err := vhostPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - vhost template render error: %v", err)
	}
}

// RedirectRow is one redirected domain on the Redirects page
type RedirectRow struct {
	Domain    string `json:"domain"`
	To        string `json:"to"`
	Permanent bool   `json:"permanent"`
	KeepPath  bool   `json:"keep_path"`
	BareOnly  bool   `json:"bare_only"`
}

// RedirectPageData is redirect.html's template context.
type RedirectPageData struct {
	web.LayoutData
	Rows []RedirectRow
	// Free are the domains without a redirect, the new-redirect picker's options
	Free []string
	// DomainCount counts every domain that can take a redirect
	DomainCount int
	// Open is the domain whose form opens on load, from ?domain=
	Open string
}

// buildRedirectPageData splits the user's domains into redirected rows and free ones, .onion ones can't redirect
func buildRedirectPageData(layout web.LayoutData, domainsList []appctx.Domain, existing map[string]redirectRule, open string) RedirectPageData {
	data := RedirectPageData{LayoutData: layout, Open: open}
	for _, d := range domainsList {
		if strings.HasSuffix(d.DomainURL, ".onion") {
			continue
		}
		data.DomainCount++
		if rule := existing[d.DomainURL]; rule.To != "" {
			data.Rows = append(data.Rows, RedirectRow{Domain: d.DomainURL, To: rule.To, Permanent: rule.Permanent, KeepPath: rule.KeepPath, BareOnly: rule.BareOnly})
		} else {
			data.Free = append(data.Free, d.DomainURL)
		}
	}
	return data
}

func renderRedirectsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainsList []appctx.Domain, existing map[string]redirectRule, open string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Redirects")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := redirectPage.Render(w, http.StatusOK, buildRedirectPageData(layout, domainsList, existing, open)); err != nil {
		log.Printf("DOMAINS - redirect template render error: %v", err)
	}
}

// SSLPageData is ssl.html's template context.
type SSLPageData struct {
	web.LayoutData
	DomainName     string
	CurrentSetting string
	Keys           string
	Domains        []appctx.Domain
	Certs          []SSLRow // landing page: every domain's certificate
	Install        string   // ?install=domain opens the custom certificate sheet for it
}

// SSLRow is one domain on the SSL overview
type SSLRow struct {
	Domain string
	HTTPS  string // Automatic, Custom or anything else for none
	SSL    SSLCertInfo
}

// SSLSummary feeds the tiles on the SSL overview
type SSLSummary struct {
	Valid, Expiring, Expired, Pending               int
	FirstExpiring, FirstExpired, FirstPending, Next string
}

func (d SSLPageData) Summary() SSLSummary {
	var s SSLSummary
	soonest := -1
	for _, c := range d.Certs {
		switch {
		case c.HTTPS != "Automatic" && c.HTTPS != "Custom", c.SSL.Problem == "missing":
			if s.FirstPending == "" {
				s.FirstPending = c.Domain
			}
			s.Pending++
		case c.SSL.Problem == "expired" || c.SSL.Problem == "unreadable":
			if s.FirstExpired == "" {
				s.FirstExpired = c.Domain
			}
			s.Expired++
		case c.SSL.Problem == "expiring":
			if s.FirstExpiring == "" {
				s.FirstExpiring = c.Domain
			}
			s.Expiring++
		default:
			s.Valid++
			if soonest < 0 || c.SSL.Days < soonest {
				soonest, s.Next = c.SSL.Days, c.SSL.Expires
			}
		}
	}
	return s
}

func renderSSLPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName, currentSetting, keys string, domainsList []appctx.Domain, certs []SSLRow) {
	title := "SSL"
	if domainName != "" {
		title = "SSL for " + domainName
	}
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := SSLPageData{LayoutData: layout, DomainName: domainName, CurrentSetting: currentSetting, Keys: keys, Domains: domainsList, Certs: certs}
	for _, c := range certs {
		if c.Domain == r.URL.Query().Get("install") {
			data.Install = c.Domain
		}
	}
	if err := sslPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - ssl template render error: %v", err)
	}
}
