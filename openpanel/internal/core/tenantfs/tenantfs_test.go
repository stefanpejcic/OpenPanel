package tenantfs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// fakeHome makes <tmp>/alice/docker-data/volumes/alice_html_data/_data and a victim dir next to it
func fakeHome(t *testing.T) (volume, victim string) {
	t.Helper()
	tmp := t.TempDir()
	old := homeRoot
	homeRoot = tmp + "/"
	t.Cleanup(func() { homeRoot = old })
	volume = filepath.Join(tmp, "alice", "docker-data", "volumes", "alice_html_data", "_data")
	victim = filepath.Join(tmp, "bob", "docker-data", "volumes", "bob_html_data", "_data")
	for _, d := range []string{volume, victim} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(victim, "wp-config.php"), []byte("VICTIM"), 0o600); err != nil {
		t.Fatal(err)
	}
	return volume, victim
}

func TestVolumeOf(t *testing.T) {
	volume, _ := fakeHome(t)
	cases := []struct {
		in      string
		inside  bool
		rel     string
		escapes bool
	}{
		{volume, true, ".", false},
		{volume + "/", true, ".", false},
		{volume + "/site/index.php", true, "site/index.php", false},
		{volume + "/site/../index.php", true, "index.php", false},
		{volume + "/../../../../bob/docker-data/volumes/bob_html_data/_data/x", false, "", true},
		{volume + "/a/../../_data2", false, "", true},
		{"/etc/openpanel/wordpress/mu-plugin.php", false, "", false},
		{filepath.Dir(volume) + "/wpcli_plugins.txt", false, "", false},
	}
	for _, c := range cases {
		_, rel, ok, err := volumeOf(c.in)
		if c.escapes {
			if !errors.Is(err, ErrEscape) {
				t.Errorf("volumeOf(%q) err = %v, want ErrEscape", c.in, err)
			}
			continue
		}
		if err != nil || ok != c.inside || rel != c.rel {
			t.Errorf("volumeOf(%q) = %q, %v, %v; want %q, %v", c.in, rel, ok, err, c.rel, c.inside)
		}
	}
}

// the installer pattern: account plants a symlink at its docroot, panel writes and chowns there as root
func TestPlantedSymlinkIsNotFollowedOut(t *testing.T) {
	volume, victim := fakeHome(t)
	if err := os.Symlink(victim, filepath.Join(volume, "example.com")); err != nil {
		t.Fatal(err)
	}
	site := filepath.Join(volume, "example.com")

	if err := WriteFile(filepath.Join(site, "wp-config.php"), []byte("PWNED"), 0o644); err == nil {
		t.Error("WriteFile followed the symlink out")
	}
	if _, err := ReadFile(filepath.Join(site, "wp-config.php")); err == nil {
		t.Error("ReadFile followed the symlink out")
	}
	if err := MkdirAll(filepath.Join(site, "wp-content"), 0o755); err == nil {
		t.Error("MkdirAll followed the symlink out")
	}
	if err := RemoveAll(filepath.Join(site, "wp-config.php")); err == nil {
		t.Error("RemoveAll followed the symlink out")
	}
	_ = ChownR(site, os.Getuid(), os.Getgid())
	_ = ChmodR(site, 0o777)
	if err := CopyDir(site, filepath.Join(volume, "copy")); err == nil {
		t.Error("CopyDir read through the symlink")
	}
	if err := Rename(filepath.Join(site, "wp-config.php"), filepath.Join(volume, "stolen.php")); err == nil {
		t.Error("Rename moved a file out of another volume")
	}

	if fi, err := os.Stat(filepath.Join(victim, "wp-config.php")); err != nil || fi.Mode().Perm() != 0o600 {
		t.Fatalf("victim file changed: %v %v", fi, err)
	}
	if got, _ := os.ReadFile(filepath.Join(victim, "wp-config.php")); string(got) != "VICTIM" {
		t.Fatalf("victim content = %q", got)
	}
	if _, err := os.Stat(filepath.Join(volume, "copy", "wp-config.php")); err == nil {
		t.Fatal("victim file was copied into the volume")
	}
}

func TestDotDotOutOfVolumeRefused(t *testing.T) {
	volume, victim := fakeHome(t)
	sneaky := volume + "/../../../../bob/docker-data/volumes/bob_html_data/_data/wp-config.php"
	if err := WriteFile(sneaky, []byte("PWNED"), 0o644); !errors.Is(err, ErrEscape) {
		t.Fatalf("WriteFile(..) err = %v, want ErrEscape", err)
	}
	if got, _ := os.ReadFile(filepath.Join(victim, "wp-config.php")); string(got) != "VICTIM" {
		t.Fatalf("victim content = %q", got)
	}
}

func TestExtractTarStripAndCopyDir(t *testing.T) {
	volume, _ := fakeHome(t)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	tw.WriteHeader(&tar.Header{Name: "wordpress/", Typeflag: tar.TypeDir, Mode: 0o755})
	tw.WriteHeader(&tar.Header{Name: "wordpress/index.php", Typeflag: tar.TypeReg, Mode: 0o644, Size: 5})
	tw.Write([]byte("hello"))
	tw.Close()
	gz.Close()
	archive := filepath.Join(t.TempDir(), "wordpress-latest.tar.gz")
	if err := os.WriteFile(archive, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	site := filepath.Join(volume, "example.com")
	if err := ExtractTar(context.Background(), archive, site, 1); err != nil {
		t.Fatalf("ExtractTar: %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(site, "index.php")); string(got) != "hello" {
		t.Fatalf("index.php = %q", got)
	}
	if err := CopyDir(site, filepath.Join(volume, "staging")); err != nil {
		t.Fatalf("CopyDir: %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(volume, "staging", "index.php")); string(got) != "hello" {
		t.Fatalf("copied index.php = %q", got)
	}
	out := filepath.Join(volume, "backup.tar.gz")
	if err := WriteTarGz(context.Background(), site, out); err != nil {
		t.Fatalf("WriteTarGz: %v", err)
	}
	if err := ExtractTar(context.Background(), out, filepath.Join(volume, "restored"), 0); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if got, _ := os.ReadFile(filepath.Join(volume, "restored", "index.php")); string(got) != "hello" {
		t.Fatalf("restored index.php = %q", got)
	}
}

func TestOutsideVolumesPassThrough(t *testing.T) {
	fakeHome(t)
	dir := t.TempDir()
	p := filepath.Join(dir, "a", "b.txt")
	if err := MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Rename(p, filepath.Join(dir, "c.txt")); err != nil {
		t.Fatal(err)
	}
	if got, err := ReadFile(filepath.Join(dir, "c.txt")); err != nil || string(got) != "x" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestOutputInStaysUnderBase(t *testing.T) {
	volume, victim := fakeHome(t)
	base := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(volume))))
	os.MkdirAll(filepath.Join(volume, "site", "sub"), 0o755)
	os.Symlink(victim, filepath.Join(volume, "link"))

	out, err := OutputIn(context.Background(), base, filepath.Join(volume, "site"), "find", ".", "-type", "d")
	if err != nil {
		t.Fatalf("OutputIn: %v", err)
	}
	if got := string(bytes.TrimSpace(out)); got != ".\n./sub" {
		t.Fatalf("find output = %q", got)
	}
	if _, err := OutputIn(context.Background(), base, filepath.Join(volume, "link"), "ls"); err == nil {
		t.Fatal("OutputIn followed a symlink out of base")
	}
	if _, err := OutputIn(context.Background(), base, base+"/../..", "ls"); err == nil {
		t.Fatal("OutputIn allowed .. out of base")
	}
}
