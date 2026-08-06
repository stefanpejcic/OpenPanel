package mysql

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
	databasesPage     = loadPage("mysql/databases.html")
	newDatabasePage   = loadPage("mysql/new.html")
	usersPage         = loadPage("mysql/users.html")
	createUserPage    = loadPage("mysql/mysql_user.html")
	passwordPage      = loadPage("mysql/password.html")
	wizardPage        = loadPage("mysql/wizard.html")
	assignPage        = loadPage("mysql/assign.html")
	removePage        = loadPage("mysql/remove.html")
	importPage        = loadPage("mysql/import.html")
	processlistPage   = loadPage("mysql/processlist.html")
	rootPasswordPage  = loadPage("mysql/root_password.html")
	remoteMySQLPage   = loadPage("mysql/remote_mysql.html")
	configurationPage = loadPage("mysql/configuration.html")
)

// ServiceStatusData is the container_state/health_status view-model shared
// by databases.html and users.html.
type ServiceStatusData struct {
	ContainerState string
	HealthStatus   string
	// StatusDetail is the longer explanatory text shown in the table's
	// empty-state row while the service isn't running/healthy ("" when
	// running+healthy, since the real rows render instead).
	StatusDetail string
}

func serviceStatusData(status docker.ContainerStatus) ServiceStatusData {
	return ServiceStatusData{
		ContainerState: status.State, HealthStatus: status.Health,
		StatusDetail: mysqlContainerStatusDetail(status.State, status.Health),
	}
}

// DatabasesPageData is mysql/databases.html's template context.
type DatabasesPageData struct {
	web.LayoutData
	ServiceStatusData
	Service        string
	Databases      []DatabaseRow
	Unit           string
	ShowAll        bool
	DBToastID      string
	DBToastMessage string
}

func renderDatabasesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, mysqlVersion string, databases []DatabaseRow, unit string, showAll bool, dbToastID, dbToastMessage string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Databases")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DatabasesPageData{
		LayoutData: layout, ServiceStatusData: serviceStatusData(status),
		Service: mysqlVersion, Databases: databases, Unit: unit, ShowAll: showAll,
		DBToastID: dbToastID, DBToastMessage: dbToastMessage,
	}
	if err := databasesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - databases template render error: %v", err)
	}
}

// NewDatabasePageData is mysql/new.html's template context.
type NewDatabasePageData struct {
	web.LayoutData
	Service string
}

func renderNewDatabasePage(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Create Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := NewDatabasePageData{LayoutData: layout, Service: mysqlVersion}
	if err := newDatabasePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - new database template render error: %v", err)
	}
}

// UsersPageData is mysql/users.html's template context.
type UsersPageData struct {
	web.LayoutData
	ServiceStatusData
	Service string
	Users   []string
	ShowAll bool
}

func renderUsersPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, mysqlVersion string, users []string, showAll bool) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Database Users")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := UsersPageData{
		LayoutData: layout, ServiceStatusData: serviceStatusData(status),
		Service: mysqlVersion, Users: users, ShowAll: showAll,
	}
	if err := usersPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - users template render error: %v", err)
	}
}

// CreateUserPageData is mysql/mysql_user.html's template context.
type CreateUserPageData struct {
	web.LayoutData
	Service string
}

func renderCreateUserPage(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Create User")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := CreateUserPageData{LayoutData: layout, Service: mysqlVersion}
	if err := createUserPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - create user template render error: %v", err)
	}
}

// PasswordPageData is mysql/password.html's template context.
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
		log.Printf("MYSQL - password template render error: %v", err)
	}
}

// WizardPageData is mysql/wizard.html's template context.
type WizardPageData struct {
	web.LayoutData
	Service          string
	Success          bool
	FormDatabaseName string
	FormDBUser       string

	CreatedDB         string
	CreatedUser       string
	CreatedPassword   string
	CreatedHost       string
	CreatedPrivileges string
}

func renderWizardForm(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion, formDatabaseName, formDBUser string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Wizard")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := WizardPageData{LayoutData: layout, Service: mysqlVersion, FormDatabaseName: formDatabaseName, FormDBUser: formDBUser}
	if err := wizardPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - wizard template render error: %v", err)
	}
}

