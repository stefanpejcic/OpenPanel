package mongodb

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// Register wires the mongodb module's routes onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "mongodb")(h)
	}

	mux.Handle("GET /mongodb", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabases(a, w, r) }))
	mux.Handle("GET /mongodb/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesNew(a, w, r) }))
	mux.Handle("POST /mongodb/new", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesNew(a, w, r) }))
	mux.Handle("POST /mongodb/export", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleExportDatabase(a, w, r) }))
	mux.Handle("POST /mongodb/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteDatabase(a, w, r) }))
	mux.Handle("POST /mongodb/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBulk(a, mux, w, r, false) }))
	mux.Handle("POST /mongodb/users/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBulk(a, mux, w, r, true) }))

	mux.Handle("GET /mongodb/users", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUsers(a, w, r) }))
	mux.Handle("GET /mongodb/user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUser(a, w, r) }))
	mux.Handle("POST /mongodb/user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesUser(a, w, r) }))
	mux.Handle("GET /mongodb/password/{db_user}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesPassword(a, w, r) }))
	mux.Handle("POST /delete_mongodb_user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteMongoUser(a, w, r) }))
	mux.Handle("POST /mongodb/change_user_password", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleChangeMongoUserPassword(a, w, r) }))

	mux.Handle("GET /mongodb/wizard", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesWizard(a, w, r) }))
	mux.Handle("POST /mongodb/wizard", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesWizard(a, w, r) }))

	mux.Handle("GET /mongodb/assign", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesAssign(a, w, r) }))
	mux.Handle("POST /mongodb/assign", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesAssign(a, w, r) }))
	mux.Handle("GET /mongodb/remove", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesRemove(a, w, r) }))
	mux.Handle("POST /mongodb/remove_user", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRemoveMongoUserFromDB(a, w, r) }))

	mux.Handle("GET /mongodb/info", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDatabasesInfo(a, w, r) }))
	mux.Handle("GET /mongodb/processlist", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleProcessList(a, w, r) }))
	mux.Handle("POST /mongodb/processlist/kill", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleKillQuery(a, w, r) }))
	mux.Handle("POST /mongodb/processlist/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		web.ServeBulkDispatch(a, mux, w, r, processlistBulkActions(web.RequestTranslator(a, r)), func(_, _, id string) (*web.BulkCall, web.BulkResult) {
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/mongodb/processlist/kill", Form: url.Values{"opid": {id}}})
		})
	}))
}

// RegisterImport wires the mongodb_import module's routes onto mux.
func RegisterImport(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "mongodb_import")(h)
	}
	mux.Handle("GET /mongodb/import", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMongoImportDB(a, w, r) }))
	mux.Handle("POST /mongodb/import", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMongoImportDB(a, w, r) }))
	mux.Handle("GET /mongodb/import/{dbname}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMongoImportDB(a, w, r) }))
	mux.Handle("POST /mongodb/import/{dbname}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMongoImportDB(a, w, r) }))
}

func processlistBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "kill", Label: t.Get("Kill"), Confirm: t.Get("Kill the selected operations?"), Danger: true},
	}
}
