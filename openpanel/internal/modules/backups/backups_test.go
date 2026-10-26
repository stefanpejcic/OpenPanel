package backups

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func TestParseUncommentedEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "backup.env")
	content := "# comment\n\nAWS_S3_BUCKET_NAME=\"mybucket\"\nAWS_ACCESS_KEY_ID=AKIA123\nBACKUP_RETENTION_DAYS=7\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	entries, err := parseUncommentedEnv(path)
	if err != nil {
		t.Fatalf("parseUncommentedEnv: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d: %+v", len(entries), entries)
	}
	if entries[0].Key != "AWS_S3_BUCKET_NAME" || entries[0].Value != "mybucket" {
		t.Errorf("expected quotes stripped, got %+v", entries[0])
	}
	if entries[1].Key != "AWS_ACCESS_KEY_ID" || entries[1].Value != "AKIA123" {
		t.Errorf("unexpected second entry: %+v", entries[1])
	}
}

func TestGroupBySections(t *testing.T) {
	entries := []KV{
		{Key: "AWS_S3_BUCKET_NAME", Value: "mybucket"},
		{Key: "AWS_ACCESS_KEY_ID", Value: "AKIA123"},
		{Key: "BACKUP_RETENTION_DAYS", Value: "7"},
		{Key: "NOTIFICATION_LEVEL", Value: "error"},
	}
	grouped := groupBySections(entries)

	if len(grouped.MatchedSections) != 1 || grouped.MatchedSections[0] != "s3" {
		t.Fatalf("expected only s3 matched, got %v", grouped.MatchedSections)
	}
	if len(grouped.SectionValues["s3"]) != 2 {
		t.Errorf("expected 2 s3 values, got %+v", grouped.SectionValues["s3"])
	}
	if len(grouped.Settings) != 2 {
		t.Errorf("expected 2 leftover settings keys, got %+v", grouped.Settings)
	}
}

func TestGroupBySectionsMultipleTargets(t *testing.T) {
	entries := []KV{
		{Key: "AWS_S3_BUCKET_NAME", Value: "mybucket"},
		{Key: "SSH_HOST_NAME", Value: "example.com"},
	}
	grouped := groupBySections(entries)
	if len(grouped.MatchedSections) != 2 {
		t.Fatalf("expected 2 matched sections, got %v", grouped.MatchedSections)
	}
	// sectionOrder is s3, webdav, ssh, azure, dropbox - s3 must come before ssh.
	if grouped.MatchedSections[0] != "s3" || grouped.MatchedSections[1] != "ssh" {
		t.Errorf("expected [s3 ssh] in sectionOrder order, got %v", grouped.MatchedSections)
	}
}

func TestHasAnyCredentialMarker(t *testing.T) {
	if hasAnyCredentialMarker(nil) {
		t.Error("expected false for empty values")
	}
	if !hasAnyCredentialMarker([]KV{{Key: "AWS_ACCESS_KEY_ID", Value: "x"}}) {
		t.Error("expected true when a marker key has a value")
	}
	if hasAnyCredentialMarker([]KV{{Key: "AWS_ACCESS_KEY_ID", Value: ""}}) {
		t.Error("expected false when marker key value is empty")
	}
	if hasAnyCredentialMarker([]KV{{Key: "AWS_S3_PATH", Value: "x"}}) {
		t.Error("expected false for a non-marker key")
	}
}

func TestClassifyTarListing(t *testing.T) {
	listing := "backup/\n" +
		"backup/html/\n" +
		"backup/html/index.php\n" +
		"backup/html/install/sample.sql\n" +
		"backup/mysql/\n" +
		"backup/mysql/wp_db_2026-01-01_00-00-00.sql\n" +
		"backup/mysql/analytics.sql.gz\n" +
		"backup/crons.ini\n"

	info := classifyTarListing("site.tar.gz", listing)

	if !info.HasFiles {
		t.Error("expected has_files true")
	}
	if !info.HasCrons {
		t.Error("expected has_crons true")
	}
	wantTypes := map[string]bool{"html": true, "mysql": true, "crons": true}
	if len(info.Types) != len(wantTypes) {
		t.Fatalf("expected 3 types, got %v", info.Types)
	}
	for _, ty := range info.Types {
		if !wantTypes[ty] {
			t.Errorf("unexpected type %q", ty)
		}
	}
	wantDBs := map[string]bool{"wp_db": true, "analytics": true}
	if len(info.Databases) != len(wantDBs) {
		t.Fatalf("expected 2 databases (timestamp suffix stripped), got %v", info.Databases)
	}
	for _, db := range info.Databases {
		if !wantDBs[db] {
			t.Errorf("unexpected database %q", db)
		}
	}
}

func TestClassifyTarListingNonBackupPrefixIgnored(t *testing.T) {
	info := classifyTarListing("site.tar.gz", "README.txt\nother/stuff.sql\n")
	if info.HasFiles || info.HasCrons {
		t.Errorf("expected no sections recognized, got %+v", info)
	}
	if len(info.Databases) != 0 {
		t.Errorf("expected no databases recognized outside backup/ prefix, got %v", info.Databases)
	}
}

