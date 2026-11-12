package phpapp

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestRenderInstallPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	userAllowed := map[string]bool{"dashboard": true, "websites": true, "autoinstaller": true, "php": true}
	layout := web.LayoutData{
		Title: "Install PHP Application", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, "/php/install"), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: "/php/install", AdminPort: "2087", T: mgr.Translator("en"),
		PageTabs: web.BuildPageTabs(userAllowed, nil, "/php/install", web.TabContext{}),
	}

	t.Run("with domains", func(t *testing.T) {
		data := InstallPageData{LayoutData: layout, Domains: []appctx.Domain{
			{DomainID: 1, DomainURL: "example.com", Docroot: "/var/www/html/example.com", PHPVersion: "8.3"},
			{DomainID: 2, DomainURL: "api.example.com", Docroot: "/var/www/html/api.example.com", PHPVersion: "8.2"},
		}}
		w := httptest.NewRecorder()
		if err := installPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{`<div class="op-content">`, `x-data="opInstallWizard(window.opIWConfig)"`, `endpoint: '/php/install'`, `"PHPVersion":"8.3"`, `'initial_project'`, `'autorun_composer_install'`, `'composer_optimize_autoloader'`, `name="csrf_token"`} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %s in the install page", want)
			}
		}
	})

	t.Run("no domains", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := installPage.Render(w, 200, InstallPageData{LayoutData: layout}); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "/domains/new") {
			t.Error("expected the no-domains card")
		}
	})
}
