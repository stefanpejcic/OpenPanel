package serverinfo

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
	serverInfoPage   = loadPage("system/server_info.html")
	statsPage        = loadPage("user/stats.html")
	usageHistoryPage = loadPage("user/history_usage.html")
)

// ServerInfoPageData is system/server_info.html's template context. The
// page is filled in almost entirely client-side via the
// /json/system/hosting/* fetches, so this only carries the layout.
type ServerInfoPageData struct {
	web.LayoutData
}

func renderServerInfoPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Server Information")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := serverInfoPage.Render(w, http.StatusOK, ServerInfoPageData{LayoutData: layout}); err != nil {
		log.Printf("SERVERINFO - server info template render error: %v", err)
	}
}

// StatsPageData is user/stats.html's template context.
type StatsPageData struct {
	web.LayoutData
	HasStats bool
	Stats    ResourceUsageLine
}

func renderStatsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, rawStats map[string]any) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Resource Usage")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := StatsPageData{LayoutData: layout}
	if _, hasCPU := rawStats["cpu"]; hasCPU {
		if _, hasMemory := rawStats["memory"]; hasMemory {
			if b, marshalErr := json.Marshal(rawStats); marshalErr == nil {
				if unmarshalErr := json.Unmarshal(b, &data.Stats); unmarshalErr == nil {
					data.HasStats = true
				}
			}
		}
	}
	if err := statsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("SERVERINFO - stats template render error: %v", err)
	}
}

// PageEntry is one rendered pagination control: either a page number link
// or an ellipsis.
type PageEntry struct {
	Number     int
	IsEllipsis bool
}

// buildPageEntries builds the pagination control list: current page
// (active), first/last page, and current+-2 render as links; page 2 and
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

// UsageHistoryPageData is user/history_usage.html's template context.
type UsageHistoryPageData struct {
	web.LayoutData
	ChartsMode   string
	ShowAll      bool
	ItemsPerPage int
	TotalPages   int
	TotalLines   int
	CurrentPage  int
	UsageData    []ResourceUsageLine
	PageEntries  []PageEntry
}

func renderUsageHistoryPage(a *appctx.App, w http.ResponseWriter, r *http.Request, chartsMode string, showAll bool, itemsPerPage, totalPages, totalLines, currentPage int, usageData []ResourceUsageLine) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Resource Usage History")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := UsageHistoryPageData{
		LayoutData: layout, ChartsMode: chartsMode, ShowAll: showAll, ItemsPerPage: itemsPerPage,
		TotalPages: totalPages, TotalLines: totalLines, CurrentPage: currentPage, UsageData: usageData,
		PageEntries: buildPageEntries(currentPage, totalPages),
	}
	if err := usageHistoryPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("SERVERINFO - usage history template render error: %v", err)
	}
}
