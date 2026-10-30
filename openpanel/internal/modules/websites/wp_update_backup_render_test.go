package websites

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

// the WP manager's Updates tab: every update goes through the backup sheet, and the automatic-update backup switch sits above the preferences form
func TestWPUpdatesBackupSheetAndSwitch(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	pd := pageData{LayoutData: baseLayout(mgr, "/website"), CurrentDomain: "example.com", Docroot: "/var/www/html/example.com"}
	w := httptest.NewRecorder()
	if err := wpSinglePage.Render(w, 200, WPSinglePageData{pageData: pd, Container: ContainerInfo{SiteName: "example.com", Version: "7.1.4"}}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()

	for _, want := range []string{
		`aria-labelledby="wp-update-sheet-title"`,
		"Back up before updating (recommended)",
		"Takes a copy of files and database; restore it from the Backups tab.",
		`x-model="backupFirst"`,
		`@click="askCore()"`,
		`@click.prevent="open = false; askItem(it)"`,
		`@click="askSelected()"`,
		`localStorage.getItem('wp-backup-before-update')`,
		`'/wordpress/tools/core-update'`,
		`stream: '1'`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("update sheet: missing %q", want)
		}
	}
	if strings.Contains(body, "/wordpress/wp-cli/update_now") || strings.Contains(body, "updateItems(it.kind") {
		t.Error("an update path still skips the backup sheet")
	}

	sw := strings.Index(body, `id="wp-auto-backup"`)
	form := strings.Index(body, `id="UpdatesForm"`)
	if sw < 0 || form < 0 || sw > form {
		t.Fatalf("auto backup switch should come right before the preferences form (switch %d, form %d)", sw, form)
	}
	card := body[sw:form]
	for _, want := range []string{"Back up before automatic updates", `role="switch"`, `@click="runNow()"`, "Run now", "statusText()"} {
		if !strings.Contains(card, want) {
			t.Errorf("auto backup card: missing %q", want)
		}
	}
	for _, want := range []string{"'/wordpress/tools/autoupdate'", "'/wordpress/tools/autoupdate-set'", "'/wordpress/tools/autoupdate-run'"} {
		if !strings.Contains(body, want) {
			t.Errorf("auto backup script: missing %s", want)
		}
	}
}

// the Maintenance tab edits maintenance.php in place, only for accounts that have the File Manager
func TestWPMaintenancePageEditor(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	render := func(fileManager bool) string {
		layout := baseLayout(mgr, "/website")
		layout.UserAllowed = map[string]bool{"wordpress": true, "filemanager": fileManager}
		pd := pageData{LayoutData: layout, CurrentDomain: "example.com", Docroot: "/var/www/html/example.com"}
		w := httptest.NewRecorder()
		if err := wpSinglePage.Render(w, 200, WPSinglePageData{pageData: pd, Container: ContainerInfo{SiteName: "example.com", Version: "7.1.4"}}); err != nil {
			t.Fatalf("Render: %v", err)
		}
		return w.Body.String()
	}

	body := render(true)
	for _, want := range []string{`x-data="wpMaintPage()"`, `class="wpm-editor"`, `'/wordpress/tools/maintenance-page'`, `'/wordpress/tools/maintenance-page-save'`, `<iframe sandbox`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "wp-content/maintenance.php?new=true") {
		t.Error("the maintenance page still opens the file manager editor")
	}
	if body := render(false); strings.Contains(body, `class="wpm-editor"`) || strings.Contains(body, `x-data="wpMaintPage()"`) {
		t.Error("editor shown without the File Manager feature")
	}
}
