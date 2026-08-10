package postgresql

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var pageFiles = []string{
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
}

func loadPage(files ...string) *web.Page {
	return web.MustLoadPage(append(append([]string{}, pageFiles...), files...)...)
}

var (
	databasesPage          = loadPage("psql/databases.html")
	newDatabasePage        = loadPage("psql/new.html")
	usersPage              = loadPage("psql/users.html")
	createUserPage         = loadPage("psql/psql_user.html")
	passwordPage           = loadPage("psql/password.html")
	wizardPage             = loadPage("psql/wizard.html")
	assignPage             = loadPage("psql/assign.html")
	removePage             = loadPage("psql/remove.html")
	importPage             = loadPage("psql/import.html")
	processlistPage        = loadPage("psql/processlist.html")
	remotePostgresPage     = loadPage("psql/remote_psql.html")
	configurationPage      = loadPage("psql/configuration.html")
)

// ServiceStatusData is the container_state/health_status view-model shared
// by databases.html and users.html.
type ServiceStatusData struct {
	ContainerState string
	HealthStatus   string
}

// DatabasesPageData is psql/databases.html's template context.
type DatabasesPageData struct {
	web.LayoutData
	ServiceStatusData
	Databases []DatabaseRow
	Unit      string
	ShowAll   bool
}

func renderDatabasesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, databases []DatabaseRow, unit string, showAll bool) {
	layout, _, err := web.BuildLayoutData(a, w, r, "PostgreSQL Databases")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DatabasesPageData{
		LayoutData:        layout,
		ServiceStatusData: ServiceStatusData{ContainerState: status.State, HealthStatus: status.Health},
		Databases:         databases, Unit: unit, ShowAll: showAll,
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
	if err := newDatabasePage.Render(w, http.StatusOK, struct{ web.LayoutData }{layout}); err != nil {
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
	layout, _, err := web.BuildLayoutData(a, w, r, "PostgreSQL Users")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
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
	data := ImportPageData{LayoutData: layout, DBName: dbName}
	if err := importPage.Render(w, status, data); err != nil {
		log.Printf("POSTGRESQL - import template render error: %v", err)
	}
}

func renderProcessListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, processlistOutput string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Process List")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := struct {
		web.LayoutData
		ProcesslistOutput string
	}{layout, processlistOutput}
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
}

func renderConfigurationPage(a *appctx.App, w http.ResponseWriter, r *http.Request, currentConfig map[string]string, defaultKeys []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit PostgreSQL configuration")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ConfigurationPageData{LayoutData: layout, CurrentConfig: currentConfig, DefaultKeys: defaultKeys}
	if err := configurationPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("POSTGRESQL - configuration template render error: %v", err)
	}
}
