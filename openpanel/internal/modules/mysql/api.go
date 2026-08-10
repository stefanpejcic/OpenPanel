package mysql

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterAPI wires the MySQL REST endpoints onto mux. The
// /databases/{db_name} and /users/{db_user} sub-resources share their
// prefix with a literal suffix - Go's http.ServeMux requires a "{...}"
// wildcard to be the final segment, so each verb gets a "{rest...}"
// catch-all where needed and the dispatch funcs below strip the known
// suffix by hand to route to the right handler.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "mysql", "GET /api/mysql/databases", func(w http.ResponseWriter, r *http.Request) { apiMySQLListDatabases(a, w, r) })
	apiregistry.Handle(mux, a, "mysql", "POST /api/mysql/databases", func(w http.ResponseWriter, r *http.Request) { apiMySQLCreateDatabase(a, w, r) })

	apiregistry.Add("GET /api/mysql/databases/{db_name}/tables")
	apiregistry.Add("GET /api/mysql/databases/{db_name}/export")
	mux.Handle("GET /api/mysql/databases/{rest...}", auth.RequireAPI(a, "mysql")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiMySQLDatabasesGetDispatch(a, w, r) })))

	apiregistry.Add("POST /api/mysql/databases/{db_name}/optimize")
	apiregistry.Add("POST /api/mysql/databases/{db_name}/repair")
	apiregistry.Add("POST /api/mysql/databases/{db_name}/import")
	mux.Handle("POST /api/mysql/databases/{rest...}", auth.RequireAPI(a, "mysql")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiMySQLDatabasesPostDispatch(a, w, r) })))

	apiregistry.Handle(mux, a, "mysql", "DELETE /api/mysql/databases/{db_name}", func(w http.ResponseWriter, r *http.Request) { apiMySQLDeleteDatabase(a, w, r) })

	apiregistry.Handle(mux, a, "mysql", "GET /api/mysql/size", func(w http.ResponseWriter, r *http.Request) { apiMySQLDatabasesSize(a, w, r) })

	apiregistry.Handle(mux, a, "mysql", "GET /api/mysql/users", func(w http.ResponseWriter, r *http.Request) { apiMySQLListUsers(a, w, r) })
	apiregistry.Handle(mux, a, "mysql", "POST /api/mysql/users", func(w http.ResponseWriter, r *http.Request) { apiMySQLCreateUser(a, w, r) })
	apiregistry.Handle(mux, a, "mysql", "DELETE /api/mysql/users/{db_user}", func(w http.ResponseWriter, r *http.Request) { apiMySQLDeleteUser(a, w, r) })

	apiregistry.Add("GET /api/mysql/users/{db_user}/privileges/{db_name}")
	mux.Handle("GET /api/mysql/users/{rest...}", auth.RequireAPI(a, "mysql")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiMySQLUsersGetDispatch(a, w, r) })))

	apiregistry.Handle(mux, a, "mysql", "PATCH /api/mysql/users/{db_user}/password", func(w http.ResponseWriter, r *http.Request) { apiMySQLChangeUserPassword(a, w, r) })

	apiregistry.Handle(mux, a, "mysql", "POST /api/mysql/grants", func(w http.ResponseWriter, r *http.Request) { apiMySQLGrant(a, w, r) })
	apiregistry.Handle(mux, a, "mysql", "DELETE /api/mysql/grants", func(w http.ResponseWriter, r *http.Request) { apiMySQLRevoke(a, w, r) })

	apiregistry.Handle(mux, a, "mysql", "GET /api/mysql/info", func(w http.ResponseWriter, r *http.Request) { apiMySQLInfo(a, w, r) })
	apiregistry.Handle(mux, a, "mysql", "GET /api/mysql/processlist", func(w http.ResponseWriter, r *http.Request) { apiMySQLProcesslist(a, w, r) })

	apiregistry.Handle(mux, a, "mysql", "GET /api/mysql/remote-access", func(w http.ResponseWriter, r *http.Request) { apiMySQLRemoteAccessStatus(a, w, r) })
	apiregistry.Handle(mux, a, "mysql", "POST /api/mysql/remote-access", func(w http.ResponseWriter, r *http.Request) { apiMySQLRemoteAccessToggle(a, w, r) })
	apiregistry.Handle(mux, a, "mysql", "POST /api/mysql/remote-access/entries", func(w http.ResponseWriter, r *http.Request) { apiMySQLRemoteAccessAdd(a, w, r) })
	apiregistry.Handle(mux, a, "mysql", "PATCH /api/mysql/remote-access/entries", func(w http.ResponseWriter, r *http.Request) { apiMySQLRemoteAccessEdit(a, w, r) })
	apiregistry.Handle(mux, a, "mysql", "DELETE /api/mysql/remote-access/entries", func(w http.ResponseWriter, r *http.Request) { apiMySQLRemoteAccessDelete(a, w, r) })

	apiregistry.Handle(mux, a, "mysql", "GET /api/mysql/configuration", func(w http.ResponseWriter, r *http.Request) { apiMySQLGetConfig(a, w, r) })
	apiregistry.Handle(mux, a, "mysql", "PUT /api/mysql/configuration", func(w http.ResponseWriter, r *http.Request) { apiMySQLUpdateConfig(a, w, r) })
}

func writeAPIMySQLJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func apiMySQLDatabasesGetDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/tables"):
		r.SetPathValue("db_name", strings.TrimSuffix(rest, "/tables"))
		apiMySQLDatabaseTables(a, w, r)
	case strings.HasSuffix(rest, "/export"):
		r.SetPathValue("db_name", strings.TrimSuffix(rest, "/export"))
		apiMySQLExportDatabase(a, w, r)
	default:
		http.NotFound(w, r)
	}
}

func apiMySQLDatabasesPostDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	switch {
	case strings.HasSuffix(rest, "/optimize"):
		r.SetPathValue("db_name", strings.TrimSuffix(rest, "/optimize"))
		apiMySQLDBMaintenance(a, w, r, "optimize")
	case strings.HasSuffix(rest, "/repair"):
		r.SetPathValue("db_name", strings.TrimSuffix(rest, "/repair"))
		apiMySQLDBMaintenance(a, w, r, "repair")
	case strings.HasSuffix(rest, "/import"):
		r.SetPathValue("db_name", strings.TrimSuffix(rest, "/import"))
		apiMySQLImportDatabase(a, w, r)
	default:
		http.NotFound(w, r)
	}
}

func apiMySQLUsersGetDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	const marker = "/privileges/"
	if idx := strings.LastIndex(rest, marker); idx != -1 {
		r.SetPathValue("db_user", rest[:idx])
		r.SetPathValue("db_name", rest[idx+len(marker):])
		apiMySQLUserPrivileges(a, w, r)
		return
	}
	http.NotFound(w, r)
}

// mysqlAPIError extracts a mysqlmanager error's message for API responses.
func mysqlAPIError(err error) string {
	return err.Error()
}

// ── Databases ────────────────────────────────────────────────────────────

// apiMySQLListDatabases lists all databases with their assigned users.
func apiMySQLListDatabases(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rows, execErr := mysqlmanager.Exec(ctx, userContext, `
		SELECT 'db' AS type, schema_name AS col1, NULL AS col2
			FROM information_schema.schemata
			WHERE schema_name NOT IN (`+restricted.dbsSQL+`)
		UNION ALL
		SELECT 'assigned', Db, GROUP_CONCAT(User SEPARATOR ', ')
			FROM mysql.db
			WHERE User NOT LIKE 'mysql.s%' AND User NOT IN (`+restricted.usersSQL+`)
			GROUP BY Db
	`, "")
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	assigned := map[string]string{}
	for _, row := range rows {
		if toStringCell(row[0]) == "assigned" {
			assigned[toStringCell(row[1])] = toStringCell(row[2])
		}
	}
	type dbEntry struct {
		Name          string `json:"name"`
		AssignedUsers string `json:"assigned_users"`
	}
	databases := []dbEntry{}
	for _, row := range rows {
		if toStringCell(row[0]) == "db" {
			name := toStringCell(row[1])
			databases = append(databases, dbEntry{Name: name, AssignedUsers: assigned[name]})
		}
	}
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"databases": databases, "total": len(databases)})
}

