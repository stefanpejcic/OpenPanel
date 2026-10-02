package fixpermissions

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func TestRenderFixPermissionsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	userAllowed := map[string]bool{"dashboard": true, "fix_permissions": true}
	layout := web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, "/fix-permissions"), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: "/fix-permissions", AdminPort: "2087", T: mgr.Translator("en"),
	}
	data := FixPermissionsPageData{LayoutData: layout}
	w := httptest.NewRecorder()
	if err := fixPermissionsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), `id="directory-select"`) {
		t.Error("expected directory input in rendered page")
	}
}

func TestIsRelativeTo(t *testing.T) {
	cases := []struct {
		path, base string
		want       bool
	}{
		{"/var/www/html", "/var/www/html", true},
		{"/var/www/html/sub", "/var/www/html", true},
		{"/etc/passwd", "/var/www/html", false},
		{"/var/www/html2", "/var/www/html", false},
	}
	for _, c := range cases {
		if got := isRelativeTo(c.path, c.base); got != c.want {
			t.Errorf("isRelativeTo(%q, %q) = %v, want %v", c.path, c.base, got, c.want)
		}
	}
}
