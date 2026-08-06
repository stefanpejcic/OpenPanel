package autoinstaller

import (
	"net/http/httptest"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{
		"dashboard": true, "autoinstaller": true, "wordpress": true, "drupal": true,
		"website_builder": true, "mautic": true, "nodejs": true, "python": true,
	}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderAutoinstallerPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("all cards, zero counts", func(t *testing.T) {
		counts := map[string]int{"wordpress": 0, "drupal": 0, "sitebuilder": 0, "mautic": 0, "node": 0, "python": 0}
		data := AutoinstallerPageData{LayoutData: baseLayout(mgr, "/auto-installer"), Counts: counts}
		w := httptest.NewRecorder()
		if err := autoinstallerPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})

	t.Run("with counts and no feature access", func(t *testing.T) {
		layout := baseLayout(mgr, "/auto-installer")
		layout.UserAllowed = map[string]bool{"dashboard": true, "wordpress": true}
		counts := map[string]int{"wordpress": 3, "drupal": 1, "sitebuilder": 0, "mautic": 0, "node": 2, "python": 1}
		data := AutoinstallerPageData{LayoutData: layout, Counts: counts}
		w := httptest.NewRecorder()
		if err := autoinstallerPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})
}
