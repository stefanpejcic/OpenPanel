package diskusage

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "disk_usage": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderDiskUsagePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DiskUsagePageData{
		LayoutData:  baseLayout(mgr, "/disk-usage/"),
		Breadcrumbs: buildBreadcrumbs("disk-usage", "/disk-usage/sub/"),
		Rows:        parseDuOutput("120M dir1\n45K dir2", "/disk-usage/"),
		ShowUpLink:  true, UpLink: "/disk-usage",
	}
	w := httptest.NewRecorder()
	if err := diskUsagePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"dir1", "120M", "dir2", "Up One Level"} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered disk usage page missing %q", want)
		}
	}
}

func TestParseDuOutput(t *testing.T) {
	rows := parseDuOutput("120M dir with spaces\n45K other\n", "/disk-usage/")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d: %+v", len(rows), rows)
	}
	if rows[0].Directory != "dir with spaces" || rows[0].Count != "120M" {
		t.Errorf("unexpected row[0]: %+v", rows[0])
	}
}

func TestBuildBreadcrumbs(t *testing.T) {
	crumbs := buildBreadcrumbs("disk-usage", "/disk-usage/a/b/")
	if len(crumbs) != 2 || crumbs[0].Name != "a" || crumbs[1].Name != "b" {
		t.Fatalf("unexpected breadcrumbs: %+v", crumbs)
	}
	if crumbs[1].Path != "/disk-usage/a/b" {
		t.Errorf("unexpected path: %q", crumbs[1].Path)
	}
}
