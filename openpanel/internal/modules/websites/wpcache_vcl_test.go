package websites

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVCLFileRefusesSymlink(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "default.vcl"), []byte(testVCL), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := readVCL(home); err != nil || string(got) != testVCL {
		t.Fatalf("regular default.vcl: %v", err)
	}
	if err := writeVCL(home, []byte(testVCL+"# changed\n")); err != nil {
		t.Fatalf("write regular default.vcl: %v", err)
	}

	// another account's vcl, which the panel would have rewritten as root through the link
	other := filepath.Join(t.TempDir(), "default.vcl")
	if err := os.WriteFile(other, []byte(testVCL), 0o644); err != nil {
		t.Fatal(err)
	}
	linked := t.TempDir()
	if err := os.Symlink(other, filepath.Join(linked, "default.vcl")); err != nil {
		t.Skip("symlinks not available:", err)
	}
	if _, err := readVCL(linked); err == nil {
		t.Error("reading through a default.vcl symlink should fail")
	}
	if err := writeVCL(linked, []byte("pwned")); err == nil {
		t.Error("writing through a default.vcl symlink should fail")
	}
	if got, _ := os.ReadFile(other); string(got) != testVCL {
		t.Errorf("the link target was changed: %q", got)
	}
}
