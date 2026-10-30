package appinstall

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
	userAllowed := map[string]bool{"dashboard": true, "python": true, "nodejs": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderInstallPage(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("python, no domains", func(t *testing.T) {
		data := InstallPageData{LayoutData: baseLayout(mgr, "/python/install"), Kind: Python, Display: displayFor(Python)}
		w := httptest.NewRecorder()
		if err := installPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
	})

	// the wizard posts the field names appinstall/install.go reads, n8n swaps the startup step for its owner account
	for _, tc := range []struct {
		kind Kind
		want []string
	}{
		{NodeJS, []string{`var kind = "nodejs"`, `ext = ".js"`, `'service_name'`, `'startup_file'`, `'requirements'`, `'custom_cmd'`, `'cpu_limit'`, `'mem_limit'`, `'pids_limit'`}},
		{N8N, []string{`var kind = "n8n"`, `'owner_first_name'`, `'owner_last_name'`, `'owner_email'`, `'owner_password'`, `value: '5678'`}},
	} {
		t.Run(tc.kind.AppType+", with domains", func(t *testing.T) {
			data := InstallPageData{
				LayoutData: baseLayout(mgr, "/"+tc.kind.AppType+"/install"), Kind: tc.kind, Display: displayFor(tc.kind),
				Domains: []appctx.Domain{{DomainID: 1, DomainURL: "example.com", Docroot: "/var/www/html/", PHPVersion: "8.3"}},
			}
			w := httptest.NewRecorder()
			if err := installPage.Render(w, 200, data); err != nil {
				t.Fatalf("Render: %v", err)
			}
			body := w.Body.String()
			if !strings.Contains(body, `<div class="op-content">`) || !strings.Contains(body, `x-data="opInstallWizard(window.opIWConfig)"`) || !strings.Contains(body, `"DomainURL":"example.com"`) {
				t.Error("expected the wizard with the domain list")
			}
			for _, want := range tc.want {
				if !strings.Contains(body, want) {
					t.Errorf("expected %s in the install page", want)
				}
			}
		})
	}
}
