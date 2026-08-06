package filemanager

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "filemanager": true, "trash": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

// TestRenderFilesPage exercises filemanager.html + every included partial
// through html/template's real executor, covering both the
// populated-directory and empty-directory branches.
func TestRenderFilesPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("populated, classic view", func(t *testing.T) {
		rows := []ViewRow{
			{Index: 0, Name: "reports", Type: "directory", Permissions: "drwxr-xr-x", Href: "/files/documents/reports?view=classic"},
			{Index: 1, Name: "notes.txt", Type: "file", Size: "123", Date: "Jan 1 12:00", Permissions: "-rw-r--r--", DownloadURL: "/file-manager/download-file/notes.txt?path_param=documents"},
			{Index: 2, Name: "link", Type: "symlink", LinkTarget: "notes.txt", Permissions: "lrwxrwxrwx"},
		}
		data := FilesPageData{
			LayoutData: baseLayout(mgr, "/files/documents"), PathParam: "documents", View: "classic",
			Breadcrumbs: buildBreadcrumbs("documents"),
			Directories: rows[:1], Files: rows[1:], Rows: rows,
			ShowUpLink: true, UpLink: "/files?view=classic",
			HasReadme: true, ReadmeContent: "# hello",
			FilemanagerUploadSize: 500, FilemanagerEditSize: 5, FilemanagerViewSize: 5, FilemanagerDownloadSize: 500,
			Extensions: ".txt .php", Images: ".jpg .png", Archives: ".zip .tar",
			CurrentPage: 1, TotalPages: 1, TotalFiles: 3, StartLineNumber: 1, EndLineNumber: 3,
			PageEntries: buildPageEntries(1, 1),
		}
		w := httptest.NewRecorder()
		if err := filesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{"reports", "notes.txt", "link", "notes.txt)", "Go Up", "README.md", "Skip the trash"} {
			if !strings.Contains(body, want) {
				t.Errorf("rendered files page missing %q", want)
			}
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		data := FilesPageData{
			LayoutData: baseLayout(mgr, "/files"), View: "classic",
			FilemanagerUploadSize: 500,
		}
		w := httptest.NewRecorder()
		if err := filesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No items found.") {
			t.Error("expected empty-directory message")
		}
	})

	t.Run("pagination", func(t *testing.T) {
		rows := []ViewRow{{Index: 0, Name: "a.txt", Type: "file"}}
		data := FilesPageData{
			LayoutData: baseLayout(mgr, "/files"), View: "classic",
			Files: rows, Rows: rows,
			CurrentPage: 5, TotalPages: 10, TotalFiles: 500,
			StartLineNumber: 401, EndLineNumber: 500,
			PrevPage: 4, NextPage: 6,
			PageEntries: buildPageEntries(5, 10),
		}
		w := httptest.NewRecorder()
		if err := filesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "page=4") || !strings.Contains(body, "page=6") {
			t.Error("expected prev/next page links")
		}
	})
}

func TestBuildPageEntries(t *testing.T) {
	entries := buildPageEntries(5, 10)
	var got []int
	var ellipses int
	for _, e := range entries {
		if e.IsEllipsis {
			ellipses++
			continue
		}
		got = append(got, e.Number)
	}
	want := []int{1, 3, 4, 5, 6, 7, 10}
	if len(got) != len(want) {
		t.Fatalf("got page numbers %v, want %v", got, want)
	}
	for i, n := range want {
		if got[i] != n {
			t.Errorf("page[%d] = %d, want %d", i, got[i], n)
		}
	}
	if ellipses != 2 {
		t.Errorf("expected 2 ellipses (page 2 and page 9), got %d", ellipses)
	}
}

func TestBuildBreadcrumbs(t *testing.T) {
	crumbs := buildBreadcrumbs("a/b/c")
	if len(crumbs) != 3 {
		t.Fatalf("expected 3 breadcrumbs, got %d: %+v", len(crumbs), crumbs)
	}
	if crumbs[0].Path != "a" || crumbs[1].Path != "a/b" || crumbs[2].Path != "a/b/c" {
		t.Errorf("unexpected breadcrumb paths: %+v", crumbs)
	}
	if !crumbs[2].Last || crumbs[0].Last || crumbs[1].Last {
		t.Errorf("unexpected Last flags: %+v", crumbs)
	}
}

func TestRenderUploadPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := UploadPageData{
		LayoutData: baseLayout(mgr, "/file-manager/upload"), PathParam: "documents", FilemanagerUploadSize: 500,
	}
	w := httptest.NewRecorder()
	if err := uploadPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"documents", "Max size:", "500 MB", "Download from URL instead"} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered upload page missing %q", want)
		}
	}
}

func TestRenderEditFilePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("regular file", func(t *testing.T) {
		data := EditFilePageData{
			LayoutData: baseLayout(mgr, "/file-manager/edit-file/documents/notes.txt"),
			FilePath:   "documents/notes.txt", FilePathJSON: jsonString("documents/notes.txt"),
			FileContent: "hello world", FileContentJSON: jsonString("hello world"),
			Editor: "monaco", EditorLabel: "Monaco",
			EditorOptions: []EditorOption{{Value: "monaco", Label: "Monaco", Selected: true}, {Value: "text", Label: "Plain Text"}},
			Breadcrumbs:   buildBreadcrumbs("documents/notes.txt"),
		}
		w := httptest.NewRecorder()
		if err := editFilePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{"Edit File", "documents", "notes.txt", "hello world", "Monaco"} {
			if !strings.Contains(body, want) {
				t.Errorf("rendered edit_file page missing %q", want)
			}
		}
	})

	t.Run("special wpcli themes file", func(t *testing.T) {
		data := EditFilePageData{
			LayoutData: baseLayout(mgr, "/file-manager/edit-file/wpcli_themes.txt"),
			FilePath:   "wpcli_themes.txt", FilePathJSON: jsonString("wpcli_themes.txt"),
			FileContent: "twentytwentyfour", FileContentJSON: jsonString("twentytwentyfour"),
			Editor: "text", EditorLabel: "Plain Text",
			EditorOptions: []EditorOption{{Value: "text", Label: "Plain Text", Selected: true}},
		}
		w := httptest.NewRecorder()
		if err := editFilePage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Edit WordPress Themes Set") {
			t.Error("expected the wpcli_themes.txt special-case title")
		}
	})
}
