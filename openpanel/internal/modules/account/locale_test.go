package account

import "testing"

func TestContains(t *testing.T) {
	if !contains([]string{"en", "de"}, "de") {
		t.Error("expected true for a present element")
	}
	if contains([]string{"en", "de"}, "fr") {
		t.Error("expected false for a missing element")
	}
	if contains(nil, "en") {
		t.Error("expected false for a nil slice")
	}
}
