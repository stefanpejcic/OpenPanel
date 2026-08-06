package mysql

import (
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestMysqlContainerStatusDetail(t *testing.T) {
	cases := []struct {
		state, health string
		wantEmpty     bool
		wantSubstr    string
	}{
		{"running", "healthy", true, ""},
		{"running", "unhealthy", false, "unhealthy"},
		{"running", "starting", false, "still initializing"},
		{"running", "", false, "Unable to retrieve"},
		{"not_found", "", false, "installation is underway"},
		{"created", "", false, "not yet started"},
		{"restarting", "", false, "restarting"},
		{"paused", "", false, "paused"},
		{"exited", "", false, "stopped"},
		{"removing", "", false, "being deleted"},
		{"dead", "", false, "crashed"},
		{"bogus", "", false, "Unable to retrieve"},
	}
	for _, c := range cases {
		got := mysqlContainerStatusDetail(c.state, c.health)
		if c.wantEmpty && got != "" {
			t.Errorf("state=%q health=%q: got %q, want empty", c.state, c.health, got)
		}
		if !c.wantEmpty && !strings.Contains(got, c.wantSubstr) {
			t.Errorf("state=%q health=%q: got %q, want substring %q", c.state, c.health, got, c.wantSubstr)
		}
	}
}

func TestMysqlWarningFlashMessage(t *testing.T) {
	cases := []struct {
		version, state, health string
		wantEmpty              bool
		wantSubstr             string
	}{
		{"mysql", "running", "healthy", true, ""},
		{"mysql", "not_found", "", false, "not yet installed"},
		{"mysql", "exited", "", false, "not accessible"},
		{"mariadb", "running", "unhealthy", false, "still initializing"},
		{"mariadb", "running", "starting", false, "not ready"},
		{"mariadb", "running", "unknown", false, "health status is unknown"},
	}
	for _, c := range cases {
		got := mysqlWarningFlashMessage(c.version, c.state, c.health)
		if c.wantEmpty && got != "" {
			t.Errorf("version=%q state=%q health=%q: got %q, want empty", c.version, c.state, c.health, got)
		}
		if !c.wantEmpty && !strings.Contains(got, c.wantSubstr) {
			t.Errorf("version=%q state=%q health=%q: got %q, want substring %q", c.version, c.state, c.health, got, c.wantSubstr)
		}
	}
}

func TestIsSystemMySQLDatabase(t *testing.T) {
	for _, name := range []string{"information_schema", "mysql", "phpmyadmin", "performance_schema", "sys", "mariadb.sys"} {
		if !isSystemMySQLDatabase(name) {
			t.Errorf("expected %q to be a system database", name)
		}
	}
	if isSystemMySQLDatabase("app_db") {
		t.Error("expected app_db to not be a system database")
	}
}

func TestLoadRestrictedNamesDefaults(t *testing.T) {
	a := &appctx.App{Config: config.Config{}}
	loadRestrictedNames(a)

	if !isRestrictedUser("ROOT") {
		t.Error("expected root (case-insensitive) to be restricted by default")
	}
	if isRestrictedUser("appuser") {
		t.Error("expected appuser to not be restricted")
	}
	if !isRestrictedDatabase("mysql") {
		t.Error("expected mysql db to be restricted by default")
	}
	if isRestrictedDatabase("app_db") {
		t.Error("expected app_db to not be restricted")
	}
}

func TestSplitTrimQuotesAndSQLQuotedList(t *testing.T) {
	got := splitTrimQuotes(`"foo" 'bar' baz`)
	want := []string{"foo", "bar", "baz"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	sql := sqlQuotedList(got)
	if sql != "'foo', 'bar', 'baz'" {
		t.Errorf("sqlQuotedList = %q", sql)
	}
}

func TestEscapeMySQLString(t *testing.T) {
	got := escapeMySQLString(`O'Brien\`)
	want := `O\'Brien\\`
	if got != want {
		t.Errorf("escapeMySQLString = %q, want %q", got, want)
	}
}

func TestToStringCell(t *testing.T) {
	if got := toStringCell(nil); got != "" {
		t.Errorf("nil: got %q", got)
	}
	if got := toStringCell("hello"); got != "hello" {
		t.Errorf("string: got %q", got)
	}
	if got := toStringCell([]byte("world")); got != "world" {
		t.Errorf("[]byte: got %q", got)
	}
	if got := toStringCell(42); got != "" {
		t.Errorf("unsupported type: got %q, want empty", got)
	}
}

func TestToFloatCell(t *testing.T) {
	if got := toFloatCell(float64(1.5)); got != 1.5 {
		t.Errorf("float64: got %v", got)
	}
	if got := toFloatCell([]byte("2.25")); got != 2.25 {
		t.Errorf("[]byte: got %v", got)
	}
	if got := toFloatCell("3.75"); got != 3.75 {
		t.Errorf("string: got %v", got)
	}
	if got := toFloatCell(nil); got != 0 {
		t.Errorf("nil: got %v, want 0", got)
	}
}
