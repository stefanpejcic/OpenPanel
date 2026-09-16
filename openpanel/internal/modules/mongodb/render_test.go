package mongodb

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "mongodb": true, "mongodb_import": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderAllPages(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	layout := baseLayout(mgr, "/mongodb")

	t.Run("databases empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := DatabasesPageData{LayoutData: layout, ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"}}
		if err := databasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("databases: %v", err)
		}
	})
	t.Run("databases with rows", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := DatabasesPageData{LayoutData: layout, ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"},
			Databases: []DatabaseRow{{Database: "mydb", SizeDisplay: "1.00 MB", IsSystem: false}, {Database: "admin", IsSystem: true}}}
		if err := databasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("databases: %v", err)
		}
	})
	t.Run("databases service not running", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := DatabasesPageData{
			LayoutData:        layout,
			ServiceStatusData: ServiceStatusData{ContainerState: "exited", HealthStatus: ""},
			StatusDetail:      mongoContainerStatusDetail("exited", ""),
		}
		if err := databasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("databases: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "Service has stopped") {
			t.Error("expected StatusDetail message in body when service isn't running")
		}
		if !strings.Contains(body, "refresh-message") || !strings.Contains(body, "window.location.reload()") {
			t.Error("expected the countdown-and-reload script when not healthy")
		}
		if !strings.Contains(body, "cursor-not-allowed") {
			t.Error("New Database button should render as disabled while the service isn't running")
		}
	})
	t.Run("new", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := newDatabasePage.Render(w, 200, struct{ web.LayoutData }{layout}); err != nil {
			t.Fatalf("new: %v", err)
		}
	})
	t.Run("users empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := UsersPageData{LayoutData: layout, ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"}}
		if err := usersPage.Render(w, 200, data); err != nil {
			t.Fatalf("users: %v", err)
		}
	})
	t.Run("users with rows", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := UsersPageData{LayoutData: layout, ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"},
			Users: []UserRow{{Name: "bob", Roles: "readWrite@mydb"}, {Name: "admin", IsSystem: true}}}
		if err := usersPage.Render(w, 200, data); err != nil {
			t.Fatalf("users: %v", err)
		}
	})
	t.Run("create user", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := createUserPage.Render(w, 200, struct{ web.LayoutData }{layout}); err != nil {
			t.Fatalf("create user: %v", err)
		}
	})
	t.Run("password", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := passwordPage.Render(w, 200, PasswordPageData{LayoutData: layout, DBUser: "bob"}); err != nil {
			t.Fatalf("password: %v", err)
		}
	})
	t.Run("wizard", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := wizardPage.Render(w, 200, WizardPageData{LayoutData: layout, Roles: mongoRoles}); err != nil {
			t.Fatalf("wizard: %v", err)
		}
	})
	t.Run("assign", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := assignPage.Render(w, 200, AssignPageData{LayoutData: layout, Roles: mongoRoles}); err != nil {
			t.Fatalf("assign: %v", err)
		}
	})
	t.Run("remove", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := removePage.Render(w, 200, AssignPageData{LayoutData: layout, Roles: mongoRoles}); err != nil {
			t.Fatalf("remove: %v", err)
		}
	})
	t.Run("import", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := importPage.Render(w, 200, ImportPageData{LayoutData: layout, DBName: "mydb"}); err != nil {
			t.Fatalf("import: %v", err)
		}
	})
}
