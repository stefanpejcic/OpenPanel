package mcptokens

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestParseMode(t *testing.T) {
	cases := []struct {
		mode     string
		readOnly bool
		want     string
	}{
		{"", false, ModeFull},
		{"", true, ModeReadOnly},
		{"modify", false, ModeModify},
		{" FULL ", false, ModeFull},
		{"read-only", false, ModeReadOnly},
		// rows written by a newer panel with a mode we don't know fail closed
		{"admin", false, ModeReadOnly},
	}
	for _, c := range cases {
		if got := ParseMode(c.mode, c.readOnly); got != c.want {
			t.Errorf("ParseMode(%q, %v) = %q, want %q", c.mode, c.readOnly, got, c.want)
		}
	}
}

func TestCleanFeatures(t *testing.T) {
	got := CleanFeatures([]string{" dns", "domains", "dns", "", "a,b", "with space"})
	if want := []string{"dns", "domains"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestScopeAllows(t *testing.T) {
	frozen := Scope{Mode: ModeFull, Features: []string{"dns"}}
	if !frozen.Allows("dns") || frozen.Allows("docker") {
		t.Error("a frozen list should allow only what's in it")
	}
	// an empty list is nothing, not everything
	if (Scope{Mode: ModeFull}).Allows("dns") || (Scope{Mode: ModeFull, Features: []string{}}).Allows("dns") {
		t.Error("an empty list allowed a feature")
	}
	if !(Scope{AllFeatures: true}).Allows("docker") {
		t.Error("an old no-list token should still allow everything until it's frozen")
	}
}

func TestFeaturesRoundTrip(t *testing.T) {
	stored := joinFeatures([]string{"domains", "dns"})
	if stored.String != "dns,domains" || !stored.Valid {
		t.Fatalf("stored %+v", stored)
	}
	if got, all := splitFeatures(stored); all || !reflect.DeepEqual(got, []string{"dns", "domains"}) {
		t.Errorf("read back %v all=%v", got, all)
	}
	// an empty list is stored as an empty string, NULL is only for old tokens
	empty := joinFeatures(nil)
	if !empty.Valid || empty.String != "" {
		t.Fatalf("empty list stored %+v", empty)
	}
	if got, all := splitFeatures(empty); all || len(got) != 0 {
		t.Errorf("empty list read back as %v all=%v", got, all)
	}
	if _, all := splitFeatures(sql.NullString{}); !all {
		t.Error("NULL should read back as an old no-list token")
	}
}

func TestDefaultFeatures(t *testing.T) {
	user := []string{"api", "mcp", "dns", "account", "twofa", "passkeys", "phpmyadmin", "mysql", "dns"}
	if got := DefaultFeatures(user); !reflect.DeepEqual(got, []string{"dns", "mysql"}) {
		t.Errorf("default = %v", got)
	}
	if got := FreezableFeatures(user); !reflect.DeepEqual(got, []string{"account", "dns", "mysql", "passkeys", "phpmyadmin", "twofa"}) {
		t.Errorf("freezable = %v", got)
	}
	if got := Missing([]string{"dns", "docker", " mysql"}, user); !reflect.DeepEqual(got, []string{"docker"}) {
		t.Errorf("missing = %v", got)
	}
}
