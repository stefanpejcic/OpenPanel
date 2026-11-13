package wordpress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWPProfiles(t *testing.T) {
	dir := t.TempDir()
	orig := wpProfilesDir
	wpProfilesDir = dir
	t.Cleanup(func() { wpProfilesDir = orig })

	os.WriteFile(filepath.Join(dir, "shop.json"), []byte(`{"name":"Shop","plugins":["woocommerce"],"themes":["storefront"]}`), 0644)
	os.WriteFile(filepath.Join(dir, "blog.json"), []byte(`{"plugins":[]}`), 0644)
	os.WriteFile(filepath.Join(dir, "broken.json"), []byte(`{`), 0644)
	os.WriteFile(filepath.Join(dir, "Bad Name.json"), []byte(`{"name":"x"}`), 0644)

	ps := listWPProfiles()
	if len(ps) != 2 || ps[0].Slug != "blog" || ps[0].Name != "blog" || ps[1].Name != "Shop" || ps[1].Plugins[0] != "woocommerce" {
		t.Fatalf("unexpected profiles: %+v", ps)
	}
	if _, ok := loadWPProfile("../shop"); ok {
		t.Fatal("path traversal slug accepted")
	}
	if _, ok := loadWPProfile("missing"); ok {
		t.Fatal("missing profile loaded")
	}
}
