package crons

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "crons": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderCronjobsTablePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("with jobs", func(t *testing.T) {
		data := CronjobsPageData{
			LayoutData: baseLayout(mgr, "/cronjobs"), View: "table", Service: "cron",
			Services: []string{"mysql", "nginx-proxy"},
			CronJobs: []CronJob{{Comment: "backup-db", Schedule: "0 0 1 * * *", Container: "mysql", Command: "/usr/local/bin/backup.sh"}},
		}
		w := httptest.NewRecorder()
		if err := cronjobsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "backup-db") || !strings.Contains(body, "/usr/local/bin/backup.sh") {
			t.Error("expected cron job row in body")
		}
	})

	t.Run("empty", func(t *testing.T) {
		data := CronjobsPageData{LayoutData: baseLayout(mgr, "/cronjobs"), View: "table", Service: "cron"}
		w := httptest.NewRecorder()
		if err := cronjobsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No cronjobs.") {
			t.Error("expected empty state text")
		}
	})

	t.Run("with schedule issues", func(t *testing.T) {
		data := CronjobsPageData{
			LayoutData: baseLayout(mgr, "/cronjobs"), View: "table", Service: "cron",
			CronJobs:       []CronJob{{Comment: "bad", Schedule: "garbage", Container: "mysql", Command: "x"}},
			ScheduleIssues: []ScheduleIssue{{ID: "cron-schedule:bad", Severity: "error", Message: `Schedule "garbage" is invalid`}},
		}
		w := httptest.NewRecorder()
		if err := cronjobsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "reportHealthIssues") {
			t.Error("expected health toast script in body")
		}
	})
}

func TestRenderCronjobsCodePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := CronjobsPageData{
		LayoutData: baseLayout(mgr, "/cronjobs?view=code"), View: "code", Service: "cron",
		CrontabContent: `[job-exec "x"]
schedule = @daily
container = mysql
command = touch /tmp/x`,
	}
	w := httptest.NewRecorder()
	if err := cronjobsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "job-exec") {
		t.Error("expected crontab content in body")
	}
}

func TestRenderCronjobsNewPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := CronjobsNewPageData{LayoutData: baseLayout(mgr, "/cronjobs/new"), Service: "cron", Containers: []string{"mysql", "nginx-proxy"}}
	w := httptest.NewRecorder()
	if err := cronjobsNewPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "mysql") || !strings.Contains(body, "nginx-proxy") {
		t.Error("expected container options in body")
	}
}
