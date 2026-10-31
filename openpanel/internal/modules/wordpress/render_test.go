package wordpress

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
		"dashboard": true, "wordpress": true, "filemanager": true,
	}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestPagesParse(t *testing.T) {
	if listPage == nil {
		t.Fatal("listPage is nil")
	}
	if installPage == nil {
		t.Fatal("installPage is nil")
	}
}

func TestRenderListPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("with sites, table view", func(t *testing.T) {
		data := ListPageData{
			LayoutData: baseLayout(mgr, "/wordpress"),
			Sites: []SiteRow{
				{SiteName: "example.com", DomainID: 1, AdminEmail: "admin@example.com", Version: "6.5", CreatedDate: "2026-01-01", Type: "WordPress", ID: 1},
			},
			ViewMode: "table",
		}
		w := httptest.NewRecorder()
		if err := listPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})

	t.Run("with sites, cards view", func(t *testing.T) {
		data := ListPageData{
			LayoutData: baseLayout(mgr, "/wordpress"),
			Sites: []SiteRow{
				{SiteName: "example.com", DomainID: 1, AdminEmail: "admin@example.com", Version: "6.5", CreatedDate: "2026-01-01", Type: "WordPress", ID: 1},
			},
			ViewMode: "cards",
		}
		w := httptest.NewRecorder()
		if err := listPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})

	t.Run("no sites", func(t *testing.T) {
		data := ListPageData{LayoutData: baseLayout(mgr, "/wordpress"), ViewMode: "cards"}
		w := httptest.NewRecorder()
		if err := listPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})
}

func TestRenderInstallPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("with domains", func(t *testing.T) {
		data := InstallPageData{
			LayoutData: baseLayout(mgr, "/wordpress/install"),
			Domains: []appctx.Domain{
				{DomainID: 1, Docroot: "/var/www/html/example.com", DomainURL: "example.com"},
			},
		}
		w := httptest.NewRecorder()
		if err := installPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		// the wizard posts the same field names wordpress/install.go reads
		for _, want := range []string{`<div class="op-content">`, `x-data="opInstallWizard(window.opIWConfig)"`, `endpoint: '/wordpress/install'`, `"DomainURL":"example.com"`, `'website_name'`, `'site_description'`, `'wordpress_version'`, `name: 'db_prefix'`, `name="csrf_token"`, `js/install-wizard.js`, `adminPath: '/wp-admin/'`} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %s in the install page", want)
			}
		}
	})

	t.Run("no domains", func(t *testing.T) {
		data := InstallPageData{LayoutData: baseLayout(mgr, "/wordpress/install")}
		w := httptest.NewRecorder()
		if err := installPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if body := w.Body.String(); !strings.Contains(body, "/domains/new") || strings.Contains(body, "opInstallWizard(") {
			t.Error("expected the no-domains card instead of the wizard")
		}
	})

	t.Run("enforced mysql prefix", func(t *testing.T) {
		data := InstallPageData{
			LayoutData: baseLayout(mgr, "/wordpress/install"),
			Domains:    []appctx.Domain{{DomainID: 1, Docroot: "/var/www/html/example.com", DomainURL: "example.com"}},
		}
		data.MySQLPrefix = "john_"
		w := httptest.NewRecorder()
		if err := installPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), `mysqlPrefix: "john_"`) {
			t.Error("expected the john_ prefix next to the db fields")
		}
	})
}
