package websites

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
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
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("with grouped sites", func(t *testing.T) {
		groups := []SiteGroup{
			{Type: "wordpress", Sites: []SiteRow{
				{SiteName: "example.com", DomainID: 1, AdminEmail: "admin@example.com", Version: "6.5", CreatedDate: "2026-01-01", Type: "WordPress", Container: "", Ports: ""},
			}},
			{Type: "static", Sites: []SiteRow{
				{SiteName: "static.example.com", DomainID: 2, AdminEmail: "", Version: "", CreatedDate: "2026-01-02", Type: "static", IsStatic: true},
			}},
		}
		data := SitesPageData{LayoutData: baseLayout(mgr, "/sites"), Groups: groups, ViewMode: "table"}
		w := httptest.NewRecorder()
		if err := sitesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
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
