package websites

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{
		"dashboard": true, "websites": true, "wordpress": true, "filemanager": true, "autoinstaller": true,
	}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
		PageTabs: web.BuildPageTabs(userAllowed, nil, path, web.TabContext{}),
	}
}

func TestSitesPageParses(t *testing.T) {
	if sitesPage == nil {
		t.Fatal("sitesPage is nil")
	}
}

func TestRenderSitesPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("with grouped sites", func(t *testing.T) {
		groups := []SiteGroup{
			{Type: "wordpress", Sites: []SiteRow{
				{SiteName: "example.com", DomainID: 1, AdminEmail: "admin@example.com", Version: "6.5", CreatedDate: "2026-01-01", Type: "WordPress", Container: "", Ports: "", PHPVersion: "8.3"},
				{SiteName: "blog.example.com", DomainID: 3, AdminEmail: "editor@example.com", Version: "6.6.1", CreatedDate: "2026-03-14", Type: "WordPress"},
			}},
			{Type: "nodejs", Sites: []SiteRow{
				{SiteName: "app.example.com", DomainID: 4, CreatedDate: "2026-05-20", Type: "nodejs", Container: "nodeapp", Ports: "3000"},
			}},
			{Type: "php", Sites: []SiteRow{
				{SiteName: "shared.example.com", DomainID: 5, CreatedDate: "2026-06-01", Type: "php", Version: "8.5", PHPVersion: "8.5"},
			}},
			{Type: "static", Sites: []SiteRow{
				{SiteName: "static.example.com", DomainID: 2, AdminEmail: "", Version: "", CreatedDate: "2026-01-02", Type: "static", IsStatic: true},
			}},
		}
		data := SitesPageData{LayoutData: baseLayout(mgr, "/sites"), Groups: groups, ViewMode: "table", SiteCount: 5, ContainerCount: 1, WebsitesLimit: 40, WebsitesPct: 12,
			FreeDomains: []string{"waf.example.com", "shop.example.org"}}
		w := httptest.NewRecorder()
		if err := sitesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		// no key: every speed cell links to the key editor and nothing asks PageSpeed
		if body := w.Body.String(); strings.Contains(body, `class="speed_data"`) || !strings.Contains(body, "/file-manager/edit-file/pagespeed_api_key.txt?new=true") {
			t.Error("without a PageSpeed key the speed column should link to the key editor")
		}
		data.HasPagespeedKey = true
		w = httptest.NewRecorder()
		if err := sitesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render with key: %v", err)
		}
		if body := w.Body.String(); !strings.Contains(body, `class="speed_data"`) || strings.Contains(body, "op-speed-setup") {
			t.Error("with a PageSpeed key the speed column should load scores")
		}
	})

	t.Run("no sites", func(t *testing.T) {
		data := SitesPageData{LayoutData: baseLayout(mgr, "/sites"), ViewMode: "table"}
		w := httptest.NewRecorder()
		if err := sitesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})
}

func TestRenderWebsiteBuilderPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := WebsiteBuilderPageData{
		pageData: pageData{
			LayoutData: baseLayout(mgr, "/website"), CurrentDomain: "example.com",
			Docroot: "/var/www/html/example.com", PagespeedAPIKeyValue: "",
		},
		Container: ContainerInfo{ID: 1, SiteName: "example.com", Type: "websitebuilder", CreatedDate: "2026-01-01"},
	}
	w := httptest.NewRecorder()
	if err := websiteBuilderPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
}

func TestRenderPythonNodeAppsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	render := func(t *testing.T, pm2Status, appType string) {
		data := PythonNodeAppsPageData{
			pageData: pageData{
				LayoutData: baseLayout(mgr, "/website"), CurrentDomain: "example.com",
				Docroot: "/var/www/html/example.com",
			},
			Container: ContainerInfo{
				ID: 1, SiteName: "example.com", Type: appType, Container: "myapp_python",
				Version: "3.12", CreatedDate: "2026-01-01",
			},
			PM2Data:        map[string]string{"prefix": "MYAPP_PY_", "status": pm2Status},
			Service:        "myapp",
			Type:           appType,
			PM2Status:      pm2Status,
			CPU:            "1",
			RAM:            "1",
			PIDs:           "100",
			CurrentVersion: "3.12",
		}
		w := httptest.NewRecorder()
		if err := pythonNodeAppsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	}

	t.Run("python running", func(t *testing.T) { render(t, "true", "python") })
	t.Run("nodejs stopped", func(t *testing.T) { render(t, "false", "nodejs") })
	t.Run("unknown status", func(t *testing.T) { render(t, "unknown", "python") })
}

