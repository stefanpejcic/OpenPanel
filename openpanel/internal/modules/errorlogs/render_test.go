package errorlogs

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func testLayout(mgr *i18n.Manager) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "error_logs": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, "/error-logs"), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: "/error-logs", AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderErrorLogsPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	q := Query{Level: "error", Domain: "a.com", FromDay: "2026-10-01", ToDay: "2026-10-03", Limit: 1}
	res := Result{
		Total: 2, Truncated: true,
		Entries: []Entry{{Time: time.Unix(1792085165, 0), Service: "php-fpm-8.2", Domain: "a.com", Level: "critical", Message: "Fatal error: <script>x</script>"}},
		Sources: []Source{{Name: "php-fpm-8.2", Label: "PHP 8.2"}, {Name: "caddy", Label: "HTTP 5xx responses"}},
	}
	layout := testLayout(mgr)
	layout.BulkActions = errorLogsBulkActions(layout.T)
	data := buildPageData(layout, q, res, []Target{{Value: "a.com"}, {Value: "a.com/blog", Sub: true}, {Value: "b.com"}}, "")

	w := httptest.NewRecorder()
	if err := errorLogsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{
		"Error Logs", `value="a.com" class="punycode" selected`, `value="a.com/blog" class="punycode" >&nbsp;&nbsp;↳ a.com/blog`, `value="php-fpm-8.2"`, "PHP 8.2",
		"Critical", "Showing 1 out of 2 rows", "&lt;script&gt;x&lt;/script&gt;",
		"[critical] [php-fpm-8.2] [a.com] Fatal error",
		`id="bulk-actions-bar"`, "bulkClient(", `data-row="0"`, "bi-clipboard", "bi-download",
		`data-sort-key="severity"`, `data-sort-col="time"`, "js/table-sort.js", `<option value="error" selected>`,
		`border-gray-200 dark:border-gray-800">`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page missing %q", want)
		}
	}
	if strings.Contains(body, "<script>x</script>") {
		t.Error("log message rendered unescaped")
	}
}

func TestRenderErrorLogsEmpty(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := buildPageData(testLayout(mgr), Query{}, Result{Entries: []Entry{}}, nil, "")
	w := httptest.NewRecorder()
	if err := errorLogsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "No errors found for these filters.") {
		t.Error("empty state missing")
	}
}
