package filemanager

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestFilesPerPageNeverZero(t *testing.T) {
	// 0 passes OpenAdmin's non-negative check and used to crash every listing with a division by zero
	for value, want := range map[string]int{"0": 500, "-5": 500, "": 500, "abc": 500, "50": 50, "1000": 1000} {
		a := &appctx.App{Config: config.Config{"filemanager_files_per_page": value}}
		if got := filesPerPage(a); got != want {
			t.Errorf("filesPerPage(%q) = %d, want %d", value, got, want)
		}
	}
}

// README.md used to be read whole on every listing, whatever its size or type
func TestReadReadme(t *testing.T) {
	root, home := withUserRoot(t)
	write := func(name string, size int) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(home, filepath.Dir(name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(home, name), []byte(strings.Repeat("x", size)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("small/README.md", 100)
	write("big/README.md", readmeMaxBytes+1)
	if err := os.MkdirAll(filepath.Join(home, "dir/README.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "fifo"), 0o755); err != nil {
		t.Fatal(err)
	}
	// reading a fifo would block the listing until something writes to it
	if err := syscall.Mkfifo(filepath.Join(home, "fifo/README.md"), 0o644); err != nil {
		t.Fatal(err)
	}

	if data, ok := readReadme(root, "small/README.md"); !ok || len(data) != 100 {
		t.Errorf("small README: ok %v, %d bytes", ok, len(data))
	}
	for _, rel := range []string{"big/README.md", "dir/README.md", "fifo/README.md", "none/README.md"} {
		if _, ok := readReadme(root, rel); ok {
			t.Errorf("%s was read", rel)
		}
	}
}
