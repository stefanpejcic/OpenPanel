package filemanager

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestHasSubfolder(t *testing.T) {
	root, home := withUserRoot(t)
	mk := func(p string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(home, p), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	touch := func(p string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(home, p), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("uploads")
	mk("mixed/sub")
	mk("empty")
	// more files than one ReadDir batch, so the answer has to come from later batches too
	for i := 0; i < 600; i++ {
		touch("uploads/f" + strconv.Itoa(i) + ".jpg")
		touch("mixed/f" + strconv.Itoa(i) + ".jpg")
	}

	for dir, want := range map[string]bool{"uploads": false, "mixed": true, "empty": false, "missing": false} {
		if got := hasSubfolder(root, dir); got != want {
			t.Errorf("hasSubfolder(%q) = %v, want %v", dir, got, want)
		}
	}
}
