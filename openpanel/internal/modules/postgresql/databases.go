package postgresql

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// DatabaseRow is one row of psql/databases.html's table.
type DatabaseRow struct {
	Database      string
	AssignedUsers string
	IsSystem      bool
}

// handleDatabases lists the user's PostgreSQL databases, starting the
// container in the background if it isn't running yet.
func handleDatabases(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	showAll := r.URL.Query().Get("show_all") != ""
	unit := r.URL.Query().Get("unit")
	if unit == "" {
		unit = "mb"
	}

	status := docker.GetContainerStatus(ctx, userContext, "postgres")
	var databaseInfo []DatabaseRow

	switch {
	case status.State == "not_found":
		flashSess(a, w, r, "warning", "Postgres service is not yet installed. Starting it in the background..")
		docker.StartOrStopContainer(ctx, userContext, "postgres", "activate", "detached")
	case status.State != "running":
		flashSess(a, w, r, "warning", "Postgres container is not running. Please allow a few moments for the initialization..")
	default:
		dbWhere, assignedWhere := "", ""
		if !showAll {
			dbWhere = "WHERE datname NOT IN ('postgres','template0','template1')"
			assignedWhere = "WHERE d.datistemplate = false AND acl_item IS NOT NULL AND split_part(acl_item::text, '=', 1) <> ''"
		} else {
			assignedWhere = "WHERE d.datistemplate = false AND acl_item IS NOT NULL AND split_part(acl_item::text, '=', 1) <> ''"
		}

		userDatabasesSQL := "SELECT datname FROM pg_database " + dbWhere + " ORDER BY datname"
		rows, execErr := postgresmanager.Exec(ctx, userContext, userDatabasesSQL, "postgres")
		if execErr != nil {
			flashSess(a, w, r, "error", "Error fetching databases: "+execErr.Error())
		} else {
			assignedSQL := `
				SELECT
					d.datname AS database,
					COALESCE(
						STRING_AGG(
							split_part(acl_item::text, '=', 1),
							', ' ORDER BY split_part(acl_item::text, '=', 1)
						),
						''
					) AS assigned_users
				FROM pg_database d
				LEFT JOIN unnest(d.datacl) AS acl_item ON TRUE
				` + assignedWhere + `
				GROUP BY d.datname
				ORDER BY d.datname
			`
			assignedRows, assignedErr := postgresmanager.Exec(ctx, userContext, assignedSQL, "postgres")
			assignedLookup := map[string]string{}
			if assignedErr == nil {
				for _, row := range assignedRows {
					assignedLookup[toStringCell(row[0])] = toStringCell(row[1])
				}
			}
			for _, row := range rows {
				dbName := toStringCell(row[0])
				databaseInfo = append(databaseInfo, DatabaseRow{
					Database: dbName, AssignedUsers: assignedLookup[dbName], IsSystem: isSystemDatabase(dbName),
				})
			}
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"databases": databaseInfo, "show_all": showAll,
			"container_state": status.State, "health_status": status.Health,
		})
		return
	}

	renderDatabasesPage(a, w, r, status, databaseInfo, unit, showAll)
}

