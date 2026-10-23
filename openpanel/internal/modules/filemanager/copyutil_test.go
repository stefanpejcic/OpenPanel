package filemanager

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeHomeFile(t *testing.T, home, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readHomeFile(home, name string) string {
	b, err := os.ReadFile(filepath.Join(home, name))
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

func TestRenameNoReplace(t *testing.T) {
	root, home := withUserRoot(t)
	writeHomeFile(t, home, "a.txt", "a")
	writeHomeFile(t, home, "b.txt", "keep me")
	if err := os.Mkdir(filepath.Join(home, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}

	// plain rename(2) would silently replace b.txt here
	if err := renameNoReplace(root.Root, "a.txt", "b.txt"); !errors.Is(err, os.ErrExist) {
		t.Fatalf("rename onto an existing file: err = %v, want ErrExist", err)
	}
	if a, b := readHomeFile(home, "a.txt"), readHomeFile(home, "b.txt"); a != "a" || b != "keep me" {
		t.Fatalf("files changed: a.txt = %q, b.txt = %q", a, b)
	}
	if err := renameNoReplace(root.Root, "a.txt", "dir"); !errors.Is(err, os.ErrExist) {
		t.Fatalf("rename onto an existing dir: err = %v, want ErrExist", err)
	}
	if err := renameNoReplace(root.Root, "a.txt", "a.txt"); err != nil {
		t.Fatalf("rename to the same name: %v", err)
	}
	if err := renameNoReplace(root.Root, "a.txt", "c.txt"); err != nil || readHomeFile(home, "c.txt") != "a" {
		t.Fatalf("rename to a new name: err = %v, c.txt = %q", err, readHomeFile(home, "c.txt"))
	}
}
