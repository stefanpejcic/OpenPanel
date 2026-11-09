package wordpress

import (
	"os"
	"path/filepath"
	"testing"
)

// https://github.com/stefanpejcic/OpenPanel/issues/1298
func TestGetWPConfigDBInfoOneLevelUp(t *testing.T) {
	vol := t.TempDir()
	domain := filepath.Join(vol, "b.com")
	site := filepath.Join(domain, "stats")
	if err := os.MkdirAll(site, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := func(dir, db string) {
		if err := os.WriteFile(filepath.Join(dir, "wp-config.php"), []byte("<?php\ndefine( 'DB_NAME', '"+db+"' );\n$table_prefix = 'wp_';\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if _, _, err := getWPConfigDBInfo(vol, site); err == nil {
		t.Error("no config anywhere should fail")
	}
	cfg(domain, "moved_db")
	if db, _, err := getWPConfigDBInfo(vol, site); err != nil || db != "moved_db" {
		t.Errorf("moved config = %q, %v", db, err)
	}
	cfg(site, "own_db")
	if db, _, _ := getWPConfigDBInfo(vol, site); db != "own_db" {
		t.Errorf("docroot config should win, got %q", db)
	}
	os.Remove(filepath.Join(site, "wp-config.php"))
	if err := os.WriteFile(filepath.Join(domain, "wp-settings.php"), []byte("<?php"), 0o644); err != nil {
		t.Fatal(err)
	}
	if db, _, err := getWPConfigDBInfo(vol, site); err == nil {
		t.Errorf("parent is its own WordPress install, its config must not be used, got %q", db)
	}

	// docroot at the volume root, the folder above isn't the user's
	root := filepath.Join(t.TempDir(), "_data")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg(filepath.Dir(root), "outside_db")
	if db, _, err := getWPConfigDBInfo(root, root); err == nil {
		t.Errorf("looked outside the html volume, got %q", db)
	}
}
