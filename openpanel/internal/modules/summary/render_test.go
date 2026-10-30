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
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
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
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	from := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	s := Summary{Username: "stefan", Date: "2026-10-15", From: from, GeneratedAt: from.Add(30*time.Hour + 15*time.Minute), PrevRequests: 8,
		Attention: []Item{{Title: "Malware <found>", Detail: "3 files", Action: "Review the quarantined files", Link: "/malware-scanner/quarantine"}},
		Look:      []Item{{Title: "Disk is 90% full", Action: "Ask your hosting provider for more disk space."}},
		Good:      []string{"SSL certificates valid"},
		Websites:  []Website{{Name: "a.com", Requests: 10, ServerErrors: 2, TopPages: []string{"/", "/shop"}, TopPageViews: 4, Share: 100, PrevRequests: 5}},
		Logins:    []Login{{Time: from.Add(9 * time.Hour), IP: "1.1.1.1", Country: "RS", New: true}},
		Changes:   []string{"09:05 added domain b.com"},
		Usage:     []UsageBar{{Title: "Memory at the busiest hour", Used: "384 MB", Total: "6.0 GB", Percent: 95, Limit: 90}},
	}
	prev := Summary{Websites: []Website{{Name: "a.com", Requests: 8, UniqueIPs: 4, ServerErrors: 1}}}
	data := PageData{LayoutData: testLayout(mgr), Summary: s, Found: true, Dates: []string{"2026-10-16", "2026-10-15", "2026-10-14"}, Prev: "2026-10-14", Next: "2026-10-16", Date: "2026-10-15",
		Tiles: tiles(s, prev, true), EmailOn: true, Email: "stefan@example.com"}
	body := render(t, data)
	for _, want := range []string{
		"Thursday, 15 October 2026", "made 16 Oct 06:15 UTC", "Malware &lt;found&gt;", `href="/malware-scanner/quarantine"`,
		"Ask your hosting provider for more disk space.", "SSL certificates valid", "What went well", "a.com",
		`href="/summary?date=2026-10-14"`, `href="/summary?date=2026-10-16"`, `<option value="2026-10-15" selected>Thu 15 Oct</option>`,
		// requests 10 against 8 the day before, errors 2 against 1 is bad news
		`<span class="op-badge is-success" title="Compared with the day before">&#43;25%</span>`, "8 the day before",
		`<span class="op-badge is-danger" title="Compared with the day before">&#43;100%</span>`,
		"Top page / · 4 views · &#43;100% vs the day before", `data-sort-key="requests"`, `data-sort-value="10"`, "2 5xx",
		`src="/static/flags/rs.png"`, "New IP", "1.1.1.1", "added domain b.com", "dsChanges(", `op-progress is-warning`,
		`x-data="dsEmail(true)"`, "stefan@example.com", `href="/account/notifications#daily-summary"`, `href="/server/usage/history"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page missing %q", want)
		}
	}

	body = render(t, PageData{LayoutData: testLayout(mgr)})
	if !strings.Contains(body, "Your first summary will be ready tomorrow morning.") || !strings.Contains(body, `x-data="dsEmail(false)"`) {
		t.Error("empty state missing")
	}

	// a day that isn't saved keeps the picker so the user can go back
	body = render(t, PageData{LayoutData: testLayout(mgr), Dates: []string{"2026-10-16"}, Date: "2026-10-01"})
	if !strings.Contains(body, "There is no summary for this day.") || !strings.Contains(body, "Pick a day") {
		t.Error("missing-day state missing")
	}
}

// TestRenderSummaryRealDays renders testinguser's three saved days, a busy one and two quiet ones
func TestRenderSummaryRealDays(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	dates := []string{"2026-10-28", "2026-10-27", "2026-10-26"}
	for i, d := range dates {
		s := realSummary(t, d)
		var prev Summary
		hasPrev := i+1 < len(dates)
		if hasPrev {
			prev = realSummary(t, dates[i+1])
		}
		data := PageData{LayoutData: testLayout(mgr), Summary: s, Found: true, Dates: dates, Date: d, Tiles: tiles(s, prev, hasPrev), Email: "user_1793020192884@openpanel.org"}
		if i > 0 {
			data.Next = dates[i-1]
		}
		if hasPrev {
			data.Prev = dates[i+1]
		}
		body := render(t, data)
		if !strings.Contains(body, s.Headline) {
			t.Errorf("%s: headline missing", d)
		}
		switch d {
		case "2026-10-26":
			for _, want := range []string{"Needs your attention · 3", "Worth a look · 1", `href="/error-logs?domain=php.tests.openpanel.org&amp;exact=1"`, "nodejs.tests.openpanel.org", "No summary for the day before", "New IP", "10.0.2.2"} {
				if !strings.Contains(body, want) {
					t.Errorf("%s: missing %q", d, want)
				}
			}
		case "2026-10-27":
			// 0 requests against 34 the day before
			if !strings.Contains(body, "-100%") || !strings.Contains(body, "34 the day before") || !strings.Contains(body, "No website traffic on this day") {
				t.Errorf("%s: day-over-day missing", d)
			}
		}
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
