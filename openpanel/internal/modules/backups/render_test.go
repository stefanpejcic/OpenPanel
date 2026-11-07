package backups

import (
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "backups": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderBackupsPage(t *testing.T) {
	webtest.ModernOnly(t)
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

	t.Run("with backups on the destination", func(t *testing.T) {
		items := restoreItems([]BackupInfo{
			{BackupFile: "testuser-2026-10-27T08-17-08.tar.gz", Types: []string{"crons", "html", "mail", "mysql", "vhosts", "dns"}, Databases: []string{"testuser_wp"}, HasFiles: true, Size: 6963},
			{BackupFile: "testuser-2026-10-26T08-17-08.tar.gz", Types: []string{"html", "vhosts"}, HasFiles: true, Size: 6477},
		}, "", nil)
		sum := BackupsSummary{Count: len(items), Latest: &items[0], TotalSize: formatSize(13440), Recent: items, Schedule: "@daily", Retention: "-1", Destination: "bkdest@10.0.2.15:/home/bkdest/backups", CPU: "1.0", RAM: "1G"}
		data := BackupsPageData{LayoutData: baseLayout(mgr, "/backups"), Target: "ssh", HasCredentials: true, ServiceActive: false, Summary: sum}
		w := httptest.NewRecorder()
		if err := backupsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{"27 Oct 2026, 08:17", "keeps every copy", "2 of 3 done", "bkdest@10.0.2.15", "Latest"} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in the overview", want)
			}
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
		for _, v := range []string{"/backup/01-mail", "/backup/01-vhosts", "/backup/01-mysql", "/backup/01-postgres"} {
			if !strings.Contains(body, `value="`+v+`"`) {
				t.Errorf("expected BACKUP_SOURCES option %s", v)
			}
		}
	})

	// a value that isn't one of the options left nothing selected, so saving any other setting sent the first option
	t.Run("values not in the dropdowns", func(t *testing.T) {
		data := BackupSettingsPageData{
			LayoutData: baseLayout(mgr, "/backups/settings"), Target: "s3",
			Settings: []KV{{Key: "BACKUP_SOURCES", Value: "/backup/mysql"}, {Key: "NOTIFICATION_LEVEL", Value: ""}},
		}
		w := httptest.NewRecorder()
		if err := backupSettingsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		selected := regexp.MustCompile(`<option value="([^"]*)" selected>`).FindAllStringSubmatch(w.Body.String(), -1)
		var got []string
		for _, m := range selected {
			got = append(got, m[1])
		}
		if !reflect.DeepEqual(got, []string{"/backup/mysql", ""}) {
			t.Errorf("selected options %q, want the saved values kept", got)
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
		data := BackupRestorePageData{LayoutData: baseLayout(mgr, "/backups/list"), PageJSON: jsonJS([]any{})}
		w := httptest.NewRecorder()
		if err := backupRestorePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No backups indexed yet") {
			t.Error("expected empty state")
		}
	})

	t.Run("with backups", func(t *testing.T) {
		items := restoreItems([]BackupInfo{{
			BackupFile: "site-2026-01-01T02-03-04.tar.gz", Types: []string{"html", "mysql"}, Databases: []string{"wp_db"}, HasFiles: true,
			Size: 48303368, DatabaseSizes: map[string]int64{"wp_db": 4504538}, DatabaseEngines: map[string]string{"wp_db": "mysql"},
		}}, "", nil)
		data := BackupRestorePageData{LayoutData: baseLayout(mgr, "/backups/list"), Backups: items, PageJSON: jsonJS(items)}
		w := httptest.NewRecorder()
		if err := backupRestorePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "site-2026-01-01T02-03-04.tar.gz") {
			t.Error("expected backup filename")
		}
		if !strings.Contains(body, "wp_db") {
			t.Error("expected database badge")
		}
		if !strings.Contains(body, "46.1 MB") || !strings.Contains(body, "4.3 MB") {
			t.Error("expected archive and database sizes")
		}
	})

	t.Run("reindex error", func(t *testing.T) {
		data := BackupRestorePageData{LayoutData: baseLayout(mgr, "/backups/list"), PageJSON: jsonJS([]any{}), ReindexError: "connection refused"}
		w := httptest.NewRecorder()
		if err := backupRestorePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "connection refused") {
			t.Error("expected reindex error message")
		}
	})
}

func TestRestoreItemsNewestFirst(t *testing.T) {
	items := restoreItems([]BackupInfo{
		{BackupFile: "u-2026-10-14T02-00-00.tar.gz", Databases: []string{"a"}, DatabaseEngines: map[string]string{"a": "postgres"}},
		{BackupFile: "u-2026-10-15T09-35-02.tar.gz", Lists: []string{"files", "mail"}},
		{BackupFile: "renamed.tar.gz"},
	}, "", nil)
	// the page enables the Choose buttons from lists
	if !reflect.DeepEqual(items[0].Lists, []string{"files", "mail"}) || items[1].Lists == nil {
		t.Errorf("lists not passed to the page: %+v", items)
	}
	if items[0].File != "u-2026-10-15T09-35-02.tar.gz" || items[0].Created != "2026-10-15T09:35:02Z" {
		t.Errorf("expected newest first with its date, got %+v", items[0])
	}
	if items[2].File != "renamed.tar.gz" || items[2].Created != "" {
		t.Errorf("archive without a date in its name goes last, got %+v", items[2])
	}
	if items[1].Databases[0].Engine != "postgres" || items[1].Types == nil {
		t.Errorf("unexpected item %+v", items[1])
	}
}

func TestRestoreItemsReadsCachedLists(t *testing.T) {
	home := t.TempDir()
	name := "u-2026-10-15T10-58-27.tar.gz"
	if err := saveList(home, name, "mail", strings.NewReader("152\t1\texample.com/info\n")); err != nil {
		t.Fatal(err)
	}
	if err := saveList(home, name, "vhosts", strings.NewReader("1800\t1\texample.com.conf\n1800\t1\tgone.com.conf\n")); err != nil {
		t.Fatal(err)
	}
	items := restoreItems([]BackupInfo{{BackupFile: name, Lists: []string{"files", "mail", "vhosts"}}}, home, map[string]bool{"example.com": true})
	it := items[0]
	if len(it.Mailboxes) != 1 || it.Mailboxes[0].Count != 1 || !it.Mailboxes[0].OnServer {
		t.Errorf("mailboxes = %+v", it.Mailboxes)
	}
	if len(it.Domains) != 2 || it.Domains[1].Name != "gone.com" || it.Domains[1].OnServer {
		t.Errorf("domains = %+v", it.Domains)
	}
	if it.Zones == nil || len(it.Zones) != 0 {
		t.Errorf("zones should be an empty list, got %+v", it.Zones)
	}
}
