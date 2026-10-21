package filemanager

import (
	"archive/tar"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
)

// withUserRoot gives a temp HOME volume for "alice" and an open root on it
func withUserRoot(t *testing.T) (*paths.UserRoot, string) {
	t.Helper()
	tmp := t.TempDir()
	home := filepath.Join(tmp, "alice")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(paths.SetHomeVolumeForTest(tmp))
	root, err := paths.OpenUserRoot("HOME", "alice")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return root, home
}

func writeTar(t *testing.T, root *paths.UserRoot, name string, add func(tw *tar.Writer)) {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	add(tw)
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := root.WriteFile(name, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestArchiveRoundTrip(t *testing.T) {
	for _, ext := range []string{"zip", "tar", "tar.gz"} {
		t.Run(ext, func(t *testing.T) {
			root, home := withUserRoot(t)
			os.MkdirAll(filepath.Join(home, "site", "css"), 0o755)
			os.WriteFile(filepath.Join(home, "site", "index.php"), []byte("<?php echo 1;"), 0o644)
			os.WriteFile(filepath.Join(home, "site", "css", "a.css"), []byte("body{}"), 0o644)
			os.Symlink("index.php", filepath.Join(home, "site", "link.php"))

			ctx := context.Background()
			if err := createArchive(ctx, root, ext, ".", "out."+ext, []string{"site"}, 1); err != nil {
				t.Fatalf("createArchive: %v", err)
			}
			if err := validateArchiveMembers(root, "out."+ext, "out."+ext, "restored"); err != nil {
				t.Fatalf("validate: %v", err)
			}
			if err := root.MkdirAll("restored", 0o755); err != nil {
				t.Fatal(err)
			}
			if err := extractArchive(ctx, root, "out."+ext, "out."+ext, "restored"); err != nil {
				t.Fatalf("extract: %v", err)
			}
			if got, _ := os.ReadFile(filepath.Join(home, "restored", "site", "css", "a.css")); string(got) != "body{}" {
				t.Fatalf("a.css = %q", got)
			}
			if link, err := os.Readlink(filepath.Join(home, "restored", "site", "link.php")); err != nil || link != "index.php" {
				t.Fatalf("link.php = %q, %v, want the symlink kept as a link", link, err)
			}
		})
	}
}

func TestCreateArchiveDoesNotFollowSymlinkOut(t *testing.T) {
	root, home := withUserRoot(t)
	victim := t.TempDir()
	os.WriteFile(filepath.Join(victim, "wp-config.php"), []byte("DB_PASSWORD"), 0o600)
	os.MkdirAll(filepath.Join(home, "site"), 0o755)
	os.Symlink(victim, filepath.Join(home, "site", "other"))

	if err := createArchive(context.Background(), root, "tar", ".", "out.tar", []string{"site"}, 1); err != nil {
		t.Fatalf("createArchive: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(home, "out.tar"))
	if bytes.Contains(data, []byte("DB_PASSWORD")) {
		t.Fatal("archive pulled in a file from outside the volume")
	}
}

// a member that sneaks through a symlink created earlier in the same archive has to stay inside
func TestExtractTarSymlinkThenFileStaysInside(t *testing.T) {
	root, home := withUserRoot(t)
	victim := t.TempDir()
	writeTar(t, root, "evil.tar", func(tw *tar.Writer) {
		tw.WriteHeader(&tar.Header{Name: "d", Typeflag: tar.TypeSymlink, Linkname: victim})
		tw.WriteHeader(&tar.Header{Name: "d/pwn.php", Typeflag: tar.TypeReg, Mode: 0o644, Size: 4})
		tw.Write([]byte("pwnd"))
	})

	if err := validateArchiveMembers(root, "evil.tar", "evil.tar", "."); err == nil {
		t.Fatal("validation passed an absolute symlink")
	}
	// even with validation skipped the root keeps it in
	if err := extractArchive(context.Background(), root, "evil.tar", "evil.tar", "."); err == nil {
		t.Fatal("extracting through an escaping symlink succeeded")
	}
	if _, err := os.Stat(filepath.Join(victim, "pwn.php")); err == nil {
		t.Fatal("member was written outside the volume")
	}
	_ = home
}

func TestExtractIntoSwappedDestinationStaysInside(t *testing.T) {
	root, home := withUserRoot(t)
	victim := t.TempDir()
	os.WriteFile(filepath.Join(victim, "index.php"), []byte("victim"), 0o600)
	writeTar(t, root, "site.tar", func(tw *tar.Writer) {
		tw.WriteHeader(&tar.Header{Name: "index.php", Typeflag: tar.TypeReg, Mode: 0o644, Size: 3})
		tw.Write([]byte("new"))
	})
	dest, err := root.Rel("dest", false)
	if err != nil {
		t.Fatal(err)
	}
	os.Mkdir(filepath.Join(home, "dest"), 0o755)
	if err := validateArchiveMembers(root, "site.tar", "site.tar", dest); err != nil {
		t.Fatal(err)
	}

	os.Rename(filepath.Join(home, "dest"), filepath.Join(home, "dest-held"))
	os.Symlink(victim, filepath.Join(home, "dest"))

	if err := extractArchive(context.Background(), root, "site.tar", "site.tar", dest); err == nil {
		t.Fatal("extract followed the swapped destination")
	}
	if got, _ := os.ReadFile(filepath.Join(victim, "index.php")); string(got) != "victim" {
		t.Fatalf("victim index.php = %q", got)
	}
}

func TestIsSafeLink(t *testing.T) {
	cases := map[[2]string]bool{
		{"a/link", "b"}:          true,
		{"a/link", "../b"}:       true,
		{"a/link", "../../b"}:    false,
		{"link", "/etc/passwd"}:  false,
		{"link", ".htaccess"}:    true,
		{"a/b/link", "../../.."}: false,
	}
	for c, want := range cases {
		if got := isSafeLink("dest", c[0], c[1]); got != want {
			t.Errorf("isSafeLink(%q -> %q) = %v, want %v", c[0], c[1], got, want)
		}
	}
}

func TestRunOnDirListsPinnedDir(t *testing.T) {
	root, home := withUserRoot(t)
	os.MkdirAll(filepath.Join(home, "site"), 0o755)
	os.WriteFile(filepath.Join(home, "site", ".htaccess"), []byte("x"), 0o644)
	out, err := runOnDir(context.Background(), root, "site", "ls", "-la")
	if err != nil {
		t.Skipf("ls on /proc/self/fd not available here: %v", err)
	}
	names := map[string]bool{}
	for _, e := range parseLsOutput(string(out)) {
		names[e.Name] = true
	}
	if !names[".htaccess"] || len(names) != 1 {
		t.Fatalf("listing = %v\n%s", names, strings.TrimSpace(string(out)))
	}
}
