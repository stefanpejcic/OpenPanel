package processmanager

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "process_manager": true, "services": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderProcessManagerPage(t *testing.T) {
	webtest.ModernOnly(t)
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
		if !strings.Contains(w.Body.String(), "Show full command") {
			t.Error("expected the full command toggle for long command")
		}
	})
	t.Run("summary and program names", func(t *testing.T) {
		data := ProcessManagerPageData{
			LayoutData: baseLayout(mgr, "/process-manager"),
			ProcessData: []Process{
				{Container: "nginx", UID: "root", PID: "1", PPID: "0", C: "0", STIME: "07:48", TTY: "?", TIME: "00:00:00", CMD: "nginx: master process nginx -g daemon off;"},
				{Container: "php-fpm-8.3", UID: "www-data", PID: "41", PPID: "1", C: "3", STIME: "07:48", TTY: "?", TIME: "00:01:12", CMD: "php-fpm: pool www"},
				{Container: "mariadb", UID: "mysql", PID: "1", PPID: "0", C: "1", STIME: "07:47", TTY: "?", TIME: "00:00:31", CMD: "/usr/sbin/mariadbd --user=mysql"},
			},
		}
		w := httptest.NewRecorder()
		if err := processManagerPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{"3 processes · 3 containers", ">/usr/sbin/mariadbd --user=mysql<", "3% CPU in php-fpm-8.3", `<option value="nginx">`, "sendPID("} {
			if !strings.Contains(body, want) {
				t.Errorf("process page missing %q", want)
			}
		}
	})
	t.Run("quote in command keeps the search expression valid", func(t *testing.T) {
		data := ProcessManagerPageData{
			LayoutData:  baseLayout(mgr, "/process-manager"),
			ProcessData: []Process{{Container: "mongodb", PID: "7", CMD: "mongosh --eval 'db.ping()'"}},
		}
		w := httptest.NewRecorder()
		if err := processManagerPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if strings.Contains(w.Body.String(), "'mongosh --eval") {
			t.Error("command still wrapped in single quotes")
		}
		if !strings.Contains(w.Body.String(), "&#34;mongosh --eval &#39;db.ping()&#39;&#34;.toLowerCase()") {
			t.Error("command not embedded as a JSON string")
		}
	})
}
