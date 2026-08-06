package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "openpanel.config")
	contents := "ns1=ns1.example.com\nns2=ns2.example.com\n# a comment\n\ntwofa_nag=yes\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writing test fixture: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got := cfg.Get("ns1", ""); got != "ns1.example.com" {
		t.Errorf("ns1 = %q, want ns1.example.com", got)
	}
	if got := cfg.Get("twofa_nag", ""); got != "yes" {
		t.Errorf("twofa_nag = %q, want yes", got)
	}
	if got := cfg.Get("missing_key", "fallback"); got != "fallback" {
		t.Errorf("missing_key = %q, want fallback", got)
	}
}

func TestLoadMissingFileIsNotAnError(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "does-not-exist.config"))
	if err != nil {
		t.Fatalf("Load on missing file returned error: %v", err)
	}
	if len(cfg) != 0 {
		t.Errorf("expected empty config, got %v", cfg)
	}
}
