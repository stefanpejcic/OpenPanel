package backups

import "testing"

func TestTargetAllowed(t *testing.T) {
	mysqlOnly := map[string]bool{"mysql": true, "crons": true}
	cases := []struct {
		allowed        map[string]bool
		target, engine string
		want           bool
	}{
		{mysqlOnly, "all", "", true},
		{mysqlOnly, "files", "", true},
		{mysqlOnly, "database", "mysql", true},
		{mysqlOnly, "database", "postgres", false},
		{mysqlOnly, "database", "", true},
		{mysqlOnly, "crons", "", true},
		{mysqlOnly, "mail", "", false},
		{mysqlOnly, "mailboxes", "", false},
		{mysqlOnly, "html", "", false},
		{mysqlOnly, "paths", "", false},
		{map[string]bool{"domains": true}, "vhosts", "", true},
		{map[string]bool{"domains": true}, "dns_zones", "", false},
		{map[string]bool{"domains": true, "dns": true}, "dns_zones", "", true},
		{map[string]bool{}, "databases", "", false},
	}
	for _, c := range cases {
		if got := targetAllowed(c.allowed, c.target, c.engine); got != c.want {
			t.Errorf("targetAllowed(%v, %q, %q) = %v, want %v", c.allowed, c.target, c.engine, got, c.want)
		}
	}
}
