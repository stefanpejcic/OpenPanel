package mongodb

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var (
	databasesPage   = web.LoadPage("mongodb/databases.html")
	newDatabasePage = web.LoadPage("mongodb/new.html")
	usersPage       = web.LoadPage("mongodb/users.html")
	createUserPage  = web.LoadPage("mongodb/mongodb_user.html")
	passwordPage    = web.LoadPage("mongodb/password.html")
	wizardPage      = web.LoadPage("mongodb/wizard.html")
	assignPage      = web.LoadPage("mongodb/assign.html")
	removePage      = web.LoadPage("mongodb/remove.html")
	importPage      = web.LoadPage("mongodb/import.html")
	processlistPage = web.LoadPage("mongodb/processlist.html")
)

// ServiceStatusData is the container_state/health_status view-model shared by databases.html and users.html.
type ServiceStatusData struct {
	ContainerState string
	HealthStatus   string
}

// DatabasesPageData is mongodb/databases.html's template context.
type DatabasesPageData struct {
	web.LayoutData
	ServiceStatusData
	Databases    []DatabaseRow
	StatusDetail string // longer explanatory text for the table's empty-state row while the service isn't running/healthy, "" when running+healthy
	UserCount    int    // database users other than admin/root
	HasUsers     bool   // false when the users lookup failed or wasn't run
}

// TotalSize sums the size of the user's own databases
func (d DatabasesPageData) TotalSize() string {
	var total int64
	for _, db := range d.Databases {
		if !db.IsSystem {
			total += db.SizeBytes
		}
	}
	return formatSize(total)
}

// LargestDatabase is the user's biggest database, "" when there's none
func (d DatabasesPageData) LargestDatabase() string {
	name, size := "", int64(-1)
	for _, db := range d.Databases {
		if !db.IsSystem && db.SizeBytes > size {
			name, size = db.Database, db.SizeBytes
		}
	}
	return name
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

// shouldCountUsers: a running server with no databases yet still gets a real Users count (0) instead of a dash
func shouldCountUsers(state string, databases int) bool {
	return databases > 0 || state == "running"
}

func renderDatabasesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, databases []DatabaseRow) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "MongoDB Databases")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var users []string
	hasUsers := false
	if shouldCountUsers(status.State, len(databases)) {
		userContext, _ := injectedData["context"].(string)
		var usersErr error
		_, users, usersErr = ComputeDatabaseAndUserNames(r.Context(), userContext)
		hasUsers = usersErr == nil
	}
	if len(databases) > 0 {
		layout.BulkActions = databasesBulkActions(layout.T, users)
	}
	layout.Service = "mongodb"
	data := DatabasesPageData{
		LayoutData:        layout,
		ServiceStatusData: ServiceStatusData{ContainerState: status.State, HealthStatus: status.Health},
		Databases:         databases,
		StatusDetail:      mongoContainerStatusDetail(status.State, status.Health),
		UserCount:         len(users), HasUsers: hasUsers,
	}
	if err := databasesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - databases template render error: %v", err)
	}
}

// CreatePageData is the template context of the shared create pages (new.html, mongodb_user.html, wizard.html, assign.html)
type CreatePageData struct {
	web.LayoutData
	Engine   web.DBEngine
	Names    web.DBNames
	DBLimit  int // plan limit, 0 is unlimited
	DBUsage  int
	FormName string // what was typed, kept after a failed submit
	Created  string // the database or user just created, shows the done state
	Roles    []string

	CreatedPassword  string // a new user's password, shown once
	Success          bool   // wizard done
	FormDatabaseName string
	FormDBUser       string
	CreatedDB        string
	CreatedUser      string
	Role             string
	DBUser           string // assign page preselection
	DatabaseName     string
}

// createPage builds CreatePageData, listing the account's names unless done says they won't be shown
func createPage(a *appctx.App, w http.ResponseWriter, r *http.Request, title string, done bool) (CreatePageData, bool) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return CreatePageData{}, false
	}
	layout.Service = "mongodb"
	data := CreatePageData{LayoutData: layout, Engine: web.MongoDBEngine, DBLimit: planDBLimit(r.Context(), a, injectedData), Roles: mongoRoles}
	if !done {
		userContext, _ := injectedData["context"].(string)
		data.Names = loadNames(r.Context(), userContext)
		data.DBUsage = len(data.Names.Databases)
	}
	return data, true
}

