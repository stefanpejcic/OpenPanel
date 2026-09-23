package account

import (
	"net/url"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestParseActivityLine(t *testing.T) {
	line := "2026-08-03 10:00:00  1.2.3.4 User bob changed password"
	row, ok := parseActivityLine(line)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if row.Timestamp != "2026-08-03 10:00:00 " {
		t.Errorf("timestamp = %q", row.Timestamp)
	}
	if row.IP != "1.2.3.4" {
		t.Errorf("ip = %q", row.IP)
	}
	if row.User != "bob" {
		t.Errorf("user = %q", row.User)
	}
	if row.Action != "changed password" {
		t.Errorf("action = %q", row.Action)
	}
}

func TestParseActivityLineTooShort(t *testing.T) {
	if _, ok := parseActivityLine("not enough fields"); ok {
		t.Error("expected ok=false for a short line")
	}
}

func TestBuildPageEntriesAccount(t *testing.T) {
	entries := buildPageEntries(5, 10)
	var numbers []int
	ellipses := 0
	for _, e := range entries {
		if e.IsEllipsis {
			ellipses++
		} else {
			numbers = append(numbers, e.Number)
		}
	}
	want := []int{1, 3, 4, 5, 6, 7, 10}
	if len(numbers) != len(want) {
		t.Fatalf("got %v, want %v", numbers, want)
	}
	for i, n := range want {
		if numbers[i] != n {
			t.Errorf("position %d: got %d, want %d", i, numbers[i], n)
		}
	}
	if ellipses != 2 {
		t.Errorf("expected 2 ellipses, got %d", ellipses)
	}
}

func TestPaginateActivityLog(t *testing.T) {
	a := &appctx.App{Config: config.Config{"activity_items_per_page": "2"}}
	lines := []string{
		"2026-08-03 10:00:00  1.2.3.4 User bob did one",
		"2026-08-03 10:01:00  1.2.3.4 User bob did two",
		"2026-08-03 10:02:00  5.6.7.8 User alice did three",
	}

	t.Run("no search, page 1", func(t *testing.T) {
		result := paginateActivityLog(a, lines, ActivityFilter{}, 1)
		if result.TotalLines != 3 {
			t.Errorf("TotalLines = %d, want 3", result.TotalLines)
		}
		if result.TotalPages != 2 {
			t.Errorf("TotalPages = %d, want 2", result.TotalPages)
		}
		if len(result.Rows) != 2 {
			t.Errorf("len(Rows) = %d, want 2", len(result.Rows))
		}
	})

	t.Run("search forces show_all", func(t *testing.T) {
		result := paginateActivityLog(a, lines, ActivityFilter{Search: "alice"}, 1)
		if !result.ShowAll {
			t.Error("expected search to force ShowAll")
		}
		if result.TotalLines != 1 {
			t.Errorf("TotalLines = %d, want 1", result.TotalLines)
		}
		if len(result.Rows) != 1 || result.Rows[0].User != "alice" {
			t.Errorf("Rows = %+v", result.Rows)
		}
	})

	t.Run("show_all ignores pagination", func(t *testing.T) {
		result := paginateActivityLog(a, lines, ActivityFilter{ShowAll: true}, 1)
		if result.TotalPages != 1 || len(result.Rows) != 3 {
			t.Errorf("TotalPages=%d len(Rows)=%d, want 1/3", result.TotalPages, len(result.Rows))
		}
	})
}

func TestClassifyActivity(t *testing.T) {
	cases := map[string]string{
		"logged in with a password":                                    ActivityKindSecurity,
		"changed password for FTP account bob":                         ActivityKindSecurity,
		"terminated session abc":                                       ActivityKindSecurity,
		"enabled WAF for domain example.com":                           ActivityKindSecurity,
		"executed command in terminal for service php: ls":             ActivityKindCommand,
		"ran composer install for app":                                 ActivityKindCommand,
		"terminated process 12 in container php using Process Manager": ActivityKindCommand,
		"deleted domain example.com":                                   ActivityKindDanger,
		"moved file /x to Trash using File Manager":                    ActivityKindDanger,
		"reset DNS zone for domain example.com":                        ActivityKindDanger,
		"created email a@example.com":                                  ActivityKindCreate,
		"Added /files to Favorites":                                    ActivityKindCreate,
		"edited cron job":                                              ActivityKindEdit,
		"moved /a to /b using File Manager":                            ActivityKindEdit,
		"unsuspended domain example.com":                               ActivityKindEdit,
		"executed git install for application site":                    ActivityKindCreate,
		"optimized MySQL database wp via API":                          ActivityKindEdit,
		"opened phpMyAdmin":                                            ActivityKindInfo,
		"exported DNS zone for domain example.com":                     ActivityKindInfo,
	}
	for action, want := range cases {
		if got := classifyActivity(action); got != want {
			t.Errorf("classifyActivity(%q) = %q, want %q", action, got, want)
		}
	}
}

func TestPaginateActivityLogFilters(t *testing.T) {
	a := &appctx.App{Config: config.Config{"activity_items_per_page": "100"}}
	lines := []string{
		"2026-08-01 09:00:00  1.2.3.4 User bob deleted domain a.com",
		"2026-08-02 09:00:00  1.2.3.4 User bob created email x@a.com",
		"2026-08-03 09:00:00  1.2.3.4 User bob deleted email x@a.com",
		"2026-08-04 09:00:00  1.2.3.4 User bob edited cron job",
	}

	t.Run("type filter", func(t *testing.T) {
		r := paginateActivityLog(a, lines, ActivityFilter{Kind: ActivityKindDanger}, 1)
		if r.TotalLines != 2 || r.AllCount != 4 {
			t.Errorf("TotalLines=%d AllCount=%d, want 2/4", r.TotalLines, r.AllCount)
		}
		for _, tab := range r.KindTabs {
			if tab.Kind == ActivityKindDanger && (!tab.Active || tab.Count != 2) {
				t.Errorf("danger tab = %+v", tab)
			}
		}
	})

	t.Run("date range is inclusive", func(t *testing.T) {
		r := paginateActivityLog(a, lines, ActivityFilter{From: "2026-08-02", To: "2026-08-03"}, 1)
		if r.TotalLines != 2 {
			t.Errorf("TotalLines = %d, want 2", r.TotalLines)
		}
	})

	t.Run("type and date together", func(t *testing.T) {
		r := paginateActivityLog(a, lines, ActivityFilter{Kind: ActivityKindDanger, From: "2026-08-02"}, 1)
		if r.TotalLines != 1 || r.Rows[0].Action != "deleted email x@a.com" {
			t.Errorf("Rows = %+v", r.Rows)
		}
	})
}

func TestParseActivityFilter(t *testing.T) {
	q, _ := url.ParseQuery("type=danger&from=2026-08-05&to=2026-08-01&search=a")
	f := parseActivityFilter(q)
	if f.Kind != "danger" || f.From != "2026-08-01" || f.To != "2026-08-05" || f.Search != "a" {
		t.Errorf("filter = %+v", f)
	}
	q, _ = url.ParseQuery("type=bogus&from=yesterday")
	if f := parseActivityFilter(q); f.Kind != "" || f.From != "" {
		t.Errorf("invalid values should be dropped, got %+v", f)
	}
}

func TestActivityFilterURL(t *testing.T) {
	f := ActivityFilter{Kind: "edit", From: "2026-08-01", Search: "a b"}
	if got := string(f.URL(2)); got != "/account/activity?from=2026-08-01&page=2&search=a+b&type=edit" {
		t.Errorf("URL = %q", got)
	}
	if got := string((ActivityFilter{}).URL(1)); got != "/account/activity" {
		t.Errorf("empty URL = %q", got)
	}
}
