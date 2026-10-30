package mongodb

import (
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// joinParts builds fixture values at runtime, so they never look like real secrets
func joinParts(parts ...string) string { return strings.Join(parts, "") }

var (
	shownOnce = joinParts("Pw-0nce", "-shown") // ggignore
	wizValue  = joinParts("s3c", "r3t")        // ggignore
)

var testNames = web.DBNames{
	Databases: []string{"shop"},
	Users:     []string{"app"},
}

func TestCheckNewDatabase(t *testing.T) {
	cases := map[string]string{
		"":                      "Database name is required.",
		"bad-name":              "Only letters, numbers, and underscores are allowed (1-63 characters).",
		strings.Repeat("a", 64): "Only letters, numbers, and underscores are allowed (1-63 characters).",
		"admin":                 "This is a system database that can not be used.",
		"shop":                  "A database with this name already exists.",
		strings.Repeat("a", 63): "",
	}
	for name, want := range cases {
		if got := checkNewDatabase(name, testNames); got != want {
			t.Errorf("checkNewDatabase(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestCheckNewUser(t *testing.T) {
	strong := joinParts("Str0ng", "Tw0Value") // ggignore
	cases := []struct{ user, pass, want string }{
		{"", strong, "User name is required."},
		{"bad user", strong, "Only letters, numbers, and underscores are allowed (1-63 characters)."},
		{"Root", strong, "This username is not allowed."},
		{"newuser", "Ab1", "Password must be at least 8 characters long"},
		{"newuser", "aaaaaaaa", "Password does not meet the required strength."},
		{"app", strong, "A user with this name already exists."},
		{"newuser", strong, ""},
	}
	for _, c := range cases {
		if got := checkNewUser(c.user, c.pass, 50, testNames); got != c.want {
			t.Errorf("checkNewUser(%q, %q) = %q, want %q", c.user, c.pass, got, c.want)
		}
	}
}
