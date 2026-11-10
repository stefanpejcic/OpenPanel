package dashboard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func TestParseContainerRows(t *testing.T) {
	rows := parseContainerRows("nginx|running|docker.io/nginx:latest\nredis|exited|redis:7\n\nbad line\n")
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0].Name != "nginx" || rows[0].State != "Running" || rows[0].Sub != "docker.io/nginx:latest" {
		t.Errorf("row 0 = %+v", rows[0])
	}
	if rows[1].State != "Stopped" {
		t.Errorf("row 1 state = %q, want Stopped", rows[1].State)
	}
}

func TestHumanSize(t *testing.T) {
	for in, want := range map[int64]string{0: "0 B", 2048: "2 KB", 1536: "1.5 KB", 50 << 20: "50 MB"} {
		if got := humanSize(in); got != want {
			t.Errorf("humanSize(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestSitesAllowed(t *testing.T) {
	if sitesAllowed(map[string]bool{"emails": true}) {
		t.Error("an email-only plan has no sites")
	}
	if !sitesAllowed(map[string]bool{"nodejs": true}) {
		t.Error("nodejs counts as sites")
	}
}

func TestRenderDashboardAltList(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	allowed := map[string]bool{"dashboard": true, "docker": true}
	base := DashboardPageData{
		LayoutData: web.LayoutData{
			Title: "Dashboard", CSRFToken: "x", PanelDir: "ltr", MenuStyle: "modern",
			NavItems: web.BuildSidebarNav(allowed, nil, "/dashboard"), UserAllowed: allowed,
			UserAllowedJSON: web.UserAllowedList(allowed), CurrentUsername: "u", RequestPath: "/dashboard",
			T: mgr.Translator("en"),
		},
		View: "overview",
	}

	alt := base
	alt.AltList = &DashboardList{Kind: "containers", Title: "Containers", Href: "/containers", Icon: "ri-stack-line", NameLabel: "Service", HasState: true, Empty: "No containers yet",
		Rows: []DashboardListRow{{Name: "redis-box", Sub: "redis:7", State: "Running", Href: "/containers"}}}
	w := httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, alt); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"redis-box", "redis:7", ">Containers<", `href="/containers"`} {
		if !strings.Contains(body, want) {
			t.Errorf("stand-in card missing %q", want)
		}
	}
	if strings.Contains(body, "No websites yet") {
		t.Error("websites card shown on a plan without sites")
	}

	none := base
	w = httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, none); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(w.Body.String(), `data-widget="sites" x-init`) {
		t.Error("a plan with none of the fallbacks should have no card there")
	}
}

func TestLimitsFull(t *testing.T) {
	d := DashboardPageData{MainDomainsCount: 3, DomainsLimit: 3, UserWebsitesCount: 6, WebsitesLimit: 10}
	if !d.DomainsFull() || d.SitesFull() {
		t.Errorf("DomainsFull=%v SitesFull=%v, want true false", d.DomainsFull(), d.SitesFull())
	}
	d.DomainsLimit = 0
	if d.DomainsFull() {
		t.Error("a limit of 0 is unlimited")
	}
}

func TestRenderAddButtonsAtLimit(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	allowed := map[string]bool{"dashboard": true, "domains": true, "autoinstaller": true, "wordpress": true}
	data := DashboardPageData{
		LayoutData: web.LayoutData{Title: "Dashboard", CSRFToken: "x", PanelDir: "ltr", MenuStyle: "modern",
			NavItems: web.BuildSidebarNav(allowed, nil, "/dashboard"), UserAllowed: allowed,
			UserAllowedJSON: web.UserAllowedList(allowed), CurrentUsername: "u", RequestPath: "/dashboard", T: mgr.Translator("en")},
		View: "overview", ShowSites: true,
		MainDomainsCount: 2, DomainsLimit: 2, UserWebsitesCount: 1, WebsitesLimit: 5,
	}
	w := httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "You have used all 2 domains on your plan.") || strings.Contains(body, `href="/domains/new"`) {
		t.Error("Add domain should be disabled with a reason at the limit")
	}
	if !strings.Contains(body, `href="/auto-installer"`) {
		t.Error("New website should stay a link below the limit")
	}
}

func TestEmailRowsSortedByUsage(t *testing.T) {
	rows := emailRows([]string{
		"* small@a.tld ( 7.0K / 10G ) [0%]",
		"* big@a.tld ( 28M / 2.0G ) [1%]",
		"* empty@b.tld ( 0 / 1.0G ) [0%]",
	})
	if len(rows) != 3 || rows[0].Name != "big@a.tld" || rows[0].Meta != "28M" || rows[0].Sub != "a.tld" || rows[2].Name != "empty@b.tld" {
		t.Errorf("rows = %+v", rows)
	}
}

func TestByTraffic(t *testing.T) {
	sites := []UserWebsite{{SiteName: "new.tld"}, {SiteName: "busy.tld"}, {SiteName: "old.tld"}}
	got := byTraffic(sites, map[string]int{"busy.tld": 900, "old.tld": 10})
	if got[0].SiteName != "busy.tld" || got[1].SiteName != "old.tld" || got[2].SiteName != "new.tld" {
		t.Errorf("order = %s %s %s", got[0].SiteName, got[1].SiteName, got[2].SiteName)
	}
}

func TestContainerRowsRunningFirst(t *testing.T) {
	rows := parseContainerRows("a|exited|x\nb|running|y\n")
	if rows[0].Name != "b" {
		t.Errorf("running container should come first, got %s", rows[0].Name)
	}
}

func TestRenderDatabaseTypes(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	allowed := map[string]bool{"dashboard": true, "mysql": true, "postgresql": true}
	data := DashboardPageData{
		LayoutData: web.LayoutData{Title: "Dashboard", CSRFToken: "x", PanelDir: "ltr", MenuStyle: "modern",
			NavItems: web.BuildSidebarNav(allowed, nil, "/dashboard"), UserAllowed: allowed,
			UserAllowedJSON: web.UserAllowedList(allowed), CurrentUsername: "u", RequestPath: "/dashboard", T: mgr.Translator("en")},
		View: "overview",
		AltList: &DashboardList{Kind: "databases", Title: "Databases", Href: "/mysql", Icon: "ri-database-2-line", NameLabel: "Database", MetaLabel: "Size", HasType: true,
			Rows: []DashboardListRow{dbRow("shop", "mariadb", "/mysql", 5<<20), dbRow("analytics", "postgresql", "/postgresql", 1<<20)}},
	}
	w := httptest.NewRecorder()
	if err := dashboardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{">Type<", `aria-label="MariaDB"`, "PostgreSQL", "5 MB", "analytics"} {
		if !strings.Contains(body, want) {
			t.Errorf("databases card missing %q", want)
		}
	}
}

func TestDashboardViewDefault(t *testing.T) {
	r := httptest.NewRequest("GET", "/dashboard", nil)
	if got := dashboardView(r); got != "overview" {
		t.Errorf("default view = %q, want overview", got)
	}
	r = httptest.NewRequest("GET", "/dashboard?view=tools", nil)
	if got := dashboardView(r); got != "tools" {
		t.Errorf("?view=tools gave %q", got)
	}
	r = httptest.NewRequest("GET", "/dashboard", nil)
	r.AddCookie(&http.Cookie{Name: "dashboard_view", Value: "tools"})
	if got := dashboardView(r); got != "tools" {
		t.Errorf("saved tools view gave %q", got)
	}
}
