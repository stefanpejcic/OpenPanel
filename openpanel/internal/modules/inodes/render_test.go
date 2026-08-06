package inodes

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "inodes": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderInodesPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := InodesPageData{
		LayoutData:  baseLayout(mgr, "/inodes-explorer/"),
		Breadcrumbs: buildBreadcrumbs("inodes-explorer", "/inodes-explorer/sub/"),
		Rows:        parseInodesOutput("150 dir1\n42 dir2", "/inodes-explorer/"),
		ShowUpLink:  true, UpLink: "/inodes-explorer",
	}
	w := httptest.NewRecorder()
	if err := inodesPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"dir1", "150", "dir2", "Up One Level"} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered inodes page missing %q", want)
		}
	}
}

func TestParseInodesOutput(t *testing.T) {
	rows := parseInodesOutput("150 dir1\n42 dir2\n5 .\n", "/inodes-explorer/")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (\".\" excluded), got %d: %+v", len(rows), rows)
	}
}
