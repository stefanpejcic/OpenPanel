package domains

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

var domainsPage = loadPage("domains/domains.html")
var newDomainPage = loadPage("domains/new.html")
var deleteDomainPage = loadPage("domains/delete.html")
var suspendPage = loadPage("domains/suspend.html")
var unsuspendPage = loadPage("domains/unsuspend.html")
var docrootPage = loadPage("domains/docroot.html")
var domainLogsPage = loadPage("domains/logs.html")
var capitalizePage = loadPage("domains/capitalize.html")
var vhostPage = loadPage("domains/vhost.html")
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
	PrevPage        int
	NextPage        int
	PageEntries     []PageEntry
}

// PageEntry is one rendered pagination control: either a page number link
// or an ellipsis.
type PageEntry struct {
	Number     int
	IsEllipsis bool
}

// buildPageEntries mirrors domains.html's pagination loop: current page
// (active), first/last page, and current±2 render as links; page 2 and
// total_pages-1 render as an ellipsis when they don't already qualify
// above; every other page renders nothing.
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

func renderDomainsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, rows []DomainRow, totalPages, currentPage, startLine, endLine, totalDomains int) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Domains")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DomainsPageData{
		LayoutData: layout, Domains: rows, TotalPages: totalPages, CurrentPage: currentPage,
		StartLineNumber: startLine, EndLineNumber: endLine, TotalDomains: totalDomains,
		PrevPage: currentPage - 1, NextPage: currentPage + 1,
		PageEntries: buildPageEntries(currentPage, totalPages),
	}
	if err := domainsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - domains template render error: %v", err)
	}
}

// NewDomainPageData is new.html's template context (stateless besides
// layout - the form posts via fetch()/SSE, not a normal redirect-driven
// flow).
type NewDomainPageData struct {
	web.LayoutData
}

func renderNewDomainPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Add Domain")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := newDomainPage.Render(w, http.StatusOK, NewDomainPageData{LayoutData: layout}); err != nil {
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

// DomainSelectorPageData is the shared shape for every "pick a domain,
// then act on it" page (suspend/unsuspend/docroot/logs/vhost/redirect/
// ssl) in their no-domain-selected state.
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
	DomainName                  string
	JSONLogs                    []AccessLogEntry
	ShowAll                     bool
	CurrentPage, ItemsPerPage   int
	TotalPages, TotalLines      int
	TotalAllowedLinesForShowAll int
	Domains                     []appctx.Domain
	LogPageEntries              []PageEntry
}

// buildLogPageEntries mirrors logs.html's own pagination window logic
// (distinct from domains.html's current±2 scheme): a fixed window of 2
// pages around current, with page 1 and total_pages always shown
// (bridged by a single ellipsis each side when there's a gap).
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

func renderDomainLogsSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainsList []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Access Logs")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DomainLogsPageData{LayoutData: layout, Domains: domainsList}
	if err := domainLogsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - logs template render error: %v", err)
	}
}

func renderDomainLogsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName string, entries []AccessLogEntry, showAll bool, currentPage, itemsPerPage, totalPages, totalLines, totalAllowedForShowAll int) {
	layout, _, err := web.BuildLayoutData(a, w, r, domainName+" Access Log")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DomainLogsPageData{
		LayoutData: layout, DomainName: domainName, JSONLogs: entries, ShowAll: showAll,
		CurrentPage: currentPage, ItemsPerPage: itemsPerPage, TotalPages: totalPages,
		TotalLines: totalLines, TotalAllowedLinesForShowAll: totalAllowedForShowAll,
		LogPageEntries: buildLogPageEntries(currentPage, totalPages),
	}
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
}

func renderVhostSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, webServerPreference string, domainsList []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit VirtualHosts File")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := VhostPageData{LayoutData: layout, WebServerPreference: webServerPreference, Domains: domainsList}
	if err := vhostPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - vhost template render error: %v", err)
	}
}

func renderVhostEditPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName, webServerPreference, vhostContent string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit VirtualHosts for "+domainName)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := VhostPageData{LayoutData: layout, DomainName: domainName, WebServerPreference: webServerPreference, VhostContent: vhostContent}
	if err := vhostPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - vhost template render error: %v", err)
	}
}

// RedirectPageData is redirect.html's template context.
type RedirectPageData struct {
	web.LayoutData
	DomainName  string
	RedirectURL string
	Domains     []appctx.Domain
}

func renderRedirectSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainsList []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Redirect Domain to URL")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := RedirectPageData{LayoutData: layout, Domains: domainsList}
	if err := redirectPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - redirect template render error: %v", err)
	}
}

func renderRedirectEditPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName, redirectURL string, domainsList []appctx.Domain) {
	title := "Create Redirect for " + domainName
	if redirectURL != "" {
		title = "Edit Redirect for " + domainName
	}
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := RedirectPageData{LayoutData: layout, DomainName: domainName, RedirectURL: redirectURL, Domains: domainsList}
	if err := redirectPage.Render(w, http.StatusOK, data); err != nil {
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
}

func renderSSLPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName, currentSetting, keys string, domainsList []appctx.Domain) {
	title := "SSL"
	if domainName != "" {
		title = "SSL for " + domainName
	}
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := SSLPageData{LayoutData: layout, DomainName: domainName, CurrentSetting: currentSetting, Keys: keys, Domains: domainsList}
	if err := sslPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOMAINS - ssl template render error: %v", err)
	}
}
