package postgresql

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/lib/pq"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterAPI wires the postgresql API routes onto mux. The
// /databases/{db_name} and /users/{db_user} sub-resources share their
// prefix with a literal suffix - Go's http.ServeMux requires a "{...}"
// wildcard to be the final segment, so GET/POST get a "{rest...}"
// catch-all where needed and the dispatch funcs below strip the known
// suffix by hand to resolve the real route.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "postgresql", "GET /api/postgresql/databases", func(w http.ResponseWriter, r *http.Request) { apiPsqlListDatabases(a, w, r) })
	apiregistry.Handle(mux, a, "postgresql", "POST /api/postgresql/databases", func(w http.ResponseWriter, r *http.Request) { apiPsqlCreateDatabase(a, w, r) })
	apiregistry.Handle(mux, a, "postgresql", "DELETE /api/postgresql/databases/{db_name}", func(w http.ResponseWriter, r *http.Request) { apiPsqlDeleteDatabase(a, w, r) })

	apiregistry.Add("GET /api/postgresql/databases/{db_name}/export")
	mux.Handle("GET /api/postgresql/databases/{rest...}", auth.RequireAPI(a, "postgresql")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiPsqlDatabasesGetDispatch(a, w, r) })))

	apiregistry.Add("POST /api/postgresql/databases/{db_name}/import")
	mux.Handle("POST /api/postgresql/databases/{rest...}", auth.RequireAPI(a, "postgresql")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { apiPsqlDatabasesPostDispatch(a, w, r) })))

	apiregistry.Handle(mux, a, "postgresql", "GET /api/postgresql/users", func(w http.ResponseWriter, r *http.Request) { apiPsqlListUsers(a, w, r) })
	apiregistry.Handle(mux, a, "postgresql", "POST /api/postgresql/users", func(w http.ResponseWriter, r *http.Request) { apiPsqlCreateUser(a, w, r) })
	apiregistry.Handle(mux, a, "postgresql", "DELETE /api/postgresql/users/{db_user}", func(w http.ResponseWriter, r *http.Request) { apiPsqlDeleteUser(a, w, r) })
	apiregistry.Handle(mux, a, "postgresql", "PATCH /api/postgresql/users/{db_user}/password", func(w http.ResponseWriter, r *http.Request) { apiPsqlChangeUserPassword(a, w, r) })

	apiregistry.Handle(mux, a, "postgresql", "POST /api/postgresql/grants", func(w http.ResponseWriter, r *http.Request) { apiPsqlGrant(a, w, r) })
	apiregistry.Handle(mux, a, "postgresql", "DELETE /api/postgresql/grants", func(w http.ResponseWriter, r *http.Request) { apiPsqlRevoke(a, w, r) })

	apiregistry.Handle(mux, a, "postgresql", "GET /api/postgresql/info", func(w http.ResponseWriter, r *http.Request) { apiPsqlInfo(a, w, r) })
	apiregistry.Handle(mux, a, "postgresql", "GET /api/postgresql/processlist", func(w http.ResponseWriter, r *http.Request) { apiPsqlProcesslist(a, w, r) })

	apiregistry.Handle(mux, a, "postgresql", "GET /api/postgresql/remote-access", func(w http.ResponseWriter, r *http.Request) { apiPsqlRemoteAccessStatus(a, w, r) })
	apiregistry.Handle(mux, a, "postgresql", "POST /api/postgresql/remote-access", func(w http.ResponseWriter, r *http.Request) { apiPsqlRemoteAccessToggle(a, w, r) })

	apiregistry.Handle(mux, a, "postgresql", "GET /api/postgresql/configuration", func(w http.ResponseWriter, r *http.Request) { apiPsqlGetConfig(a, w, r) })
	apiregistry.Handle(mux, a, "postgresql", "PUT /api/postgresql/configuration", func(w http.ResponseWriter, r *http.Request) { apiPsqlUpdateConfig(a, w, r) })
}

func writeAPIPsqlJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func apiPsqlDatabasesGetDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	if strings.HasSuffix(rest, "/export") {
		r.SetPathValue("db_name", strings.TrimSuffix(rest, "/export"))
		apiPsqlExportDatabase(a, w, r)
		return
	}
	http.NotFound(w, r)
}

func apiPsqlDatabasesPostDispatch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	rest := r.PathValue("rest")
	if strings.HasSuffix(rest, "/import") {
		r.SetPathValue("db_name", strings.TrimSuffix(rest, "/import"))
		apiPsqlImportDatabase(a, w, r)
		return
	}
	http.NotFound(w, r)
}

// ── Databases ────────────────────────────────────────────────────────────

// apiPsqlListDatabases returns every non-system database along with which
// users have been granted access to each.
func apiPsqlListDatabases(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	dbRows, execErr := postgresmanager.Exec(ctx, userContext, `
		SELECT datname FROM pg_database
		WHERE datname NOT IN ('postgres','template0','template1')
		ORDER BY datname
	`, "postgres")
	if execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}
	assignedRows, execErr := postgresmanager.Exec(ctx, userContext, `
		SELECT
			d.datname,
			COALESCE(
				STRING_AGG(split_part(acl_item::text, '=', 1), ', '
					ORDER BY split_part(acl_item::text, '=', 1)),
				''
			) AS assigned_users
		FROM pg_database d
		LEFT JOIN unnest(d.datacl) AS acl_item ON TRUE
		WHERE d.datistemplate = false
		  AND acl_item IS NOT NULL
		  AND split_part(acl_item::text, '=', 1) <> ''
		GROUP BY d.datname
		ORDER BY d.datname
	`, "postgres")
	if execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	assigned := map[string]string{}
	for _, row := range assignedRows {
		assigned[toStringCell(row[0])] = toStringCell(row[1])
	}
	type dbEntry struct {
		Name          string `json:"name"`
		AssignedUsers string `json:"assigned_users"`
	}
	databases := make([]dbEntry, 0, len(dbRows))
	for _, row := range dbRows {
		name := toStringCell(row[0])
		databases = append(databases, dbEntry{Name: name, AssignedUsers: assigned[name]})
	}
	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"databases": databases, "total": len(databases)})
}

// apiPsqlCreateDatabase creates a new database, enforcing the plan's
// database limit first.
func apiPsqlCreateDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Database name is required."})
		return
	}
	if !validators.IsValidIdentifier(name) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Name '" + name + "' is not allowed. Use alphanumeric characters and '_' only."})
		return
	}
	if isRestrictedDatabase(name) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "This database name is reserved."})
		return
	}

	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	dbLimit := 1000000
	if plan, planErr := a.QueryPlanDetailsByID(ctx, planID); planErr == nil {
		if v := atoiDefault(plan.DBLimit, 0); v != 0 {
			dbLimit = v
		}
	}

	rows, countErr := postgresmanager.Exec(ctx, userContext, "SELECT COUNT(*) FROM pg_database WHERE datname NOT IN ('postgres', 'template0', 'template1')", "postgres")
	if countErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": countErr.Error()})
		return
	}
	dbUsage := 0
	if len(rows) > 0 {
		dbUsage = postgresmanager.ToInt(rows[0][0])
	}
	if dbUsage >= dbLimit {
		writeAPIPsqlJSON(w, http.StatusConflict, map[string]string{"error": "Database limit reached for your hosting plan."})
		return
	}

	docker.StartComposeServiceIfNotRunning(ctx, userContext, "postgres")

	if _, execErr := postgresmanager.Exec(ctx, userContext, `CREATE DATABASE "`+name+`"`, "postgres"); execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "created PostgreSQL database "+name+" via API", reqip.ClientIP(r))
	writeAPIPsqlJSON(w, http.StatusCreated, map[string]string{"name": name})
}

