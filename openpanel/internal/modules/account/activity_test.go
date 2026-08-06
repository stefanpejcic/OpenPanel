package account

import (
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
		result := paginateActivityLog(a, lines, "", false, 1)
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
		result := paginateActivityLog(a, lines, "alice", false, 1)
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
		result := paginateActivityLog(a, lines, "", true, 1)
		if result.TotalPages != 1 || len(result.Rows) != 3 {
			t.Errorf("TotalPages=%d len(Rows)=%d, want 1/3", result.TotalPages, len(result.Rows))
		}
	})
}
