package filemanager

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecursiveModes(t *testing.T) {
	cases := []struct {
		perm, folder string
		mode, dir    os.FileMode
		bad          bool
	}{
		{"644", "755", 0o644, 0o755, false},
		{"0640", "", 0o640, 0o640, false}, // no folder mode: folders get the file mode, like before
		{"644", "4755", 0, 0, true},       // special bits refused for folders too
		{"644", "75", 0, 0, true},
		{"999", "755", 0, 0, true},
	}
	for _, c := range cases {
		mode, dir, msg := recursiveModes(c.perm, c.folder)
		if (msg != "") != c.bad {
			t.Errorf("%s/%s: msg %q", c.perm, c.folder, msg)
			continue
		}
		if !c.bad && (mode != c.mode || dir != c.dir) {
			t.Errorf("%s/%s: got %o/%o, want %o/%o", c.perm, c.folder, mode, dir, c.mode, c.dir)
		}
	}
}

func TestChmodPathSplitsFilesAndFolders(t *testing.T) {
	root, home := withUserRoot(t)
	site := filepath.Join(home, "site")
	for _, d := range []string{"site", "site/wp-content", "site/wp-content/uploads"} {
		if err := os.MkdirAll(filepath.Join(home, d), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range []string{"site/index.php", "site/wp-content/a.php", "site/wp-content/uploads/b.jpg"} {
		if err := os.WriteFile(filepath.Join(home, f), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(home, "site/index.php"), filepath.Join(site, "link.php")); err != nil {
		t.Skip("symlinks not available:", err)
	}
	rel, err := root.Rel("site", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := chmodPath(root, rel, 0o644, 0o755, true); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]os.FileMode{
		"site": 0o755, "site/wp-content": 0o755, "site/wp-content/uploads": 0o755,
		"site/index.php": 0o644, "site/wp-content/a.php": 0o644, "site/wp-content/uploads/b.jpg": 0o644,
	} {
		st, err := os.Stat(filepath.Join(home, path))
		if err != nil {
			t.Fatal(err)
		}
		if got := st.Mode().Perm(); got != want {
			t.Errorf("%s: %o, want %o", path, got, want)
		}
	}

	// not recursive: only the selected item, with the file mode
	if err := chmodPath(root, rel, 0o750, 0o711, false); err != nil {
		t.Fatal(err)
	}
	if st, _ := os.Stat(site); st.Mode().Perm() != 0o750 {
		t.Errorf("non-recursive on a folder: %o, want 750", st.Mode().Perm())
	}
	if st, _ := os.Stat(filepath.Join(home, "site/index.php")); st.Mode().Perm() != 0o644 {
		t.Errorf("non-recursive touched a file inside: %o", st.Mode().Perm())
	}
}

func TestChmodAction(t *testing.T) {
	for _, c := range []struct {
		perm, folder string
		rec          bool
		want         string
	}{
		{"644", "755", true, "changed permissions to 644 for files and 755 for folders recursively"},
		{"755", "755", true, "changed permissions to 755 recursively"},
		{"755", "", true, "changed permissions to 755 recursively"},
		{"644", "755", false, "changed permissions to 644"},
	} {
		if got := chmodAction(c.perm, c.folder, c.rec); got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}
}
