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
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
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

// links were built from the decoded path, so a folder named a#b sent the browser to /inodes-explorer/a
func TestParseInodesOutputEscapesCurrentPath(t *testing.T) {
	cases := map[string]string{
		"/inodes-explorer/a#b/":  "/inodes-explorer/a%23b/sub/",
		"/inodes-explorer/a?b/":  "/inodes-explorer/a%3Fb/sub/",
		"/inodes-explorer/x%41/": "/inodes-explorer/x%2541/sub/",
	}
	for urlPath, want := range cases {
		if rows := parseInodesOutput("3 sub", urlPath); len(rows) != 1 || rows[0].Href != want {
			t.Errorf("parseInodesOutput under %q: %+v, want Href %q", urlPath, rows, want)
		}
	}
}

// a folder named " lead" used to link to /inodes-explorer/lead/
func TestParseInodesOutputKeepsLeadingSpaces(t *testing.T) {
	rows := parseInodesOutput("7  lead\n3 plain\n", "/inodes-explorer/")
	if len(rows) != 2 || rows[0].Directory != " lead" || rows[0].Href != "/inodes-explorer/%20lead/" || rows[1].Directory != "plain" {
		t.Fatalf("got %+v", rows)
	}
}

// the File Manager link only showed under /inodes-explorer/files/, which isn't the File Manager's root
func TestFileManagerLinkForDocroot(t *testing.T) {
	cases := map[string]string{
		"/inodes-explorer/docker-data/volumes/u_html_data/_data/":     "/files/site",
		"/inodes-explorer/docker-data/volumes/u_html_data/_data/a b/": "/files/a%20b/site",
		"/inodes-explorer/files/":                                     "",
		"/inodes-explorer/":                                           "",
		"/inodes-explorer/docker-data/volumes/u_mysql_data/_data/":    "",
	}
	for urlPath, want := range cases {
		if rows := parseInodesOutput("3 site", urlPath); len(rows) != 1 || rows[0].FileManagerHref != want {
			t.Errorf("under %q: %+v, want FileManagerHref %q", urlPath, rows, want)
		}
	}
}
