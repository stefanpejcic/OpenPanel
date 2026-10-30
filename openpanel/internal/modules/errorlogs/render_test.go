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
	q := Query{Domain: "a.com", Service: "php-fpm-8.2", Search: "fatal", Exact: true, FromDay: "2026-10-01", ToDay: "2026-10-03", Limit: 1}
	res := Result{
		Total: 2, Truncated: true,
		Entries: []Entry{{Time: time.Unix(1792085165, 0), Source: "php", Service: "php-fpm-8.2", Domain: "a.com", Level: "critical", Message: "Fatal error: </script><script>alert(1)</script>\nStack trace:"}},
		Sources: []Source{{Name: "php-fpm-8.2", Label: "PHP 8.2"}, {Name: "caddy", Label: "HTTP 5xx responses"}},
	}
	data := buildPageData(testLayout(mgr), q, res, []Target{{Value: "a.com"}, {Value: "a.com/blog", Sub: true}, {Value: "b.com"}}, "")

	w := httptest.NewRecorder()
	if err := errorLogsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{
		"Error Logs", `x-data="opErrorLogs()"`, `value="a.com/blog" class="punycode"`, "&nbsp;&nbsp;↳ a.com/blog", `value="php-fpm-8.2"`, "PHP 8.2",
		`window.EL_ROWS = [[1792085165000,"critical","php-fpm-8.2","php","a.com","Fatal error: `,
		`domain: "a.com", exact: true, service: "php-fpm-8.2"`, `q: "fatal", from: "2026-10-01", to: "2026-10-03"`,
		"Download .txt", "opBusy.show(", "format: 'json'", "Grouped", `class="op-bulkbar"`, `class="op-th-sort"`, "sortBy('rank')",
		"Only this website", "Only this service", "Copy line", "Quick date range", "Last 7 days",
		"newest 1 of 2 entries loaded",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page missing %q", want)
		}
	}
	if strings.Contains(body, "<script>alert(1)") {
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
	body := w.Body.String()
	for _, want := range []string{"window.EL_ROWS = []", "window.EL_TARGETS = []", "No errors logged", "No errors found for these filters."} {
		if !strings.Contains(body, want) {
			t.Errorf("empty page missing %q", want)
		}
	}
}

func TestRenderErrorLogsForbidden(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := buildPageData(testLayout(mgr), Query{Domain: "x.com"}, Result{}, nil, "You do not own this domain.")
	w := httptest.NewRecorder()
	if err := errorLogsPage.Render(w, 403, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if body := w.Body.String(); !strings.Contains(body, `class="op-alert is-danger"`) || !strings.Contains(body, "You do not own this domain.") {
		t.Error("error alert missing")
	}
}
