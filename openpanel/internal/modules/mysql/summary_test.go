package mysql

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestShortServerVersion(t *testing.T) {
	cases := map[string]string{
		"13.0.2-MariaDB-ubu2604":      "13.0.2",
		"8.4.3":                       "8.4.3",
		" 11.4.5-MariaDB-log ":        "11.4.5",
		"":                            "",
		"MariaDB":                     "",
		"5.7.44-percona-sure1-log-xx": "5.7.44",
	}
	for in, want := range cases {
		if got := shortServerVersion(in); got != want {
			t.Errorf("shortServerVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGroupUserGrants(t *testing.T) {
	all := "ALTER,ALTER ROUTINE,CREATE,CREATE ROUTINE,CREATE TEMPORARY TABLES,CREATE VIEW,DELETE,DELETE HISTORY,DROP,EVENT,EXECUTE,INDEX,INSERT,LOCK TABLES,REFERENCES,SELECT,SHOW CREATE ROUTINE,SHOW VIEW,TRIGGER,UPDATE"
	// rows as information_schema.SCHEMA_PRIVILEGES returns them on MariaDB 13
	rows := [][3]string{
		{"'app'@'%'", "testinguser_bkfix", all},
		{"'reader'@'%'", "shop", "INSERT,SELECT"},
		{"'reader'@'localhost'", "shop", "SELECT,UPDATE"},
		{"'reader'@'%'", "blog", "DELETE,INSERT,SELECT,UPDATE"},
		{"broken", "x", "SELECT"},
	}
	got := groupUserGrants(rows)
	if len(got) != 2 {
		t.Fatalf("got %d users, want 2: %v", len(got), got)
	}
	if g := got["app"]; len(g) != 1 || g[0].Database != "testinguser_bkfix" || g[0].Summary() != "ALL" {
		t.Errorf("app grants = %+v", g)
	}
	reader := got["reader"]
	if len(reader) != 2 || reader[0].Database != "shop" || reader[1].Database != "blog" {
		t.Fatalf("reader grants = %+v", reader)
	}
	if s := reader[0].Summary(); s != "INSERT, SELECT, UPDATE" {
		t.Errorf("shop summary = %q", s)
	}
	if s := reader[1].Summary(); s != "DELETE, INSERT, SELECT…" {
		t.Errorf("blog summary = %q", s)
	}
}

func TestGranteeUser(t *testing.T) {
	cases := map[string]string{"'app'@'%'": "app", "'a@b'@'localhost'": "a@b", "nouser": "", "'x'@''": "x"}
	for in, want := range cases {
		if got := granteeUser(in); got != want {
			t.Errorf("granteeUser(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderDatabasesSummaryTiles(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := DatabasesPageData{
		LayoutData:        baseLayout(mgr, "/mysql"),
		ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"},
		Service:           "mariadb", ServerVersion: "13.0.2", Uptime: "2d 14h", DBLimit: 10,
		Databases: []DatabaseRow{{Database: "app_db"}, {Database: "shop"}, {Database: "mysql", IsSystem: true}},
		Unit:      "mb",
	}
	w := httptest.NewRecorder()
	if err := databasesPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"MariaDB 13.0.2 · up 2d 14h", "2 <small>/ 10</small>", "width:20%", "db_tables_cell", ">Tables<"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
}

func TestRenderUsersPageGrants(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := UsersPageData{
		LayoutData:        baseLayout(mgr, "/mysql/users"),
		ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"},
		Service:           "mariadb", Users: []string{"app", "lonely"},
		Grants: map[string][]UserGrant{"app": {{Database: "shop", Privileges: []string{"SELECT", "INSERT"}}}},
	}
	w := httptest.NewRecorder()
	if err := usersPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "shop · SELECT, INSERT") {
		t.Error("expected grant line for app")
	}
	if !strings.Contains(body, "/mysql/assign?db_user=app") {
		t.Error("expected manage privileges link with the user")
	}
}

func TestRenderWizardQuotaAndPort(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	layout := baseLayout(mgr, "/mysql/wizard")
	layout.HostingPlanName = "Developer plus"
	data := WizardPageData{LayoutData: layout, Service: "mariadb", Engine: web.MySQLEngine, DBUsage: 1, DBLimit: 40, RemotePort: "32770"}
	w := httptest.NewRecorder()
	if err := wizardPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"1 of 40 databases used", "on the Developer plus plan", "32770"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
}
