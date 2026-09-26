package mongodb

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
	databasesPage   = loadPage("mongodb/databases.html")
	newDatabasePage = loadPage("mongodb/new.html")
	usersPage       = loadPage("mongodb/users.html")
	createUserPage  = loadPage("mongodb/mongodb_user.html")
	passwordPage    = loadPage("mongodb/password.html")
	wizardPage      = loadPage("mongodb/wizard.html")
	assignPage      = loadPage("mongodb/assign.html")
	removePage      = loadPage("mongodb/remove.html")
	importPage      = loadPage("mongodb/import.html")
	processlistPage = loadPage("mongodb/processlist.html")
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
}

func renderDatabasesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, databases []DatabaseRow) {
	layout, injectedData, err := web.BuildLayoutData(a, w, r, "MongoDB Databases")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if len(databases) > 0 {
		userContext, _ := injectedData["context"].(string)
		_, users, _ := ComputeDatabaseAndUserNames(r.Context(), userContext)
		layout.BulkActions = databasesBulkActions(layout.T, users)
	}
	layout.Service = "mongodb"
	data := DatabasesPageData{
		LayoutData:        layout,
		ServiceStatusData: ServiceStatusData{ContainerState: status.State, HealthStatus: status.Health},
		Databases:         databases,
		StatusDetail:      mongoContainerStatusDetail(status.State, status.Health),
	}
	if err := databasesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - databases template render error: %v", err)
	}
}

func renderNewDatabasePage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Create Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "mongodb"
	if err := newDatabasePage.Render(w, http.StatusOK, struct{ web.LayoutData }{layout}); err != nil {
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

func renderCreateUserPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Create User")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "mongodb"
	if err := createUserPage.Render(w, http.StatusOK, struct{ web.LayoutData }{layout}); err != nil {
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

// WizardPageData is mongodb/wizard.html's template context.
type WizardPageData struct {
	web.LayoutData
	Roles []string
}

func renderWizardPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Database Wizard")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "mongodb"
	data := WizardPageData{LayoutData: layout, Roles: mongoRoles}
	if err := wizardPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - wizard template render error: %v", err)
	}
}

// AssignPageData is mongodb/assign.html's template context.
type AssignPageData struct {
	web.LayoutData
	Roles []string
}

func renderAssignPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Assign User to Database")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.Service = "mongodb"
	data := AssignPageData{LayoutData: layout, Roles: mongoRoles}
	if err := assignPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MONGODB - assign template render error: %v", err)
	}
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
