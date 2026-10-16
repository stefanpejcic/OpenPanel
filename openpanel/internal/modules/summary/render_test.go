package summary

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func testLayout(mgr *i18n.Manager) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "summary": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, "/summary"), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "stefan",
		RequestPath: "/summary", AdminPort: "2087", T: mgr.Translator("en"),
	}
}

func TestRenderSummaryPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	s := Summary{Username: "stefan", Date: "2026-10-15", From: from,
		Attention: []Item{{Title: "Malware <found>", Detail: "3 files", Action: "Review the quarantined files", Link: "/malware-scanner/quarantine"}},
		Look:      []Item{{Title: "Disk is 90% full", Action: "Ask your hosting provider for more disk space."}},
		Good:      []string{"SSL certificates valid"},
		Websites:  []Website{{Name: "a.com", Requests: 10, ServerErrors: 2, TopPages: []string{"/", "/shop"}}},
		Logins:    []Login{{Time: from.Add(9 * time.Hour), IP: "1.1.1.1"}},
		Changes:   []string{"09:05 added domain b.com"},
	}
	data := PageData{LayoutData: testLayout(mgr), Summary: s, Found: true, Dates: []string{"2026-10-16", "2026-10-15", "2026-10-14"}, Prev: "2026-10-14", Next: "2026-10-16"}
	w := httptest.NewRecorder()
	if err := summaryPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{
		"Thursday, 15 October 2026", "Malware &lt;found&gt;", `href="/malware-scanner/quarantine"`,
		"Ask your hosting provider for more disk space.", "SSL certificates valid", "a.com",
		`href="/summary/2026-10-14"`, `href="/summary/2026-10-16"`, `<option value="2026-10-15" selected>`,
		">1.1.1.1</p>", "added domain b.com", "09:05", "border-gray-200 dark:border-gray-800",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page missing %q", want)
		}
	}

	w = httptest.NewRecorder()
	if err := summaryPage.Render(w, 200, PageData{LayoutData: testLayout(mgr)}); err != nil {
		t.Fatalf("Render empty: %v", err)
	}
	if !strings.Contains(w.Body.String(), "Your first summary will be ready tomorrow morning.") {
		t.Error("empty state missing")
	}
}
