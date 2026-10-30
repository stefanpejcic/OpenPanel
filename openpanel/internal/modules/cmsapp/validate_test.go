package cmsapp

import (
	"strings"
	"testing"
)

// joinParts builds fixture values at runtime, so they never look like real secrets
func joinParts(parts ...string) string { return strings.Join(parts, "") }

var (
	goodDB     = joinParts("Ab1!x#y%", "Z^*()-_=+") // ggignore
	goodAdmin  = joinParts("Xk9!fQ2#", "mP7@wL4z")  // ggignore
	adminDash  = joinParts("-Xk9!fQ2#", "mP7@wL4z") // ggignore
	adminSpace = joinParts("Xk9!fQ2 #", "mP7@wL4z") // ggignore
)

func TestValidateInstallForm(t *testing.T) {
	ok := map[string]string{"db_name": "joomla_ab12cd", "db_user": "qwertyuiop", "db_password": goodDB, "admin_email": "admin@example.com", "admin_username": "admin", "admin_password": goodAdmin} // ggignore
	get := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	if msg := ValidateInstallForm(get(ok), 50); msg != "" {
		t.Fatalf("valid form rejected: %s", msg)
	}
	if msg := ValidateInstallForm(get(map[string]string{}), 50); msg != "" {
		t.Fatalf("empty form should fall back to defaults, got %s", msg)
	}
	bad := []map[string]string{
		{"db_name": "x`; DROP DATABASE y; --"},
		{"db_user": "a'@'%"},
		{"db_name": "this_name_is_way_longer_than_thirty_two_chars"},
		{"db_password": "it's"},
		{"db_password": "back\\slash"},
		{"db_password": "tick`tick"},
		{"db_password": "has space"},
		{"admin_email": "not-an-email"},
		{"admin_username": "--help"},
		{"admin_login": "a b"},
		{"admin_user": "x;rm"},
		{"admin_password": adminDash},  // ggignore
		{"admin_password": adminSpace}, // ggignore
		{"admin_password": "weak"},
	}
	for _, m := range bad {
		if ValidateInstallForm(get(m), 50) == "" {
			t.Errorf("expected %v to be rejected", m)
		}
	}
	// strength 0 skips only the strength check
	if ValidateInstallForm(get(map[string]string{"admin_password": "weak"}), 0) != "" {
		t.Error("strength 0 should not check strength")
	}
}

func TestIsValidSubdirectory(t *testing.T) {
	for _, s := range []string{"", "blog", "shop/v2", "my-site_1.0", "blog/"} {
		if !IsValidSubdirectory(s) {
			t.Errorf("IsValidSubdirectory(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"../etc", "a/../../b", "/abs", "a b", "a;b", "a/./b", ".", "a//b", "$(id)"} {
		if IsValidSubdirectory(s) {
			t.Errorf("IsValidSubdirectory(%q) = true, want false", s)
		}
	}
}
