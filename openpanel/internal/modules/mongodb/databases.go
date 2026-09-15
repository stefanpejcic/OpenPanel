package mongodb

import (
	"context"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// DatabaseRow is one row of mongodb/databases.html's table.
type DatabaseRow struct {
	Database    string
	SizeDisplay string
	IsSystem    bool
}

// normalizeHealth treats "no healthcheck configured" (podman reports "none") as healthy for a
// running container - unlike postgres/mysql, the mongodb compose service has no HEALTHCHECK
// defined, so its health would otherwise sit at "none" forever and never satisfy the "healthy"
// gate that unlocks the databases table.
func normalizeHealth(status docker.ContainerStatus) docker.ContainerStatus {
	if status.State == "running" && status.Health == "none" {
		status.Health = "healthy"
	}
	return status
}

// handleDatabases lists the user's MongoDB databases, starting the container in the background if it isn't running yet
func handleDatabases(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	status := normalizeHealth(docker.GetContainerStatus(ctx, userContext, "mongodb"))
	var databaseInfo []DatabaseRow

	switch {
	case status.State == "not_found":
		flashSess(a, w, r, "warning", "MongoDB service is not yet installed. Starting it in the background..")
		docker.StartOrStopContainer(ctx, userContext, "mongodb", "activate", "detached")
	case status.State != "running":
		flashSess(a, w, r, "warning", "MongoDB container is not running. Please allow a few moments for the initialization..")
	default:
		dbs, listErr := mongomanager.ListDatabases(ctx, userContext)
		if listErr != nil {
			flashSess(a, w, r, "error", "Error fetching databases: "+listErr.Error())
		} else {
			for _, d := range dbs {
				databaseInfo = append(databaseInfo, DatabaseRow{Database: d.Name, SizeDisplay: formatSize(d.SizeBytes), IsSystem: mongomanager.IsSystemDatabase(d.Name)})
			}
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"databases":       databaseInfo,
			"container_state": status.State, "health_status": status.Health,
		})
		return
	}

	renderDatabasesPage(a, w, r, status, databaseInfo)
}

// handleDatabasesNew creates a new MongoDB database, enforcing the plan's database limit first
func handleDatabasesNew(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	status := docker.GetContainerStatus(ctx, userContext, "mongodb")
	if status.State != "running" {
		flashAndRedirect(a, w, r, "warning", "MongoDB service is not ready yet. Please wait for the installation to finish before creating a database.", "/mongodb")
		return
	}

	if !checkMongoInsideContainer(ctx, userContext) {
		http.Redirect(w, r, "/mongodb", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		databaseName := r.Form.Get("database_name")
		if databaseName == "" {
			flashAndRedirect(a, w, r, "error", "Database name is required.", "/mongodb/new")
			return
		}
		if !validators.IsValidIdentifier(databaseName) {
			flashAndRedirect(a, w, r, "error", "Name "+databaseName+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/mongodb/new")
			return
		}
		if isRestrictedDatabase(databaseName) {
			flashAndRedirect(a, w, r, "error", "This is a system database that can not be used.", "/mongodb/new")
			return
		}

		docker.StartComposeServiceIfNotRunning(ctx, userContext, "mongodb")

		injectedData, _ := a.InjectData(ctx, userID)
		planID, _ := injectedData["hosting_plan"].(int)
		plan, _ := a.QueryPlanDetailsByID(ctx, planID)
		dbLimit := 100
		if v := atoiDefault(plan.DBLimit, 0); v != 0 {
			dbLimit = v
		} else {
			dbLimit = 1000000
		}

		dbUsage := 0
		if dbs, countErr := mongomanager.ListDatabases(ctx, userContext); countErr == nil {
			dbUsage = len(dbs)
		}

		if dbUsage >= dbLimit {
			flashAndRedirect(a, w, r, "error", "You have reached the maximum number of databases allowed."+plan.UpgradeMessage(), "/mongodb/new")
			return
		}

		if createErr := mongomanager.CreateDatabase(ctx, userContext, databaseName); createErr != nil {
			flashAndRedirect(a, w, r, "error", "Failed to create database: "+createErr.Error(), "/mongodb/new")
			return
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "created a MongoDB database "+databaseName, ipAddress)
		flashSess(a, w, r, "success", "Successfully created a MongoDB database "+databaseName)
		http.Redirect(w, r, "/mongodb", http.StatusFound)
		return
	}

	renderNewDatabasePage(a, w, r)
}

// handleDeleteDatabase drops a MongoDB database.
func handleDeleteDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	databaseName := r.Form.Get("database_name")

	switch {
	case databaseName == "":
		flashAndRedirect(a, w, r, "error", "Database name is required.", "/mongodb")
		return
	case !validators.IsValidIdentifier(databaseName):
		flashAndRedirect(a, w, r, "error", "Name "+databaseName+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/mongodb")
		return
	case isRestrictedDatabase(databaseName):
		flashAndRedirect(a, w, r, "error", "This is a system database that cannot be deleted.", "/mongodb")
		return
	}

	if dropErr := mongomanager.DropDatabase(ctx, userContext, databaseName); dropErr != nil {
		flashAndRedirect(a, w, r, "error", "Error deleting database "+databaseName+": "+dropErr.Error(), "/mongodb")
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted a MongoDB database "+databaseName, ipAddress)
	flashSess(a, w, r, "success", "Successfully deleted a MongoDB database "+databaseName)
	http.Redirect(w, r, "/mongodb", http.StatusFound)
}

// ComputeDatabaseAndUserNames returns the plain database/user name lists, reused by both /mongodb/info and the global entity search
func ComputeDatabaseAndUserNames(ctx context.Context, userContext string) (databases, users []string, err error) {
	dbs, listErr := mongomanager.ListDatabases(ctx, userContext)
	if listErr != nil {
		return nil, nil, listErr
	}
	for _, d := range dbs {
		databases = append(databases, d.Name)
	}

	userList, usersErr := mongomanager.ListUsers(ctx, userContext)
	if usersErr != nil {
		return nil, nil, usersErr
	}
	for _, u := range userList {
		if isRestrictedUser(u.Username) {
			continue
		}
		users = append(users, u.Username)
	}
	return databases, users, nil
}

func handleDatabasesInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	docker.StartComposeServiceIfNotRunning(ctx, userContext, "mongodb")

	databaseNames, users, namesErr := ComputeDatabaseAndUserNames(ctx, userContext)
	if namesErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error executing MongoDB query: " + namesErr.Error()})
		return
	}
	var databases []map[string]string
	for _, dbName := range databaseNames {
		databases = append(databases, map[string]string{"database": dbName})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"databases": databases, "users": users,
	})
}
