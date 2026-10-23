package appkit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProtectBackupsDir(t *testing.T) {
	vol := t.TempDir()
	ProtectBackupsDir(vol)
	got, err := os.ReadFile(filepath.Join(vol, "backups", ".htaccess"))
	if err != nil {
		t.Fatalf("no .htaccess written: %v", err)
	}
	if !strings.Contains(string(got), "Require all denied") {
		t.Errorf(".htaccess = %q", got)
	}

	// a user's own .htaccess there is left alone
	custom := filepath.Join(t.TempDir(), "backups")
	if err := os.MkdirAll(custom, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(custom, ".htaccess"), []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ProtectBackupsDir(filepath.Dir(custom))
	if got, _ := os.ReadFile(filepath.Join(custom, ".htaccess")); string(got) != "mine\n" {
		t.Errorf("existing .htaccess was overwritten: %q", got)
	}
}
