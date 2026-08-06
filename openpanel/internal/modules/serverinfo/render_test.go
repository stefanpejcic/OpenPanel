package serverinfo

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "usage": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderServerInfoPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := ServerInfoPageData{LayoutData: baseLayout(mgr, "/server/info")}
	w := httptest.NewRecorder()
	if err := serverInfoPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "Server Information") {
		t.Error("expected page title in body")
	}
}

func TestRenderStatsPage(t *testing.T) {
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
		data := StatsPageData{LayoutData: baseLayout(mgr, "/server/usage"), HasStats: true, Stats: stats}
		w := httptest.NewRecorder()
		if err := statsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "0.5 cores") || !strings.Contains(body, "800MB") {
			t.Error("expected stats values in body")
		}
	})
}

func TestRenderUsageHistoryPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	var row ResourceUsageLine
	row.Timestamp = "2026-08-02 12:00:00"
	row.CPU.Usage.Pct = 10
	row.CPU.Total.Human = "4 cores"
	row.Memory.UsagePct = 20
	row.Memory.Used.Human = "400MB"
	row.Memory.Total.Human = "2GB"
	row.Bandwidth.TotalSent.Human = "1GB"
	row.Bandwidth.Limit.Human = "10GB"
	row.Bandwidth.UsagePct = 10
	row.Warning = "disk almost full"

	data := UsageHistoryPageData{
		LayoutData: baseLayout(mgr, "/server/usage/history"), ChartsMode: "one", ShowAll: false,
		ItemsPerPage: 25, TotalPages: 1, TotalLines: 1, CurrentPage: 1,
		UsageData: []ResourceUsageLine{row}, PageEntries: buildPageEntries(1, 1),
	}
	w := httptest.NewRecorder()
	if err := usageHistoryPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "disk almost full") || !strings.Contains(body, "400MB") {
		t.Error("expected usage row in body")
	}
}
