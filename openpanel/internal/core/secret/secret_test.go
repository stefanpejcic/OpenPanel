package secret

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsExistingFileTrimmed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.key")
	if err := os.WriteFile(path, []byte("  abc123def456  \n"), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	key, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if string(key) != "abc123def456" {
		t.Errorf("key = %q, want abc123def456", key)
	}
}

func TestLoadGeneratesEphemeralKeyWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.key")

	key, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(key) != 64 { // hex.EncodeToString(32 random bytes)
		t.Errorf("generated key length = %d, want 64", len(key))
	}

	key2, err := Load(path)
	if err != nil {
		t.Fatalf("Load (second call): %v", err)
	}
	if string(key) == string(key2) {
		t.Error("expected two ephemeral generations to differ")
	}
}
