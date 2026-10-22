package mysql

import "testing"

func TestParseMySQLConfigContent(t *testing.T) {
	content := `# a comment
[mysqld]
max_connections=150
wait_timeout = 300
skip-log-bin
skip_name_resolve

only_flag_no_value
`
	entries := parseMySQLConfigContent(content)

	want := []ConfigEntry{
		{Key: "max_connections", Value: "150", HasValue: true},
		{Key: "wait_timeout", Value: "300", HasValue: true},
		{Key: "skip_name_resolve"},
		{Key: "only_flag_no_value"},
	}
	if len(entries) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(entries), len(want), entries)
	}
	for i, e := range want {
		if entries[i] != e {
			t.Errorf("entries[%d] = %+v, want %+v", i, entries[i], e)
		}
	}
}

func TestConfigEntriesToMap(t *testing.T) {
	entries := []ConfigEntry{
		{Key: "max_connections", Value: "150", HasValue: true},
		{Key: "flag_only"},
	}
	m := configEntriesToMap(entries)
	if m["max_connections"] != "150" {
		t.Errorf("max_connections = %q", m["max_connections"])
	}
	if _, ok := m["flag_only"]; ok {
		if m["flag_only"] != "" {
			t.Errorf("flag_only = %q, want empty", m["flag_only"])
		}
	}
}

func TestGlobalValue(t *testing.T) {
	cases := map[string]string{
		"150":                     "150",
		"64M":                     "67108864",
		"1g":                      "1073741824",
		"16K":                     "16384",
		"0.5":                     "0.5",
		"ON":                      "ON",
		"off":                     "off",
		"FILE,TABLE":              "'FILE,TABLE'",
		"/var/log/slow.log":       "'/var/log/slow.log'",
		"STRICT_TRANS_TABLES":     "'STRICT_TRANS_TABLES'",
		"x'; DROP TABLE users;--": `'x\'; DROP TABLE users;--'`,
	}
	for in, want := range cases {
		if got := globalValue(in); got != want {
			t.Errorf("globalValue(%q) = %q, want %q", in, got, want)
		}
	}
}
