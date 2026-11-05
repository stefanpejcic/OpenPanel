package installmanifest

import (
	"os"
	"path/filepath"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

func volume(t *testing.T, home, ctx string) string {
	t.Helper()
	d := filepath.Join(home, ctx, "docker-data", "volumes", ctx+"_html_data", "_data")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestRecordAndRemoveOwnSite(t *testing.T) {
	home := t.TempDir()
	t.Cleanup(tenantfs.SetHomeRootForTest(home))
	site := filepath.Join(volume(t, home, "alice"), "alice.test")
	if err := os.MkdirAll(filepath.Join(site, "wp-admin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(site, "index.php"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Record(site); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(site, "mine.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	entries := Entries(site)
	if len(entries) != 2 {
		t.Fatalf("entries = %v", entries)
	}
	if err := Remove(site, entries); err != nil {
		t.Fatal(err)
	}
	left, _ := os.ReadDir(site)
	if len(left) != 1 || left[0].Name() != "mine.txt" {
		t.Fatalf("left after remove: %v", left)
	}
}

func TestRemoveDoesNotFollowSwappedDir(t *testing.T) {
	home := t.TempDir()
	t.Cleanup(tenantfs.SetHomeRootForTest(home))
	victim := filepath.Join(volume(t, home, "vic"), "vic.test")
	if err := os.MkdirAll(victim, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(victim, "index.php"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Record(victim); err != nil {
		t.Fatal(err)
	}
	// the attacker's own install dir swapped for a link into another account's site
	link := filepath.Join(volume(t, home, "att"), "tfm")
	if err := os.Symlink(victim, link); err != nil {
		t.Fatal(err)
	}
	if got := Entries(link); got != nil {
		t.Errorf("manifest read through the link: %v", got)
	}
	_ = Remove(link, []string{"index.php"})
	_ = ClearContents(link)
	if _, err := os.Stat(filepath.Join(victim, "index.php")); err != nil {
		t.Fatalf("victim file was deleted through the link: %v", err)
	}
	if err := Record(link); err == nil {
		t.Error("Record wrote a manifest through the link")
	}
}
