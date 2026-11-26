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

func TestContainerLimitsRejectEnvInjection(t *testing.T) {
	dir := t.TempDir()
	old := homeDirOverride
	homeDirOverride = dir
	defer func() { homeDirOverride = old }()
	envPath := filepath.Join(dir, ".env")
	orig := "USER_ID=\"1000\"\nTOTAL_CPU=\"2\"\nNGINX_CPU=\"1\"\nNGINX_PIDS=\"100\"\n"
	if err := os.WriteFile(envPath, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	evil := "1\"\nTOTAL_CPU=\"64"
	if setContainerEnvLimit("u", "NGINX_CPU", evil) == "" {
		t.Error("setContainerEnvLimit wrote a value with a quote and newline")
	}
	if setContainerEnvLimit("u", "NGINX_CPU\nUSER_ID", "0") == "" {
		t.Error("setContainerEnvLimit took a key with a newline")
	}
	// validation runs before podman is touched, so no container is needed here
	if msg, ok := updateContainerPIDs(t.Context(), "u", "nginx", "100\"\nUSER_ID=\"0"); ok {
		t.Errorf("updateContainerPIDs accepted an injected value: %s", msg)
	}
	data, _ := os.ReadFile(envPath)
	if string(data) != orig {
		t.Fatalf(".env changed:\n%s", data)
	}
	if msg := setContainerEnvLimit("u", "NGINX_CPU", "1.5"); msg != "" {
		t.Fatalf("rejected a normal value: %s", msg)
	}
	data, _ = os.ReadFile(envPath)
	if !strings.Contains(string(data), "NGINX_CPU=\"1.5\"") || strings.Count(string(data), "TOTAL_CPU=") != 1 {
		t.Fatalf("normal update went wrong:\n%s", data)
	}
}

func TestLimitValueError(t *testing.T) {
	cases := []struct {
		action, value string
		ok            bool
	}{
		{"cpu", "1", true}, {"cpu", "0.5", true}, {"cpu", "0", true}, {"cpu", "4", true},
		{"cpu", "5", false}, {"cpu", "1e3", false}, {"cpu", "NaN", false}, {"cpu", "-1", false}, {"cpu", "", false},
		{"cpu", "1\"\nX=\"2", false}, {"cpu", "1 ", false},
		{"ram", "1.5", true}, {"ram", "2G", true}, {"ram", "2g", true}, {"ram", "0", true},
		{"ram", "9", false}, {"ram", "512m", false}, {"ram", "1GG", false}, {"ram", "Inf", false},
		{"pids", "512", true}, {"pids", "0", true}, {"pids", "-1", false}, {"pids", "1.5", false}, {"pids", "1\n", false},
		{"disk", "1", false},
	}
	for _, c := range cases {
		msg := limitValueError(c.action, c.value, 4, 8)
		if (msg == "") != c.ok {
			t.Errorf("%s %q: got %q, want ok=%v", c.action, c.value, msg, c.ok)
		}
	}
	// a plan total of 0 means no cap
	if msg := limitValueError("cpu", "64", 0, 0); msg != "" {
		t.Errorf("uncapped plan rejected 64 cores: %s", msg)
	}
}

func TestBulkLimitRejectsNonNumbers(t *testing.T) {
	for _, v := range []string{"NaN", "Inf", "1e3", "0x10"} {
		if msg, _ := validateBulkLimit("cpu", v, 0, 0); msg == "" {
			t.Errorf("bulk cpu accepted %q", v)
		}
	}
	if msg, _ := validateBulkLimit("pids", "+5", 0, 0); msg == "" {
		t.Error("bulk pids accepted +5")
	}
	if msg, _ := validateBulkLimit("ram", "1.5", 0, 0); msg != "" {
		t.Errorf("bulk ram rejected 1.5: %s", msg)
	}
}
