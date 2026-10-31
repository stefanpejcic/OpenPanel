package serverinfo

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
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
	History  []UsagePoint
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

func renderStatsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, rawStats map[string]any, history []UsagePoint) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Resource Usage")
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

// UsageHistoryPageData is user/history_usage.html's template context, Samples are oldest first
type UsageHistoryPageData struct {
	web.LayoutData
	Samples   []HistorySample
	Limits    UsageLimits
	Retention int
	Ranges    []string
	RowLimit  int
}

// Rows are the samples newest first, the table's default order
func (d UsageHistoryPageData) Rows() []HistorySample {
	out := make([]HistorySample, len(d.Samples))
	for i, s := range d.Samples {
		out[len(d.Samples)-1-i] = s
	}
	return out
}

// Thresholds are collect_stats' warning levels, for the page's notes
func (UsageHistoryPageData) Thresholds() map[string]int {
	return map[string]int{"mem": memWarnPct, "cpu": cpuWarnPct, "bw": bwWarnPct}
}

// LimitCores prints the plan's cores
func (d UsageHistoryPageData) LimitCores() string { return fmtCores(d.Limits.Cores) }

func renderUsageHistoryPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data UsageHistoryPageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Usage History")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData = layout
	if err := usageHistoryPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("SERVERINFO - usage history template render error: %v", err)
	}
}