// apiMySQLCreateDatabase creates a new database.
func apiMySQLCreateDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	name := strings.TrimSpace(body.Name)

	if name == "" {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Database name is required."})
		return
	}
	if !validators.IsValidIdentifier(name) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Name '" + name + "' is not allowed. Use alphanumeric characters and '_' only."})
		return
	}
	if isRestrictedDatabase(name) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This database name is reserved."})
		return
	}

	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	dbLimit := 0
	if plan, planErr := a.QueryPlanDetailsByID(ctx, planID); planErr == nil {
		dbLimit = atoiDefault(plan.DBLimit, 0)
	}

	invalidateDatabaseCount(ctx, a, currentUsername)
	dbUsage := getDatabaseCount(ctx, a, currentUsername, userContext)
	if dbLimit != 0 && dbUsage >= dbLimit {
		writeAPIMySQLJSON(w, http.StatusConflict, map[string]string{"error": "Database limit reached for your hosting plan."})
		return
	}

	docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "CREATE DATABASE IF NOT EXISTS `"+name+"`", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	invalidateDatabasesInfo(ctx, a, userContext)
	invalidateDatabaseCount(ctx, a, currentUsername)

	_ = logger.RecordUserAction(a.Config, currentUsername, "created MySQL database "+name+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusCreated, map[string]string{"name": name})
}

// apiMySQLDeleteDatabase drops a database.
func apiMySQLDeleteDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbName := r.PathValue("db_name")

	if !validators.IsValidIdentifier(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid database name."})
		return
	}
	if isRestrictedDatabase(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This database cannot be deleted."})
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+dbName+"`", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	invalidateDatabasesInfo(ctx, a, userContext)
	invalidateDatabaseCount(ctx, a, currentUsername)

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted MySQL database "+dbName+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"name": dbName, "deleted": true})
}

// apiMySQLDatabaseTables lists tables with row/size info for one database.
func apiMySQLDatabaseTables(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbName := r.PathValue("db_name")
	if !validators.IsValidIdentifier(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid database name."})
		return
	}

	exists, execErr := mysqlmanager.Exec(ctx, userContext, "SELECT schema_name FROM information_schema.schemata WHERE schema_name NOT IN ("+restricted.dbsSQL+") AND schema_name = '"+dbName+"'", "")
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	if len(exists) == 0 {
		writeAPIMySQLJSON(w, http.StatusNotFound, map[string]string{"error": "Database '" + dbName + "' not found."})
		return
	}

	rows, execErr := mysqlmanager.Exec(ctx, userContext, `
		SELECT table_name, table_rows, data_length, index_length, data_free
		FROM information_schema.tables
		WHERE table_schema = DATABASE()
		ORDER BY table_name
	`, dbName)
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	type tableEntry struct {
		Table       string `json:"table"`
		Rows        int    `json:"rows"`
		DataLength  int    `json:"data_length"`
		IndexLength int    `json:"index_length"`
		DataFree    int    `json:"data_free"`
	}
	tables := make([]tableEntry, 0, len(rows))
	for _, row := range rows {
		tables = append(tables, tableEntry{
			Table: toStringCell(row[0]), Rows: mysqlmanager.ToInt(row[1]),
			DataLength: mysqlmanager.ToInt(row[2]), IndexLength: mysqlmanager.ToInt(row[3]), DataFree: mysqlmanager.ToInt(row[4]),
		})
	}
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"database": dbName, "tables": tables})
}

// apiMySQLDBMaintenance runs OPTIMIZE or REPAIR TABLE against every table in a database.
func apiMySQLDBMaintenance(a *appctx.App, w http.ResponseWriter, r *http.Request, action string) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbName := r.PathValue("db_name")
	if !validators.IsValidIdentifier(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid database name."})
		return
	}

	exists, execErr := mysqlmanager.Exec(ctx, userContext, "SELECT schema_name FROM information_schema.schemata WHERE schema_name NOT IN ("+restricted.dbsSQL+") AND schema_name = '"+dbName+"'", "")
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	if len(exists) == 0 {
		writeAPIMySQLJSON(w, http.StatusNotFound, map[string]string{"error": "Database '" + dbName + "' not found."})
		return
	}

	tableRows, execErr := mysqlmanager.Exec(ctx, userContext, "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() ORDER BY table_name", dbName)
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	sqlVerb := "REPAIR"
	if action == "optimize" {
		sqlVerb = "OPTIMIZE"
	}

	type detail struct {
		Op      string `json:"op"`
		MsgType string `json:"msg_type"`
		MsgText string `json:"msg_text"`
	}
	type result struct {
		Table   string   `json:"table"`
		Status  string   `json:"status"`
		Details []detail `json:"details,omitempty"`
		Error   string   `json:"error,omitempty"`
	}
	results := make([]result, 0, len(tableRows))
	for _, tr := range tableRows {
		table := toStringCell(tr[0])
		opRows, opErr := mysqlmanager.Exec(ctx, userContext, sqlVerb+" TABLE `"+table+"`", dbName)
		if opErr != nil {
			results = append(results, result{Table: table, Status: "error", Error: mysqlAPIError(opErr)})
			continue
		}
		details := make([]detail, 0, len(opRows))
		for _, or := range opRows {
			details = append(details, detail{Op: toStringCell(or[1]), MsgType: toStringCell(or[2]), MsgText: toStringCell(or[3])})
		}
		results = append(results, result{Table: table, Status: "ok", Details: details})
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, action+"d MySQL database "+dbName+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"database": dbName, "action": action, "results": results})
}

