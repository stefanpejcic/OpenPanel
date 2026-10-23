package trash

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
)

func TestSplitRestoreLine(t *testing.T) {
	cases := []struct {
		line, name, path, date string
		ok                     bool
	}{
		{"site.php=/home/u/_data/site.php|deletion_date=2026-10-22T10:00:00", "site.php", "/home/u/_data/site.php", "2026-10-22T10:00:00", true},
		// this one used to be keyed "a" and could never be restored
		{"a=b.txt=/home/u/_data/a=b.txt|deletion_date=2026-10-22T10:00:00", "a=b.txt", "/home/u/_data/a=b.txt", "2026-10-22T10:00:00", true},
		{"x==/home/u/_data/x=|deletion_date=d", "x=", "/home/u/_data/x=", "d", true},
		{"old.txt=/var/www/html/old.txt", "old.txt", "/var/www/html/old.txt", "", true},
		{"rel=some/relative/path", "rel", "some/relative/path", "", true},
		{"no equals sign", "", "", "", false},
		{"=/home/u/_data/x", "", "", "", false},
		{"", "", "", "", false},
	}
	for _, c := range cases {
		name, meta, ok := splitRestoreLine(c.line)
		if ok != c.ok || name != c.name || meta.originalPath != c.path || meta.deletionDate != c.date {
			t.Errorf("splitRestoreLine(%q) = %q, %+v, %v, want %q, {%q %q}, %v", c.line, name, meta, ok, c.name, c.path, c.date, c.ok)
		}
	}
}

func TestDropRestoreLine(t *testing.T) {
	content := "a=/d/a|deletion_date=1\na=b.txt=/d/a=b.txt|deletion_date=2\nab=/d/ab|deletion_date=3\n"
	// deleting "a" used to drop a=b.txt's line too, since both start with "a="
	got := dropRestoreLine(content, "a")
	want := []string{"a=b.txt=/d/a=b.txt|deletion_date=2", "ab=/d/ab|deletion_date=3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dropRestoreLine = %q, want %q", got, want)
	}
}

func writeTestFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

func TestRestoreItemRefusesExistingTarget(t *testing.T) {
	dir := t.TempDir()
	trashed := filepath.Join(dir, "Trash", "notes.txt")
	dest := filepath.Join(dir, "site", "notes.txt")
	writeTestFile(t, trashed, "old version")
	writeTestFile(t, dest, "new version")

	// rename(2) would replace the newer file without a word
	if err := restoreItem(trashed, dest); !errors.Is(err, errRestoreExists) {
		t.Fatalf("restoreItem onto an existing file: err = %v, want errRestoreExists", err)
	}
	if got := readTestFile(dest); got != "new version" {
		t.Fatalf("existing file = %q, want it untouched", got)
	}
	if got := readTestFile(trashed); got != "old version" {
		t.Fatalf("trashed file = %q, want it still in the trash", got)
	}
}

func TestRestoreItemRecreatesDeletedFolder(t *testing.T) {
	dir := t.TempDir()
	trashed := filepath.Join(dir, "Trash", "notes.txt")
	dest := filepath.Join(dir, "site", "blog", "notes.txt")
	writeTestFile(t, trashed, "x")
	asRoot := os.Geteuid() == 0
	if asRoot {
		if err := os.Lchown(trashed, 4321, 4321); err != nil {
			t.Fatal(err)
		}
	}

	// site/blog is gone, this used to fail with a raw "rename ...: no such file or directory"
	if err := restoreItem(trashed, dest); err != nil {
		t.Fatalf("restoreItem into a deleted folder: %v", err)
	}
	if got := readTestFile(dest); got != "x" {
		t.Fatalf("restored file = %q", got)
	}
	if !asRoot {
		t.Log("not root, skipping the owner check on recreated folders")
		return
	}
	for _, d := range []string{"site", "site/blog"} {
		info, err := os.Stat(filepath.Join(dir, d))
		if err != nil {
			t.Fatal(err)
		}
		if st := info.Sys().(*syscall.Stat_t); st.Uid != 4321 || st.Gid != 4321 {
			t.Errorf("%s owned by %d:%d, want the restored item's owner 4321:4321", d, st.Uid, st.Gid)
		}
	}
}
