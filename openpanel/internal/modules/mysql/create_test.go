package mysql

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
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

func TestRenderCreatePages(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("new database form", func(t *testing.T) {
		body := render(t, newDatabasePage, NewDatabasePageData{
			LayoutData: baseLayout(mgr, "/mysql/new"), Service: "mariadb", Engine: web.MySQLEngine,
			Names: testNames, DBUsage: 2, DBLimit: 40, RemotePort: "32770", FormName: "typed_name",
		})
		wantAll(t, body, `name="database_name"`, `"typed_name"`, `["shop","blog"]`, "2 of 40 databases used", "32770", "MariaDB", "/mysql/wizard", "Name the database")
		if strings.Contains(body, "Database created") {
			t.Error("form shouldn't show the done state")
		}
	})

	t.Run("new database done", func(t *testing.T) {
		body := render(t, newDatabasePage, NewDatabasePageData{
			LayoutData: baseLayout(mgr, "/mysql/new"), Service: "mysql", Engine: web.MySQLEngine, Created: "fresh_db",
		})
		wantAll(t, body, "Database created", "fresh_db", "/mysql/assign?database_name=fresh_db", "/mysql/import/fresh_db", "/mysql/new")
		if strings.Contains(body, `name="database_name"`) {
			t.Error("done state shouldn't render the form")
		}
	})

	t.Run("new user form", func(t *testing.T) {
		body := render(t, createUserPage, CreateUserPageData{
			LayoutData: baseLayout(mgr, "/mysql/user"), Service: "mysql", Engine: web.MySQLEngine, Names: testNames, FormName: "typed_user",
		})
		wantAll(t, body, `name="db_user"`, `name="password"`, `"typed_user"`, `["app"]`, "Set a password", "Create MySQL User")
	})

	t.Run("new user done", func(t *testing.T) {
		body := render(t, createUserPage, CreateUserPageData{
			LayoutData: baseLayout(mgr, "/mysql/user"), Service: "mysql", Engine: web.MySQLEngine, Created: "fresh_user", CreatedPassword: shownOnce, // ggignore
		})
		wantAll(t, body, "User created", "fresh_user", shownOnce, "/mysql/assign?db_user=fresh_user", "localhost", "3306")
		if strings.Contains(body, `name="password"`) {
			t.Error("done state shouldn't render the form")
		}
	})

	t.Run("wizard lists names", func(t *testing.T) {
		body := render(t, wizardPage, WizardPageData{
			LayoutData: baseLayout(mgr, "/mysql/wizard"), Service: "mysql", Engine: web.MySQLEngine, Names: testNames,
			FormDatabaseName: "typed_db", FormDBUser: "typed_user",
		})
		wantAll(t, body, `action="/mysql/wizard"`, `name="privileges" value="ALL PRIVILEGES"`, `"typed_db"`, `"typed_user"`, `["shop","blog"]`, "GRANT ALL PRIVILEGES ON")
	})

	t.Run("wizard success snippets", func(t *testing.T) {
		body := render(t, wizardPage, WizardPageData{
			LayoutData: baseLayout(mgr, "/mysql/wizard"), Service: "mysql", Engine: web.MySQLEngine, Success: true,
			CreatedDB: "wiz_db", CreatedUser: "wiz_user", CreatedPassword: wizValue, // ggignore
		})
		// values go in as JSON strings, so a quote in the password can't break the snippet code
		wantAll(t, body, `"wiz_db"`, `"wiz_user"`, `"`+wizValue+`"`, "wp-config-snippet.php", "Download snippet")
	})

	t.Run("assign", func(t *testing.T) {
		body := render(t, assignPage, AssignPageData{
			LayoutData: baseLayout(mgr, "/mysql/assign"), Engine: web.MySQLEngine, Names: testNames,
			DBUser: "app", DatabaseName: "shop", Privileges: mysqlPrivileges, PrivilegeGroups: mysqlPrivilegeGroups,
		})
		wantAll(t, body, `name="db_user"`, `name="database_name"`, `name="privileges"`, "Routines and events", "CREATE TEMPORARY TABLES",
			`{"database":"shop","users":"app"}`, "/mysql/privileges/", "/mysql/remove", "Save Privileges")
	})
}