// apiPsqlDeleteDatabase drops a database.
func apiPsqlDeleteDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbName := r.PathValue("db_name")

	if !validators.IsValidIdentifier(dbName) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid database name."})
		return
	}
	if isRestrictedDatabase(dbName) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "This database cannot be deleted."})
		return
	}

	postgresmanager.InvalidatePool(userContext, dbName)

	if _, execErr := postgresmanager.Exec(ctx, userContext, `DROP DATABASE IF EXISTS "`+dbName+`"`, "postgres"); execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted PostgreSQL database "+dbName+" via API", reqip.ClientIP(r))
	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"name": dbName, "deleted": true})
}

// apiPsqlExportDatabase streams a `pg_dump` of the database as a
// downloadable .sql file.
func apiPsqlExportDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbName := r.PathValue("db_name")

	if !validators.IsValidIdentifier(dbName) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid database name."})
		return
	}
	if isRestrictedDatabase(dbName) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "This database cannot be exported."})
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", "postgres", "pg_dump", "-U", "postgres", dbName)
	dumpOutput, dumpErr := podmanmanager.Command(ctx, userContext, argv).Output()
	if dumpErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to export database '" + dbName + "'."})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "exported PostgreSQL database "+dbName+" via API", reqip.ClientIP(r))
	w.Header().Set("Content-Type", "application/sql")
	w.Header().Set("Content-Disposition", `attachment; filename="`+dbName+`.sql"`)
	_, _ = w.Write(dumpOutput)
}

// apiPsqlImportDatabase imports an uploaded .sql file into a database.
func apiPsqlImportDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbName := r.PathValue("db_name")

	if !validators.IsValidIdentifier(dbName) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid database name."})
		return
	}

	if mpErr := r.ParseMultipartForm(1 << 30); mpErr != nil {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "SQL file required (multipart field: 'file')."})
		return
	}
	file, header, fileErr := r.FormFile("file")
	if fileErr != nil || header.Filename == "" {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "SQL file required (multipart field: 'file')."})
		return
	}
	defer file.Close()

	targetDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_mysql_dumps/_data/"
	if mkErr := os.MkdirAll(targetDir, 0o755); mkErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
		return
	}
	tempPath := targetDir + header.Filename

	out, createErr := os.Create(tempPath)
	if createErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": createErr.Error()})
		return
	}
	if _, copyErr := out.ReadFrom(file); copyErr != nil {
		_ = out.Close()
		_ = os.Remove(tempPath)
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": copyErr.Error()})
		return
	}
	_ = out.Close()
	defer os.Remove(tempPath)

	containerPath := "/tmp/" + header.Filename
	cpArgv := podmanmanager.PodmanArgv(userContext, "cp", tempPath, "postgres:"+containerPath)
	if cpErr := podmanmanager.Command(ctx, userContext, cpArgv).Run(); cpErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": "Import failed: " + cpErr.Error()})
		return
	}
	execArgv := podmanmanager.PodmanArgv(userContext, "exec", "-i", "postgres", "psql", "-U", "postgres", "-d", dbName, "-f", containerPath)
	if execErr := podmanmanager.Command(ctx, userContext, execArgv).Run(); execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": "Import failed: " + execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "imported "+header.Filename+" into PostgreSQL database "+dbName+" via API", reqip.ClientIP(r))
	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"database": dbName, "file": header.Filename, "imported": true})
}

// ── Users ────────────────────────────────────────────────────────────────

// apiPsqlListUsers returns every non-system PostgreSQL role.
func apiPsqlListUsers(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rows, execErr := postgresmanager.Exec(ctx, userContext, "SELECT rolname FROM pg_roles WHERE rolname NOT IN ("+filteredRolesSQL()+") ORDER BY rolname", "postgres")
	if execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}
	users := make([]string, 0, len(rows))
	for _, row := range rows {
		users = append(users, toStringCell(row[0]))
	}
	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"users": users, "total": len(users)})
}

