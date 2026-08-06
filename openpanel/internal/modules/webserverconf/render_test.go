package webserverconf

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "webserver_conf": true}
	return web.LayoutData{
		Title: "Nginx Configuration Editor", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderWebserverConfPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	data := WebserverConfPageData{
		LayoutData: baseLayout(mgr, "/server/webserver_conf"), Service: "nginx", Path: "nginx.conf",
		CanRestoreDefault: true, FileContent: "server { listen 80; }",
	}
	w := httptest.NewRecorder()
	if err := webserverConfPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "server { listen 80; }") {
		t.Error("expected file content in body")
	}
	if !strings.Contains(body, "Restore Default") {
		t.Error("expected restore-default button when CanRestoreDefault is true")
	}
}

func TestRenderWebserverConfPageNoRestoreOption(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	data := WebserverConfPageData{
		LayoutData: baseLayout(mgr, "/server/webserver_conf"), Service: "nginx", Path: "nginx.conf",
		CanRestoreDefault: false, FileContent: "",
	}
	w := httptest.NewRecorder()
	if err := webserverConfPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(w.Body.String(), "Restore Default") {
		t.Error("did not expect restore-default button when CanRestoreDefault is false")
	}
}
