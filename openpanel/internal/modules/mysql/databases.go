package mysql

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// invalidateDatabasesInfo busts the per-account databases_info cache entry - keyed per userContext, not a single global entry, so invalidating one account's cache never affects another's
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
	_, _, userContext, err := auth.Injected(a, r)
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
	switch status.State {
	case "not_found":
		docker.StartOrStopContainer(ctx, userContext, mysqlVersion, "activate", "detached")
	case "exited", "created":
		// sentinel stops the database of an account that has none after a while, opening this page brings it back; not asked
		// whether it's up yet, that waits on the socket for up to 15s while the page can list from disk and refresh
		startContainerInBackground(userContext, mysqlVersion)
		status = docker.ContainerStatus{State: "running", Health: "starting"}
	default:
		status.Health = resolveContainerHealth(ctx, userContext, status)
	}

	var databaseInfo []DatabaseRow
	var serverVersion string
	fromDisk := status.Health != "healthy"
	if fromDisk {
		// listed from the datadir until the service answers, so the user sees their databases are still there
		databaseInfo = databasesFromDisk(userContext, showAll)
	}
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
			UNION ALL
			SELECT 'version', VERSION(), NULL
		`, dbWhere, assignedWhere)

		rows, execErr := mysqlmanager.Exec(ctx, userContext, query, "")
		if execErr != nil {
			web.Flash(a, w, r, "error", web.Tr(a, r, "Error fetching databases: %(error)s", "error", execErr.Error()))
		} else {
			assignedLookup := map[string]string{}
			for _, row := range rows {
				switch toStringCell(row[0]) {
				case "assigned":
					assignedLookup[toStringCell(row[1])] = toStringCell(row[2])
				case "version":
					serverVersion = shortServerVersion(toStringCell(row[1]))
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
		web.Flash(a, w, r, "warning", web.Tr(a, r, mysqlWarningFlashMessage(mysqlVersion, status.State, status.Health), "service", mysqlVersion, "state", status.State))
	}

	if r.URL.Query().Get("output") == "json" {
		web.WriteJSON(w, http.StatusOK, map[string]any{
			"container_state": status.State, "health_status": status.Health,
			"databases": databaseInfo, "show_all": showAll, "from_disk": fromDisk,
		})
		return
	}

	// users aren't known while listing from disk, every database would look unassigned
	var dbToastID, dbToastMessage string
	if !fromDisk {
		dbToastID, dbToastMessage = zeroUserDatabasesToast(databaseInfo)
	}
	renderDatabasesPage(a, w, r, status, mysqlVersion, serverVersion, databaseInfo, unit, showAll, fromDisk, dbToastID, dbToastMessage)
}

// startContainerInBackground starts a stopped container as it is, quicker than a compose up which recreates it
func startContainerInBackground(userContext, name string) {
	cmd := podmanmanager.Command(context.Background(), userContext, podmanmanager.PodmanArgv(userContext, "start", name))
	if cmd.Start() == nil {
		go func() { _ = cmd.Wait() }()
	}
}

var mysqlEncodedCharRE = regexp.MustCompile(`@([0-9a-fA-F]{4})`)

// databasesFromDisk lists the database directories in the account's datadir, mysqld stores a name like my-db as my@002ddb
func databasesFromDisk(userContext string, showAll bool) []DatabaseRow {
	entries, err := tenantfs.ReadDir("/home/" + userContext + "/docker-data/volumes/" + userContext + "_mysql_data/_data")
	if err != nil {
		return nil
	}
	var rows []DatabaseRow
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), "#") || e.Name() == "lost+found" {
			continue
		}
		name := mysqlEncodedCharRE.ReplaceAllStringFunc(e.Name(), func(m string) string {
			n, err := strconv.ParseUint(m[1:], 16, 32)
			if err != nil {
				return m
			}
			return string(rune(n))
		})
		if !showAll && slices.Contains(restricted.databases, name) {
			continue
		}
		rows = append(rows, DatabaseRow{Database: name, IsSystem: isSystemMySQLDatabase(name)})
	}
	return rows
}

// zeroUserDatabasesToast builds the "these databases have no users assigned" warning toast text, pluralized for 1/2/many databases, returns ("", "") when there's nothing to warn about
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
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := GetMySQLVersion(ctx, a, userContext)

	status := docker.GetContainerStatus(ctx, userContext, mysqlVersion)
	if status.State != "running" {
		web.FlashRedirect(a, w, r, "warning", web.Tr(a, r, "%(mysql_version)s service is not ready yet. Please wait for the installation to finish before creating a database.", "mysql_version", mysqlVersion), "/mysql")
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		typed := strings.TrimSpace(r.Form.Get("database_name"))
		databaseName := mysqlmanager.WithPrefix(a.Config, userContext, typed)
		// errors keep the typed name in the form instead of bouncing to an empty one
		fail := func(message string) {
			web.Flash(a, w, r, "error", message)
			renderNewDatabasePage(a, w, r, mysqlVersion, typed, "")
		}
		if msg := checkNewDatabase(databaseName, loadNames(ctx, userContext)); msg != "" {
			fail(web.Tr(a, r, msg))
			return
		}

		docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")

		injectedData, _ := a.InjectData(ctx, userID)
		planID, _ := injectedData["hosting_plan"].(int)
		plan, _ := a.QueryPlanDetailsByID(ctx, planID)
		dbLimit := web.AtoiDefault(plan.DBLimit, 0)

		invalidateDatabaseCount(ctx, a, currentUsername)
		dbUsage := getDatabaseCount(ctx, a, currentUsername, userContext)

		if dbLimit != 0 && dbUsage >= dbLimit {
			fail(web.Tr(a, r, "Error creating database: '%(database_name)s' - You have reached the maximum number of databases allowed.%(upgrade_message)s", "database_name", databaseName, "upgrade_message", plan.UpgradeMessage()))
			return
		}

		if _, execErr := mysqlmanager.Exec(ctx, userContext, "CREATE DATABASE `"+databaseName+"`", ""); execErr != nil {
			fail(web.Tr(a, r, "Error creating database: %(error)s", "error", execErr.Error()))
			return
		}
		invalidateDatabasesInfo(ctx, a, userContext)
		invalidateDatabaseCount(ctx, a, currentUsername)

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "created a MySQL database "+databaseName, ipAddress)
		web.Flash(a, w, r, "success", web.Tr(a, r, "Successfully created a database %(database_name)s", "database_name", databaseName))
		renderNewDatabasePage(a, w, r, mysqlVersion, "", databaseName)
		return
	}

	renderNewDatabasePage(a, w, r, mysqlVersion, "", "")
}

// handleDeleteDatabase drops a database.
func handleDeleteDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	databaseName := r.Form.Get("database_name")

	if databaseName == "" {
		web.FlashRedirect(a, w, r, "error", "Database name is required.", "/mysql/users")
		return
	}
	if !validators.IsValidIdentifier(databaseName) {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Name %(database_name)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "database_name", databaseName), "/mysql/users")
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+databaseName+"`", ""); execErr != nil {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Error deleting database: %(error)s", "error", execErr.Error()), "/mysql")
		return
	}
	invalidateDatabasesInfo(ctx, a, userContext)
	invalidateDatabaseCount(ctx, a, currentUsername)

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted a MYSQL database "+databaseName, ipAddress)
	web.Flash(a, w, r, "success", web.Tr(a, r, "Successfully deleted a database %(database_name)s", "database_name", databaseName))
	http.Redirect(w, r, "/mysql", http.StatusFound)
}

// handleDatabasesSizeInfo reports on-disk size per database.
func handleDatabasesSizeInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, _, userContext, err := auth.Injected(a, r)
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
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": `Invalid unit parameter. Use "bytes", "kb", "mb", or "gb".`})
		return
	}
	showAll := r.URL.Query().Get("show_all") != ""

	whereClause := ""
	if !showAll {
		whereClause = "WHERE table_schema NOT IN (" + restricted.dbsSQL + ") "
	}
	query := fmt.Sprintf("SELECT table_schema, ROUND(SUM(data_length + index_length) / %d, 2), COUNT(*) FROM information_schema.TABLES %sGROUP BY table_schema", divisor, whereClause)

	rows, execErr := mysqlmanager.Exec(ctx, userContext, query, "")
	if execErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	sizeKey := "Size (" + strings.ToUpper(unit) + ")"
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		entry := map[string]any{"Database": toStringCell(row[0]), sizeKey: toFloatCell(row[1])}
		if len(row) > 2 {
			entry["Tables"] = mysqlmanager.ToInt(row[2])
		}
		result = append(result, entry)
	}
	web.WriteJSON(w, http.StatusOK, result)
}
