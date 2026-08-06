package backupwizard

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "backup_wizard": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderBackupWizardPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no backups", func(t *testing.T) {
		data := BackupWizardPageData{LayoutData: baseLayout(mgr, "/backup-wizard"), BackupIncludes: []string{"The home directory"}}
		w := httptest.NewRecorder()
		if err := backupWizardPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No backups found") {
			t.Error("expected empty-state message")
		}
	})

	t.Run("in progress", func(t *testing.T) {
		data := BackupWizardPageData{
			LayoutData: baseLayout(mgr, "/backup-wizard"), InProgress: true,
			InProgressStarted: "2026-01-01 00:00:00", InProgressSize: "12.3 MB",
			Backups: []BackupFile{{Name: "site.tar.gz", Size: "5.0 MB", Mtime: "2026-01-01 00:00:00", InProgress: true}},
		}
		w := httptest.NewRecorder()
		if err := backupWizardPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "site.tar.gz") {
			t.Error("expected backup row in rendered page")
		}
		if !strings.Contains(body, "disabled") {
			t.Error("expected the Generate Backup button to be disabled while in progress")
		}
	})
}

func TestFormatSize(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0.0 B"},
		{1023, "1023.0 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}
	for _, c := range cases {
		if got := formatSize(c.in); got != c.want {
			t.Errorf("formatSize(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsWithin(t *testing.T) {
	cases := []struct {
		candidate, base string
		want            bool
	}{
		{"/home/u/_backups/x.tar.gz", "/home/u/_backups", true},
		{"/home/u/_backups", "/home/u/_backups", true},
		{"/etc/passwd", "/home/u/_backups", false},
	}
	for _, c := range cases {
		if got := isWithin(c.candidate, c.base); got != c.want {
			t.Errorf("isWithin(%q, %q) = %v, want %v", c.candidate, c.base, got, c.want)
		}
	}
}
