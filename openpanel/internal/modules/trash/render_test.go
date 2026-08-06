package trash

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "trash": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderTrashPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("empty trash", func(t *testing.T) {
		data := TrashPageData{LayoutData: baseLayout(mgr, "/files.trash"), Title: "Trash"}
		w := httptest.NewRecorder()
		if err := trashPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No files in the Trash.") {
			t.Error("expected empty-state message")
		}
	})

	t.Run("with entries", func(t *testing.T) {
		rows := []Row{
			toRow(Entry{Name: "deleted.txt", Type: "file", Size: "123", OriginalPath: "/home/user/docker-data/volumes/user_html_data/_data/deleted.txt", DeletionDate: "2026-01-01T00:00:00"}),
			toRow(Entry{Name: "old_folder", Type: "directory", Size: "4096"}),
		}
		data := TrashPageData{LayoutData: baseLayout(mgr, "/files.trash"), Title: "Trash", Directories: rows[1:], Files: rows[:1], Rows: rows}
		w := httptest.NewRecorder()
		if err := trashPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "deleted.txt") {
			t.Error("expected file row in rendered page")
		}
		if !strings.Contains(body, "old_folder") {
			t.Error("expected directory row in rendered page")
		}
		if !strings.Contains(body, "/var/www/html/_data/deleted.txt") {
			t.Errorf("expected shortened original_path display, got:\n%s", body)
		}
	})
}

func TestDisplayOriginalPath(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		// First "_data/" match lands inside "..._html_data/", not the later
		// literal "_data" folder - see displayOriginalPath's doc comment.
		{"/home/user/docker-data/volumes/user_html_data/_data/foo.txt", "/var/www/html/_data/foo.txt"},
		{"/home/user/.local/share/Trash/foo.txt", "/home/user/.local/share/Trash/foo.txt"},
	}
	for _, c := range cases {
		if got := displayOriginalPath(c.in); got != c.want {
			t.Errorf("displayOriginalPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseLsOutputTrash(t *testing.T) {
	out := "total 8\n" +
		"-rw-r--r-- 1 user user 123 Jan  1 12:00 deleted.txt\n" +
		"drwxr-xr-x 2 user user 4096 Jan  1 12:00 old_folder\n"
	info := "deleted.txt=/home/user/docker-data/volumes/user_html_data/_data/deleted.txt|deletion_date=2026-01-01T00:00:00\n"

	entries := parseLsOutputTrash(out, info)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "deleted.txt" || entries[0].Type != "file" {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[0].OriginalPath != "/home/user/docker-data/volumes/user_html_data/_data/deleted.txt" {
		t.Errorf("expected original_path lookup to populate, got %q", entries[0].OriginalPath)
	}
	if entries[0].DeletionDate != "2026-01-01T00:00:00" {
		t.Errorf("expected deletion_date lookup to populate, got %q", entries[0].DeletionDate)
	}
	if entries[1].Name != "old_folder" || entries[1].Type != "directory" {
		t.Errorf("unexpected second entry: %+v", entries[1])
	}
	if entries[1].OriginalPath != "" {
		t.Errorf("expected no metadata match for old_folder, got %q", entries[1].OriginalPath)
	}
}

func TestIsWithin(t *testing.T) {
	cases := []struct {
		candidate, base string
		want            bool
	}{
		{"/home/user/.local/share/Trash", "/home/user/.local/share/Trash", true},
		{"/home/user/.local/share/Trash/sub", "/home/user/.local/share/Trash", true},
		{"/home/user/.local/share/Trash/../../../etc/passwd", "/home/user/.local/share/Trash", false},
		{"/etc/passwd", "/home/user/.local/share/Trash", false},
	}
	for _, c := range cases {
		if got := isWithin(c.candidate, c.base); got != c.want {
			t.Errorf("isWithin(%q, %q) = %v, want %v", c.candidate, c.base, got, c.want)
		}
	}
}
