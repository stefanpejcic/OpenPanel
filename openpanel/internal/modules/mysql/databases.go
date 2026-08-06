package mysql

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// invalidateDatabasesInfo busts the per-account databases_info cache entry
// - keyed per userContext, not a single global entry, so invalidating one
// account's cache never affects another's - see handleDatabasesInfo's doc
// comment.
func invalidateDatabasesInfo(ctx context.Context, a *appctx.App, userContext string) {
	_ = a.Cache.Delete(ctx, "databases_info:"+userContext)
}

func invalidateDatabaseCount(ctx context.Context, a *appctx.App, currentUsername string) {
	_ = a.Cache.Delete(ctx, "get_database_count:"+currentUsername)
}

// getDatabaseCount counts non-system databases for this account, cached 1h.
func getDatabaseCount(ctx context.Context, a *appctx.App, currentUsername, userContext string) int {
	count, _ := cache.Memoize(ctx, a.Cache, "get_database_count:"+currentUsername, time.Hour, func() (int, error) {
		result, err := mysqlmanager.Exec(ctx, userContext,
			"SELECT COUNT(*) AS total FROM information_schema.schemata WHERE schema_name NOT IN ("+restricted.dbsSQL+")", "")
		if err != nil || len(result) == 0 {
			return 0, nil
		}
		return mysqlmanager.ToInt(result[0][0]), nil
	})
	return count
}

// DatabaseRow is one row of mysql/databases.html's table.
type DatabaseRow struct {
	Database      string
	AssignedUsers string
	IsSystem      bool
}

// handleDatabases lists databases (and their assigned users) for this account.
func handleDatabases(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := GetMySQLVersion(ctx, a, userContext)

	showAll := r.URL.Query().Get("show_all") != ""
	unit := r.URL.Query().Get("unit")
	if unit == "" {
		unit = "mb"
	}

	status := docker.GetContainerStatus(ctx, userContext, mysqlVersion)
	if status.State == "not_found" {
		docker.StartOrStopContainer(ctx, userContext, mysqlVersion, "activate", "detached")
	}
	status.Health = resolveContainerHealth(ctx, userContext, status)

	var databaseInfo []DatabaseRow
	if status.Health == "healthy" {
		dbWhere, assignedWhere := "", ""
		if !showAll {
			dbWhere = "WHERE schema_name NOT IN (" + restricted.dbsSQL + ")"
			assignedWhere = "WHERE User NOT LIKE 'mysql.s%' AND User NOT IN (" + restricted.usersSQL + ")"
		}
		query := fmt.Sprintf(`
			SELECT 'db' AS type, schema_name AS col1, NULL AS col2
				FROM information_schema.schemata
				%s
			UNION ALL
			SELECT 'assigned', Db, GROUP_CONCAT(User SEPARATOR ', ')
				FROM mysql.db
				%s
				GROUP BY Db
		`, dbWhere, assignedWhere)

		rows, execErr := mysqlmanager.Exec(ctx, userContext, query, "")
		if execErr != nil {
			flashSess(a, w, r, "error", "Error fetching databases: "+execErr.Error())
		} else {
			assignedLookup := map[string]string{}
			for _, row := range rows {
				if toStringCell(row[0]) == "assigned" {
					assignedLookup[toStringCell(row[1])] = toStringCell(row[2])
				}
			}
			for _, row := range rows {
				if toStringCell(row[0]) == "db" {
					dbName := toStringCell(row[1])
					databaseInfo = append(databaseInfo, DatabaseRow{Database: dbName, AssignedUsers: assignedLookup[dbName], IsSystem: isSystemMySQLDatabase(dbName)})
				}
			}
		}
	} else {
		flashSess(a, w, r, "warning", mysqlWarningFlashMessage(mysqlVersion, status.State, status.Health))
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"container_state": status.State, "health_status": status.Health,
			"databases": databaseInfo, "show_all": showAll,
		})
		return
	}

	dbToastID, dbToastMessage := zeroUserDatabasesToast(databaseInfo)
	renderDatabasesPage(a, w, r, status, mysqlVersion, databaseInfo, unit, showAll, dbToastID, dbToastMessage)
}

