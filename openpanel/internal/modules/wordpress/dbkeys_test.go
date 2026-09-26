package wordpress

import (
	"sort"
	"strings"
	"testing"
)

func TestHighPerformanceKeySQLPrefix(t *testing.T) {
	stmts := highPerformanceKeySQL("wp_")
	if len(stmts) != len(highPerformanceKeys) {
		t.Fatalf("got %d statements, want %d", len(stmts), len(highPerformanceKeys))
	}
	if !strings.HasPrefix(stmts[0], "ALTER TABLE `wp_options` ADD UNIQUE KEY option_id") {
		t.Errorf("unexpected first statement: %s", stmts[0])
	}
	if got := highPerformanceKeySQL("x9_")[2]; !strings.HasPrefix(got, "ALTER TABLE `x9_postmeta` ") {
		t.Errorf("custom prefix not used: %s", got)
	}
	for _, bad := range []string{"", "wp`_", "wp_; DROP TABLE x;", "wp-", "wp _"} {
		if highPerformanceKeySQL(bad) != nil {
			t.Errorf("prefix %q should be refused", bad)
		}
	}
}

// swapping a primary key away from the auto-increment column needs a unique key on it added first, in the same ALTER
func TestHighPerformanceKeysUniqueBeforePrimarySwap(t *testing.T) {
	for _, k := range highPerformanceKeys {
		drop := strings.Index(k.alter, "DROP PRIMARY KEY")
		if drop == -1 {
			continue
		}
		unique := strings.Index(k.alter, "ADD UNIQUE KEY")
		if unique == -1 || unique > drop {
			t.Errorf("%s drops its primary key without adding a unique key first", k.table)
		}
		if !strings.Contains(k.alter[drop:], "ADD PRIMARY KEY") {
			t.Errorf("%s drops its primary key without adding a new one", k.table)
		}
	}
}

func TestHighPerformanceKeysTables(t *testing.T) {
	var got []string
	for _, k := range highPerformanceKeys {
		got = append(got, k.table)
	}
	sort.Strings(got)
	want := "commentmeta comments options postmeta posts termmeta usermeta users"
	if strings.Join(got, " ") != want {
		t.Errorf("tables = %v, want %s", got, want)
	}
}
