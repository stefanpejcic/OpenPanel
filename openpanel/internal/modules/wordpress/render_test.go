package wordpress

import (
	"net/http/httptest"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{
		"dashboard": true, "wordpress": true, "filemanager": true,
	}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
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
	})

	t.Run("no domains", func(t *testing.T) {
		data := InstallPageData{LayoutData: baseLayout(mgr, "/wordpress/install")}
		w := httptest.NewRecorder()
		if err := installPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})
}
