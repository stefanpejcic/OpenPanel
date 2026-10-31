package goaccess

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// the GoAccess HTML report carries all of its numbers in one "var json_data={...}" script line
var jsonDataMarker = []byte("var json_data=")

type gaCount struct {
	Count int64 `json:"count"`
}

type gaItem struct {
	Hits     gaCount         `json:"hits"`
	Visitors gaCount         `json:"visitors"`
	Bytes    gaCount         `json:"bytes"`
	Data     string          `json:"data"`
	Items    json.RawMessage `json:"items"`
}

type gaPanel struct {
	Data []gaItem `json:"data"`
}

type gaReport struct {
	General struct {
		StartDate      string `json:"start_date"`
		EndDate        string `json:"end_date"`
		DateTime       string `json:"date_time"`
		TotalRequests  int64  `json:"total_requests"`
		ValidRequests  int64  `json:"valid_requests"`
		FailedRequests int64  `json:"failed_requests"`
		UniqueVisitors int64  `json:"unique_visitors"`
		Bandwidth      int64  `json:"bandwidth"`
	} `json:"general"`
	Visitors       gaPanel `json:"visitors"`
	Requests       gaPanel `json:"requests"`
	ReferringSites gaPanel `json:"referring_sites"`
	Geolocation    gaPanel `json:"geolocation"`
	Browsers       gaPanel `json:"browsers"`
	StatusCodes    gaPanel `json:"status_codes"`
}

// StatBar is one row of a top-N list, Pct is relative to the biggest row
type StatBar struct {
	Label string
	Count int64
	Pct   int
}

// StatusCount is one status code class like 2xx with the kit tone it shows in
type StatusCount struct {
	Class string
	Count int64
	Tone  string
}

// StatsSummary is what the Visitor Statistics page shows for one domain
type StatsSummary struct {
	Updated       time.Time
	Period        string
	Visitors      int64
	Requests      int64
	PerVisitor    string
	Bandwidth     string
	BandwidthUnit string
	Failed        int64
	FailedPct     string
	Days          int
	ChartVisitors string
	ChartHits     string
	ChartArea     string
	AxisLabels    []string
	TopPages      []StatBar
	Referrers     []StatBar
	Countries     []StatBar
	Browsers      []StatBar
	Statuses      []StatusCount
}

// parseReport pulls the json_data object out of a GoAccess HTML report and boils it down for the page
func parseReport(html []byte) (*StatsSummary, error) {
	i := bytes.Index(html, jsonDataMarker)
	if i < 0 {
		return nil, errors.New("no json_data in report")
	}
	var rep gaReport
	if err := json.NewDecoder(bytes.NewReader(html[i+len(jsonDataMarker):])).Decode(&rep); err != nil {
		return nil, fmt.Errorf("decode json_data: %w", err)
	}
	g := rep.General
	s := &StatsSummary{
		Visitors: g.UniqueVisitors, Requests: g.ValidRequests, Failed: g.FailedRequests,
	}
	s.Bandwidth, s.BandwidthUnit = humanBytes(g.Bandwidth)
	if t, err := time.Parse("2006-01-02 15:04:05 -0700", g.DateTime); err == nil {
		s.Updated = t
	}
	if g.StartDate != "" {
		s.Period = shortDate(g.StartDate)
		if g.EndDate != "" && g.EndDate != g.StartDate {
			s.Period += " – " + shortDate(g.EndDate)
		}
	}
	if g.UniqueVisitors > 0 {
		s.PerVisitor = fmt.Sprintf("%.1f", float64(g.ValidRequests)/float64(g.UniqueVisitors))
	}
	if g.TotalRequests > 0 {
		s.FailedPct = fmt.Sprintf("%.1f%%", float64(g.FailedRequests)*100/float64(g.TotalRequests))
	}

	days := append([]gaItem(nil), rep.Visitors.Data...)
	sort.Slice(days, func(a, b int) bool { return days[a].Data < days[b].Data })
	s.Days = len(days)
	s.ChartVisitors, s.ChartHits, s.ChartArea, s.AxisLabels = chartPaths(days)

	s.TopPages = topBars(rep.Requests.Data, 5)
	s.Referrers = topBars(rep.ReferringSites.Data, 5)
	s.Countries = topBars(rep.Geolocation.Data, 5)
	s.Browsers = topBars(rep.Browsers.Data, 5)
	for _, it := range rep.StatusCodes.Data {
		class := strings.Fields(it.Data + " ")[0]
		s.Statuses = append(s.Statuses, StatusCount{Class: class, Count: it.Hits.Count, Tone: statusTone(class)})
	}
	sort.Slice(s.Statuses, func(a, b int) bool { return s.Statuses[a].Class < s.Statuses[b].Class })
	return s, nil
}