// apiMySQLExportDatabase streams a database dump (sql or gzip) for download.
func apiMySQLExportDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbName := r.PathValue("db_name")
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "sql"
	}

	if !validators.IsValidIdentifier(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid database name."})
		return
	}
	if isRestrictedDatabase(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This database cannot be exported."})
		return
	}
	if format != "sql" && format != "gzip" {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid format. Use 'sql' or 'gzip'."})
		return
	}

	mysqlVersion := GetMySQLVersion(ctx, a, userContext)
	dumpCmd := "mysqldump"
	if mysqlVersion == "mariadb" {
		dumpCmd = "mariadb-dump"
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, dumpCmd, "-u", "root", dbName)
	dumpOutput, dumpErr := podmanmanager.Command(ctx, userContext, argv).Output()
	if dumpErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to export database '" + dbName + "'."})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "exported MySQL database "+dbName+" via API", reqip.ClientIP(r))

	if format == "gzip" {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, _ = gz.Write(dumpOutput)
		_ = gz.Close()
		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Disposition", `attachment; filename="`+dbName+`.sql.gz"`)
		_, _ = w.Write(buf.Bytes())
		return
	}
	w.Header().Set("Content-Type", "application/sql")
	w.Header().Set("Content-Disposition", `attachment; filename="`+dbName+`.sql"`)
	_, _ = w.Write(dumpOutput)
}

// apiMySQLDatabasesSize returns per-database disk usage (data_length +
// index_length, summed across every table) in a configurable unit. This is
// the API equivalent of GET /json/mysql-size - distinct from
// GET /api/mysql/databases (names + assigned users, no size) and
// GET /api/mysql/databases/{db_name}/tables (per-table size for one db).
func apiMySQLDatabasesSize(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": `Invalid unit parameter. Use "bytes", "kb", "mb", or "gb".`})
		return
	}
	showAll := r.URL.Query().Get("show_all") != ""

	whereClause := ""
	if !showAll {
		whereClause = "WHERE table_schema NOT IN (" + restricted.dbsSQL + ") "
	}
	query := "SELECT table_schema, ROUND(SUM(data_length + index_length) / " + strconv.FormatInt(divisor, 10) + ", 2) FROM information_schema.TABLES " + whereClause + "GROUP BY table_schema"

	rows, execErr := mysqlmanager.Exec(ctx, userContext, query, "")
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	type dbSizeEntry struct {
		Database string  `json:"database"`
		Size     float64 `json:"size"`
	}
	sizes := make([]dbSizeEntry, 0, len(rows))
	for _, row := range rows {
		sizes = append(sizes, dbSizeEntry{Database: toStringCell(row[0]), Size: toFloatCell(row[1])})
	}
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"unit": strings.ToUpper(unit), "sizes": sizes, "total": len(sizes)})
}

// ── Users ────────────────────────────────────────────────────────────────

// apiMySQLListUsers lists all database users.
func apiMySQLListUsers(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rows, execErr := mysqlmanager.Exec(ctx, userContext, "SELECT User FROM mysql.user WHERE User NOT IN ("+restricted.usersSQL+") ORDER BY User", "")
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	users := make([]string, 0, len(rows))
	for _, row := range rows {
		users = append(users, toStringCell(row[0]))
	}
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"users": users, "total": len(users)})
}

