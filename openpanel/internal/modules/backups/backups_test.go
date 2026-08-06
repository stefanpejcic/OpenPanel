package backups

import (
	"os"
	"path/filepath"
	"testing"
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