// apiPsqlCreateUser creates a new PostgreSQL database user.
func apiPsqlCreateUser(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dbUser := strings.TrimSpace(body.Username)
	password := body.Password

	if dbUser == "" {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Username is required."})
		return
	}
	if !validators.IsValidIdentifier(dbUser) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Username '" + dbUser + "' is not allowed. Use alphanumeric characters and '_' only."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "This username is reserved."})
		return
	}
	if password == "" {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Password is required."})
		return
	}

	if _, execErr := postgresmanager.Exec(ctx, userContext, `CREATE USER "`+dbUser+`" WITH PASSWORD `+pq.QuoteLiteral(password), "postgres"); execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "created PostgreSQL user "+dbUser+" via API", reqip.ClientIP(r))
	writeAPIPsqlJSON(w, http.StatusCreated, map[string]string{"username": dbUser})
}

// apiPsqlDeleteUser revokes a user's privileges on every database and
// drops the role.
func apiPsqlDeleteUser(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbUser := r.PathValue("db_user")

	if !validators.IsValidIdentifier(dbUser) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid username."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "This is a system user that cannot be deleted."})
		return
	}

	dbRows, execErr := postgresmanager.Exec(ctx, userContext, "SELECT datname FROM pg_database WHERE datistemplate = false", "postgres")
	if execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}
	for _, row := range dbRows {
		db := toStringCell(row[0])
		_, _ = postgresmanager.Exec(ctx, userContext, `REVOKE ALL PRIVILEGES ON DATABASE "`+db+`" FROM "`+dbUser+`"`, "postgres")
		_, _ = postgresmanager.Exec(ctx, userContext, `REVOKE ALL PRIVILEGES ON SCHEMA public FROM "`+dbUser+`"`, db)
	}

	if _, execErr := postgresmanager.Exec(ctx, userContext, `DROP ROLE IF EXISTS "`+dbUser+`"`, "postgres"); execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted PostgreSQL user "+dbUser+" via API", reqip.ClientIP(r))
	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"username": dbUser, "deleted": true})
}

// apiPsqlChangeUserPassword updates an existing user's password.
func apiPsqlChangeUserPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbUser := r.PathValue("db_user")

	var body struct {
		Password string `json:"password"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	newPassword := body.Password

	if !validators.IsValidIdentifier(dbUser) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid username."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "This system user's password cannot be changed."})
		return
	}
	if newPassword == "" {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "New password is required."})
		return
	}

	if _, execErr := postgresmanager.Exec(ctx, userContext, `ALTER USER "`+dbUser+`" WITH PASSWORD `+pq.QuoteLiteral(newPassword), "postgres"); execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}
	postgresmanager.InvalidatePool(userContext, "")

	_ = logger.RecordUserAction(a.Config, currentUsername, "changed PostgreSQL user "+dbUser+" password via API", reqip.ClientIP(r))
	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"username": dbUser, "updated": true})
}

// ── Grants ───────────────────────────────────────────────────────────────

// apiPsqlGrant grants a user full privileges on a database plus USAGE/
// CREATE on its public schema.
func apiPsqlGrant(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username string `json:"username"`
		Database string `json:"database"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dbUser := strings.TrimSpace(body.Username)
	dbName := strings.TrimSpace(body.Database)

	if dbUser == "" || !validators.IsValidIdentifier(dbUser) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid username is required."})
		return
	}
	if dbName == "" || !validators.IsValidIdentifier(dbName) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid database name is required."})
		return
	}
	if isRestrictedDatabase(dbName) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "This is a system database."})
		return
	}

	grantSQL := `GRANT ALL PRIVILEGES ON DATABASE "` + dbName + `" TO "` + dbUser + `"; ` +
		`GRANT USAGE ON SCHEMA public TO "` + dbUser + `"; ` +
		`GRANT CREATE ON SCHEMA public TO "` + dbUser + `";`
	if _, execErr := postgresmanager.Exec(ctx, userContext, grantSQL, dbName); execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "granted all privileges on PostgreSQL database "+dbName+" to "+dbUser+" via API", reqip.ClientIP(r))
	writeAPIPsqlJSON(w, http.StatusCreated, map[string]any{"username": dbUser, "database": dbName, "granted": true})
}

