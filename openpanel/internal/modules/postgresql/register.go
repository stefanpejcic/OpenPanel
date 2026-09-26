package postgresql

import (
	"net/http"
	"net/url"
	"sync"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var initOnce sync.Once

// ensureInit lazily loads the config-derived package state (available config keys) exactly once, regardless of which Register* function - or what order RegisterAll calls them in - triggers it first
func ensureInit() {
	initOnce.Do(func() {
		loadConfKeys()
	})
}

// Register wires the postgresql module's routes onto mux, plus the always-on /json/postgresql-size route (login-only, no enabled_modules gate) since it has no registrar of its own - same pragmatic stashing as /json/mysql-size in the mysql package
func Register(mux *http.ServeMux, a *appctx.App) {
	ensureInit()
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "postgresql")(h)
	}

	mux.Handle("GET /postgresql", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabases(a, w, r) }))
	mux.Handle("GET /postgresql/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesNew(a, w, r) }))
	mux.Handle("POST /postgresql/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesNew(a, w, r) }))
	mux.Handle("POST /postgresql/export", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleExportDatabase(a, w, r) }))
	mux.Handle("POST /postgresql/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteDatabase(a, w, r) }))
	mux.Handle("POST /postgresql/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBulk(a, mux, w, r, false) }))
	mux.Handle("POST /postgresql/users/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBulk(a, mux, w, r, true) }))

	mux.Handle("GET /postgresql/users", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUsers(a, w, r) }))
	mux.Handle("GET /postgresql/user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUser(a, w, r) }))
	mux.Handle("POST /postgresql/user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUser(a, w, r) }))
	mux.Handle("GET /postgresql/password/{db_user}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesPassword(a, w, r) }))
	mux.Handle("POST /delete_postgres_user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeletePostgresUser(a, w, r) }))
	mux.Handle("POST /postgresql/change_user_password", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleChangePostgresUserPassword(a, w, r) }))

	mux.Handle("GET /postgresql/wizard", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesWizard(a, w, r) }))
	mux.Handle("POST /postgresql/wizard", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesWizard(a, w, r) }))

	mux.Handle("GET /postgresql/assign", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesAssign(a, w, r) }))
	mux.Handle("POST /postgresql/assign", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesAssign(a, w, r) }))
	mux.Handle("GET /postgresql/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesRemove(a, w, r) }))
	mux.Handle("POST /postgresql/remove_user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemovePostgresUserFromDB(a, w, r) }))

	mux.Handle("GET /postgresql/info", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesInfo(a, w, r) }))
	mux.Handle("GET /json/postgresql-size", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesSizeInfo(a, w, r) }))
	mux.Handle("GET /postgresql/processlist", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleProcessList(a, w, r) }))
	mux.Handle("POST /postgresql/processlist/kill", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleKillQuery(a, w, r) }))
	mux.Handle("POST /postgresql/processlist/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		web.ServeBulkDispatch(a, mux, w, r, processlistBulkActions(web.RequestTranslator(a, r)), func(_, _, id string) (*web.BulkCall, web.BulkResult) {
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/postgresql/processlist/kill", Form: url.Values{"pid": {id}}})
		})
	}))
}

// RegisterConf wires the postgresql_conf module's route onto mux.
func RegisterConf(mux *http.ServeMux, a *appctx.App) {
	ensureInit()
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "postgresql_conf")(h)
	}
	mux.Handle("GET /postgresql/configuration", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEditPostgresConfig(a, w, r) }))
	mux.Handle("POST /postgresql/configuration", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleEditPostgresConfig(a, w, r) }))
	mux.Handle("GET /postgresql/configuration/recommendations", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePgConfigRecommendations(a, w, r) }))
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

func processlistBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "kill", Label: t.Get("Kill"), Confirm: t.Get("Kill the selected queries?"), Danger: true},
	}
}
