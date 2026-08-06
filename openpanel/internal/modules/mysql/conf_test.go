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