// apiMySQLCreateUser creates a database user.
func apiMySQLCreateUser(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username string `json:"username"`
		Host     string `json:"host"`
		Password string `json:"password"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dbUser := strings.TrimSpace(body.Username)
	dbHost := strings.TrimSpace(body.Host)
	if dbHost == "" {
		dbHost = "%"
	}
	password := body.Password

	if dbUser == "" {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Username is required."})
		return
	}
	if !validators.IsValidIdentifier(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Username '" + dbUser + "' is not allowed. Use alphanumeric characters and '_' only."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This username is reserved."})
		return
	}
	if !validators.IsValidHost(dbHost) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid host format."})
		return
	}
	if password == "" {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Password is required."})
		return
	}

	docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "CREATE USER '"+dbUser+"'@'"+dbHost+"' IDENTIFIED BY '"+escapeMySQLString(password)+"'", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	invalidateDatabasesInfo(ctx, a, userContext)

	_ = logger.RecordUserAction(a.Config, currentUsername, "created MySQL user "+dbUser+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusCreated, map[string]string{"username": dbUser, "host": dbHost})
}

// apiMySQLDeleteUser deletes a database user.
func apiMySQLDeleteUser(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbUser := r.PathValue("db_user")
	dbHost := r.URL.Query().Get("host")
	if dbHost == "" {
		dbHost = "%"
	}

	if !validators.IsValidIdentifier(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid username."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This is a system user that cannot be deleted."})
		return
	}
	if !validators.IsValidHost(dbHost) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid host format."})
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'"+dbHost+"'", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	invalidateDatabasesInfo(ctx, a, userContext)

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted MySQL user "+dbUser+"@"+dbHost+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"username": dbUser, "host": dbHost, "deleted": true})
}

// apiMySQLChangeUserPassword changes a database user's password.
func apiMySQLChangeUserPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbUser := r.PathValue("db_user")

	var body struct {
		Host     string `json:"host"`
		Password string `json:"password"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dbHost := strings.TrimSpace(body.Host)
	if dbHost == "" {
		dbHost = "%"
	}
	newPassword := body.Password

	if !validators.IsValidIdentifier(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid username."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This system user's password cannot be changed."})
		return
	}
	if !validators.IsValidHost(dbHost) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid host format."})
		return
	}
	if newPassword == "" {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "New password is required."})
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "ALTER USER '"+dbUser+"'@'"+dbHost+"' IDENTIFIED BY '"+escapeMySQLString(newPassword)+"'", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	mysqlmanager.InvalidatePool(userContext)

	_ = logger.RecordUserAction(a.Config, currentUsername, "changed MySQL user "+dbUser+" password via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"username": dbUser, "host": dbHost, "updated": true})
}

// ── Grants ───────────────────────────────────────────────────────────────

// apiMySQLUserPrivileges returns a user's privileges on a database.
func apiMySQLUserPrivileges(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbUser := r.PathValue("db_user")
	dbName := r.PathValue("db_name")
	dbHost := r.URL.Query().Get("host")
	if dbHost == "" {
		dbHost = "%"
	}

	if !validators.IsValidIdentifier(dbUser) || !validators.IsValidIdentifier(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid username or database name."})
		return
	}

	rows, execErr := mysqlmanager.Exec(ctx, userContext, "SHOW GRANTS FOR '"+dbUser+"'@'"+dbHost+"'", "")
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	privilegesSet := map[string]bool{}
	needle := "`" + dbName + "`."
	for _, row := range rows {
		line := toStringCell(row[0])
		if !strings.Contains(line, needle) && !strings.HasPrefix(line, "GRANT ALL PRIVILEGES ON *.*") {
			continue
		}
		onIdx := strings.Index(line, " ON ")
		if onIdx == -1 {
			continue
		}
		privPart := strings.TrimPrefix(line[:onIdx], "GRANT ")
		for _, p := range strings.Split(privPart, ",") {
			privilegesSet[strings.TrimSpace(p)] = true
		}
	}
	if privilegesSet["ALL PRIVILEGES"] {
		privilegesSet = map[string]bool{"ALL PRIVILEGES": true}
	}
	privileges := make([]string, 0, len(privilegesSet))
	for p := range privilegesSet {
		privileges = append(privileges, p)
	}

	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{
		"username": dbUser, "host": dbHost, "database": dbName, "privileges": privileges,
	})
}

