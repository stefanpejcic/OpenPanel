package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidServiceName(t *testing.T) {
	valid := []string{"abc", "mysql1", "a12"}
	invalid := []string{"ab", "1abc", "ABC", "a-b", ""}
	for _, s := range valid {
		if !IsValidServiceName(s) {
			t.Errorf("expected %q to be a valid service name", s)
		}
	}
	for _, s := range invalid {
		if IsValidServiceName(s) {
			t.Errorf("expected %q to be an invalid service name", s)
		}
	}
}

func TestIsValidCPULimit(t *testing.T) {
	valid := []string{"1", "0.5", "2.5", "10"}
	invalid := []string{"0", "-1", "abc", "", "0.0"}
	for _, s := range valid {
		if !IsValidCPULimit(s) {
			t.Errorf("expected %q to be a valid CPU limit", s)
		}
	}
	for _, s := range invalid {
		if IsValidCPULimit(s) {
			t.Errorf("expected %q to be an invalid CPU limit", s)
		}
	}
}

func TestIsValidRAMLimit(t *testing.T) {
	valid := []string{"512M", "1.5G", "1g", "100m"}
	invalid := []string{"512", "1.5", "G", "", "512X"}
	for _, s := range valid {
		if !IsValidRAMLimit(s) {
			t.Errorf("expected %q to be a valid RAM limit", s)
		}
	}
	for _, s := range invalid {
		if IsValidRAMLimit(s) {
			t.Errorf("expected %q to be an invalid RAM limit", s)
		}
	}
}

func TestIsValidPIDsLimit(t *testing.T) {
	valid := []string{"1", "100", "5000"}
	invalid := []string{"0", "-1", "abc", "", "1.5"}
	for _, s := range valid {
		if !IsValidPIDsLimit(s) {
			t.Errorf("expected %q to be a valid PIDs limit", s)
		}
	}
	for _, s := range invalid {
		if IsValidPIDsLimit(s) {
			t.Errorf("expected %q to be an invalid PIDs limit", s)
		}
	}
}