func renderWizardSuccess(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion, createdDB, createdUser, createdPassword, createdHost, createdPrivileges string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Wizard")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := WizardPageData{
		LayoutData: layout, Service: mysqlVersion, Success: true,
		CreatedDB: createdDB, CreatedUser: createdUser, CreatedPassword: createdPassword,
		CreatedHost: createdHost, CreatedPrivileges: createdPrivileges,
	}
	if err := wizardPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - wizard success template render error: %v", err)
	}
}

// mysqlPrivileges mirrors assign.html's hardcoded privileges list.
var mysqlPrivileges = []string{
	"ALL PRIVILEGES", "ALTER", "ALTER ROUTINE", "CREATE", "CREATE ROUTINE", "CREATE TEMPORARY TABLES",
	"CREATE VIEW", "DELETE", "DROP", "EVENT", "EXECUTE", "INDEX", "INSERT", "LOCK TABLES",
	"REFERENCES", "SELECT", "SHOW VIEW", "TRIGGER", "UPDATE",
}

// AssignPageData is mysql/assign.html's template context.
type AssignPageData struct {
	web.LayoutData
	DBUser       string
	DatabaseName string
	Privileges   []string
}

func renderAssignPage(a *appctx.App, w http.ResponseWriter, r *http.Request, dbUser, databaseName string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Assign User to Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := AssignPageData{LayoutData: layout, DBUser: dbUser, DatabaseName: databaseName, Privileges: mysqlPrivileges}
	if err := assignPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - assign template render error: %v", err)
	}
}

func renderRemovePage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Remove User from Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := removePage.Render(w, http.StatusOK, struct{ web.LayoutData }{layout}); err != nil {
		log.Printf("MYSQL - remove template render error: %v", err)
	}
}

// ImportPageData is mysql/import.html's template context.
type ImportPageData struct {
	web.LayoutData
	Service      string
	MySQLVersion string
	DBName       string
}

func renderImportPage(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion, dbName string, status int) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Import")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ImportPageData{LayoutData: layout, Service: mysqlVersion, MySQLVersion: mysqlVersion, DBName: dbName}
	if err := importPage.Render(w, status, data); err != nil {
		log.Printf("MYSQL - import template render error: %v", err)
	}
}

// ProcessListPageData is mysql/processlist.html's template context.
type ProcessListPageData struct {
	web.LayoutData
	ProcessList []ProcessRow
}

func renderProcessListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, processList []ProcessRow) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Process List")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ProcessListPageData{LayoutData: layout, ProcessList: processList}
	if err := processlistPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - processlist template render error: %v", err)
	}
}

// RootPasswordPageData is mysql/root_password.html's template context.
type RootPasswordPageData struct {
	web.LayoutData
	Service string
}

func renderRootPasswordPage(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Change Root Password")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := RootPasswordPageData{LayoutData: layout, Service: mysqlVersion}
	if err := rootPasswordPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - root password template render error: %v", err)
	}
}

// RemoteMySQLPageData is mysql/remote_mysql.html's template context.
type RemoteMySQLPageData struct {
	web.LayoutData
	Service            string
	ServerIP           string
	ContainerPort      string
	RemoteMySQLDisplay string
	MySQLPort          int
	UserAccess         []RemoteUserAccess
}

func renderRemoteMySQLPage(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion, serverIP, containerPort, remoteMySQLDisplay string, mysqlPort int, userAccess []RemoteUserAccess) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Remote MySQL")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := RemoteMySQLPageData{
		LayoutData: layout, Service: mysqlVersion, ServerIP: serverIP, ContainerPort: containerPort,
		RemoteMySQLDisplay: remoteMySQLDisplay, MySQLPort: mysqlPort, UserAccess: userAccess,
	}
	if err := remoteMySQLPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - remote mysql template render error: %v", err)
	}
}

// ConfigurationPageData is mysql/configuration.html's template context.
type ConfigurationPageData struct {
	web.LayoutData
	Service       string
	CurrentConfig map[string]string
	DefaultKeys   []string
}

func renderConfigurationPage(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion string, currentConfig map[string]string, defaultKeys []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit "+mysqlVersion+" configuration")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ConfigurationPageData{LayoutData: layout, Service: mysqlVersion, CurrentConfig: currentConfig, DefaultKeys: defaultKeys}
	if err := configurationPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - configuration template render error: %v", err)
	}
}
