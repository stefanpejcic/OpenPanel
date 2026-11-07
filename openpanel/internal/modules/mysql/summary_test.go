package mysql

import (
	"testing"
)

func TestShortServerVersion(t *testing.T) {
	cases := map[string]string{
		"13.0.2-MariaDB-ubu2604":      "13.0.2",
		"8.4.3":                       "8.4.3",
		" 11.4.5-MariaDB-log ":        "11.4.5",
		"":                            "",
		"MariaDB":                     "",
		"5.7.44-percona-sure1-log-xx": "5.7.44",
	}
	for in, want := range cases {
		if got := shortServerVersion(in); got != want {
			t.Errorf("shortServerVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGroupUserGrants(t *testing.T) {
	all := "ALTER,ALTER ROUTINE,CREATE,CREATE ROUTINE,CREATE TEMPORARY TABLES,CREATE VIEW,DELETE,DELETE HISTORY,DROP,EVENT,EXECUTE,INDEX,INSERT,LOCK TABLES,REFERENCES,SELECT,SHOW CREATE ROUTINE,SHOW VIEW,TRIGGER,UPDATE"
	// rows as information_schema.SCHEMA_PRIVILEGES returns them on MariaDB 13
	rows := [][3]string{
		{"'app'@'%'", "testinguser_bkfix", all},
		{"'reader'@'%'", "shop", "INSERT,SELECT"},
		{"'reader'@'localhost'", "shop", "SELECT,UPDATE"},
		{"'reader'@'%'", "blog", "DELETE,INSERT,SELECT,UPDATE"},
		{"broken", "x", "SELECT"},
	}
	got := groupUserGrants(rows)
	if len(got) != 2 {
		t.Fatalf("got %d users, want 2: %v", len(got), got)
	}
	if g := got["app"]; len(g) != 1 || g[0].Database != "testinguser_bkfix" || g[0].Summary() != "ALL" {
		t.Errorf("app grants = %+v", g)
	}
	reader := got["reader"]
	if len(reader) != 2 || reader[0].Database != "shop" || reader[1].Database != "blog" {
		t.Fatalf("reader grants = %+v", reader)
	}
	if s := reader[0].Summary(); s != "INSERT, SELECT, UPDATE" {
		t.Errorf("shop summary = %q", s)
	}
	if s := reader[1].Summary(); s != "DELETE, INSERT, SELECT…" {
		t.Errorf("blog summary = %q", s)
	}
}

func TestGranteeUser(t *testing.T) {
	cases := map[string]string{"'app'@'%'": "app", "'a@b'@'localhost'": "a@b", "nouser": "", "'x'@''": "x"}
	for in, want := range cases {
		if got := granteeUser(in); got != want {
			t.Errorf("granteeUser(%q) = %q, want %q", in, got, want)
		}
	}
}
