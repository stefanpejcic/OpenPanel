package appinstall

import (
	"net/http/httptest"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "python": true, "nodejs": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderInstallPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("python, no domains", func(t *testing.T) {
		data := InstallPageData{LayoutData: baseLayout(mgr, "/python/install"), Kind: Python, Display: displayFor(Python)}
		w := httptest.NewRecorder()
		if err := installPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})

	t.Run("nodejs, with domains", func(t *testing.T) {
		data := InstallPageData{
			LayoutData: baseLayout(mgr, "/nodejs/install"), Kind: NodeJS, Display: displayFor(NodeJS),
			Domains: []appctx.Domain{{DomainID: 1, DomainURL: "example.com", Docroot: "/var/www/html/", PHPVersion: "8.3"}},
		}
		w := httptest.NewRecorder()
		if err := installPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})
}
