package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBindSourceEscapes(t *testing.T) {
	home := t.TempDir()
	home, err := filepath.EvalSymlinks(home) // canonicalize so Rel comparisons are stable
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, "evil")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(home, "data"), filepath.Join(home, "innerlink")); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		rel  string
		want bool
	}{
		{"data", false},      // a real folder inside the account
		{"data/sub", false},  // not created yet, nothing to follow
		{"newfolder", false}, // doesn't exist yet
		{"innerlink", false}, // symlink that stays inside the account
		{"evil", true},       // symlink pointing out of the account
		{"evil/x", true},     // path through a symlink that points out
	}
	for _, c := range cases {
		if got := bindSourceEscapes(home, c.rel); got != c.want {
			t.Errorf("bindSourceEscapes(%q) = %v, want %v", c.rel, got, c.want)
		}
	}
}
