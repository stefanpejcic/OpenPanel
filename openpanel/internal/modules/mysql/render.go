package mysql

import (
	"log"
	"net/http"
	"slices"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var (
	databasesPage     = web.LoadPage("mysql/databases.html")
	newDatabasePage   = web.LoadPage("db/new.html", "partials/_db_create.html")
	usersPage         = web.LoadPage("mysql/users.html")
	createUserPage    = web.LoadPage("mysql/mysql_user.html", "partials/_db_create.html")
	passwordPage      = web.LoadPage("mysql/password.html")
	wizardPage        = web.LoadPage("mysql/wizard.html", "partials/_db_create.html")
	assignPage        = web.LoadPage("mysql/assign.html", "partials/_db_create.html")
	removePage        = web.LoadPage("mysql/remove.html")
	importPage        = web.LoadPage("mysql/import.html")
	processlistPage   = web.LoadPage("mysql/processlist.html")
	rootPasswordPage  = web.LoadPage("mysql/root_password.html")
	remoteMySQLPage   = web.LoadPage("mysql/remote_mysql.html")
	configurationPage = web.LoadPage("db/configuration.html", "partials/_db_tuning.html")
)

// ServiceStatusData is the container_state/health_status view-model shared by databases.html and users.html.
type ServiceStatusData struct {
	ContainerState string
	HealthStatus   string
	// StatusDetail is the longer explanatory text shown in the table's empty-state row while the service isn't running/healthy ("" when running+healthy, since the real rows render instead)
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
	FromDisk       bool
	DBToastID      string
	DBToastMessage string
	ServerVersion  string // number from VERSION(), "" while listed from disk
	Uptime         string
	DBLimit        int // plan limit, 0 is unlimited
}

// DBQuotaPct is how full the plan's database quota is, capped at 100
func (d DatabasesPageData) DBQuotaPct() int {
	return quotaPct(d.UserDatabaseCount(), d.DBLimit)
}

func quotaPct(used, limit int) int {
	if limit <= 0 {
		return 0
	}
	return min(used*100/limit, 100)
}

// UserDatabaseCount is the number of non-system databases, for the summary tile
func (d DatabasesPageData) UserDatabaseCount() int {
	n := 0
	for _, db := range d.Databases {
		if !db.IsSystem {
			n++
		}
	}
	return n
}

// UserCount is the number of distinct users assigned to any database
func (d DatabasesPageData) UserCount() int {
	seen := map[string]bool{}
	for _, db := range d.Databases {
		for _, u := range strings.Split(db.AssignedUsers, ",") {
			if u = strings.TrimSpace(u); u != "" {
				seen[u] = true
			}
		}
	}
	return len(seen)
}

func renderDatabasesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, mysqlVersion, serverVersion string, databases []DatabaseRow, unit string, showAll, fromDisk bool, dbToastID, dbToastMessage string) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "Databases")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// the bulk actions list the database users, which needs the service up
	if len(databases) > 0 && !fromDisk {
		userContext, _ := injectedData["context"].(string)
		layout.BulkActions = databasesBulkActions(layout.T, mysqlUserOptions(r.Context(), userContext))
	}
	data := DatabasesPageData{
		LayoutData: layout, ServiceStatusData: serviceStatusData(status),
		Service: mysqlVersion, Databases: databases, Unit: unit, ShowAll: showAll, FromDisk: fromDisk,
		DBToastID: dbToastID, DBToastMessage: dbToastMessage,
		ServerVersion: serverVersion, Uptime: uptimeSince(status.Started),
		DBLimit: planDBLimit(r.Context(), a, injectedData),
	}
	if err := databasesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - databases template render error: %v", err)
	}
}

// NewDatabasePageData is db/new.html's template context for MySQL.
type NewDatabasePageData struct {
	web.LayoutData
	Service    string
	Engine     web.DBEngine
	Names      web.DBNames
	DBUsage    int
	DBLimit    int    // plan limit, 0 is unlimited
	RemotePort string // published host port, "" when unknown
	FormName   string // what was typed, kept after a failed submit
	Created    string // the database just created, shows the done state
}

func renderNewDatabasePage(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion, formName, created string) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "Create Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injectedData["context"].(string)
	data := NewDatabasePageData{
		LayoutData: layout, Service: mysqlVersion, Engine: web.MySQLEngine, FormName: formName, Created: created,
		DBLimit: planDBLimit(r.Context(), a, injectedData), RemotePort: remoteMySQLPort(userContext),
	}
	if created == "" {
		data.Names = loadNames(r.Context(), userContext)
		data.DBUsage = getDatabaseCount(r.Context(), a, layout.CurrentUsername, userContext)
	}
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
	Grants  map[string][]UserGrant // nil when the privileges lookup failed
}

func renderUsersPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, mysqlVersion string, users []string, showAll bool, grants map[string][]UserGrant) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "Database Users")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if len(users) > 0 {
		userContext, _ := injectedData["context"].(string)
		layout.BulkActions = usersBulkActions(layout.T, mysqlDatabaseOptions(r.Context(), userContext))
	}
	data := UsersPageData{
		LayoutData: layout, ServiceStatusData: serviceStatusData(status),
		Service: mysqlVersion, Users: users, ShowAll: showAll, Grants: grants,
	}
	if err := usersPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - users template render error: %v", err)
	}
}