func TestBelongsToOtherSection(t *testing.T) {
	if !belongsToOtherSection("s3", "SSH_HOST_NAME") {
		t.Error("expected SSH_HOST_NAME to belong to ssh, not s3")
	}
	if belongsToOtherSection("ssh", "SSH_HOST_NAME") {
		t.Error("expected SSH_HOST_NAME to belong to ssh itself, not 'other'")
	}
	if belongsToOtherSection("s3", "SOME_UNKNOWN_KEY") {
		t.Error("expected an unrecognized key to belong to no section")
	}
}

func TestBackupNamePrefix(t *testing.T) {
	cases := []struct {
		config map[string]string
		want   string
	}{
		{map[string]string{"BACKUP_FILENAME": "$USERNAME-%Y-%m-%dT%H-%M-%S.{{ .Extension }}"}, "alice-"},
		{map[string]string{"BACKUP_FILENAME": "${USERNAME}_db-%Y.tar.gz"}, "alice_db-"},
		{map[string]string{"BACKUP_FILENAME": "backup-{{ .Extension }}"}, "backup-"},
		{map[string]string{"BACKUP_FILENAME": "$USERNAME-%Y.tar.gz", "BACKUP_PRUNING_PREFIX": "custom-"}, "custom-"},
		{map[string]string{"BACKUP_FILENAME": "$HOSTNAME-%Y.tar.gz"}, ""},
		{map[string]string{}, ""},
	}
	for _, c := range cases {
		if got := backupNamePrefix(c.config, "alice"); got != c.want {
			t.Errorf("backupNamePrefix(%v) = %q, want %q", c.config, got, c.want)
		}
	}
}

func TestBackupNameAllowed(t *testing.T) {
	cfg := map[string]string{"BACKUP_FILENAME": "$USERNAME-%Y-%m-%dT%H-%M-%S.tar.gz"} // prefix "alice-"
	cases := []struct {
		name         string
		adminManaged bool
		config       map[string]string
		safeName     string
		want         bool
	}{
		{"own destination allows any name", false, cfg, "bob-2026-10-08.tar.gz", true},
		{"shared: own prefix allowed", true, cfg, "alice-2026-10-08.tar.gz", true},
		{"shared: other account's prefix rejected", true, cfg, "bob-2026-10-08.tar.gz", false},
		{"shared: unknown prefix fails closed", true, map[string]string{"BACKUP_FILENAME": "$HOSTNAME-%Y.tar.gz"}, "alice-2026-10-08.tar.gz", false},
		{"shared: no config fails closed", true, map[string]string{}, "alice-2026-10-08.tar.gz", false},
	}
	for _, c := range cases {
		if got := backupNameAllowed(c.adminManaged, c.config, "alice", c.safeName); got != c.want {
			t.Errorf("%s: backupNameAllowed(%v, %q) = %v, want %v", c.name, c.adminManaged, c.safeName, got, c.want)
		}
	}
}

func TestFilterBackupNames(t *testing.T) {
	names := []string{"alice-2026-10-07T00-00-00.tar.gz", "bob-2026-10-07T00-00-00.tar.gz", "alice-2026-10-08T00-00-00.tar.gz"}
	got := filterBackupNames(names, "alice-")
	if len(got) != 2 || got[0] != names[0] || got[1] != names[2] {
		t.Errorf("filterBackupNames kept %v", got)
	}
	if got := filterBackupNames(names, ""); len(got) != 3 {
		t.Errorf("empty prefix should keep everything, kept %v", got)
	}
}

// a refused or failed save has to come back as an error, the API used to report the unchanged settings as success
func TestUpdateBackupSettingsErrors(t *testing.T) {
	bin := t.TempDir()
	podman := filepath.Join(bin, "podman")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	path := filepath.Join(t.TempDir(), "backup.env")
	if err := os.WriteFile(path, []byte("BACKUP_RETENTION_DAYS=7\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	form := url.Values{"settings[BACKUP_RETENTION_DAYS]": {"30"}}
	r := web.WithForm(httptest.NewRequest(http.MethodPut, "/api/backups/settings", nil), form)
	a := &appctx.App{}

	// podman exec ... test -f <lock> succeeding means a backup is running
	if err := os.WriteFile(podman, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if status, _ := updateBackupSettings(a, r, path, "u-test", "u"); status != http.StatusConflict {
		t.Errorf("backup running: status %d, want 409", status)
	}
	if b, _ := os.ReadFile(path); string(b) != "BACKUP_RETENTION_DAYS=7\n" {
		t.Errorf("settings changed while a backup was running: %q", b)
	}

	if err := os.WriteFile(podman, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if status, _ := updateBackupSettings(a, r, t.TempDir(), "u-test", "u"); status != http.StatusInternalServerError {
		t.Errorf("unreadable env file: status %d, want 500", status)
	}
}

// Trim used to strip every quote at either end, so a password ending in " lost it
func TestUnquoteEnv(t *testing.T) {
	cases := map[string]string{
		`"mybucket"`: "mybucket",
		`'x y'`:      "x y",
		`secret"`:    `secret"`,
		`"secret`:    `"secret`,
		`'"x"'`:      `"x"`,
		`"a'`:        `"a'`,
		`""`:         "",
		`"`:          `"`,
		`plain`:      "plain",
	}
	for in, want := range cases {
		if got := unquoteEnv(in); got != want {
			t.Errorf("unquoteEnv(%s) = %q, want %q", in, got, want)
		}
	}
}
