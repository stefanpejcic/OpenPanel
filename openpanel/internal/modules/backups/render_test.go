package backups

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "backups": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderBackupsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("no target", func(t *testing.T) {
		data := BackupsPageData{LayoutData: baseLayout(mgr, "/backups")}
		w := httptest.NewRecorder()
		if err := backupsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Connect a Destination") {
			t.Error("expected step 1 prompt")
		}
	})

	t.Run("target with credentials, service inactive", func(t *testing.T) {
		data := BackupsPageData{LayoutData: baseLayout(mgr, "/backups"), Target: "s3", HasCredentials: true, ServiceActive: false}
		w := httptest.NewRecorder()
		if err := backupsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Current destination") {
			t.Error("expected destination-configured step")
		}
		if !strings.Contains(body, "Edit Backup Limits") {
			t.Error("expected step 3 prompt when service inactive")
		}
	})

	t.Run("fully configured", func(t *testing.T) {
		data := BackupsPageData{LayoutData: baseLayout(mgr, "/backups"), Target: "ssh", HasCredentials: true, ServiceActive: true}
		w := httptest.NewRecorder()
		if err := backupsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Backup service is active") {
			t.Error("expected step 3 done state")
		}
	})
}

func TestRenderBackupSettingsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("error state", func(t *testing.T) {
		data := BackupSettingsPageData{LayoutData: baseLayout(mgr, "/backups/settings"), Error: "no backup target configured"}
		w := httptest.NewRecorder()
		if err := backupSettingsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "no backup target configured") {
			t.Error("expected error message")
		}
	})

	t.Run("target with values and settings", func(t *testing.T) {
		data := BackupSettingsPageData{
			LayoutData: baseLayout(mgr, "/backups/settings"), Target: "s3",
			Values:   []KV{{Key: "AWS_S3_BUCKET_NAME", Value: "mybucket"}},
			Settings: []KV{{Key: "BACKUP_SOURCES", Value: "/backup"}, {Key: "NOTIFICATION_LEVEL", Value: "error"}},
		}
		w := httptest.NewRecorder()
		if err := backupSettingsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "AWS_S3_BUCKET_NAME") || !strings.Contains(body, "mybucket") {
			t.Error("expected target values rendered")
		}
		if !strings.Contains(body, `value="/backup"`) {
			t.Error("expected BACKUP_SOURCES select to preserve the raw value")
		}
	})
}

func TestRenderBackupDestinationsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	cards := make([]DestinationCard, len(sectionOrder))
	for i, target := range sectionOrder {
		cards[i] = DestinationCard{Target: target, Title: target, Active: target == "ssh"}
	}
	w := httptest.NewRecorder()
	data := BackupDestinationsPageData{LayoutData: baseLayout(mgr, "/backups/destination"), Active: "ssh", Cards: cards}
	if err := backupDestinationsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, target := range sectionOrder {
		if !strings.Contains(body, `value="`+target+`"`) {
			t.Errorf("expected destination card for %q", target)
		}
	}
}

func TestRenderBackupRestorePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty, not reindexing", func(t *testing.T) {
		data := BackupRestorePageData{LayoutData: baseLayout(mgr, "/backups/list")}
		w := httptest.NewRecorder()
		if err := backupRestorePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No backups indexed yet") {
			t.Error("expected empty state")
		}
	})

	t.Run("with backups", func(t *testing.T) {
		rows := []RestoreBackupRow{{
			BackupInfo: BackupInfo{BackupFile: "site_2026-01-01.tar.gz", Types: []string{"html", "mysql"}, Databases: []string{"wp_db"}, HasFiles: true},
		}}
		rows[0].JSON = jsonJS(rows[0].BackupInfo)
		data := BackupRestorePageData{LayoutData: baseLayout(mgr, "/backups/list"), Backups: rows}
		w := httptest.NewRecorder()
		if err := backupRestorePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "site_2026-01-01.tar.gz") {
			t.Error("expected backup filename")
		}
		if !strings.Contains(body, "wp_db") {
			t.Error("expected database badge")
		}
	})

	t.Run("reindex error", func(t *testing.T) {
		data := BackupRestorePageData{LayoutData: baseLayout(mgr, "/backups/list"), ReindexError: "connection refused"}
		w := httptest.NewRecorder()
		if err := backupRestorePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "connection refused") {
			t.Error("expected reindex error message")
		}
	})
}
