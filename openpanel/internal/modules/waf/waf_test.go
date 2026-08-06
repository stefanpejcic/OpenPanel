package waf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFirstPathSegment(t *testing.T) {
	if got := firstPathSegment("example.com/foo"); got != "example.com" {
		t.Errorf("got %q", got)
	}
	if got := firstPathSegment("example.com"); got != "example.com" {
		t.Errorf("got %q", got)
	}
}

func TestFilterOut(t *testing.T) {
	got := filterOut([]string{"a", "007", "b", "007"}, "007")
	want := []string{"a", "b"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestReadLinesKeepEnds(t *testing.T) {
	lines := readLinesKeepEnds("a\nb\nc")
	if len(lines) != 3 || lines[0] != "a\n" || lines[1] != "b\n" || lines[2] != "c" {
		t.Errorf("lines = %q", lines)
	}
	if lines := readLinesKeepEnds(""); lines != nil {
		t.Errorf("empty input: got %v, want nil", lines)
	}
}

func TestRewriteDirectivesBlock(t *testing.T) {
	content := `domain example.com {
    directives ` + "`" + `
            SecRuleEngine On
            SecRuleRemoveById 007 123
            SecRuleRemoveByTag example
    ` + "`" + `
}
`
	got := rewriteDirectivesBlock(content, []string{"007", "123", "456"}, []string{"example", "foo"})
	if !containsLine(got, "            SecRuleRemoveById 007 123 456") {
		t.Errorf("expected new SecRuleRemoveById line, got:\n%s", got)
	}
	if !containsLine(got, "            SecRuleRemoveByTag example foo") {
		t.Errorf("expected new SecRuleRemoveByTag line, got:\n%s", got)
	}
	if !containsLine(got, "            SecRuleEngine On") {
		t.Errorf("expected SecRuleEngine line preserved, got:\n%s", got)
	}
}

func containsLine(content, line string) bool {
	for _, l := range readLinesKeepEnds(content) {
		if l == line || l == line+"\n" {
			return true
		}
	}
	return false
}

func TestBuildLogPageEntries(t *testing.T) {
	t.Run("small total, no ellipsis needed", func(t *testing.T) {
		entries := buildLogPageEntries(1, 3)
		for _, e := range entries {
			if e.IsEllipsis {
				t.Errorf("did not expect an ellipsis for a 3-page total, got %+v", entries)
			}
		}
	})

	t.Run("large total, both ellipses", func(t *testing.T) {
		entries := buildLogPageEntries(10, 20)
		if entries[0].Number != 1 {
			t.Errorf("expected first entry to be page 1, got %+v", entries[0])
		}
		if !entries[1].IsEllipsis {
			t.Errorf("expected an ellipsis after page 1, got %+v", entries[1])
		}
		last := entries[len(entries)-1]
		if last.Number != 20 {
			t.Errorf("expected last entry to be page 20, got %+v", last)
		}
	})
}

func TestReadWAFLogs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "example.com.log")

	now := time.Now()
	recent := now.Format("2006/01/02 15:04:05")
	old := now.Add(-2 * time.Hour).Format("2006/01/02 15:04:05")

	lines := []string{
		mustLogLine(t, old, false),
		mustLogLine(t, recent, false),
		mustLogLine(t, recent, true),
	}
	if err := os.WriteFile(path, []byte(lines[0]+"\n"+lines[1]+"\n"+lines[2]+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stats := readWAFLogs(path, 3600)
	if stats.Checks != 2 {
		t.Errorf("checks = %d, want 2 (old entry outside window should stop the scan)", stats.Checks)
	}
	if stats.Blocks != 1 {
		t.Errorf("blocks = %d, want 1", stats.Blocks)
	}
}

func TestReadWAFLogsMissingFile(t *testing.T) {
	stats := readWAFLogs("/nonexistent/path.log", 60)
	if stats.Checks != 0 || stats.Blocks != 0 {
		t.Errorf("expected zero stats for a missing file, got %+v", stats)
	}
}

func mustLogLine(t *testing.T, timestamp string, interrupted bool) string {
	t.Helper()
	entry := map[string]any{
		"transaction": map[string]any{
			"timestamp":      timestamp,
			"is_interrupted": interrupted,
		},
	}
	b, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