func TestServiceKeyPrefix(t *testing.T) {
	cases := map[string]string{
		"mysql":      "MYSQL",
		"my-service": "MY_SERVICE",
		"my.service": "MY_SERVICE",
		"web-1.2":    "WEB_1_2",
	}
	for in, want := range cases {
		if got := ServiceKeyPrefix(in); got != want {
			t.Errorf("ServiceKeyPrefix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseEnvVars(t *testing.T) {
	envVars, newEnvVars := ParseEnvVars([]string{"FOO: bar", "  BAZ  :  qux  ", "malformed line", ""}, "SVC")
	if envVars["FOO"] != "${SVC_FOO}" {
		t.Errorf("envVars[FOO] = %q, want ${SVC_FOO}", envVars["FOO"])
	}
	if newEnvVars["SVC_FOO"] != "bar" {
		t.Errorf("newEnvVars[SVC_FOO] = %q, want bar", newEnvVars["SVC_FOO"])
	}
	if envVars["BAZ"] != "${SVC_BAZ}" || newEnvVars["SVC_BAZ"] != "qux" {
		t.Errorf("BAZ not parsed correctly: envVars=%v newEnvVars=%v", envVars, newEnvVars)
	}
	if len(envVars) != 2 {
		t.Errorf("expected 2 env vars parsed, got %d: %v", len(envVars), envVars)
	}
}

func TestVolumeEntryComposeVolumeString(t *testing.T) {
	rw := VolumeEntry{Name: "data", Mount: "/var/data"}
	if got := rw.ComposeVolumeString(); got != "data:/var/data" {
		t.Errorf("got %q, want data:/var/data", got)
	}
	ro := VolumeEntry{Name: "data", Mount: "/var/data", ReadOnly: true}
	if got := ro.ComposeVolumeString(); got != "data:/var/data:ro" {
		t.Errorf("got %q, want data:/var/data:ro", got)
	}
}

func TestParseVolumeEntries(t *testing.T) {
	names := []string{"vol1", "vol2", ""}
	mounts := []string{"/mnt1", "/mnt2", "/mnt3"}
	readonly := []string{"on", "", ""}

	got := ParseVolumeEntries(names, mounts, readonly, false)
	want := []string{"vol1:/mnt1:ro", "vol2:/mnt2"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseVolumeEntriesAddSocket(t *testing.T) {
	got := ParseVolumeEntries(nil, nil, nil, true)
	if len(got) != 1 || got[0] != "/run/user/${USER_ID}/podman/podman.sock:/var/run/docker.sock:ro" {
		t.Errorf("expected the docker socket entry, got %v", got)
	}
}

func TestResolveEnvPlaceholder(t *testing.T) {
	env := map[string]string{"SVC_CPU": "2"}
	if got := ResolveEnvPlaceholder("${SVC_CPU}", env); got != "2" {
		t.Errorf("got %q, want 2", got)
	}
	if got := ResolveEnvPlaceholder("${SVC_RAM:-512M}", env); got != "512M" {
		t.Errorf("got %q, want 512M (default, not in env)", got)
	}
	if got := ResolveEnvPlaceholder("plain-value", env); got != "plain-value" {
		t.Errorf("got %q, want plain-value unchanged", got)
	}
}

func TestLoadSaveCompose(t *testing.T) {
	dir := t.TempDir()
	oldHome := homeDirOverride
	homeDirOverride = dir
	defer func() { homeDirOverride = oldHome }()

	got, err := LoadCompose("testuser")
	if err != nil {
		t.Fatalf("LoadCompose (missing file): %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map for missing compose file, got %v", got)
	}

	data := map[string]any{"services": map[string]any{"web": map[string]any{"image": "nginx"}}}
	if err := SaveCompose("testuser", data); err != nil {
		t.Fatalf("SaveCompose: %v", err)
	}

	reloaded, err := LoadCompose("testuser")
	if err != nil {
		t.Fatalf("LoadCompose (reload): %v", err)
	}
	services := servicesOf(reloaded)
	if services["web"]["image"] != "nginx" {
		t.Errorf("reloaded compose data = %v, want services.web.image=nginx", reloaded)
	}
}

func TestLoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("FOO=\"bar\"\nBAZ=qux\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldHome := homeDirOverride
	homeDirOverride = dir
	defer func() { homeDirOverride = oldHome }()

	env := LoadEnvFile("testuser")
	if env["FOO"] != "bar" || env["BAZ"] != "qux" {
		t.Errorf("LoadEnvFile() = %v", env)
	}
}

func TestGetSetEnvValue(t *testing.T) {
	dir := t.TempDir()
	oldHome := homeDirOverride
	homeDirOverride = dir
	defer func() { homeDirOverride = oldHome }()

	if _, ok := GetEnvValue("testuser", "MISSING"); ok {
		t.Error("expected ok=false for a missing .env file")
	}

	if msg := SetEnvValue("testuser", "web_server", "nginx"); msg != "" {
		t.Fatalf("SetEnvValue: %v", msg)
	}
	if v, ok := GetEnvValue("testuser", "WEB_SERVER"); !ok || v != "nginx" {
		t.Errorf("GetEnvValue after set = (%q, %v), want (nginx, true)", v, ok)
	}

	if msg := SetEnvValue("testuser", "web_server", "apache"); msg != "" {
		t.Fatalf("SetEnvValue (update): %v", msg)
	}
	if v, _ := GetEnvValue("testuser", "WEB_SERVER"); v != "apache" {
		t.Errorf("expected updated value apache, got %q", v)
	}
}

func TestSetEnvValueRestrictedKey(t *testing.T) {
	dir := t.TempDir()
	oldHome := homeDirOverride
	homeDirOverride = dir
	defer func() { homeDirOverride = oldHome }()

	if msg := SetEnvValue("testuser", "username", "hacker"); msg == "" {
		t.Error("expected an error message when setting a restricted key")
	}
}

func TestSaveEnvFileKeepsCommentsAndOrder(t *testing.T) {
	dir := t.TempDir()
	old := homeDirOverride
	homeDirOverride = dir
	defer func() { homeDirOverride = old }()

	in := "# REDIS\nREDIS_CPU=\"0.1\"\nMONGODB_VERSION=8.0\n#PROXY_HTTP_PORT=\"1:80\"\n\n# MY APP\nMY_APP_CPU=\"1\"\nMY_APP_RAM=\"1G\"\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(in), 0o644); err != nil {
		t.Fatal(err)
	}
	env := LoadEnvFile("u")
	delete(env, "MY_APP_CPU")
	delete(env, "MY_APP_RAM")
	env["REDIS_CPU"] = "0.2"
	env["B_NEW"] = "x"
	env["A_NEW"] = "y"
	if err := SaveEnvFile("u", env); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(dir, ".env"))
	want := "# REDIS\nREDIS_CPU=\"0.2\"\nMONGODB_VERSION=8.0\n#PROXY_HTTP_PORT=\"1:80\"\n\n# MY APP\nA_NEW=\"y\"\nB_NEW=\"x\"\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
