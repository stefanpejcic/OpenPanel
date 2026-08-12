package validators

import (
	"strings"
	"testing"
)

func TestIsValidIdentifier(t *testing.T) {
	if !IsValidIdentifier("user_123") {
		t.Error("expected user_123 to be valid")
	}
	if IsValidIdentifier("bad name!") {
		t.Error("expected 'bad name!' to be invalid")
	}
}

func TestIsValidHost(t *testing.T) {
	for _, h := range []string{"%", "localhost", "127.0.0.1", "db.example.com"} {
		if !IsValidHost(h) {
			t.Errorf("expected %q to be a valid host", h)
		}
	}
	if IsValidHost("bad host!") {
		t.Error("expected 'bad host!' to be invalid")
	}
}

func TestIsValidEmailUsername(t *testing.T) {
	valid := []string{"john", "john.doe", "john_doe", "john+tag", "john%doe", "john-doe", "j123"}
	invalid := []string{"", "john@doe", "@proba.digital", "john@", "john doe", "john\ndoe"}
	for _, s := range valid {
		if !IsValidEmailUsername(s) {
			t.Errorf("expected %q to be a valid email username", s)
		}
	}
	for _, s := range invalid {
		if IsValidEmailUsername(s) {
			t.Errorf("expected %q to be an invalid email username", s)
		}
	}
}

func TestIsValidPanelUsername(t *testing.T) {
	valid := []string{"joh", "johndoe", "john123", "JohnDoe2026"}
	invalid := []string{"", "jo", strings.Repeat("a", 21), "john doe", "john-doe", "john_doe", "john@doe", "john.doe"}
	for _, s := range valid {
		if !IsValidPanelUsername(s) {
			t.Errorf("expected %q to be a valid panel username", s)
		}
	}
	for _, s := range invalid {
		if IsValidPanelUsername(s) {
			t.Errorf("expected %q to be an invalid panel username", s)
		}
	}
}

func TestClampPasswordStrength(t *testing.T) {
	cases := []struct {
		raw  string
		def  int
		want int
	}{
		{"50", 50, 50},
		{"", 50, 50},
		{"not-a-number", 50, 50},
		{"0", 50, 1},
		{"1000", 50, 100},
	}
	for _, c := range cases {
		if got := ClampPasswordStrength(c.raw, c.def); got != c.want {
			t.Errorf("ClampPasswordStrength(%q, %d) = %d, want %d", c.raw, c.def, got, c.want)
		}
	}
}

func TestPasswordStrengthScore(t *testing.T) {
	if got := PasswordStrengthScore(""); got != 0 {
		t.Errorf("empty password score = %d, want 0", got)
	}
	// "Aa1!Aa1!Aa1!" (>=12 chars, lower, upper, digit, special) = full 6/6.
	if got := PasswordStrengthScore("Aa1!Aa1!Aa1!"); got != 100 {
		t.Errorf("strong password score = %d, want 100", got)
	}
	// "aaaaaaaa": len>=8 + has-lower = 2/6 checks -> round(2/6*100) = 33.
	if got := PasswordStrengthScore("aaaaaaaa"); got != 33 {
		t.Errorf("lowercase-only 8-char password score = %d, want 33", got)
	}
}

func TestIsPasswordStrongEnough(t *testing.T) {
	if !IsPasswordStrongEnough("Aa1!Aa1!Aa1!", 100) {
		t.Error("expected strong password to meet threshold 100")
	}
	if IsPasswordStrongEnough("aaaaaaaa", 50) {
		t.Error("expected weak password to fail threshold 50")
	}
}
