package mysql

import (
	"net/http"
	"sync"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

var initOnce sync.Once

// ensureInit lazily loads the config-derived package state
// (restricted user/database lists, tuning knobs, available conf keys)
// exactly once, regardless of which Register* function - or which order
// they're called in by RegisterAll - triggers it first.
func ensureInit(a *appctx.App) {
	initOnce.Do(func() {
		loadRestrictedNames(a)
		loadTuningConfig(a)
		loadConfKeys()
	})
}

// Register wires the core MySQL database/user management routes onto mux,
// plus the always-on (login-only, no feature gate) /json/mysql-size route,
// since it has no registrar of its own to live in.
func Register(mux *http.ServeMux, a *appctx.App) {
	ensureInit(a)
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "mysql")(h)
	}

	mux.Handle("GET /mysql", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabases(a, w, r) }))
	mux.Handle("GET /mysql/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesNew(a, w, r) }))
	mux.Handle("POST /mysql/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesNew(a, w, r) }))
	mux.Handle("POST /mysql/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteDatabase(a, w, r) }))

	mux.Handle("GET /mysql/users", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUsers(a, w, r) }))
	mux.Handle("GET /mysql/user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUser(a, w, r) }))
	mux.Handle("POST /mysql/user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUser(a, w, r) }))
	mux.Handle("GET /mysql/password/{db_user}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesPassword(a, w, r) }))
	mux.Handle("GET /delete_db_user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteDBUser(a, w, r) }))
	mux.Handle("POST /delete_db_user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteDBUser(a, w, r) }))
	mux.Handle("POST /mysql/change_user_password", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleChangeMySQLUserPassword(a, w, r) }))

	mux.Handle("GET /mysql/wizard", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesWizard(a, w, r) }))
	mux.Handle("POST /mysql/wizard", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesWizard(a, w, r) }))

	mux.Handle("GET /mysql/assign", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesAssign(a, w, r) }))
	mux.Handle("POST /mysql/assign", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesAssign(a, w, r) }))
	mux.Handle("GET /mysql/privileges/{db_user}/{database_name}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMySQLUserPrivileges(a, w, r) }))
	mux.Handle("GET /mysql/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesRemove(a, w, r) }))
	mux.Handle("POST /mysql/remove_user_from_db", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveUserFromDB(a, w, r) }))

	mux.Handle("POST /mysql/export", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleExportDatabase(a, w, r) }))
	mux.Handle("GET /mysql/{action}/{db_name}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabaseAction(a, w, r) }))
	mux.Handle("POST /mysql/{action}/{db_name}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabaseAction(a, w, r) }))

	mux.Handle("GET /mysql/info", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesInfo(a, w, r) }))
	mux.Handle("GET /json/mysql-size", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesSizeInfo(a, w, r) }))
}

// RegisterConf wires the MySQL server configuration route onto mux.
func RegisterConf(mux *http.ServeMux, a *appctx.App) {
	ensureInit(a)
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "mysql_conf")(h)
	}
	mux.Handle("GET /mysql/configuration", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEditMySQLConfig(a, w, r) }))
	mux.Handle("POST /mysql/configuration", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEditMySQLConfig(a, w, r) }))
}

// RegisterImport wires the database-import routes onto mux.
func RegisterImport(mux *http.ServeMux, a *appctx.App) {
	ensureInit(a)
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "mysql_import")(h)
	}
	mux.Handle("GET /mysql/import", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMySQLImportDB(a, w, r) }))
	mux.Handle("POST /mysql/import", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMySQLImportDB(a, w, r) }))
	mux.Handle("GET /mysql/import/{dbname}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMySQLImportDB(a, w, r) }))
	mux.Handle("POST /mysql/import/{dbname}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMySQLImportDB(a, w, r) }))
}

// RegisterProcesslist wires the MySQL processlist route onto mux.
func RegisterProcesslist(mux *http.ServeMux, a *appctx.App) {
	ensureInit(a)
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "mysql_processlist")(h)
	}
	mux.Handle("GET /mysql/processlist", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMySQLProcessList(a, w, r) }))
}

// RegisterRootPassword wires the MySQL root-password route onto mux.
func RegisterRootPassword(mux *http.ServeMux, a *appctx.App) {
	ensureInit(a)
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "mysql_root_password")(h)
	}
	mux.Handle("GET /mysql/root-password", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRootPasswordMySQL(a, w, r) }))
	mux.Handle("POST /mysql/root-password", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRootPasswordMySQL(a, w, r) }))

	// API equivalent, gated by the same "mysql_root_password" feature as the
	// web route above (unlike the import API route, this path doesn't
	// collide with any wildcard registered elsewhere, so it can carry its
	// own feature gate instead of borrowing RegisterAPI's "mysql" one).
	apiregistry.Handle(mux, a, "mysql_root_password", "PUT /api/mysql/root-password", func(w http.ResponseWriter, r *http.Request) { apiMySQLSetRootPassword(a, w, r) })
}

// RegisterRemote wires the remote-MySQL-access routes onto mux.
func RegisterRemote(mux *http.ServeMux, a *appctx.App) {
	ensureInit(a)
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "remote_mysql")(h)
	}
	mux.Handle("GET /mysql/remote-mysql", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoteMySQL(a, w, r) }))
	mux.Handle("POST /mysql/remote-mysql", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoteMySQL(a, w, r) }))
	mux.Handle("POST /mysql/remote-mysql/access/add", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoteMySQLAccessAdd(a, w, r) }))
	mux.Handle("POST /mysql/remote-mysql/access/edit", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoteMySQLAccessEdit(a, w, r) }))
	mux.Handle("POST /mysql/remote-mysql/access/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoteMySQLAccessDelete(a, w, r) }))
}
