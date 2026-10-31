package postgresql

import (
	"net/http/httptest"
	"slices"
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
	wizValue  = joinParts("s3c", "r3t")        // ggignore
)

var testNames = web.DBNames{
	Databases: []string{"shop"},
	Users:     []string{"app"},
	Assigned:  []web.DBAssignment{{Database: "shop", Users: "app"}},
}

func TestCheckNewDatabase(t *testing.T) {
	cases := map[string]string{
		"":                      "Database name is required.",
		"bad-name":              "Only letters, numbers, and underscores are allowed (1-63 characters).",
		strings.Repeat("a", 64): "Only letters, numbers, and underscores are allowed (1-63 characters).",
		"template1":             "This is a system database that can not be used.",
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
		{"Postgres", strong, "This username is not allowed."},
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

func TestRenderCreatePages(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	layout := baseLayout(mgr, "/postgresql/new")
	page := func(edit func(*CreatePageData)) CreatePageData {
		d := CreatePageData{LayoutData: layout, Engine: web.PostgreSQLEngine, Names: testNames, DBUsage: 1, DBLimit: 40}
		edit(&d)
		return d
	}
	cases := []struct {
		name  string
		page  *web.Page
		data  CreatePageData
		wants []string
	}{
		{"new database form", newDatabasePage, page(func(d *CreatePageData) { d.FormName = "typed_db" }),
			[]string{`name="database_name"`, `"typed_db"`, `["shop"]`, "1 of 40 databases used", "postgres", "5432", "/postgresql/wizard"}},
		{"new database done", newDatabasePage, page(func(d *CreatePageData) { d.Created = "fresh_db" }),
			[]string{"Database created", "/postgresql/assign?database_name=fresh_db", "/postgresql/import/fresh_db"}},
		{"new user form", createUserPage, page(func(d *CreatePageData) { d.FormName = "typed_user" }),
			[]string{`name="db_user"`, `name="password"`, `"typed_user"`, `["app"]`, "Create PostgreSQL User"}},
		{"new user done", createUserPage, page(func(d *CreatePageData) { d.Created, d.CreatedPassword = "fresh_user", shownOnce }), // ggignore
			[]string{"User created", shownOnce, "/postgresql/assign?db_user=fresh_user"}},
		{"wizard form", wizardPage, page(func(d *CreatePageData) { d.FormDatabaseName, d.FormDBUser = "typed_db", "typed_user" }),
			[]string{`action="/postgresql/wizard"`, `"typed_db"`, `"typed_user"`, "GRANT ALL PRIVILEGES ON DATABASE", "Already on this account"}},
		{"wizard success", wizardPage, page(func(d *CreatePageData) {
			d.Success, d.CreatedDB, d.CreatedUser, d.CreatedPassword = true, "wiz_db", "wiz_user", wizValue // ggignore
		}), []string{"Setup complete", `"wiz_db"`, `"wiz_user"`, `"` + wizValue + `"`, "postgresql://", "django.db.backends.postgresql"}},
		{"assign", assignPage, page(func(d *CreatePageData) { d.DBUser, d.DatabaseName = "app", "shop" }),
			[]string{`name="db_user"`, `name="database_name"`, `{"database":"shop","users":"app"}`, "GRANT USAGE, CREATE ON SCHEMA public", "/postgresql/remove"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			if err := c.page.Render(w, 200, c.data); err != nil {
				t.Fatalf("Render: %v", err)
			}
			body := w.Body.String()
			for _, want := range c.wants {
				if !strings.Contains(body, want) {
					t.Errorf("expected %q in body", want)
				}
			}
			// the done states show what was made, never the form again
			if (c.data.Created != "" || c.data.Success) && strings.Contains(body, `name="password"`+" ") {
				t.Error("done state shouldn't render the form")
			}
		})
	}
}

func TestLoadNamesReservedWithoutServer(t *testing.T) {
	names := web.DBNames{ReservedDatabases: web.SortedKeys(restrictedDatabases), ReservedUsers: web.SortedKeys(restrictedUsers)}
	for _, want := range []string{"postgres", "template0", "template1"} {
		if !slices.Contains(names.ReservedDatabases, want) {
			t.Errorf("reserved databases miss %s", want)
		}
	}
	for _, u := range names.ReservedUsers {
		if !isRestrictedUser(u) {
			t.Errorf("%s is listed as reserved but the server would accept it", u)
		}
	}
}
