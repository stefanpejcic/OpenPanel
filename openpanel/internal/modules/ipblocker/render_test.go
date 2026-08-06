package ipblocker

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "ip_blocker": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderIPBlockerPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no ips", func(t *testing.T) {
		data := IPBlockerPageData{LayoutData: baseLayout(mgr, "/security/ip-blocker")}
		w := httptest.NewRecorder()
		if err := ipBlockerPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "IP Blocker") {
			t.Error("expected page title in body")
		}
	})

	t.Run("with ips", func(t *testing.T) {
		data := IPBlockerPageData{LayoutData: baseLayout(mgr, "/security/ip-blocker"), IPs: []string{"1.2.3.4", "10.0.0.0/24"}}
		w := httptest.NewRecorder()
		if err := ipBlockerPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "1.2.3.4") || !strings.Contains(body, "10.0.0.0/24") {
			t.Error("expected IPs in body")
		}
	})
}
