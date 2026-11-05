package wordpress

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaintenancePageReadWrite(t *testing.T) {
	site := t.TempDir()
	host := maintenancePagePath(site)

	// no wp-content at all: reading is fine, saving says why it can't
	if res, err := readMaintenancePage(host); err != nil || res.Exists || res.Path != "wp-content/maintenance.php" {
		t.Fatalf("missing file: %+v, %v", res, err)
	}
	if err := writeMaintenancePage(host, "<?php echo 1;"); !errors.Is(err, errNoWPContent) {
		t.Fatalf("expected errNoWPContent, got %v", err)
	}

	if err := os.Mkdir(filepath.Join(site, "wp-content"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeMaintenancePage(host, "first"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := os.Chmod(host, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeMaintenancePage(host, "second, shorter"); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	res, err := readMaintenancePage(host)
	if err != nil || !res.Exists || res.Code != "second, shorter" || res.Modified == 0 {
		t.Fatalf("after rewrite: %+v, %v", res, err)
	}
	if st, _ := os.Stat(host); st.Mode().Perm() != 0o600 {
		t.Errorf("rewrite changed the mode to %v", st.Mode().Perm())
	}

	if err := writeMaintenancePage(host, strings.Repeat("x", maintenancePageMax+1)); !errors.Is(err, errMaintenanceTooBig) {
		t.Errorf("expected errMaintenanceTooBig, got %v", err)
	}
	if err := writeMaintenancePage(host, "nul\x00byte"); !errors.Is(err, errMaintenanceText) {
		t.Errorf("expected errMaintenanceText for a NUL byte, got %v", err)
	}
	if err := writeMaintenancePage(host, "bad \xff utf8"); !errors.Is(err, errMaintenanceText) {
		t.Errorf("expected errMaintenanceText for invalid UTF-8, got %v", err)
	}
	if res, _ := readMaintenancePage(host); res.Code != "second, shorter" {
		t.Errorf("a rejected save changed the file: %q", res.Code)
	}
}

func TestMaintenancePageRefusesSymlink(t *testing.T) {
	site := t.TempDir()
	if err := os.Mkdir(filepath.Join(site, "wp-content"), 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	host := maintenancePagePath(site)
	if err := os.Symlink(outside, host); err != nil {
		t.Skip("symlinks not available:", err)
	}
	if _, err := readMaintenancePage(host); !errors.Is(err, errNotRegular) {
		t.Errorf("read through a symlink: %v", err)
	}
	if err := writeMaintenancePage(host, "changed"); !errors.Is(err, errNotRegular) {
		t.Errorf("write through a symlink: %v", err)
	}
	if b, _ := os.ReadFile(outside); string(b) != "keep" {
		t.Errorf("the symlink target changed: %q", b)
	}
}
