package postgresql

import (
	"net/http"
	"sync"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

var initOnce sync.Once

// ensureInit lazily loads the config-derived package state (available
// config keys) exactly once, regardless of which Register* function - or
// which order they're called in by RegisterAll - triggers it first.
func ensureInit() {
	initOnce.Do(func() {
		loadConfKeys()
	})
}

// Register wires the postgresql module's routes onto mux, plus the
// always-on (login-only, no enabled_modules gate) /json/postgresql-size
// route, since it has no registrar of its own to live in - the same
// pragmatic stashing already used for /json/mysql-size in the mysql
// package.
func Register(mux *http.ServeMux, a *appctx.App) {
	ensureInit()
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "postgresql")(h)
	}

	requirePGAdminLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "pgadmin")(h)
	}
	mux.Handle("GET /pgadmin/", requirePGAdminLogin(func(w http.ResponseWriter, r *http.Request) { handlePGAdminRedirect(a, w, r) }))
	mux.Handle("GET /postgresql/pgadmin", requirePGAdminLogin(func(w http.ResponseWriter, r *http.Request) { handlePGAdminRedirect(a, w, r) }))

	mux.Handle("GET /postgresql", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabases(a, w, r) }))
	mux.Handle("GET /postgresql/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesNew(a, w, r) }))
	mux.Handle("POST /postgresql/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesNew(a, w, r) }))
	mux.Handle("POST /postgresql/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteDatabase(a, w, r) }))

	mux.Handle("GET /postgresql/users", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUsers(a, w, r) }))
	mux.Handle("GET /postgresql/user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUser(a, w, r) }))
	mux.Handle("POST /postgresql/user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUser(a, w, r) }))
	mux.Handle("GET /postgresql/password/{db_user}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesPassword(a, w, r) }))
	mux.Handle("GET /delete_postgres_user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeletePostgresUser(a, w, r) }))
	mux.Handle("POST /delete_postgres_user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeletePostgresUser(a, w, r) }))
	mux.Handle("POST /postgresql/change_user_password", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleChangePostgresUserPassword(a, w, r) }))

	mux.Handle("GET /postgresql/wizard", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesWizard(a, w, r) }))

	mux.Handle("GET /postgresql/assign", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesAssign(a, w, r) }))
	mux.Handle("POST /postgresql/assign", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesAssign(a, w, r) }))
	mux.Handle("GET /postgresql/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesRemove(a, w, r) }))
	mux.Handle("POST /postgresql/remove_user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemovePostgresUserFromDB(a, w, r) }))

	mux.Handle("GET /postgresql/info", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesInfo(a, w, r) }))
	mux.Handle("GET /json/postgresql-size", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesSizeInfo(a, w, r) }))
	mux.Handle("GET /postgresql/processlist", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleProcessList(a, w, r) }))
}

// RegisterConf wires the postgresql_conf module's route onto mux.
func RegisterConf(mux *http.ServeMux, a *appctx.App) {
	ensureInit()
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "postgresql_conf")(h)
	}
	mux.Handle("GET /postgresql/configuration", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEditPostgresConfig(a, w, r) }))
	mux.Handle("POST /postgresql/configuration", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEditPostgresConfig(a, w, r) }))
}

// RegisterImport wires the postgresql_import module's routes onto mux.
func RegisterImport(mux *http.ServeMux, a *appctx.App) {
	ensureInit()
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "postgresql_import")(h)
	}
	mux.Handle("GET /postgresql/import", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePostgresImportDB(a, w, r) }))
	mux.Handle("POST /postgresql/import", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePostgresImportDB(a, w, r) }))
	mux.Handle("GET /postgresql/import/{dbname}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePostgresImportDB(a, w, r) }))
	mux.Handle("POST /postgresql/import/{dbname}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePostgresImportDB(a, w, r) }))
}

// RegisterRemote wires the remote_postgresql module's route onto mux.
func RegisterRemote(mux *http.ServeMux, a *appctx.App) {
	ensureInit()
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "remote_postgresql")(h)
	}
	mux.Handle("GET /postgresql/remote-postgresql", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemotePostgres(a, w, r) }))
	mux.Handle("POST /postgresql/remote-postgresql", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemotePostgres(a, w, r) }))
}
