package crons

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "crons": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
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
		LayoutData: baseLayout(mgr, "/cronjobs/editor"), View: "code", Service: "cron",
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
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("with containers", func(t *testing.T) {
		data := CronjobsNewPageData{
			LayoutData: baseLayout(mgr, "/cronjobs/new"), Service: "cron", Containers: []string{"mariadb", "php-fpm-8.5"},
			PHPContainer: "php-fpm-8.5", DBContainer: "mariadb", JobNames: []string{"Nightly"}, TimeZone: "Europe/Belgrade",
		}
		w := httptest.NewRecorder()
		if err := cronjobsNewPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{
			`action="/cronjobs/save"`, `name="command"`, `name="comment"`, `name="schedule"`, `name="container"`, `name="no_overlap"`,
			`<option value="php-fpm-8.5"`, `op-cf-presets`, `Europe/Belgrade`, `/cronjobs/new?`, `Schedule Cron Job`, `Added to crons.ini`, `Nightly`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("missing %q", want)
			}
		}
		if !strings.Contains(body, `<option value="php-fpm-8.5" :selected="container === &#34;php-fpm-8.5&#34;" selected>`) {
			t.Error("the default PHP container should be preselected")
		}
	})

	t.Run("no containers", func(t *testing.T) {
		data := CronjobsNewPageData{LayoutData: baseLayout(mgr, "/cronjobs/new"), Service: "cron", TimeZone: "UTC"}
		w := httptest.NewRecorder()
		if err := cronjobsNewPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "No containers to run jobs in") || strings.Contains(body, `action="/cronjobs/save"`) {
			t.Error("expected the empty state without a form")
		}
	})
}

func TestRenderCronjobsTableCreated(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := CronjobsPageData{
		LayoutData: baseLayout(mgr, "/cronjobs"), View: "table", Service: "cron", Created: "fresh",
		CronJobs: []CronJob{{Comment: "old", Schedule: "@daily", Container: "php", Command: "a"}, {Comment: "fresh", Schedule: "@hourly", Container: "php", Command: "b"}},
	}
	w := httptest.NewRecorder()
	if err := cronjobsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if strings.Count(body, "op-cf-new-row") != 1 || strings.Count(body, "op-cf-tag-new") != 1 {
		t.Error("expected exactly the created row highlighted")
	}
}

func TestRenderCronjobsLogsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := CronjobsPageData{
		LayoutData: baseLayout(mgr, "/cronjobs/logs"), View: "logs", Service: "cron", LogJob: "db dump",
		CronJobs: []CronJob{{Schedule: "@daily", Command: "mysqldump x", Container: "mysql", Comment: "db dump"}},
	}
	w := httptest.NewRecorder()
	if err := cronjobsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{`cronLogs(&#34;db dump&#34;)`, "Cron Job Logs", `<option value="db dump">`} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
	if strings.Contains(body, "Switch to File Editor") || strings.Contains(body, "Need Help?") {
		t.Error("logs tab shouldn't show the editor switch or help button")
	}
}

func TestRenderCronjobsTableHasLogLinks(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := CronjobsPageData{
		LayoutData: baseLayout(mgr, "/cronjobs"), View: "table", Service: "cron",
		CronJobs: []CronJob{{Schedule: "@daily", Command: "mysqldump x", Container: "mysql", Comment: "db dump"}},
	}
	w := httptest.NewRecorder()
	if err := cronjobsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if strings.Contains(body, "Switch to File Editor") {
		t.Error("the switch button should be gone")
	}
	if !strings.Contains(body, `href="/cronjobs/logs?job=db%20dump"`) {
		t.Error("each job should link to its logs tab")
	}
}

func TestRenderCronjobsTableRows(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	jobs := []CronJob{
		{Comment: "Nightly database dump", Schedule: "0 0 3 * * *", Container: "mysql", Command: "mysqldump --all-databases > /backups/db.sql"},
		{Comment: "WordPress cron", Schedule: "0 */5 * * * *", Container: "php-fpm-8.3", Command: "php /var/www/html/example.com/wp-cron.php", NoOverlap: true},
		{Comment: "Shop cleanup", Schedule: "0 30 2 * * 0", Container: "node-shop", Command: "node /app/scripts/cleanup.js"},
		{Comment: "Rebuild sitemap", Schedule: "0 0 */6 * * *", Container: "nginx", Command: "curl -fsS https://example.com/sitemap/rebuild", Disabled: true},
	}
	data := CronjobsPageData{
		LayoutData: baseLayout(mgr, "/cronjobs"), View: "table", Service: "cron", CronJobs: jobs,
		Services: []string{"mysql", "php-fpm-8.3", "node-shop", "nginx"}, Summary: cronSummary(jobs, time.Now(), time.UTC),
		Timezones: []string{"UTC", "Europe/Belgrade"},
	}
	w := httptest.NewRecorder()
	if err := cronjobsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"Every day at 03:00", "Sundays at 02:30", "Paused", `action="/cronjobs/edit"`, `action="/cronjobs/toggle"`, `action="/cronjobs/delete"`, `action="/cronjobs/timezone"`, `id="editForm"`, `form="cron-edit-1"`, "4 jobs, 3 active"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
}
