package paths

import (
	"os"
	"path/filepath"
	"testing"
)

// swapForSymlink renames dir away and leaves a symlink to target in its place, like a tenant racing a request
func swapForSymlink(t *testing.T, dir, target string) {
	t.Helper()
	if err := os.Rename(dir, dir+"-held"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, dir); err != nil {
		t.Fatal(err)
	}
}

func TestUserRootRelRejectsSymlinks(t *testing.T) {
	home, context := withHomeVolume(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, "escape")); err != nil {
		t.Fatal(err)
	}
	root, err := OpenUserRoot("HOME", context)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	for _, p := range []string{"escape", "escape/x.php", "../x", "a/../../x"} {
		if _, err := root.Rel(p, false); err == nil {
			t.Errorf("Rel(%q) passed, want an error", p)
		}
	}
	for in, want := range map[string]string{"": ".", "/": ".", "a/b.txt": "a/b.txt", "./a//b": "a/b", "/etc/passwd": "etc/passwd"} {
		if got, err := root.Rel(in, false); err != nil || got != want {
			t.Errorf("Rel(%q) = %q, %v, want %q", in, got, err, want)
		}
	}
}

// the reported bug: path checked, parent swapped for a symlink before the write
func TestUserRootWriteAfterParentSwapStaysInside(t *testing.T) {
	home, context := withHomeVolume(t)
	victim := t.TempDir()
	canary := filepath.Join(victim, "canary.php")
	if err := os.WriteFile(canary, []byte("OLD_INERT_MARKER"), 0o600); err != nil {
		t.Fatal(err)
	}
	probe := filepath.Join(home, "probe")
	if err := os.Mkdir(probe, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(probe, "canary.php"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}

	root, err := OpenUserRoot("HOME", context)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	rel, err := root.Rel("probe/canary.php", false)
	if err != nil {
		t.Fatalf("Rel: %v", err)
	}

	swapForSymlink(t, probe, victim)

	if err := root.WriteFileAtomic(rel, []byte("NEW_INERT_MARKER"), 0o644); err == nil {
		t.Fatal("write through a swapped symlink succeeded")
	}
	if err := root.Lchown(rel, os.Getuid(), os.Getgid()); err == nil {
		t.Fatal("lchown through a swapped symlink succeeded")
	}
	if _, err := root.ReadFile(rel); err == nil {
		t.Fatal("read through a swapped symlink succeeded")
	}
	if got, _ := os.ReadFile(canary); string(got) != "OLD_INERT_MARKER" {
		t.Fatalf("victim canary = %q, want it untouched", got)
	}
	if _, err := os.Stat(filepath.Join(victim, "canary.php.tmp")); err == nil {
		t.Fatal("temp file landed in the victim dir")
	}
}

func TestRenameBetweenRefusesSwappedParent(t *testing.T) {
	home, context := withHomeVolume(t)
	victim := t.TempDir()
	if err := os.WriteFile(filepath.Join(victim, "site.php"), []byte("victim"), 0o600); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(home, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	trash, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer trash.Close()

	root, err := OpenUserRoot("HOME", context)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	swapForSymlink(t, sub, victim)

	if err := RenameBetween(root.Root, "sub/site.php", trash, "site.php"); err == nil {
		t.Fatal("rename followed a swapped symlink")
	}
	if _, err := os.Stat(filepath.Join(victim, "site.php")); err != nil {
		t.Fatal("victim file was moved")
	}
}

func TestCopyDirIntoStaysInside(t *testing.T) {
	home, context := withHomeVolume(t)
	victim := t.TempDir()
	if err := os.Symlink(victim, filepath.Join(home, "site")); err != nil {
		t.Fatal(err)
	}
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "index.php"), []byte("deployed"), 0o644); err != nil {
		t.Fatal(err)
	}
	root, err := OpenUserRoot("HOME", context)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()

	if err := CopyDirInto(root.Root, "site", src); err == nil {
		t.Fatal("CopyDirInto through an escaping symlink succeeded")
	}
	if _, err := os.Stat(filepath.Join(victim, "index.php")); err == nil {
		t.Fatal("deploy wrote into the symlink target")
	}
}
