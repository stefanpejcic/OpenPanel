package processmanager

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "process_manager": true, "services": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderProcessManagerPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty", func(t *testing.T) {
		data := ProcessManagerPageData{LayoutData: baseLayout(mgr, "/process-manager")}
		w := httptest.NewRecorder()
		if err := processManagerPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No processes") {
			t.Error("expected empty state text")
		}
	})

	t.Run("with processes", func(t *testing.T) {
		data := ProcessManagerPageData{
			LayoutData: baseLayout(mgr, "/process-manager"),
			ProcessData: []Process{
				{Container: "mysql", UID: "0", PID: "123", PPID: "1", C: "0.5", STIME: "12:00", TTY: "?", TIME: "00:00:01", CMD: "mysqld"},
			},
		}
		w := httptest.NewRecorder()
		if err := processManagerPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "mysqld") || !strings.Contains(body, "services/mysql") {
			t.Error("expected process row and services link in body")
		}
	})

	t.Run("long command truncated", func(t *testing.T) {
		longCmd := strings.Repeat("x", 150)
		data := ProcessManagerPageData{
			LayoutData:  baseLayout(mgr, "/process-manager"),
			ProcessData: []Process{{Container: "mysql", PID: "5", CMD: longCmd}},
		}
		w := httptest.NewRecorder()
		if err := processManagerPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "View full command") {
			t.Error("expected truncation link for long command")
		}
	})
}
