package wordpress

import "testing"

func TestBackupFolderRE(t *testing.T) {
	cases := map[string]bool{
		"2026-01-01_10-00-00": true,
		"2099-12-31_23-59-59": true,
		"wp-content":          false,
		"1999-01-01_00-00-00": false,
	}
	for name, want := range cases {
		if got := backupFolderRE.MatchString(name); got != want {
			t.Errorf("backupFolderRE.MatchString(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestBackupDateRE(t *testing.T) {
	if !backupDateRE.MatchString("2026-01-01_10-00-00") {
		t.Error("expected valid backup date to match")
	}
	if backupDateRE.MatchString("../../etc/passwd") {
		t.Error("path traversal string must not match backup date pattern")
	}
	if backupDateRE.MatchString("2026-01-01") {
		t.Error("date without time component must not match")
	}
}

func TestDBNameAndTablePrefixRE(t *testing.T) {
	content := `
define( 'DB_NAME', 'wp_db_example' );
define( 'DB_USER', 'someuser' );
$table_prefix = 'wp7x_';
`
	m := dbNameRE.FindStringSubmatch(content)
	if m == nil || m[1] != "wp_db_example" {
		t.Fatalf("dbNameRE match = %v, want wp_db_example", m)
	}
	m2 := tablePrefixValueRE.FindStringSubmatch(content)
	if m2 == nil || m2[1] != "wp7x_" {
		t.Fatalf("tablePrefixValueRE match = %v, want wp7x_", m2)
	}
}

func TestToStringCell(t *testing.T) {
	if toStringCell(nil) != "" {
		t.Error("nil should stringify to empty")
	}
	if toStringCell("abc") != "abc" {
		t.Error("string passthrough failed")
	}
	if toStringCell([]byte("abc")) != "abc" {
		t.Error("[]byte should stringify")
	}
	// INTEGER columns (e.g. wp_users.ID) scan into int64 via
	// mysqlmanager.Exec()'s interface{} destinations, not []byte/string -
	// this must stringify correctly, not silently go empty (see
	// toStringCell's doc comment for the autologin bug this caused).
	if toStringCell(int64(42)) != "42" {
		t.Error("int64 should stringify to its decimal representation")
	}
	if toStringCell(uint64(42)) != "42" {
		t.Error("uint64 should stringify to its decimal representation")
	}
	if toStringCell(42) != "42" {
		t.Error("int should stringify to its decimal representation")
	}
	if toStringCell(3.14) != "3.14" {
		t.Error("float64 should stringify to its decimal representation")
	}
	if toStringCell(struct{}{}) != "" {
		t.Error("genuinely unsupported type should stringify to empty")
	}
}