func statusTone(class string) string {
	switch {
	case strings.HasPrefix(class, "2"):
		return "is-success"
	case strings.HasPrefix(class, "4"):
		return "is-warning"
	case strings.HasPrefix(class, "5"):
		return "is-danger"
	}
	return "is-muted"
}

func topBars(items []gaItem, n int) []StatBar {
	list := append([]gaItem(nil), items...)
	sort.SliceStable(list, func(a, b int) bool { return list[a].Hits.Count > list[b].Hits.Count })
	if len(list) > n {
		list = list[:n]
	}
	var out []StatBar
	for _, it := range list {
		if it.Hits.Count == 0 {
			continue
		}
		pct := 100
		if list[0].Hits.Count > 0 {
			pct = int(it.Hits.Count * 100 / list[0].Hits.Count)
		}
		out = append(out, StatBar{Label: it.Data, Count: it.Hits.Count, Pct: max(pct, 2)})
	}
	return out
}

// chartPaths draws visitors and hits per day into a 600x200 box, the same one the kit's .op-chart svg uses
func chartPaths(days []gaItem) (visitors, hits, area string, axis []string) {
	if len(days) == 0 {
		return "", "", "", nil
	}
	// one day can't make a line, so it gets two bars, visitors filled and page views outlined
	if len(days) == 1 {
		top := max(1, days[0].Hits.Count, days[0].Visitors.Count)
		bar := func(x0 float64, v int64) string {
			y := 200 - float64(v)*180/float64(top)
			return fmt.Sprintf("M%.1f,200 L%.1f,%.1f L%.1f,%.1f L%.1f,200 Z", x0, x0, y, x0+90, y, x0+90)
		}
		visitors = bar(206, days[0].Visitors.Count)
		return visitors, bar(304, days[0].Hits.Count), visitors, []string{dayLabel(days[0].Data)}
	}
	pts := days
	var top int64 = 1
	for _, d := range pts {
		top = max(top, d.Hits.Count, d.Visitors.Count)
	}
	x := func(i int) float64 { return float64(i) * 600 / float64(len(pts)-1) }
	y := func(v int64) float64 { return 190 - float64(v)*170/float64(top) }
	var vb, hb strings.Builder
	for i, d := range pts {
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&vb, "%s%.1f,%.1f ", cmd, x(i), y(d.Visitors.Count))
		fmt.Fprintf(&hb, "%s%.1f,%.1f ", cmd, x(i), y(d.Hits.Count))
	}
	visitors = strings.TrimSpace(vb.String())
	hits = strings.TrimSpace(hb.String())
	area = visitors + " L600,200 L0,200 Z"
	// first, middle and last day under the chart
	for _, i := range uniqueInts(0, len(days)/2, len(days)-1) {
		axis = append(axis, dayLabel(days[i].Data))
	}
	return visitors, hits, area, axis
}

func uniqueInts(v ...int) []int {
	var out []int
	for _, n := range v {
		if len(out) == 0 || out[len(out)-1] != n {
			out = append(out, n)
		}
	}
	return out
}

// the general block writes dates as 26/Oct/2026
func shortDate(s string) string {
	if t, err := time.Parse("02/Jan/2006", s); err == nil {
		return t.Format("2 Jan")
	}
	return s
}

// GoAccess keys days as 20261026
func dayLabel(s string) string {
	if t, err := time.Parse("20060102", s); err == nil {
		return t.Format("2 Jan")
	}
	return s
}

func humanBytes(n int64) (string, string) {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	f := float64(n)
	i := 0
	for f >= 1024 && i < len(units)-1 {
		f /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d", n), "B"
	}
	return fmt.Sprintf("%.1f", f), units[i]
}
