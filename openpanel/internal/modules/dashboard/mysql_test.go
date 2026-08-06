package dashboard

import (
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestStripQuotes(t *testing.T) {
	cases := map[string]string{
		`"hello"`: "hello",
		`'hello'`: "hello",
		"hello":   "hello",
		`"`:       `"`,
		"":        "",
		`'a"b'`:   `a"b`,
	}
	for in, want := range cases {
		if got := stripQuotes(in); got != want {
			t.Errorf("stripQuotes(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRestrictedDatabasesSQLDefault(t *testing.T) {
	got := restrictedDatabasesSQL(config.Config{})
	want := "'information_schema', 'performance_schema', 'mysql', 'phpmyadmin', 'sys', 'mariadb.sys'"
	if got != want {
		t.Errorf("restrictedDatabasesSQL(default) = %q, want %q", got, want)
	}
}

func TestRestrictedDatabasesSQLFromConfig(t *testing.T) {
	cfg := config.Config{"mysql_restricted_databases": `"foo bar"`}
	got := restrictedDatabasesSQL(cfg)
	want := "'foo', 'bar'"
	if got != want {
		t.Errorf("restrictedDatabasesSQL(custom) = %q, want %q", got, want)
	}
}
