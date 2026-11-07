package serverinfo

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "usage": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderServerInfoPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := ServerInfoPageData{LayoutData: baseLayout(mgr, "/server/info")}
	w := httptest.NewRecorder()
	if err := serverInfoPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Server Information") {
		t.Error("expected page title in body")
	}
	for _, want := range []string{`x-show="tab === 'server'"`, `x-show="tab === 'plan'" x-cloak`, `x-show="tab === 'panel'" x-cloak`, "openpanel: 'panel'", "/json/system/hosting/plan", "/json/system/features", "raw.githubusercontent.com", `id="plan_domains_limit"`, "use('domains')", `id="plan_inodes_limit"`} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
}

func TestRenderStatsPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no stats", func(t *testing.T) {
		data := StatsPageData{LayoutData: baseLayout(mgr, "/server/usage")}
		w := httptest.NewRecorder()
		if err := statsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No resource usage data available yet.") {
			t.Error("expected empty state text")
		}
	})

	t.Run("with stats", func(t *testing.T) {
		var stats ResourceUsageLine
		stats.Timestamp = "2026-08-02 12:00:00"
		stats.CPU.Usage.Human = "0.5 cores"
		stats.CPU.Usage.Pct = 12.5
		stats.CPU.Total.Human = "4 cores"
		stats.CPU.Total.Pct = 100
		stats.Memory.UsagePct = 40
		stats.Memory.Used.Human = "800MB"
		stats.Memory.Total.Human = "2GB"
		stats.Memory.Available.Human = "1.2GB"
		layout := baseLayout(mgr, "/server/usage")
		layout.UserAllowed = map[string]bool{"dashboard": true, "usage": true, "docker": true, "summary": true}
		var lines []ResourceUsageLine
		for i := 0; i < 48; i++ {
			var l ResourceUsageLine
			l.Timestamp = time.Date(2026, 8, 2, 12, 0, 0, 0, time.Local).Add(time.Duration(i-47) * 30 * time.Minute).Format("2006-01-02 15:04:05")
			l.CPU.Usage.Pct = float64(5 + i%7)
			l.Memory.UsagePct = float64(35 + i%5)
			lines = append(lines, l)
		}
		data := StatsPageData{LayoutData: layout, HasStats: true, Stats: stats, History: usageHistory(lines)}
		w := httptest.NewRecorder()
		if err := statsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{"0.5 cores of 4 cores", "12.5<small>%</small>", "800 <small>MB</small>", "1.2 GB free", `width:40%`, `&#34;c&#34;:11`, "/json/services", `href="/summary"`, "Usage per container"} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in body", want)
			}
		}
	})
}

// realLines are testinguser's hourly samples from the test VM
func realLines(t *testing.T) []ResourceUsageLine {
	t.Helper()
	raw, err := os.ReadFile("testdata/resource_usage.txt")
	if err != nil {
		t.Fatal(err)
	}
	var lines []ResourceUsageLine
	for _, l := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var u ResourceUsageLine
		if err := json.Unmarshal([]byte(l), &u); err != nil {
			t.Fatalf("bad line %q: %v", l, err)
		}
		lines = append(lines, u)
	}
	return lines
}

func TestRenderUsageHistoryPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	samples, limits := historySamples(realLines(t))
	data := UsageHistoryPageData{LayoutData: baseLayout(mgr, "/server/usage/history"), Samples: samples, Limits: limits, Retention: 168, Ranges: historyRanges(168), RowLimit: 24}
	data.UserAllowed["summary"] = true
	w := httptest.NewRecorder()
	if err := usageHistoryPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{
		"Usage History", "24 hours", "3 days", "7 days", `data-sort-key="cpu"`, "opSortToggle", "js/table-sort.js",
		// newest sample first, 18:00 on the 29th with 1.3 GB and 329 processes
		`<b>18:00</b><span class="op-sub"><time data-day datetime="2026-10-29T18:00:02Z">Thu 29 Oct</time>`, "1.33 GB", `data-sort-value="329"`,
		"The newest 168 samples are kept.", "No bandwidth limit is set on this account.", `href="/summary"`, "Only samples over a threshold",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
	if strings.Index(body, `data-t="1793296802000"`) > strings.Index(body, `data-t="1793019234000"`) {
		t.Error("table should list the newest sample first")
	}

	w = httptest.NewRecorder()
	empty := UsageHistoryPageData{LayoutData: baseLayout(mgr, "/server/usage/history"), Retention: 24, Ranges: historyRanges(24), RowLimit: 24}
	if err := usageHistoryPage.Render(w, 200, empty); err != nil {
		t.Fatalf("Render empty: %v", err)
	}
	if body := w.Body.String(); !strings.Contains(body, "No usage samples yet") || strings.Contains(body, "3 days") {
		t.Error("expected the empty state without range buttons")
	}
}

