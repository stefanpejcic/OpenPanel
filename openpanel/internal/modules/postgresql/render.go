package postgresql

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var (
	databasesPage      = web.LoadPage("psql/databases.html")
	newDatabasePage    = web.LoadPage("db/new.html")
	usersPage          = web.LoadPage("psql/users.html")
	createUserPage     = web.LoadPage("psql/psql_user.html")
	passwordPage       = web.LoadPage("psql/password.html")
	wizardPage         = web.LoadPage("psql/wizard.html")
	assignPage         = web.LoadPage("psql/assign.html")
	removePage         = web.LoadPage("psql/remove.html")
	importPage         = web.LoadPage("psql/import.html")
	processlistPage    = web.LoadPage("psql/processlist.html")
	remotePostgresPage = web.LoadPage("psql/remote_psql.html")
	configurationPage  = web.LoadPage("db/configuration.html", "partials/_db_tuning.html")
)

// ServiceStatusData is the container_state/health_status view-model shared by databases.html and users.html.
type ServiceStatusData struct {
	ContainerState string
	HealthStatus   string
}

// DatabasesPageData is psql/databases.html's template context.
type DatabasesPageData struct {
	web.LayoutData
	ServiceStatusData
	Databases    []DatabaseRow
	Unit         string
	ShowAll      bool
	StatusDetail string // longer explanatory text for the table's empty-state row while the service isn't running/healthy, "" when running+healthy
}

func renderDatabasesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, databases []DatabaseRow, unit string, showAll bool) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "PostgreSQL Databases")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if len(databases) > 0 {
		userContext, _ := injectedData["context"].(string)
		_, users, _ := ComputeDatabaseAndUserNames(r.Context(), userContext)
		layout.BulkActions = databasesBulkActions(layout.T, users)
	}
	layout.Service = "postgres"
	data := DatabasesPageData{
		LayoutData:        layout,
		ServiceStatusData: ServiceStatusData{ContainerState: status.State, HealthStatus: status.Health},
		Databases:         databases, Unit: unit, ShowAll: showAll,
		StatusDetail: postgresContainerStatusDetail(status.State, status.Health),
	}
	if err := databasesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("POSTGRESQL - databases template render error: %v", err)
	}
}

func renderNewDatabasePage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Create Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "postgres"
	data := struct {
		web.LayoutData
		Engine web.DBEngine
	}{layout, web.PostgreSQLEngine}
	if err := newDatabasePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("POSTGRESQL - new database template render error: %v", err)
	}
}

// UserRow is one row of psql/users.html's table.
type UserRow struct {
	Name     string
	IsSystem bool
}

// UsersPageData is psql/users.html's template context.
type UsersPageData struct {
	web.LayoutData
	ServiceStatusData
	Users   []UserRow
	ShowAll bool
}

func renderUsersPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, users []UserRow, showAll bool) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "PostgreSQL Users")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if len(users) > 0 {
		userContext, _ := injectedData["context"].(string)
		databases, _, _ := ComputeDatabaseAndUserNames(r.Context(), userContext)
		layout.BulkActions = usersBulkActions(layout.T, databases)
	}
	layout.Service = "postgres"
	data := UsersPageData{
		LayoutData:        layout,
		ServiceStatusData: ServiceStatusData{ContainerState: status.State, HealthStatus: status.Health},
		Users:             users, ShowAll: showAll,
	}
	if err := usersPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("POSTGRESQL - users template render error: %v", err)
	}
}

func renderCreateUserPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Create User")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "postgres"
	if err := createUserPage.Render(w, http.StatusOK, struct{ web.LayoutData }{layout}); err != nil {
		log.Printf("POSTGRESQL - create user template render error: %v", err)
	}
}

// PasswordPageData is psql/password.html's template context.
type PasswordPageData struct {
	web.LayoutData
	DBUser string
}

func renderChangePasswordPage(a *appctx.App, w http.ResponseWriter, r *http.Request, dbUser string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Change Password")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "postgres"
	data := PasswordPageData{LayoutData: layout, DBUser: dbUser}
	if err := passwordPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("POSTGRESQL - password template render error: %v", err)
	}
}

func renderWizardPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Database Wizard")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "postgres"
	if err := wizardPage.Render(w, http.StatusOK, struct{ web.LayoutData }{layout}); err != nil {
		log.Printf("POSTGRESQL - wizard template render error: %v", err)
	}
}

func renderAssignPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Assign User to Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "postgres"
	if err := assignPage.Render(w, http.StatusOK, struct{ web.LayoutData }{layout}); err != nil {
		log.Printf("POSTGRESQL - assign template render error: %v", err)
	}
}

func renderRemovePage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Remove User from Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "postgres"
	if err := removePage.Render(w, http.StatusOK, struct{ web.LayoutData }{layout}); err != nil {
		log.Printf("POSTGRESQL - remove template render error: %v", err)
	}
}

// ImportPageData is psql/import.html's template context.
type ImportPageData struct {
	web.LayoutData
	DBName string
}

func renderImportPage(a *appctx.App, w http.ResponseWriter, r *http.Request, dbName string, status int) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Import")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "postgres"
	data := ImportPageData{LayoutData: layout, DBName: dbName}
	if err := importPage.Render(w, status, data); err != nil {
		log.Printf("POSTGRESQL - import template render error: %v", err)
	}
}

// ProcessListPageData is psql/processlist.html's template context.
type ProcessListPageData struct {
	web.LayoutData
	ProcessList []ProcessRow
}

func renderProcessListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, processList []ProcessRow) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Running Queries")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "postgres"
	layout.BulkActions = processlistBulkActions(layout.T)
	data := ProcessListPageData{LayoutData: layout, ProcessList: processList}
	if err := processlistPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("POSTGRESQL - processlist template render error: %v", err)
	}
}

// RemotePostgresPageData is psql/remote_psql.html's template context.
type RemotePostgresPageData struct {
	web.LayoutData
	ServerIP                string
	ContainerPort           string
	RemotePostgreSQLDisplay string
	PostgresPort            int
}

func renderRemotePostgresPage(a *appctx.App, w http.ResponseWriter, r *http.Request, serverIP, containerPort, display string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Remote PostgreSQL")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "postgres"
	data := RemotePostgresPageData{
		LayoutData: layout, ServerIP: serverIP, ContainerPort: containerPort,
		RemotePostgreSQLDisplay: display, PostgresPort: 5432,
	}
	if err := remotePostgresPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("POSTGRESQL - remote postgres template render error: %v", err)
	}
}

// ConfigurationPageData is psql/configuration.html's template context.
type ConfigurationPageData struct {
	web.LayoutData
	CurrentConfig map[string]string
	DefaultKeys   []string
	Engine        web.DBEngine
}

func renderConfigurationPage(a *appctx.App, w http.ResponseWriter, r *http.Request, currentConfig map[string]string, defaultKeys []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit PostgreSQL configuration")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "postgres"
	data := ConfigurationPageData{LayoutData: layout, CurrentConfig: currentConfig, DefaultKeys: defaultKeys, Engine: web.PostgreSQLEngine}
	if err := configurationPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("POSTGRESQL - configuration template render error: %v", err)
	}
}
