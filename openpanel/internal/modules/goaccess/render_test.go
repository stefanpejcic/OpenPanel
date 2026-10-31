package goaccess

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "domains": true, "goaccess": true}
	return web.LayoutData{
		Title: "Visitor Statistics", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
		PageTabs: web.BuildPageTabs(userAllowed, nil, path, web.TabContext{}),
	}
}

func TestRenderGoaccessSelectPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("with domains", func(t *testing.T) {
		data := GoaccessPageData{LayoutData: baseLayout(mgr, "/domains/stats"), Domains: []appctx.Domain{{DomainURL: "example.com"}, {DomainURL: "shop.example.com"}}}
		w := httptest.NewRecorder()
		if err := goaccessPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "shop.example.com") {
			t.Error("expected the domains in the picker")
		}
	})

	t.Run("no domains", func(t *testing.T) {
		data := GoaccessPageData{LayoutData: baseLayout(mgr, "/domains/stats")}
		w := httptest.NewRecorder()
		if err := goaccessPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No domains yet") {
			t.Error("expected the empty state")
		}
	})
}

// trimmed from a real report on the test install
const sampleReport = `<html><script>var json_data={"general": {"start_date": "24/Oct/2026","end_date": "26/Oct/2026","date_time": "2026-10-27 02:30:12 +0000","total_requests": 40,"valid_requests": 38,"failed_requests": 2,"unique_visitors": 9,"bandwidth": 1468006},
"visitors": {"data": [{"hits": {"count": 20}, "visitors": {"count": 5}, "data": "20261026"}, {"hits": {"count": 8}, "visitors": {"count": 2}, "data": "20261024"}, {"hits": {"count": 10}, "visitors": {"count": 3}, "data": "20261025"}]},
"requests": {"data": [{"hits": {"count": 2}, "data": "/restored"}, {"hits": {"count": 12}, "data": "/"}, {"hits": {"count": 5}, "data": "/shop/"}]},
"referring_sites": {"data": []},
"browsers": {"data": [{"hits": {"count": 30}, "data": "Chrome", "items": []}, {"hits": {"count": 8}, "data": "Others", "items": [{"hits": {"count": 8}, "data": "curl/8.12.1"}]}]},
"status_codes": {"data": [{"hits": {"count": 30}, "data": "2xx Success"}, {"hits": {"count": 2}, "data": "4xx Client Errors"}, {"hits": {"count": 6}, "data": "3xx Redirection"}]}};</script></html>`

func TestParseReport(t *testing.T) {
	s, err := parseReport([]byte(sampleReport))
	if err != nil {
		t.Fatalf("parseReport: %v", err)
	}
	if s.Visitors != 9 || s.Requests != 38 || s.Failed != 2 || s.Days != 3 {
		t.Errorf("totals = %+v", s)
	}
	if s.Bandwidth != "1.4" || s.BandwidthUnit != "MB" {
		t.Errorf("bandwidth = %q %q", s.Bandwidth, s.BandwidthUnit)
	}
	if s.Period != "24 Oct – 26 Oct" || s.PerVisitor != "4.2" || s.FailedPct != "5.0%" {
		t.Errorf("period/per visitor/failed = %q %q %q", s.Period, s.PerVisitor, s.FailedPct)
	}
	if len(s.TopPages) != 3 || s.TopPages[0].Label != "/" || s.TopPages[0].Pct != 100 || s.TopPages[2].Label != "/restored" {
		t.Errorf("top pages = %+v", s.TopPages)
	}
	if len(s.Referrers) != 0 || len(s.Countries) != 0 {
		t.Error("empty panels should stay empty")
	}
	if len(s.Statuses) != 3 || s.Statuses[0].Class != "2xx" || s.Statuses[1].Tone != "is-muted" || s.Statuses[2].Tone != "is-warning" {
		t.Errorf("statuses = %+v", s.Statuses)
	}
	if strings.Join(s.AxisLabels, ",") != "24 Oct,25 Oct,26 Oct" || !strings.HasPrefix(s.ChartVisitors, "M0.0,") {
		t.Errorf("chart = %q %v", s.ChartVisitors, s.AxisLabels)
	}
	if _, err := parseReport([]byte("<html>nothing here</html>")); err == nil {
		t.Error("expected an error for a page without json_data")
	}
}
