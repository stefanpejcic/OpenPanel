package serverinfo

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var (
	serverInfoPage   = web.LoadPage("system/server_info.html")
	statsPage        = web.LoadPage("user/stats.html")
	usageHistoryPage = web.LoadPage("user/history_usage.html")
)

// ServerInfoPageData is system/server_info.html's template context - the page is filled in almost entirely client-side via /json/system/hosting/* fetches, so this only carries the layout
type ServerInfoPageData struct {
	web.LayoutData
	Features []Feature
}

// Feature is one entry of OpenAdmin's features.json, only what the page lists
type Feature struct {
	Name     string `json:"name"`
	HelpLink string `json:"help_link"`
}

const featuresFile = "/etc/openpanel/openadmin/config/features.json"

// enabledFeatures reads the local features.json the configuration repo ships, the page used to fetch it from github on every load
func enabledFeatures(allowed map[string]bool) []Feature {
	content, err := os.ReadFile(featuresFile)
	if err != nil {
		return nil
	}
	var all []Feature
	if json.Unmarshal(content, &all) != nil {
		return nil
	}
	var out []Feature
	for _, f := range all {
		if allowed[f.Name] {
			out = append(out, f)
		}
	}
	return out
}

func renderServerInfoPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Server information")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := serverInfoPage.Render(w, http.StatusOK, ServerInfoPageData{LayoutData: layout, Features: enabledFeatures(layout.UserAllowed)}); err != nil {
		log.Printf("SERVERINFO - server info template render error: %v", err)
	}
}

// StatsPageData is user/stats.html's template context.
type StatsPageData struct {
	web.LayoutData
	HasStats bool
	Stats    ResourceUsageLine
	// the modern theme's CPU and memory chart
	History []UsagePoint
}

func renderStatsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, rawStats map[string]any, history []UsagePoint) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Resource usage")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := StatsPageData{LayoutData: layout, History: history}
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

// PageEntry is one rendered pagination control: either a page number link or an ellipsis
type PageEntry struct {
	Number     int
	IsEllipsis bool
}

// buildPageEntries builds the pagination control list: current page (active), first/last page, and current+-2 render as links; page 2 and total_pages-1 render as an ellipsis when they don't already qualify above; every other page renders nothing
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
	layout, _, err := web.BuildLayoutData(a, w, r, "Resource usage history")
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

// UsagePoint is one sample for the CPU and memory chart: unix seconds, CPU % and memory %
type UsagePoint struct {
	T   int64   `json:"t"`
	CPU float64 `json:"c"`
	Mem float64 `json:"m"`
}

// usageHistory turns the usage file's lines into chart points from the 30 days before the newest sample
func usageHistory(lines []ResourceUsageLine) []UsagePoint {
	points := make([]UsagePoint, 0, len(lines))
	for _, l := range lines {
		ts, ok := parseUsageTime(l.Timestamp)
		if !ok {
			continue
		}
		points = append(points, UsagePoint{T: ts.Unix(), CPU: l.CPU.Usage.Pct, Mem: l.Memory.UsagePct})
	}
	sort.SliceStable(points, func(i, j int) bool { return points[i].T < points[j].T })
	if len(points) > 0 {
		cutoff := points[len(points)-1].T - 30*24*3600
		points = points[sort.Search(len(points), func(i int) bool { return points[i].T >= cutoff }):]
	}
	return points
}

// Num is the number part of a human value like "800MB" or "0.5 cores"
func (StatsPageData) Num(s string) string {
	num, _ := splitHuman(s)
	return num
}

// Unit is the unit part of a human value like "800MB" or "0.5 cores", a bare "G" reads as "GB"
func (StatsPageData) Unit(s string) string {
	_, unit := splitHuman(s)
	if len(unit) == 1 && strings.Contains("KMGTP", strings.ToUpper(unit)) {
		unit = strings.ToUpper(unit) + "B"
	}
	return unit
}

// Human prints a human value with a space before its unit, "6.0G" as "6.0 GB"
func (d StatsPageData) Human(s string) string {
	num, unit := splitHuman(s)
	if unit = d.Unit(s); unit == "" {
		return num
	}
	return num + " " + unit
}

// Pct prints a percentage with at most one decimal
func (StatsPageData) Pct(f float64) string {
	return strconv.FormatFloat(math.Round(f*10)/10, 'f', -1, 64)
}

// Bar clamps a percentage to 0-100 for a progress bar width
func (StatsPageData) Bar(f float64) float64 {
	return math.Max(0, math.Min(100, f))
}

func splitHuman(s string) (string, string) {
	s = strings.TrimSpace(s)
	i := strings.IndexFunc(s, func(r rune) bool { return (r < '0' || r > '9') && r != '.' && r != ',' })
	if i < 0 {
		return s, ""
	}
	return s[:i], strings.TrimSpace(s[i:])
}

// ModernUsageHistoryPageData is the modern theme's user/history_usage.html context, Samples are oldest first
type ModernUsageHistoryPageData struct {
	web.LayoutData
	Samples   []HistorySample
	Limits    UsageLimits
	Retention int
	Ranges    []string
	RowLimit  int
}

// Rows are the samples newest first, the table's default order
func (d ModernUsageHistoryPageData) Rows() []HistorySample {
	out := make([]HistorySample, len(d.Samples))
	for i, s := range d.Samples {
		out[len(d.Samples)-1-i] = s
	}
	return out
}

// Thresholds are collect_stats' warning levels, for the page's notes
func (ModernUsageHistoryPageData) Thresholds() map[string]int {
	return map[string]int{"mem": memWarnPct, "cpu": cpuWarnPct, "bw": bwWarnPct}
}

// LimitCores prints the plan's cores
func (d ModernUsageHistoryPageData) LimitCores() string { return fmtCores(d.Limits.Cores) }

// renderModernUsageHistoryPage is the history page with the hourly samples the modern theme charts
func renderModernUsageHistoryPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data ModernUsageHistoryPageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Resource usage history")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData = layout
	if err := usageHistoryPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("SERVERINFO - usage history template render error: %v", err)
	}
}
