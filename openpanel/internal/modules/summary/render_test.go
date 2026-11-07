package summary

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func testLayout(mgr *i18n.Manager) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "summary": true, "notifications": true, "usage": true, "activity": true, "login_history": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, "/summary"), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "stefan",
		RequestPath: "/summary", AdminPort: "2087", T: mgr.Translator("en"),
	}
}

// realSummary is one of testinguser's saved days from the test VM
func realSummary(t *testing.T, date string) Summary {
	t.Helper()
	raw, err := os.ReadFile("testdata/" + date + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var s Summary
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	return s
}

func render(t *testing.T, data PageData) string {
	t.Helper()
	w := httptest.NewRecorder()
	if err := summaryPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return w.Body.String()
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

func TestTiles(t *testing.T) {
	s := Summary{PrevRequests: -1, Websites: []Website{{Requests: 30, UniqueIPs: 3, Bytes: 2048, ServerErrors: 0, WAFBlocked: 4}}}
	prev := Summary{Websites: []Website{{Requests: 20, UniqueIPs: 4, Bytes: 1024, ServerErrors: 2, WAFBlocked: 2}}}
	got := tiles(s, prev, true)
	if got.Requests.Delta() != "+50%" || got.Requests.Tone() != "is-success" {
		t.Errorf("requests = %q %q", got.Requests.Delta(), got.Requests.Tone())
	}
	if got.Visitors.Delta() != "-25%" || got.Visitors.Tone() != "is-danger" {
		t.Errorf("visitors = %q %q", got.Visitors.Delta(), got.Visitors.Tone())
	}
	if got.Errors.Delta() != "-100%" || got.Errors.Tone() != "is-success" {
		t.Errorf("errors = %q %q, fewer errors is good", got.Errors.Delta(), got.Errors.Tone())
	}
	if got.Blocked.Delta() != "+100%" || got.Blocked.Tone() != "" {
		t.Errorf("blocked = %q %q", got.Blocked.Delta(), got.Blocked.Tone())
	}
	if none := tiles(s, Summary{}, false); none.Requests.Delta() != "" || none.Requests.HasPrev {
		t.Error("no day before should give no delta")
	}
	// the saved prev_requests wins over the loaded day for requests
	s.PrevRequests = 60
	if got := tiles(s, prev, true); got.Requests.Delta() != "-50%" || got.Requests.PrevNum() != "60" {
		t.Errorf("requests with prev_requests = %q", got.Requests.Delta())
	}
}

func TestSetPref(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "notify_daily_summary=1\n"},
		{"notify_account_login=1\n", "notify_account_login=1\nnotify_daily_summary=1\n"},
		{"notify_account_login=1\nnotify_daily_summary=0\nnotify_ssl_expiry=1", "notify_account_login=1\nnotify_daily_summary=1\nnotify_ssl_expiry=1\n"},
	}
	for _, c := range cases {
		if got := setPref(c.in, emailKey, "1"); got != c.want {
			t.Errorf("setPref(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEmailOn(t *testing.T) {
	usersDir = t.TempDir() + "/"
	if emailOn("stefan") {
		t.Error("no file should be off")
	}
	write(t, usersDir+"stefan/notifications.yaml", "notify_account_login=1\nnotify_daily_summary=1\n")
	if !emailOn("stefan") {
		t.Error("expected on")
	}
	write(t, usersDir+"stefan/notifications.yaml", "notify_daily_summary=0\n")
	if emailOn("stefan") {
		t.Error("expected off")
	}
}

func TestDayHelpers(t *testing.T) {
	if got := dayBefore("2026-11-01"); got != "2026-10-31" {
		t.Errorf("dayBefore = %q", got)
	}
	if got := dayBefore("../x"); got != "" {
		t.Errorf("dayBefore on junk = %q", got)
	}
	if got := (PageData{}).DayLabel("2026-10-28"); got != "Wed 28 Oct" {
		t.Errorf("DayLabel = %q", got)
	}
	w := Website{Requests: 15, PrevRequests: 10, Share: 0}
	if w.Change() != "+50%" || w.Bar() != 2 || (Website{Requests: 5, PrevRequests: -1}).Change() != "" {
		t.Errorf("website helpers = %q %d", w.Change(), w.Bar())
	}
}

func TestSizeParts(t *testing.T) {
	for in, want := range map[int64]string{0: "0 B", 529: "529 B", 1840: "1.8 KB", 15561: "15.2 KB", 3 << 20: "3.0 MB"} {
		if got := (Website{Bytes: in}).Size(); got != want {
			t.Errorf("Size(%d) = %q, want %q", in, got, want)
		}
	}
}