// apiPsqlRevoke revokes a user's privileges on a database.
func apiPsqlRevoke(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Username string `json:"username"`
		Database string `json:"database"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	dbUser := strings.TrimSpace(body.Username)
	dbName := strings.TrimSpace(body.Database)

	if dbUser == "" || !validators.IsValidIdentifier(dbUser) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid username is required."})
		return
	}
	if dbName == "" || !validators.IsValidIdentifier(dbName) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "Valid database name is required."})
		return
	}
	if isRestrictedDatabase(dbName) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "This is a system database."})
		return
	}
	if isRestrictedUser(dbUser) {
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "This is a system user."})
		return
	}

	if _, execErr := postgresmanager.Exec(ctx, userContext, `REVOKE ALL PRIVILEGES ON DATABASE "`+dbName+`" FROM "`+dbUser+`"`, "postgres"); execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "revoked PostgreSQL user "+dbUser+" from database "+dbName+" via API", reqip.ClientIP(r))
	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"username": dbUser, "database": dbName, "revoked": true})
}

// ── Info & processlist ───────────────────────────────────────────────────

// apiPsqlInfo returns the combined databases/users/assigned-databases
// payload used to populate client-side selects.
func apiPsqlInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	dbRows, execErr := postgresmanager.Exec(ctx, userContext, `
		SELECT datname FROM pg_database
		WHERE datname NOT IN ('postgres','template0','template1')
		ORDER BY datname
	`, "postgres")
	if execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}
	userRows, execErr := postgresmanager.Exec(ctx, userContext, `
		SELECT rolname FROM pg_roles
		WHERE rolname NOT IN (`+filteredRolesSQL()+`)
		ORDER BY rolname
	`, "postgres")
	if execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}
	assignedRows, execErr := postgresmanager.Exec(ctx, userContext, `
		SELECT d.datname, STRING_AGG(r.rolname, ', ')
		FROM pg_database d
		JOIN pg_roles r ON has_database_privilege(r.rolname, d.datname, 'CONNECT')
		WHERE r.rolname NOT LIKE 'pg_%'
		  AND r.rolname NOT IN (`+filteredRolesSQL()+`)
		GROUP BY d.datname
		ORDER BY d.datname
	`, "postgres")
	if execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	databases := make([]string, 0, len(dbRows))
	for _, row := range dbRows {
		databases = append(databases, toStringCell(row[0]))
	}
	users := make([]string, 0, len(userRows))
	for _, row := range userRows {
		users = append(users, toStringCell(row[0]))
	}
	type assignedEntry struct {
		Database string `json:"database"`
		Users    string `json:"users"`
	}
	assigned := make([]assignedEntry, 0, len(assignedRows))
	for _, row := range assignedRows {
		assigned = append(assigned, assignedEntry{Database: toStringCell(row[0]), Users: toStringCell(row[1])})
	}

	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"databases": databases, "users": users, "assigned_databases": assigned})
}

// apiPsqlProcesslist returns the current pg_stat_activity rows, excluding
// the connection running the query itself.
func apiPsqlProcesslist(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rows, execErr := postgresmanager.Exec(ctx, userContext, `
		SELECT pid, usename, application_name, client_addr, state,
		       wait_event_type, wait_event, query, backend_type,
		       now() - xact_start AS duration
		FROM pg_stat_activity
		WHERE pid <> pg_backend_pid()
		ORDER BY pid
	`, "postgres")
	if execErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
		return
	}

	type procEntry struct {
		PID           int    `json:"pid"`
		User          string `json:"user"`
		Application   string `json:"application"`
		ClientAddr    string `json:"client_addr"`
		State         string `json:"state"`
		WaitEventType string `json:"wait_event_type"`
		WaitEvent     string `json:"wait_event"`
		Query         string `json:"query"`
		BackendType   string `json:"backend_type"`
		Duration      string `json:"duration"`
	}
	processlist := make([]procEntry, 0, len(rows))
	for _, row := range rows {
		processlist = append(processlist, procEntry{
			PID: postgresmanager.ToInt(row[0]), User: genericCellString(row[1]), Application: genericCellString(row[2]),
			ClientAddr: genericCellString(row[3]), State: genericCellString(row[4]), WaitEventType: genericCellString(row[5]),
			WaitEvent: genericCellString(row[6]), Query: genericCellString(row[7]), BackendType: genericCellString(row[8]),
			Duration: genericCellString(row[9]),
		})
	}
	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"processlist": processlist, "total": len(processlist)})
}

// ── Remote access ────────────────────────────────────────────────────────

// apiPsqlRemoteAccessStatus reports whether PostgreSQL's port is currently
// exposed for remote access.
func apiPsqlRemoteAccessStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rawPort := webserver.GetEnvFileValue(userContext, "POSTGRES_PORT")
	port := rawPort
	if parts := strings.Split(rawPort, ":"); len(parts) >= 2 {
		port = parts[len(parts)-2]
	}
	enabled := !strings.Contains(rawPort, "127.0.0.1")
	serverIP := a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername)

	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{
		"enabled": enabled, "server_ip": serverIP, "port": port, "postgres_port": 5432,
	})
}

// apiPsqlRemoteAccessToggle enables or disables remote access by rebinding
// PostgreSQL's exposed port between 0.0.0.0 and 127.0.0.1.
func apiPsqlRemoteAccessToggle(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be 'enable' or 'disable'."})
		return
	}

	rawPort := webserver.GetEnvFileValue(userContext, "POSTGRES_PORT")
	port := rawPort
	if parts := strings.Split(rawPort, ":"); len(parts) >= 2 {
		port = parts[len(parts)-2]
	}

	var enabled bool
	if action == "enable" {
		docker.SetEnvValue(userContext, "POSTGRES_PORT", port+":5432")
		enabled = true
	} else {
		if strings.Contains(rawPort, "127.0.0.1") {
			writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"enabled": false, "message": "Remote access is already disabled."})
			return
		}
		docker.SetEnvValue(userContext, "POSTGRES_PORT", "127.0.0.1:"+port+":5432")
		enabled = false
	}

	if result := docker.RestartContainer(ctx, userContext, "postgres"); !result.Success {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": "Port changed but the PostgreSQL service failed to restart. Try restarting it manually from Services."})
		return
	}
	action2 := "disabled"
	if enabled {
		action2 = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, action2+" remote PostgreSQL via API", reqip.ClientIP(r))
	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"enabled": enabled})
}

// ── Configuration ────────────────────────────────────────────────────────

// apiPsqlGetConfig returns the current custom PostgreSQL config values
// plus the set of keys allowed to be edited.
func apiPsqlGetConfig(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !docker.IsServiceRunning(ctx, userContext, "postgres") {
		writeAPIPsqlJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "PostgreSQL container is not running."})
		return
	}

	content, readErr := readPostgresConfigFile(userContext)
	if readErr != nil {
		writeAPIPsqlJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}
	currentConfig := parsePostgresConfigContent(content)
	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"configuration": currentConfig, "available_keys": availableConfKeys})
}

// apiPsqlUpdateConfig writes submitted config values (filtered to the
// allowed key set) and restarts the postgres container to apply them.
func apiPsqlUpdateConfig(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		writeAPIPsqlJSON(w, http.StatusBadRequest, map[string]any{"error": "No valid configuration keys provided.", "available_keys": availableConfKeys})
		return
	}

	updatePostgresConfigFile(userContext, newConfig, availableConfKeys)
	_ = logger.RecordUserAction(a.Config, currentUsername, "updated PostgreSQL configuration via API", reqip.ClientIP(r))

	argv := podmanmanager.PodmanArgv(userContext, "restart", "postgres")
	restarted := false
	if runErr := exec.CommandContext(ctx, argv[0], argv[1:]...).Run(); runErr == nil {
		restarted = true
	}

	writeAPIPsqlJSON(w, http.StatusOK, map[string]any{"updated": true, "restarted": restarted, "configuration": newConfig})
}