func TestUsageHistory(t *testing.T) {
	line := func(ts string, cpu, mem float64) ResourceUsageLine {
		var l ResourceUsageLine
		l.Timestamp, l.CPU.Usage.Pct, l.Memory.UsagePct = ts, cpu, mem
		return l
	}
	got := usageHistory([]ResourceUsageLine{
		line("2026-08-02 12:00:00", 3, 40),
		line("bad", 1, 1),
		line("2026-06-01 12:00:00", 9, 9),
		line("2026-08-02 11:30:00", 2, 39),
	})
	if len(got) != 2 || got[0].CPU != 2 || got[1].Mem != 40 {
		t.Fatalf("usageHistory = %+v, want the two recent samples in order", got)
	}

	// collect_stats writes RFC3339, which used to fail every line
	real := usageHistory(realLines(t))
	if len(real) != 77 || real[0].T != 1793019234 || real[76].Mem != 22 {
		t.Fatalf("usageHistory on the real file = %d points, first %+v, last %+v", len(real), real[0], real[len(real)-1])
	}
}

func TestHistorySamples(t *testing.T) {
	samples, limits := historySamples(realLines(t))
	if len(samples) != 77 {
		t.Fatalf("got %d samples, want 77", len(samples))
	}
	if limits != (UsageLimits{MemBytes: 6442450944, Cores: 4, Tasks: 6000}) {
		t.Errorf("limits = %+v", limits)
	}
	first, last := samples[0], samples[76]
	if first.T != 1793019234000 || first.CPU != 0.05 || first.CPUPct != 1 || first.Mem != 45465600 || first.Tasks != 25 || first.Warn != "" {
		t.Errorf("first sample = %+v", first)
	}
	if last.Tasks != 329 || last.MemPct != 22 || last.GB() != "1.33" || last.Cores() != "0.00" || last.Time().Format("2006-01-02 15:04") != "2026-10-29 18:00" {
		t.Errorf("last sample = %+v", last)
	}
	for i := 1; i < len(samples); i++ {
		if samples[i].T < samples[i-1].T {
			t.Fatal("samples should be oldest first")
		}
	}
}

func TestWarningFor(t *testing.T) {
	var l ResourceUsageLine
	l.CPU.Usage.Pct, l.CPU.Total.Pct = 360, 400
	l.Memory.UsagePct = 85
	l.Bandwidth.UsagePct = 89
	if got := warningFor(l); got != "Memory at 85%, CPU at 90% — above threshold" {
		t.Errorf("warningFor = %q", got)
	}
	l.Memory.UsagePct, l.CPU.Usage.Pct = 84, 336
	if got := warningFor(l); got != "" {
		t.Errorf("under the thresholds got %q", got)
	}
	l.Warning = "Bandwidth at 95% — above threshold"
	if got := warningFor(l); got != l.Warning {
		t.Errorf("a saved warning should win, got %q", got)
	}
}

func TestHistoryRanges(t *testing.T) {
	for retention, want := range map[int]string{0: "24h 3d 7d", 24: "24h", 72: "24h 3d", 100: "24h 3d 7d", 168: "24h 3d 7d"} {
		if got := strings.Join(historyRanges(retention), " "); got != want {
			t.Errorf("historyRanges(%d) = %q, want %q", retention, got, want)
		}
	}
}

func TestHumanBinary(t *testing.T) {
	for in, want := range map[int64]string{0: "0 B", 1023: "1023 B", 1536: "1.5 KB", 5 << 30: "5.0 GB"} {
		if got := humanBinary(in); got != want {
			t.Errorf("humanBinary(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestSplitHuman(t *testing.T) {
	for in, want := range map[string][2]string{"800MB": {"800", "MB"}, "0.5 cores": {"0.5", "cores"}, "2.75 GB": {"2.75", "GB"}, "": {"", ""}} {
		if n, u := splitHuman(in); n != want[0] || u != want[1] {
			t.Errorf("splitHuman(%q) = %q, %q", in, n, u)
		}
	}
}