// apiMySQLGrant grants privileges to a user on a database.
func apiMySQLGrant(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username   string   `json:"username"`
		Host       string   `json:"host"`
		Database   string   `json:"database"`
		Privileges []string `json:"privileges"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dbUser := strings.TrimSpace(body.Username)
	dbHost := strings.TrimSpace(body.Host)
	if dbHost == "" {
		dbHost = "%"
	}
	dbName := strings.TrimSpace(body.Database)

	if dbUser == "" || !validators.IsValidIdentifier(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid username is required."})
		return
	}
	if dbName == "" || !validators.IsValidIdentifier(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid database name is required."})
		return
	}
	if !validators.IsValidHost(dbHost) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid host format."})
		return
	}
	if isRestrictedDatabase(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This is a system database."})
		return
	}
	if len(body.Privileges) == 0 {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "At least one privilege is required."})
		return
	}

	privilegesSQL := strings.Join(body.Privileges, ", ")
	for _, p := range body.Privileges {
		if p == "ALL PRIVILEGES" {
			privilegesSQL = "ALL PRIVILEGES"
			break
		}
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "REVOKE ALL PRIVILEGES ON `"+dbName+"`.* FROM '"+dbUser+"'@'"+dbHost+"'", ""); execErr != nil {
		// MySQL error 1141 ("no grants for user") is swallowed - there's
		// simply nothing to revoke yet.
		if !strings.Contains(execErr.Error(), "1141") {
			writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
			return
		}
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "GRANT "+privilegesSQL+" ON `"+dbName+"`.* TO '"+dbUser+"'@'"+dbHost+"'", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "granted "+privilegesSQL+" on "+dbName+" to MySQL user "+dbUser+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusCreated, map[string]string{"username": dbUser, "host": dbHost, "database": dbName, "privileges": privilegesSQL})
}

// apiMySQLRevoke revokes all privileges from a user on a database.
func apiMySQLRevoke(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username string `json:"username"`
		Host     string `json:"host"`
		Database string `json:"database"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dbUser := strings.TrimSpace(body.Username)
	dbHost := strings.TrimSpace(body.Host)
	if dbHost == "" {
		dbHost = "%"
	}
	dbName := strings.TrimSpace(body.Database)

	if dbUser == "" || !validators.IsValidIdentifier(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid username is required."})
		return
	}
	if dbName == "" || !validators.IsValidIdentifier(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid database name is required."})
		return
	}
	if isRestrictedDatabase(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This is a system database."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This is a system user."})
		return
	}
	if !validators.IsValidHost(dbHost) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid host format."})
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "REVOKE ALL PRIVILEGES ON `"+dbName+"`.* FROM '"+dbUser+"'@'"+dbHost+"'", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "revoked MySQL user "+dbUser+" from database "+dbName+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"username": dbUser, "host": dbHost, "database": dbName, "revoked": true})
}

// ── Info & processlist ───────────────────────────────────────────────────

// apiMySQLInfo returns databases, users, and assigned-databases summary info.
func apiMySQLInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rows, execErr := mysqlmanager.Exec(ctx, userContext, `
		SELECT 'db' AS type, schema_name AS col1, NULL AS col2
			FROM information_schema.schemata
			WHERE schema_name NOT IN (`+restricted.dbsSQL+`)
		UNION ALL
		SELECT 'user', User, NULL
			FROM mysql.user
			WHERE User NOT IN (`+restricted.usersSQL+`)
		UNION ALL
		SELECT 'assigned', Db, GROUP_CONCAT(User SEPARATOR ', ')
			FROM mysql.db
			WHERE User NOT LIKE 'mysql.s%' AND User NOT IN (`+restricted.usersSQL+`)
			GROUP BY Db
	`, "")
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	type assignedEntry struct {
		Database string `json:"database"`
		Users    string `json:"users"`
	}
	databases := []string{}
	users := []string{}
	assigned := []assignedEntry{}
	for _, row := range rows {
		switch toStringCell(row[0]) {
		case "db":
			databases = append(databases, toStringCell(row[1]))
		case "user":
			users = append(users, toStringCell(row[1]))
		case "assigned":
			assigned = append(assigned, assignedEntry{Database: toStringCell(row[1]), Users: toStringCell(row[2])})
		}
	}
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"databases": databases, "users": users, "assigned_databases": assigned})
}

