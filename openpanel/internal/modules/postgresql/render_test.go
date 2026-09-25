package postgresql

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "postgresql": true, "postgresql_import": true}
	return web.LayoutData{
		Title: "Test", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderAllPages(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	layout := baseLayout(mgr, "/postgresql")

	t.Run("databases empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := DatabasesPageData{LayoutData: layout, ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"}, Unit: "mb"}
		if err := databasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("databases: %v", err)
		}
	})
	t.Run("databases with rows", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := DatabasesPageData{LayoutData: layout, ServiceStatusData: ServiceStatusData{ContainerState: "running", HealthStatus: "healthy"}, Unit: "mb",
			Databases: []DatabaseRow{{Database: "mydb", AssignedUsers: "bob", IsSystem: false}, {Database: "postgres", IsSystem: true}}}
		if err := databasesPage.Render(w, 200, data); err != nil {
			t.Fatalf("databases: %v", err)
		}
	})
	t.Run("databases service not running", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := DatabasesPageData{
			LayoutData:        layout,
			ServiceStatusData: ServiceStatusData{ContainerState: "exited", HealthStatus: ""},
			Unit:              "mb",
			StatusDetail:      postgresContainerStatusDetail("exited", ""),
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
			Users: []UserRow{{Name: "bob"}, {Name: "postgres", IsSystem: true}}}
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
		if err := wizardPage.Render(w, 200, struct{ web.LayoutData }{layout}); err != nil {
			t.Fatalf("wizard: %v", err)
		}
	})
	t.Run("assign", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := assignPage.Render(w, 200, struct{ web.LayoutData }{layout}); err != nil {
			t.Fatalf("assign: %v", err)
		}
	})
	t.Run("remove", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := removePage.Render(w, 200, struct{ web.LayoutData }{layout}); err != nil {
			t.Fatalf("remove: %v", err)
		}
	})
	t.Run("import", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := importPage.Render(w, 200, ImportPageData{LayoutData: layout, DBName: "mydb"}); err != nil {
			t.Fatalf("import: %v", err)
		}
	})
	t.Run("processlist empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		if err := processlistPage.Render(w, 200, ProcessListPageData{LayoutData: layout}); err != nil {
			t.Fatalf("processlist: %v", err)
		}
		if !strings.Contains(w.Body.String(), "No active processes.") {
			t.Error("expected empty state")
		}
	})
	t.Run("processlist with rows", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ProcessListPageData{LayoutData: layout, ProcessList: []ProcessRow{
			{PID: "123", DB: "postgres", User: "postgres", State: "active", Query: "SELECT a | b", BackendType: "client backend"},
			{PID: "124", DB: "postgres", User: "postgres", State: "idle", Query: "SELECT 1", BackendType: "client backend"},
			{PID: "50", State: "", BackendType: "autovacuum launcher"},
		}}
		if err := processlistPage.Render(w, 200, data); err != nil {
			t.Fatalf("processlist: %v", err)
		}
		body := w.Body.String()
		if !strings.Contains(body, "SELECT a | b") {
			t.Error("expected query with a pipe to render intact")
		}
		if n := strings.Count(body, `action="/postgresql/processlist/kill"`); n != 1 {
			t.Errorf("expected kill button only on the active client query, got %d", n)
		}
		if !strings.Contains(body, `name="pid" value="123"`) {
			t.Error("expected kill form for pid 123")
		}
	})
	t.Run("configuration", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := ConfigurationPageData{LayoutData: layout, CurrentConfig: map[string]string{"max_connections": "100"}, DefaultKeys: []string{"max_connections", "shared_buffers"}}
		if err := configurationPage.Render(w, 200, data); err != nil {
			t.Fatalf("configuration: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{`id="db-tuning"`, "dbTuning('/postgresql/configuration/recommendations')", "Confirm Changes", `x-ref="form"`} {
			if !strings.Contains(body, want) {
				t.Errorf("expected %q in body", want)
			}
		}
		if n := strings.Count(body, "review(&#34;"); n != 2 {
			t.Errorf("expected an optimize button per key, got %d", n)
		}
	})
	t.Run("remote", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := RemotePostgresPageData{LayoutData: layout, ServerIP: "1.2.3.4", ContainerPort: "5433", RemotePostgreSQLDisplay: "ON", PostgresPort: 5432}
		if err := remotePostgresPage.Render(w, 200, data); err != nil {
			t.Fatalf("remote: %v", err)
		}
	})
}