// zeroUserDatabasesToast builds the "these databases have no users assigned"
// warning toast text, pluralized for 1/2/many databases. Returns ("", "")
// when there's nothing to warn about.
func zeroUserDatabasesToast(databases []DatabaseRow) (id, message string) {
	var zeroUserDBs []string
	for _, db := range databases {
		if !db.IsSystem && db.AssignedUsers == "" {
			zeroUserDBs = append(zeroUserDBs, db.Database)
		}
	}
	switch len(zeroUserDBs) {
	case 0:
		return "", ""
	case 1:
		return "no-users:" + zeroUserDBs[0], "Database " + zeroUserDBs[0] + " has no users assigned."
	case 2:
		joined := strings.Join(zeroUserDBs, ",")
		return "no-users:" + joined, "Databases " + strings.Join(zeroUserDBs, ", ") + " have no users assigned."
	default:
		return "no-users:multiple", strconv.Itoa(len(zeroUserDBs)) + " databases have no users assigned."
	}
}

// handleDatabasesNew creates a new database, enforcing the account's plan limit.
func handleDatabasesNew(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := GetMySQLVersion(ctx, a, userContext)

	status := docker.GetContainerStatus(ctx, userContext, mysqlVersion)
	if status.State != "running" {
		flashAndRedirect(a, w, r, "warning", fmt.Sprintf("%s service is not ready yet. Please wait for the installation to finish before creating a database.", mysqlVersion), "/mysql")
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		databaseName := r.Form.Get("database_name")
		if databaseName == "" {
			flashAndRedirect(a, w, r, "error", "Database name is required.", "/mysql/new")
			return
		}
		if !validators.IsValidIdentifier(databaseName) {
			flashAndRedirect(a, w, r, "error", fmt.Sprintf("Name %s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", databaseName), "/mysql/new")
			return
		}

		docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")

		injectedData, _ := a.InjectData(ctx, userID)
		planID, _ := injectedData["hosting_plan"].(int)
		dbLimit := 0
		if plan, planErr := a.QueryPlanDetailsByID(ctx, planID); planErr == nil {
			dbLimit = atoiDefault(plan.DBLimit, 0)
		}

		invalidateDatabaseCount(ctx, a, currentUsername)
		dbUsage := getDatabaseCount(ctx, a, currentUsername, userContext)

		if dbLimit != 0 && dbUsage >= dbLimit {
			flashAndRedirect(a, w, r, "error", fmt.Sprintf("Error creating database: '%s' - You have reached the maximum number of databases allowed.", databaseName), "/mysql/new")
			return
		}

		if _, execErr := mysqlmanager.Exec(ctx, userContext, "CREATE DATABASE IF NOT EXISTS `"+databaseName+"`", ""); execErr != nil {
			flashAndRedirect(a, w, r, "error", "Error creating database: "+execErr.Error(), "/mysql/new")
			return
		}
		invalidateDatabasesInfo(ctx, a, userContext)
		invalidateDatabaseCount(ctx, a, currentUsername)

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "created a MySQL database "+databaseName, ipAddress)
		flashSess(a, w, r, "success", "Successfully created a database "+databaseName)
		http.Redirect(w, r, "/mysql", http.StatusFound)
		return
	}

	renderNewDatabasePage(a, w, r, mysqlVersion)
}

// handleDeleteDatabase drops a database.
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
		flashAndRedirect(a, w, r, "error", "Database name is required.", "/mysql/users")
		return
	}
	if !validators.IsValidIdentifier(databaseName) {
		flashAndRedirect(a, w, r, "error", fmt.Sprintf("Name %s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", databaseName), "/mysql/users")
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+databaseName+"`", ""); execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error deleting database: "+execErr.Error(), "/mysql")
		return
	}
	invalidateDatabasesInfo(ctx, a, userContext)
	invalidateDatabaseCount(ctx, a, currentUsername)

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted a MYSQL database "+databaseName, ipAddress)
	flashSess(a, w, r, "success", "Successfully deleted a database "+databaseName)
	http.Redirect(w, r, "/mysql", http.StatusFound)
}

// handleDatabasesSizeInfo reports on-disk size per database.
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
		whereClause = "WHERE table_schema NOT IN (" + restricted.dbsSQL + ") "
	}
	query := fmt.Sprintf("SELECT table_schema, ROUND(SUM(data_length + index_length) / %d, 2) FROM information_schema.TABLES %sGROUP BY table_schema", divisor, whereClause)

	rows, execErr := mysqlmanager.Exec(ctx, userContext, query, "")
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
