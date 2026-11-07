package phpapp

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// a missing partial renders as an empty 200, so check the body and not only the error
func TestRenderInstallPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	userAllowed := map[string]bool{"dashboard": true, "php": true}
	layout := web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, "/php/install"), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: "/php/install", AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}

	cases := []struct {
		name    string
		domains []appctx.Domain
		want    string
	}{
		{"no domains", nil, "/domains/new"},
		{"with domains", []appctx.Domain{{DomainID: 1, DomainURL: "example.com", Docroot: "/var/www/html/", PHPVersion: "8.3"}}, `id="installForm"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			if err := installPage.Render(w, 200, InstallPageData{LayoutData: layout, Domains: c.domains}); err != nil {
				t.Fatalf("Render: %v", err)
			}
			if !strings.Contains(w.Body.String(), c.want) {
				t.Errorf("page has no %q, got %d bytes", c.want, w.Body.Len())
			}
		})
	}
}
