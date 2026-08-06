package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLastLineOf(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resource_usage.txt")
	if err := os.WriteFile(path, []byte("{\"cpu\":1}\n{\"cpu\":2}\n{\"cpu\":3}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := lastLineOf(path)
	if err != nil {
		t.Fatalf("lastLineOf: %v", err)
	}
	if got != `{"cpu":3}` {
		t.Errorf("lastLineOf() = %q, want the last JSON line", got)
	}
}

func TestLastLineOfMissingFile(t *testing.T) {
	if _, err := lastLineOf(filepath.Join(t.TempDir(), "nope.txt")); err == nil {
		t.Error("expected an error for a missing file")
	}
}

func TestGetLastLoginDataParsing(t *testing.T) {
	// GetLastLoginData hardcodes /etc/openpanel/..., so this test verifies
	// the parsing logic directly rather than the full read-file path.
	line := "IP: 1.2.3.4 - Country: US - Login Time: 2026-01-01 12:00:00"
	entries := parseLastLoginLines([]string{line, "garbage line", ""})
	if len(entries) != 1 {
		t.Fatalf("expected 1 valid entry, got %d: %+v", len(entries), entries)
	}
	if entries[0].IP != "1.2.3.4" || entries[0].CountryCode != "US" || entries[0].LoginTime != "2026-01-01 12:00:00" {
		t.Errorf("unexpected entry: %+v", entries[0])
	}
}