// CreateUserPageData is mysql/mysql_user.html's template context.
type CreateUserPageData struct {
	web.LayoutData
	Service         string
	Engine          web.DBEngine
	Names           web.DBNames
	RemotePort      string
	FormName        string
	Created         string // the user just created, shows the done state with CreatedPassword once
	CreatedPassword string
}

func renderCreateUserPage(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion, formName, created, password string) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "Create User")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injectedData["context"].(string)
	data := CreateUserPageData{
		LayoutData: layout, Service: mysqlVersion, Engine: web.MySQLEngine, RemotePort: remoteMySQLPort(userContext),
		FormName: formName, Created: created, CreatedPassword: password,
	}
	if created == "" {
		data.Names = loadNames(r.Context(), userContext)
	}
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
	Engine           web.DBEngine
	Names            web.DBNames
	Role             string // only MongoDB has roles, the shared wizard template reads it
	Success          bool
	FormDatabaseName string
	FormDBUser       string

	CreatedDB         string
	CreatedUser       string
	CreatedPassword   string
	CreatedHost       string
	CreatedPrivileges string

	DBUsage    int
	DBLimit    int    // plan limit, 0 is unlimited
	RemotePort string // published host port, "" when unknown
}

func renderWizardForm(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion, formDatabaseName, formDBUser string) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "Wizard")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injectedData["context"].(string)
	data := WizardPageData{
		LayoutData: layout, Service: mysqlVersion, Engine: web.MySQLEngine, Names: loadNames(r.Context(), userContext),
		FormDatabaseName: formDatabaseName, FormDBUser: formDBUser,
		DBUsage: getDatabaseCount(r.Context(), a, layout.CurrentUsername, userContext),
		DBLimit: planDBLimit(r.Context(), a, injectedData), RemotePort: remoteMySQLPort(userContext),
	}
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
		LayoutData: layout, Service: mysqlVersion, Engine: web.MySQLEngine, Success: true,
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

// buildPrivilegesSQL checks selected against mysqlPrivileges since the result goes straight into a GRANT
func buildPrivilegesSQL(selected []string) (string, bool) {
	var out []string
	for _, p := range selected {
		p = strings.ToUpper(strings.TrimSpace(p))
		if !slices.Contains(mysqlPrivileges, p) {
			return "", false
		}
		if !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return "", false
	}
	if slices.Contains(out, "ALL PRIVILEGES") {
		return "ALL PRIVILEGES", true
	}
	return strings.Join(out, ", "), true
}

// PrivilegeGroup is one box of privilege checkboxes on assign.html
type PrivilegeGroup struct {
	Label string
	Privs []string
}

// mysqlPrivilegeGroups is mysqlPrivileges minus ALL PRIVILEGES, grouped by what they touch
var mysqlPrivilegeGroups = []PrivilegeGroup{
	{"Data", []string{"SELECT", "INSERT", "UPDATE", "DELETE"}},
	{"Structure", []string{"CREATE", "ALTER", "DROP", "INDEX", "REFERENCES", "CREATE VIEW", "SHOW VIEW", "CREATE TEMPORARY TABLES", "LOCK TABLES"}},
	{"Routines and events", []string{"CREATE ROUTINE", "ALTER ROUTINE", "EXECUTE", "EVENT", "TRIGGER"}},
}

// AssignPageData is mysql/assign.html's template context.
type AssignPageData struct {
	web.LayoutData
	Engine          web.DBEngine
	Names           web.DBNames
	DBUser          string
	DatabaseName    string
	Privileges      []string
	PrivilegeGroups []PrivilegeGroup
}

func renderAssignPage(a *appctx.App, w http.ResponseWriter, r *http.Request, dbUser, databaseName string) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "Assign User to Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injectedData["context"].(string)
	data := AssignPageData{
		LayoutData: layout, Engine: web.MySQLEngine, Names: loadNames(r.Context(), userContext),
		DBUser: dbUser, DatabaseName: databaseName, Privileges: mysqlPrivileges, PrivilegeGroups: mysqlPrivilegeGroups,
	}
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
	layout, _, err := web.BuildLayoutData(a, w, r, "Running Queries")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.BulkActions = processlistBulkActions(layout.T)
	data := ProcessListPageData{LayoutData: layout, ProcessList: processList}
	if err := processlistPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - processlist template render error: %v", err)
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
	layout.BulkActions = remoteAccessBulkActions(layout.T)
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
	Engine        web.DBEngine
}

func renderConfigurationPage(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlVersion string, currentConfig map[string]string, defaultKeys []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Edit "+mysqlVersion+" configuration")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ConfigurationPageData{LayoutData: layout, Service: mysqlVersion, CurrentConfig: currentConfig, DefaultKeys: defaultKeys, Engine: web.MySQLEngine}
	if err := configurationPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - configuration template render error: %v", err)
	}
}
