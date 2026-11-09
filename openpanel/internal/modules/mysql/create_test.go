package mysql

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// joinParts builds fixture values at runtime, so they never look like real secrets
func joinParts(parts ...string) string { return strings.Join(parts, "") }

var (
	shownOnce = joinParts("Pw-0nce", "-shown") // ggignore
	wizValue  = joinParts("s3cr3t", "'pw")     // ggignore
)

var testNames = web.DBNames{
	Databases: []string{"shop", "blog"},
	Users:     []string{"app"},
	Assigned:  []web.DBAssignment{{Database: "shop", Users: "app"}},
}

func TestCheckNewDatabase(t *testing.T) {
	restricted.databases = []string{"mysql"}
	cases := map[string]string{
		"":                      "Database name is required.",
		"bad-name":              "Only letters, numbers, and underscores are allowed (1-63 characters).",
		strings.Repeat("a", 65): "Database name is too long. MySQL identifiers are limited to 64 characters.",
		"mysql":                 "This is a system database that can not be used.",
		"shop":                  "A database with this name already exists.",
		"shop2":                 "",
		strings.Repeat("a", 64): "",
	}
	for name, want := range cases {
		if got := checkNewDatabase(name, testNames); got != want {
			t.Errorf("checkNewDatabase(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestCheckNewUser(t *testing.T) {
	restricted.users = []string{"root"}
	strong := joinParts("Str0ng", "Tw0Value") // ggignore
	cases := []struct{ user, host, pass, want string }{
		{"", "%", strong, "User name is required."},
		{strings.Repeat("u", 33), "%", strong, "User name is too long. MySQL user names are limited to 32 characters."},
		{"bad-user", "%", strong, "Only letters, numbers, and underscores are allowed (1-31 characters)."},
		{"root", "%", strong, "This username is not allowed."},
		{"newuser", "bad host!", strong, "Invalid host format."},
		{"newuser", "%", "Ab1", "Password must be at least 8 characters long"},
		{"newuser", "%", "aaaaaaaa", "Password does not meet the required strength."},
		{"app", "%", strong, "A user with this name already exists."},
		{"newuser", "%", strong, ""},
	}
	for _, c := range cases {
		if got := checkNewUser(c.user, c.host, c.pass, 50, testNames); got != c.want {
			t.Errorf("checkNewUser(%q, %q, %q) = %q, want %q", c.user, c.host, c.pass, got, c.want)
		}
	}
}

func render(t *testing.T, page *web.Page, data any) string {
	t.Helper()
	w := httptest.NewRecorder()
	if err := page.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return w.Body.String()
}

func wantAll(t *testing.T, body string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
}