// apiMySQLProcesslist returns the full MySQL processlist.
func apiMySQLProcesslist(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rows, execErr := mysqlmanager.Exec(ctx, userContext, "SHOW FULL PROCESSLIST", "")
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	type procEntry struct {
		ID      string `json:"id"`
		User    string `json:"user"`
		Host    string `json:"host"`
		DB      string `json:"db"`
		Command string `json:"command"`
		Time    string `json:"time"`
		State   string `json:"state"`
		Info    string `json:"info"`
	}
	processlist := make([]procEntry, 0, len(rows))
	for _, row := range rows {
		processlist = append(processlist, procEntry{
			ID: toStringCell(row[0]), User: toStringCell(row[1]), Host: toStringCell(row[2]), DB: toStringCell(row[3]),
			Command: toStringCell(row[4]), Time: toStringCell(row[5]), State: toStringCell(row[6]), Info: toStringCell(row[7]),
		})
	}
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"processlist": processlist, "total": len(processlist)})
}

// ── Remote access ────────────────────────────────────────────────────────

// apiMySQLRemoteAccessStatus returns remote-access status, port, and per-user allowed hosts.
func apiMySQLRemoteAccessStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rawPort := webserver.GetEnvFileValue(userContext, "MYSQL_PORT")
	m := mysqlPortRE.FindStringSubmatch(rawPort)
	if m == nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not parse MYSQL_PORT value."})
		return
	}
	enabled := !strings.Contains(rawPort, "127.0.0.1")
	port := m[2]
	serverIP := a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername)

	type userAccessEntry struct {
		Username string   `json:"username"`
		Hosts    []string `json:"hosts"`
	}
	userAccess := []userAccessEntry{}
	if enabled {
		rows, execErr := mysqlmanager.Exec(ctx, userContext, "SELECT User, Host FROM mysql.user WHERE User NOT IN ("+restricted.usersSQL+") ORDER BY User, Host", "")
		if execErr != nil {
			writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
			return
		}
		var order []string
		grouped := map[string][]string{}
		for _, row := range rows {
			u, h := toStringCell(row[0]), toStringCell(row[1])
			if _, ok := grouped[u]; !ok {
				order = append(order, u)
			}
			grouped[u] = append(grouped[u], h)
		}
		for _, u := range order {
			userAccess = append(userAccess, userAccessEntry{Username: u, Hosts: grouped[u]})
		}
	}

	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{
		"enabled": enabled, "server_ip": serverIP, "port": port, "user_access": userAccess,
	})
}

// apiMySQLRemoteAccessToggle enables or disables remote access.
func apiMySQLRemoteAccessToggle(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Action string `json:"action"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	action := strings.TrimSpace(body.Action)
	if action != "enable" && action != "disable" {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be 'enable' or 'disable'."})
		return
	}

	rawPort := webserver.GetEnvFileValue(userContext, "MYSQL_PORT")
	m := mysqlPortRE.FindStringSubmatch(rawPort)
	if m == nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not parse MYSQL_PORT value."})
		return
	}
	port := m[2]

	var enabled bool
	if action == "enable" {
		docker.SetEnvValue(userContext, "MYSQL_PORT", port+":3306")
		enabled = true
	} else {
		if strings.Contains(rawPort, "127.0.0.1") {
			writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"enabled": false, "message": "Remote access is already disabled."})
			return
		}
		docker.SetEnvValue(userContext, "MYSQL_PORT", "127.0.0.1:"+port+":3306")
		enabled = false
	}

	docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")
	action2 := "disabled"
	if enabled {
		action2 = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, action2+" remote MySQL via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"enabled": enabled})
}

// apiMySQLRemoteAccessAdd grants a user remote access from a host.
func apiMySQLRemoteAccessAdd(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username string `json:"username"`
		Host     string `json:"host"`
		Password string `json:"password"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dbUser := strings.TrimSpace(body.Username)
	dbHost := strings.TrimSpace(body.Host)
	password := body.Password

	if dbUser == "" || !validators.IsValidIdentifier(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid username is required."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This username is reserved."})
		return
	}
	if dbHost == "" || !validators.IsValidHost(dbHost) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid host format."})
		return
	}
	if password == "" {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Password is required."})
		return
	}

	existing, execErr := mysqlmanager.Exec(ctx, userContext, "SELECT Host FROM mysql.user WHERE User = '"+dbUser+"'", "")
	if execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	existingHosts := make([]string, 0, len(existing))
	for _, row := range existing {
		existingHosts = append(existingHosts, toStringCell(row[0]))
	}
	for _, h := range existingHosts {
		if h == dbHost {
			writeAPIMySQLJSON(w, http.StatusConflict, map[string]string{"error": "User '" + dbUser + "' already has access from '" + dbHost + "'."})
			return
		}
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "CREATE USER '"+dbUser+"'@'"+dbHost+"' IDENTIFIED BY '"+escapeMySQLString(password)+"'", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	if len(existingHosts) > 0 {
		cloneMySQLGrants(ctx, userContext, dbUser, existingHosts[0], dbHost)
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "granted MySQL remote access for "+dbUser+"@"+dbHost+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusCreated, map[string]string{"username": dbUser, "host": dbHost})
}

