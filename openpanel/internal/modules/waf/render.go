package waf

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

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
	// profile recommendations, shown as toasts at most once a day
	Recommendations []WAFIssue
	Rows            map[string]DomainWAFRow
	// default for new domains, the /home/<context>/waf.disabled marker opencli domains-add reads
	AccountEnabled bool
	// choices for the level and profile filters next to search
	LevelOptions   []LevelView
	ProfileOptions []ProfileView
}

// Mixed is true when some domains don't match the default, so the page offers to apply it to all
func (d WAFListPageData) Mixed() bool {
	want := "Off"
	if d.AccountEnabled {
		want = "On"
	}
	for _, s := range d.ModsecStatus {
		if (s == "On" || s == "Off") && s != want {
			return true
		}
	}
	return false
}

// CountOn is how many domains have the firewall on
func (d WAFListPageData) CountOn() int {
	n := 0
	for _, s := range d.ModsecStatus {
		if s == "On" {
			n++
		}
	}
	return n
}

func renderWAFListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data WAFListPageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Web Firewall")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.BulkActions = wafBulkActions(layout.T)
	data.LayoutData = layout
	for _, p := range availableProfiles() {
		data.ProfileOptions = append(data.ProfileOptions, ProfileView{Key: p.Key, Name: p.Name})
	}
	for _, l := range levelCatalog {
		data.LevelOptions = append(data.LevelOptions, LevelView{Key: l.Key, Name: layout.T.Get(l.Name)})
	}
	for k, row := range data.Rows {
		row.LevelKey = row.Level
		row.LevelStrength = levelStrength(row.Level)
		row.LevelDots = make([]bool, 3)
		for i := range row.LevelDots {
			row.LevelDots[i] = i < row.LevelStrength
		}
		if l, ok := levelByKey(row.Level); ok {
			row.Level = layout.T.Get(l.Name)
		} else {
			row.Level = layout.T.Get("Custom")
		}
		data.Rows[k] = row
	}
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
	Profiles     []ProfileView
	// recommended profiles that aren't on yet, drives the "we detected..." banner
	Suggested []ProfileView
	Stats     wafLogStats
	// standard, compatibility, strict, or custom when an admin hand-edited it
	Level  string
	Levels []LevelView
	// plain-words summary of the newest log entries, the "Recent blocked requests" section
	Summary LogSummary
}

// LevelView is one protection level option on the domain's WAF page
type LevelView struct {
	Key, Name, Description string
}

// ProfileView is one profile card on the domain's WAF page
type ProfileView struct {
	Key, Name, Initials, Color, Description string
	Active, Recommended                     bool
	Sites                                   []string
}

func buildProfileViews(active []string, detected map[string][]string) (all, suggested []ProfileView) {
	on := map[string]bool{}
	for _, k := range active {
		on[k] = true
	}
	for _, p := range availableProfiles() {
		v := ProfileView{
			Key: p.Key, Name: p.Name, Initials: p.Initials, Color: p.Color, Description: p.Description,
			Active: on[p.Key], Recommended: len(detected[p.Key]) > 0, Sites: detected[p.Key],
		}
		all = append(all, v)
		if v.Recommended && !v.Active {
			suggested = append(suggested, v)
		}
	}
	return all, suggested
}

func renderWAFDomainPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data WAFDomainPageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Web Firewall for "+data.Domain)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData = layout
	for i := range data.Profiles {
		data.Profiles[i].Description = layout.T.Get(data.Profiles[i].Description)
	}
	for _, l := range levelCatalog {
		data.Levels = append(data.Levels, LevelView{Key: l.Key, Name: layout.T.Get(l.Name), Description: layout.T.Get(l.Description)})
	}
	for i := range data.Summary.Groups {
		data.Summary.Groups[i].Category = layout.T.Get(data.Summary.Groups[i].Category)
	}
	for i := range data.Summary.Events {
		for j := range data.Summary.Events[i].Rules {
			data.Summary.Events[i].Rules[j].Category = layout.T.Get(data.Summary.Events[i].Rules[j].Category)
		}
	}
	if err := wafDomainPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WAF - domain template render error: %v", err)
	}
}

// PageEntry is one rendered pagination control: either a page number link or an ellipsis
type PageEntry struct {
	Number     int
	IsEllipsis bool
}

// buildLogPageEntries mirrors waf_logs.html's own pagination window logic: a fixed window of 2 pages around current, with page 1 and total_pages always shown, bridged by a single ellipsis each side when there's a gap
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

// WAFLogsPageData is system/waf_logs.html's template context, covering both the domain-picker state (Domains set, JSONLogs nil) and the log-viewer state for one domain (JSONLogs set)
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
	DisabledRules               []string
}

func renderWAFLogSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Web Firewall Logs")
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
	if content, err := os.ReadFile(domainConfigPath(domainName)); err == nil {
		data.DisabledRules, _ = parseWAFRemovals(string(content))
	}
	if err := wafLogsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("WAF - logs template render error: %v", err)
	}
}
