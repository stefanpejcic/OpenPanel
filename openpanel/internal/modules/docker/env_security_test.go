package docker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetEnvValueRejectsLineInjection(t *testing.T) {
	dir := t.TempDir()
	old := homeDirOverride
	homeDirOverride = dir
	defer func() { homeDirOverride = old }()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("USER_ID=\"1000\"\nMYSQL_ROOT_PASSWORD=\"x\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, v := range []string{"p\"\nUSER_ID=\"0", "p\rUSER_ID=0", "p\\", "a\"b"} {
		if SetEnvValue("u", "MYSQL_ROOT_PASSWORD", v) == "" {
			t.Errorf("accepted %q", v)
		}
	}
	if SetEnvValue("u", "BAD KEY", "x") == "" || SetEnvValue("u", "A\nUSER_ID", "x") == "" {
		t.Error("accepted an invalid key")
	}
	if msg := SetEnvValue("u", "MYSQL_ROOT_PASSWORD", "Str0ng!pass$#"); msg != "" {
		t.Fatalf("rejected a normal password: %s", msg)
	}
	data, _ := os.ReadFile(envPath)
	if strings.Count(string(data), "USER_ID=") != 1 {
		t.Fatalf("USER_ID line count changed:\n%s", data)
	}
}

func TestSyncMyCnfPasswordIsNotASedScript(t *testing.T) {
	p := filepath.Join(t.TempDir(), "my.cnf")
	if err := os.WriteFile(p, []byte("[client]\nuser=root\npassword=old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	pass := "x/;1e id > /tmp/marker;#"
	if err := syncMyCnfPassword(p, pass); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), "password="+pass+"\n") || !strings.Contains(string(data), "user=root") {
		t.Fatalf("unexpected my.cnf:\n%s", data)
	}
	if syncMyCnfPassword(p, "a\nuser=evil") == nil {
		t.Fatal("accepted a multi-line password")
	}
}