// apiMySQLRemoteAccessEdit changes the host a remote-access entry is valid from.
func apiMySQLRemoteAccessEdit(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username string `json:"username"`
		OldHost  string `json:"old_host"`
		NewHost  string `json:"new_host"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dbUser := strings.TrimSpace(body.Username)
	oldHost := strings.TrimSpace(body.OldHost)
	newHost := strings.TrimSpace(body.NewHost)

	if dbUser == "" || !validators.IsValidIdentifier(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid username is required."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "System users cannot be edited."})
		return
	}
	if !validators.IsValidHost(oldHost) || !validators.IsValidHost(newHost) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid host format."})
		return
	}
	if oldHost == newHost {
		writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"username": dbUser, "host": newHost, "updated": false, "message": "No changes made."})
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "RENAME USER '"+dbUser+"'@'"+oldHost+"' TO '"+dbUser+"'@'"+newHost+"'", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "changed MySQL remote access host for "+dbUser+" from "+oldHost+" to "+newHost+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"username": dbUser, "old_host": oldHost, "host": newHost, "updated": true})
}

// apiMySQLRemoteAccessDelete removes a remote-access entry.
func apiMySQLRemoteAccessDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username string `json:"username"`
		Host     string `json:"host"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dbUser := strings.TrimSpace(body.Username)
	dbHost := strings.TrimSpace(body.Host)

	if dbUser == "" || !validators.IsValidIdentifier(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid username is required."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "System users cannot be deleted."})
		return
	}
	if !validators.IsValidHost(dbHost) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid host format."})
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'"+dbHost+"'", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "removed MySQL remote access for "+dbUser+"@"+dbHost+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"username": dbUser, "host": dbHost, "deleted": true})
}

// ── Configuration ────────────────────────────────────────────────────────

// apiMySQLGetConfig returns the MySQL server configuration.
func apiMySQLGetConfig(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	content := readMySQLConfigFile(userContext)
	currentConfig := configEntriesToMap(parseMySQLConfigContent(content))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"configuration": currentConfig, "available_keys": availableConfKeys})
}

// apiMySQLUpdateConfig updates the MySQL server configuration and restarts the service.
func apiMySQLUpdateConfig(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var raw map[string]json.RawMessage
	_ = json.NewDecoder(r.Body).Decode(&raw)

	allowedKeys := map[string]bool{}
	for _, k := range availableConfKeys {
		allowedKeys[k] = true
	}

	newConfig := map[string]string{}
	for k, v := range raw {
		if !allowedKeys[k] {
			continue
		}
		var s string
		if json.Unmarshal(v, &s) == nil {
			newConfig[k] = s
			continue
		}
		newConfig[k] = strings.Trim(string(v), `"`)
	}
	if len(newConfig) == 0 {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]any{"error": "No valid configuration keys provided.", "available_keys": availableConfKeys})
		return
	}

	updateMySQLConfigFile(userContext, newConfig, availableConfKeys)
	_ = logger.RecordUserAction(a.Config, currentUsername, "updated MySQL configuration via API", reqip.ClientIP(r))

	mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
	restarted := false
	if mysqlVersion == "mysql" || mysqlVersion == "mariadb" {
		argv := podmanmanager.PodmanArgv(userContext, "restart", mysqlVersion)
		if runErr := exec.CommandContext(ctx, argv[0], argv[1:]...).Run(); runErr == nil {
			restarted = true
		}
	}

	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"updated": true, "restarted": restarted, "configuration": newConfig})
}