// handleDatabasesNew creates a new PostgreSQL database, enforcing the
// plan's database limit first.
func handleDatabasesNew(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	status := docker.GetContainerStatus(ctx, userContext, "postgres")
	if status.State != "running" {
		flashAndRedirect(a, w, r, "warning", "Postgres service is not ready yet. Please wait for the installation to finish before creating a database.", "/postgresql")
		return
	}

	if !checkPostgresInsideContainer(ctx, userContext) {
		http.Redirect(w, r, "/postgresql", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		databaseName := r.Form.Get("database_name")
		if databaseName == "" {
			flashAndRedirect(a, w, r, "error", "Database name is required.", "/postgresql/new")
			return
		}
		if !validators.IsValidIdentifier(databaseName) {
			flashAndRedirect(a, w, r, "error", "Name "+databaseName+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/postgresql/new")
			return
		}

		docker.StartComposeServiceIfNotRunning(ctx, userContext, "postgres")

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
		if rows, countErr := postgresmanager.Exec(ctx, userContext,
			"SELECT COUNT(*) FROM pg_database WHERE datname NOT IN ('postgres', 'template0', 'template1')", "postgres"); countErr == nil && len(rows) > 0 {
			dbUsage = postgresmanager.ToInt(rows[0][0])
		}

		if dbUsage >= dbLimit {
			flashAndRedirect(a, w, r, "error", "You have reached the maximum number of databases allowed."+plan.UpgradeMessage(), "/postgresql/new")
			return
		}

		if _, execErr := postgresmanager.Exec(ctx, userContext, `CREATE DATABASE "`+databaseName+`"`, "postgres"); execErr != nil {
			flashAndRedirect(a, w, r, "error", "Failed to create database: "+execErr.Error(), "/postgresql/new")
			return
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "created a PostgreSQL database "+databaseName, ipAddress)
		flashSess(a, w, r, "success", "Successfully created a PostgreSQL database "+databaseName)
		http.Redirect(w, r, "/postgresql", http.StatusFound)
		return
	}

	renderNewDatabasePage(a, w, r)
}

// handleDeleteDatabase drops a PostgreSQL database.
func handleDeleteDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	databaseName := r.Form.Get("database_name")

	if databaseName == "" {
		flashAndRedirect(a, w, r, "error", "Database name is required.", "/postgresql")
		return
	}
	if !validators.IsValidIdentifier(databaseName) {
		flashAndRedirect(a, w, r, "error", "Name "+databaseName+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/postgresql")
		return
	}

	postgresmanager.InvalidatePool(userContext, databaseName)

	if _, execErr := postgresmanager.Exec(ctx, userContext, `DROP DATABASE IF EXISTS "`+databaseName+`" WITH (FORCE)`, "postgres"); execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error deleting database "+databaseName+": "+execErr.Error(), "/postgresql")
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted a PostgreSQL database "+databaseName, ipAddress)
	flashSess(a, w, r, "success", "Successfully deleted a PostgreSQL database "+databaseName)
	http.Redirect(w, r, "/postgresql", http.StatusFound)
}

// handleDatabasesInfo serves the JSON payload the assign/remove/import
// pages' client-side <select> population fetches from.
// ComputeDatabaseAndUserNames returns the plain database/user name lists,
// reused by both /postgresql/info and the global entity search.
func ComputeDatabaseAndUserNames(ctx context.Context, userContext string) (databases, users []string, err error) {
	userDatabaseRows, execErr := postgresmanager.Exec(ctx, userContext,
		"SELECT datname FROM pg_database WHERE datname NOT IN ('postgres','template0','template1') ORDER BY datname", "postgres")
	if execErr != nil {
		return nil, nil, execErr
	}
	for _, row := range userDatabaseRows {
		databases = append(databases, toStringCell(row[0]))
	}

	userRows, execErr := postgresmanager.Exec(ctx, userContext,
		"SELECT rolname FROM pg_roles WHERE rolname NOT IN ("+filteredRolesSQL()+")", "postgres")
	if execErr != nil {
		return nil, nil, execErr
	}
	for _, row := range userRows {
		users = append(users, toStringCell(row[0]))
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
	docker.StartComposeServiceIfNotRunning(ctx, userContext, "postgres")

	userDatabaseNames, users, namesErr := ComputeDatabaseAndUserNames(ctx, userContext)
	if namesErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error executing PostgreSQL query: " + namesErr.Error()})
		return
	}
	var userDatabases []map[string]string
	for _, dbName := range userDatabaseNames {
		userDatabases = append(userDatabases, map[string]string{"database": dbName})
	}

	assignedRows, execErr := postgresmanager.Exec(ctx, userContext, `
		SELECT
			d.datname,
			STRING_AGG(r.rolname, ', ')
		FROM pg_database d
		JOIN pg_roles r ON has_database_privilege(r.rolname, d.datname, 'CONNECT')
		WHERE r.rolname NOT LIKE 'pg_%'
		  AND r.rolname NOT IN (`+filteredRolesSQL()+`)
		GROUP BY d.datname
		ORDER BY d.datname
	`, "postgres")
	if execErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error executing PostgreSQL query: " + execErr.Error()})
		return
	}
	var assignedDatabases []map[string]string
	for _, row := range assignedRows {
		assignedDatabases = append(assignedDatabases, map[string]string{"database": toStringCell(row[0]), "users": toStringCell(row[1])})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"databases": userDatabases, "users": users, "assigned_databases": assignedDatabases,
	})
}

// handleDatabasesSizeInfo serves the /json/postgresql-size route: each
// database's size, converted to the requested unit.
func handleDatabasesSizeInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	unit := strings.ToLower(r.URL.Query().Get("unit"))
	if unit == "" {
		unit = "mb"
	}
	divisors := map[string]int64{"bytes": 1, "kb": 1024, "mb": 1024 * 1024, "gb": 1024 * 1024 * 1024}
	divisor, ok := divisors[unit]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": `Invalid unit parameter. Use "bytes", "kb", "mb", or "gb".`})
		return
	}
	showAll := r.URL.Query().Get("show_all") != ""

	whereClause := ""
	if !showAll {
		whereClause = "WHERE datname NOT IN ('postgres','template0','template1','information_schema','pg_catalog') "
	}
	query := "SELECT datname, ROUND(pg_database_size(datname) / " + strconv.FormatInt(divisor, 10) + ", 2) FROM pg_database " + whereClause + "ORDER BY datname"

	rows, execErr := postgresmanager.Exec(ctx, userContext, query, "postgres")
	if execErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	sizeKey := "Size (" + strings.ToUpper(unit) + ")"
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		result = append(result, map[string]any{"Database": toStringCell(row[0]), sizeKey: toFloatCell(row[1])})
	}
	writeJSON(w, http.StatusOK, result)
}
