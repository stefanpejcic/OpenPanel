package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClientOptions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "my.cnf")
	contents := "[client]\nuser = panel\ndatabase = panel\npassword = secret123\nhost = 185.7.32.112\nprotocol = tcp\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	opts, err := clientOptions(path)
	if err != nil {
		t.Fatalf("clientOptions: %v", err)
	}

	want := map[string]string{
		"user":     "panel",
		"database": "panel",
		"password": "secret123",
		"host":     "185.7.32.112",
		"protocol": "tcp",
	}
	for k, v := range want {
		if opts[k] != v {
			t.Errorf("opts[%q] = %q, want %q", k, opts[k], v)
		}
	}
}

func TestClientOptionsIgnoresOtherSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "my.cnf")
	contents := "[mysqld]\nuser = mysql\n[client]\nuser = panel\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	opts, err := clientOptions(path)
	if err != nil {
		t.Fatalf("clientOptions: %v", err)
	}
	if opts["user"] != "panel" {
		t.Errorf("user = %q, want panel (from [client], not [mysqld])", opts["user"])
	}
}

func TestOpenBuildsDSNFromRealFileOnBox(t *testing.T) {
	if _, err := os.Stat(OptionFile); err != nil {
		t.Skipf("%s not present: %v", OptionFile, err)
	}
	pool, err := Open(OptionFile)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer pool.Close()
}
