package waf

import (
	"encoding/json"
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
}

func loadPage(files ...string) *web.Page {
	return web.MustLoadPage(append(append([]string{}, pageFiles...), files...)...)
}

var (
	wafListPage   = loadPage("system/waf.html")
	wafDomainPage = loadPage("system/waf_domain.html")
	wafLogsPage   = loadPage("system/waf_logs.html")
)

// WAFListPageData is system/waf.html's template context.
type WAFListPageData struct {
	web.LayoutData
	Domains      []appctx.Domain
	ModsecStatus map[string]string
	Issues       []WAFIssue
}

func renderWAFListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain, modsecStatus map[string]string, issues []WAFIssue) {
	layout, _, err := web.BuildLayoutData(a, w, r, "WAF")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := WAFListPageData{LayoutData: layout, Domains: domains, ModsecStatus: modsecStatus, Issues: issues}
	if err := wafListPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WAF - list template render error: %v", err)
	}
}

// WAFDomainPageData is system/waf_domain.html's template context.
type WAFDomainPageData struct {
	web.LayoutData
	Domain       string
	Status       string
	RemovedRules []string
	RemovedTags  []string
}

func renderWAFDomainPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domain, status string, removedRules, removedTags []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "WAF for "+domain)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := WAFDomainPageData{
		LayoutData: layout, Domain: domain, Status: status,
		RemovedRules: removedRules, RemovedTags: removedTags,
	}
	if err := wafDomainPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WAF - domain template render error: %v", err)
	}
}

// PageEntry is one rendered pagination control: either a page number link
// or an ellipsis.
type PageEntry struct {
	Number     int
	IsEllipsis bool
}

// buildLogPageEntries mirrors waf_logs.html's own pagination window logic:
// a fixed window of 2 pages around current, with page 1 and total_pages
// always shown (bridged by a single ellipsis each side when there's a
// gap).
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

// WAFLogsPageData is system/waf_logs.html's template context, covering
// both the domain-picker state (Domains set, JSONLogs nil) and the
// log-viewer state for one domain (JSONLogs set).
type WAFLogsPageData struct {
	web.LayoutData
	DomainName                  string
	JSONLogs                    []json.RawMessage
	ShowAll                     bool
	CurrentPage, ItemsPerPage   int
	TotalPages, TotalLines      int
	TotalAllowedLinesForShowAll int
	Domains                     []appctx.Domain
	PageEntries                 []PageEntry
}

func renderWAFLogSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "WAF Logs")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := WAFLogsPageData{LayoutData: layout, Domains: domains}
	if err := wafLogsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WAF - logs select template render error: %v", err)
	}
}

func renderWAFLogPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName string, jsonLogs []json.RawMessage, showAll bool, currentPage, itemsPerPage, totalPages, totalLines, totalAllowedForShowAll int) {
	layout, _, err := web.BuildLayoutData(a, w, r, domainName+" WAF Log")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := WAFLogsPageData{
		LayoutData: layout, DomainName: domainName, JSONLogs: jsonLogs, ShowAll: showAll,
		CurrentPage: currentPage, ItemsPerPage: itemsPerPage, TotalPages: totalPages,
		TotalLines: totalLines, TotalAllowedLinesForShowAll: totalAllowedForShowAll,
		PageEntries: buildLogPageEntries(currentPage, totalPages),
	}
	if err := wafLogsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WAF - logs template render error: %v", err)
	}
}
