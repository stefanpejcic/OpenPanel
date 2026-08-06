package mysql

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "mysql": true, "mysql_import": true, "phpmyadmin": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderDatabasesPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("running healthy with rows", func(t *testing.T) {
		data := DatabasesPageData{
			LayoutData:        baseLayout(mgr, "/mysql"),
			ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"},
			Service:           "mysql",
			Databases: []DatabaseRow{
				{Database: "app_db", AssignedUsers: "app_user", IsSystem: false},
				{Database: "mysql", AssignedUsers: "", IsSystem: true},
			},
			Unit: "mb",
		}
		w := httptest.NewRecorder()
		if err := databasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "app_db") || !strings.Contains(body, "app_user") {
			t.Error("expected database row in body")
		}
		if !strings.Contains(body, "System Database") {
			t.Error("expected system database badge")
		}
	})

	t.Run("not running shows status detail", func(t *testing.T) {
		data := DatabasesPageData{
			LayoutData:        baseLayout(mgr, "/mysql"),
			ServiceStatusData: ServiceStatusData{ContainerState: "not_found", HealthStatus: "", StatusDetail: mysqlContainerStatusDetail("not_found", "")},
			Service:           "mysql",
		}
		w := httptest.NewRecorder()
		if err := databasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "installation is underway") {
			t.Error("expected not_found status detail text")
		}
	})

	t.Run("zero-user-database toast", func(t *testing.T) {
		data := DatabasesPageData{
			LayoutData:        baseLayout(mgr, "/mysql"),
			ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"},
			Service:           "mysql",
			Databases:         []DatabaseRow{{Database: "orphan_db"}},
			DBToastID:         "no-users:orphan_db",
			DBToastMessage:    "Database orphan_db has no users assigned.",
		}
		w := httptest.NewRecorder()
		if err := databasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "reportHealthIssues") {
			t.Error("expected health toast script when DBToastMessage is set")
		}
	})
}

func TestRenderUsersPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := UsersPageData{
		LayoutData:        baseLayout(mgr, "/mysql/users"),
		ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"},
		Service:           "mysql",
		Users:             []string{"appuser", "root"},
	}
	w := httptest.NewRecorder()
	if err := usersPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "appuser") {
		t.Error("expected user row in body")
	}
	if !strings.Contains(body, "System User") {
		t.Error("expected system user badge for root")
	}
}

func TestRenderNewDatabasePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := NewDatabasePageData{LayoutData: baseLayout(mgr, "/mysql/new"), Service: "mysql"}
	w := httptest.NewRecorder()
	if err := newDatabasePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
}

func TestRenderCreateUserPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := CreateUserPageData{LayoutData: baseLayout(mgr, "/mysql/user"), Service: "mysql"}
	w := httptest.NewRecorder()
	if err := createUserPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
}

func TestRenderChangePasswordPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := PasswordPageData{LayoutData: baseLayout(mgr, "/mysql/password/appuser"), DBUser: "appuser"}
	w := httptest.NewRecorder()
	if err := passwordPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "appuser") {
		t.Error("expected DBUser in body")
	}
}

func TestRenderWizardPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("form", func(t *testing.T) {
		data := WizardPageData{LayoutData: baseLayout(mgr, "/mysql/wizard"), Service: "mysql"}
		w := httptest.NewRecorder()
		if err := wizardPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "databaseWizardForm") {
			t.Error("expected wizard form in body")
		}
	})

	t.Run("success", func(t *testing.T) {
		data := WizardPageData{
			LayoutData: baseLayout(mgr, "/mysql/wizard"), Service: "mysql", Success: true,
			CreatedDB: "wiz_db", CreatedUser: "wiz_user", CreatedPassword: "s3cr3t", CreatedHost: "mysql",
		}
		w := httptest.NewRecorder()
		if err := wizardPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "wiz_db") || !strings.Contains(body, "wiz_user") {
			t.Error("expected created credentials in body")
		}
		if !strings.Contains(body, "Setup complete") {
			t.Error("expected success header")
		}
	})
}

func TestRenderAssignPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := AssignPageData{
		LayoutData: baseLayout(mgr, "/mysql/assign"), DBUser: "appuser", DatabaseName: "app_db",
		Privileges: mysqlPrivileges,
	}
	w := httptest.NewRecorder()
	if err := assignPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "ALL PRIVILEGES") {
		t.Error("expected privileges list in body")
	}
}

func TestRenderRemovePage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := struct{ web.LayoutData }{baseLayout(mgr, "/mysql/remove")}
	w := httptest.NewRecorder()
	if err := removePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
}

func TestRenderImportPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := ImportPageData{LayoutData: baseLayout(mgr, "/mysql/import"), Service: "mysql", MySQLVersion: "mysql", DBName: "app_db"}
	w := httptest.NewRecorder()
	if err := importPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "app_db") {
		t.Error("expected preselected db name in body")
	}
}

func TestRenderProcessListPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := ProcessListPageData{
		LayoutData: baseLayout(mgr, "/mysql/processlist"),
		ProcessList: []ProcessRow{
			{ID: "1", User: "app", Host: "localhost", DB: "app_db", Command: "Query", Time: "0", State: "", Info: "SELECT 1"},
		},
	}
	w := httptest.NewRecorder()
	if err := processlistPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "SELECT 1") {
		t.Error("expected process row in body")
	}
}

func TestRenderRootPasswordPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := RootPasswordPageData{LayoutData: baseLayout(mgr, "/mysql/root-password"), Service: "mysql"}
	w := httptest.NewRecorder()
	if err := rootPasswordPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
}

func TestRenderRemoteMySQLPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	t.Run("on with access rows", func(t *testing.T) {
		data := RemoteMySQLPageData{
			LayoutData: baseLayout(mgr, "/mysql/remote-mysql"), Service: "mysql",
			ServerIP: "1.2.3.4", ContainerPort: "3306", RemoteMySQLDisplay: "ON", MySQLPort: 3306,
			UserAccess: []RemoteUserAccess{{Username: "appuser", Hosts: []string{"%", "10.0.0.5"}}},
		}
		w := httptest.NewRecorder()
		if err := remoteMySQLPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "appuser") || !strings.Contains(body, "10.0.0.5") {
			t.Error("expected user access row in body")
		}
		if !strings.Contains(body, "Click to Disable") {
			t.Error("expected disable button when ON")
		}
	})

	t.Run("off hides access table", func(t *testing.T) {
		data := RemoteMySQLPageData{
			LayoutData: baseLayout(mgr, "/mysql/remote-mysql"), Service: "mysql",
			ServerIP: "1.2.3.4", ContainerPort: "3306", RemoteMySQLDisplay: "OFF", MySQLPort: 3306,
		}
		w := httptest.NewRecorder()
		if err := remoteMySQLPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		if !strings.Contains(w.Body.String(), "Click to Enable") {
			t.Error("expected enable button when OFF")
		}
	})
}

func TestRenderConfigurationPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := ConfigurationPageData{
		LayoutData: baseLayout(mgr, "/mysql/configuration"), Service: "mysql",
		CurrentConfig: map[string]string{"max_connections": "150"},
		DefaultKeys:   []string{"max_connections", "wait_timeout"},
	}
	w := httptest.NewRecorder()
	if err := configurationPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "max_connections") || !strings.Contains(body, "150") {
		t.Error("expected config key/value in body")
	}
}
