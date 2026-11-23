package mcptokens

import (
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

func TestScopeLimited(t *testing.T) {
	if (Scope{Mode: ModeFull}).Limited() {
		t.Error("plain full token counted as limited")
	}
	for _, s := range []Scope{{Mode: ModeReadOnly}, {Mode: ModeModify}, {Mode: ModeFull, Features: []string{"dns"}}} {
		if !s.Limited() {
			t.Errorf("%+v should be limited", s)
		}
	}
}

func TestFeaturesRoundTrip(t *testing.T) {
	stored := joinFeatures([]string{"domains", "dns"})
	if stored.String != "dns,domains" {
		t.Fatalf("stored %q", stored.String)
	}
	if got := splitFeatures(stored); !reflect.DeepEqual(got, []string{"dns", "domains"}) {
		t.Errorf("read back %v", got)
	}
	if joinFeatures(nil).Valid {
		t.Error("no features should store NULL")
	}
}