func renderNewDatabasePage(a *appctx.App, w http.ResponseWriter, r *http.Request, formName, created string) {
	data, ok := createPage(a, w, r, "Create Database", created != "")
	if !ok {
		return
	}
	data.FormName, data.Created = formName, created
	if err := newDatabasePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - new database template render error: %v", err)
	}
}

// UserRow is one row of mongodb/users.html's table.
type UserRow struct {
	Name     string
	Roles    string
	IsSystem bool
}

// UsersPageData is mongodb/users.html's template context.
type UsersPageData struct {
	web.LayoutData
	ServiceStatusData
	Users []UserRow
}

func renderUsersPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, users []UserRow) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "MongoDB Users")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if len(users) > 0 {
		userContext, _ := injectedData["context"].(string)
		databases, _, _ := ComputeDatabaseAndUserNames(r.Context(), userContext)
		layout.BulkActions = usersBulkActions(layout.T, databases)
	}
	layout.Service = "mongodb"
	data := UsersPageData{
		LayoutData:        layout,
		ServiceStatusData: ServiceStatusData{ContainerState: status.State, HealthStatus: status.Health},
		Users:             users,
	}
	if err := usersPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - users template render error: %v", err)
	}
}

func renderCreateUserPage(a *appctx.App, w http.ResponseWriter, r *http.Request, formName, created, password string) {
	data, ok := createPage(a, w, r, "Create User", created != "")
	if !ok {
		return
	}
	data.FormName, data.Created, data.CreatedPassword = formName, created, password
	if err := createUserPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - create user template render error: %v", err)
	}
}

// PasswordPageData is mongodb/password.html's template context.
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
	layout.Service = "mongodb"
	data := PasswordPageData{LayoutData: layout, DBUser: dbUser}
	if err := passwordPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - password template render error: %v", err)
	}
}

func renderWizardPage(a *appctx.App, w http.ResponseWriter, r *http.Request, formDB, formUser, role string) {
	data, ok := createPage(a, w, r, "Database Wizard", false)
	if !ok {
		return
	}
	data.FormDatabaseName, data.FormDBUser, data.Role = formDB, formUser, role
	if err := wizardPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - wizard template render error: %v", err)
	}
}

func renderWizardSuccess(a *appctx.App, w http.ResponseWriter, r *http.Request, createdDB, createdUser, password, role string) {
	data, ok := createPage(a, w, r, "Database Wizard", true)
	if !ok {
		return
	}
	data.Success, data.CreatedDB, data.CreatedUser, data.CreatedPassword, data.Role = true, createdDB, createdUser, password, role
	if err := wizardPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - wizard success template render error: %v", err)
	}
}

func renderAssignPage(a *appctx.App, w http.ResponseWriter, r *http.Request, dbUser, databaseName string) {
	data, ok := createPage(a, w, r, "Assign User to Database", false)
	if !ok {
		return
	}
	data.DBUser, data.DatabaseName = dbUser, databaseName
	if err := assignPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - assign template render error: %v", err)
	}
}

// AssignPageData is mongodb/remove.html's template context.
type AssignPageData struct {
	web.LayoutData
	Roles []string
}

func renderRemovePage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Remove User from Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "mongodb"
	data := AssignPageData{LayoutData: layout, Roles: mongoRoles}
	if err := removePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - remove template render error: %v", err)
	}
}

// ImportPageData is mongodb/import.html's template context.
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
	layout.Service = "mongodb"
	data := ImportPageData{LayoutData: layout, DBName: dbName}
	if err := importPage.Render(w, status, data); err != nil {
		log.Printf("MONGODB - import template render error: %v", err)
	}
}

// ProcessListPageData is mongodb/processlist.html's template context.
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
	layout.Service = "mongodb"
	layout.BulkActions = processlistBulkActions(layout.T)
	data := ProcessListPageData{LayoutData: layout, ProcessList: processList}
	if err := processlistPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - processlist template render error: %v", err)
	}
}