// every site manager page with a Git tab shows it only when the git module is allowed
func TestSiteManagerGitTab(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	for _, gitOn := range []bool{true, false} {
		layout := baseLayout(mgr, "/website")
		layout.UserAllowed["git"] = gitOn
		pd := pageData{LayoutData: layout, CurrentDomain: "example.com", Docroot: "/var/www/html/example.com"}

		pages := map[string]func(*httptest.ResponseRecorder) error{
			"wordpress": func(w *httptest.ResponseRecorder) error {
				return wpSinglePage.Render(w, 200, WPSinglePageData{pageData: pd, Container: ContainerInfo{SiteName: "example.com"}})
			},
			"php": func(w *httptest.ResponseRecorder) error {
				return phpAppPage.Render(w, 200, PHPAppPageData{pageData: pd})
			},
		}
		for cmsType, p := range cmsPages {
			cmsType, p := cmsType, p
			pages[cmsType] = func(w *httptest.ResponseRecorder) error {
				return p.page.Render(w, 200, CMSAppPageData{pageData: pd, App: web.NewAppNames(cmsType, p.name), DBInfo: map[string]string{}})
			}
		}

		for name, render := range pages {
			w := httptest.NewRecorder()
			if err := render(w); err != nil {
				t.Fatalf("%s: Render: %v", name, err)
			}
			body := w.Body.String()
			hasTab := strings.Contains(body, `id="git-tab"`) && strings.Contains(body, `id="git" role="tabpanel"`)
			if hasTab != gitOn {
				t.Errorf("%s: git tab shown = %v, want %v", name, hasTab, gitOn)
			}
			// the vertical section rail has room for every tab, only the horizontal bar folds into More
			if !strings.Contains(body, "op-mgr-rail") && !strings.Contains(body, "data-tab-overflow") {
				t.Errorf("%s: tab bar is missing data-tab-overflow", name)
			}
		}
	}
}

// the WordPress debugging tab shows the site's latest errors only when error_logs is allowed, below the debug.log card
func TestWPDebuggingErrorLogs(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	for _, on := range []bool{true, false} {
		layout := baseLayout(mgr, "/website")
		layout.UserAllowed["error_logs"] = on
		pd := pageData{LayoutData: layout, CurrentDomain: "example.com/blog", Docroot: "/var/www/html/example.com/blog"}
		w := httptest.NewRecorder()
		if err := wpSinglePage.Render(w, 200, WPSinglePageData{pageData: pd, Container: ContainerInfo{SiteName: "example.com/blog"}}); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		shown := strings.Contains(body, `href="/error-logs?domain=example.com%2fblog&exact=1"`) && strings.Contains(body, "exact: 1, limit: 10, format: 'json'")
		if shown != on {
			t.Errorf("error logs block shown = %v, want %v", shown, on)
		}
		if on && strings.Index(body, "exact: 1, limit: 10") < strings.Index(body, `x-data="wpDebugLog()"`) {
			t.Error("error logs block should come after the debug.log card")
		}
	}
}

func TestRenderN8nAppPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	pd := pageData{LayoutData: baseLayout(mgr, "/website"), CurrentDomain: "flows.example.com", Docroot: "/var/www/html/flows.example.com"}
	for _, status := range []string{"true", "false"} {
		data := N8nAppPageData{
			pageData:  pd,
			Container: ContainerInfo{SiteName: "flows.example.com", Type: "n8n", Version: "1.64.0", CreatedDate: "2026-04-02", Port: "5678"},
			Service:   "n8n", PM2Status: status, CPU: "1", RAM: "1", CurrentVersion: "1.64.0",
			EnvVars: "N8N_HOST=flows.example.com\nGENERIC_TIMEZONE=UTC",
		}
		w := httptest.NewRecorder()
		if err := n8nAppPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render (%s): %v", status, err)
		}
		if !strings.Contains(w.Body.String(), "N8N_HOST=flows.example.com") {
			t.Errorf("expected env vars in the page (%s)", status)
		}
	}
}

func TestRenderWebsiteBuilderInstallPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	for _, domains := range [][]appctx.Domain{
		{{DomainID: 1, DomainURL: "example.com", Docroot: "/var/www/html/example.com"}, {DomainID: 2, DomainURL: "landing.example.com", Docroot: "/var/www/html/landing.example.com"}},
		nil,
	} {
		data := WebsiteBuilderInstallPageData{LayoutData: baseLayout(mgr, "/website-builder/install"), Domains: domains}
		w := httptest.NewRecorder()
		if err := websiteBuilderInstallPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if domains != nil && (!strings.Contains(body, `<div class="op-content">`) || !strings.Contains(body, `endpoint: '/website-builder/install'`) || !strings.Contains(body, `"DomainURL":"landing.example.com"`) || !strings.Contains(body, "/website-builder/edit?domain=")) {
			t.Error("expected the wizard posting to /website-builder/install and opening the builder when done")
		}
		if domains == nil && !strings.Contains(body, "/domains/new") {
			t.Error("expected the no-domains card")
		}
	}
}
