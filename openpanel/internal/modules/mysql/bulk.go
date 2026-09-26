package mysql

import (
	"context"
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/dbexport"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// mysqlNames runs a one-column query for the bulk bar's user/database dropdowns
func mysqlNames(ctx context.Context, userContext, query string) []web.BulkOption {
	rows, err := mysqlmanager.Exec(ctx, userContext, query, "")
	if err != nil {
		return nil
	}
	options := make([]web.BulkOption, 0, len(rows))
	for _, row := range rows {
		name := toStringCell(row[0])
		options = append(options, web.BulkOption{Value: name, Label: name})
	}
	return options
}

func mysqlUserOptions(ctx context.Context, userContext string) []web.BulkOption {
	return mysqlNames(ctx, userContext, "SELECT DISTINCT User FROM mysql.user WHERE User NOT LIKE 'mysql.s%' AND User NOT IN ("+restricted.usersSQL+") ORDER BY User")
}

func mysqlDatabaseOptions(ctx context.Context, userContext string) []web.BulkOption {
	return mysqlNames(ctx, userContext, "SELECT schema_name FROM information_schema.schemata WHERE schema_name NOT IN ("+restricted.dbsSQL+") ORDER BY schema_name")
}

func databasesBulkActions(t i18n.Translator, users []web.BulkOption) []web.BulkAction {
	return []web.BulkAction{
		{Key: "export", Label: t.Get("Export"), Confirm: t.Get("Export the selected databases as .sql.gz files to this folder:"),
			Input: &web.BulkInput{Type: "text", Default: dbexport.WebRoot, Placeholder: dbexport.WebRoot}},
		{Key: "optimize", Label: t.Get("Optimize"), Confirm: t.Get("Optimize all tables in the selected databases?")},
		{Key: "repair", Label: t.Get("Repair"), Confirm: t.Get("Repair all tables in the selected databases?")},
		{Key: "assign", Label: t.Get("Assign user"), Confirm: t.Get("Give this user all privileges on the selected databases:"), Input: &web.BulkInput{Type: "select", Options: users}},
		{Key: "unassign", Label: t.Get("Remove user"), Confirm: t.Get("Remove this user from the selected databases:"), Input: &web.BulkInput{Type: "select", Options: users}},
		{Key: "delete", Label: t.Get("Delete"), Confirm: t.Get("Permanently delete the selected databases? This cannot be undone."), Danger: true},
	}
}

func usersBulkActions(t i18n.Translator, databases []web.BulkOption) []web.BulkAction {
	return []web.BulkAction{
		{Key: "password", Label: t.Get("Change password"), Confirm: t.Get("Set this password for the selected users:"), Input: &web.BulkInput{Type: "password", Placeholder: t.Get("New password")}},
		{Key: "assign", Label: t.Get("Add to database"), Confirm: t.Get("Give the selected users all privileges on:"), Input: &web.BulkInput{Type: "select", Options: databases}},
		{Key: "unassign", Label: t.Get("Remove from database"), Confirm: t.Get("Remove the selected users from:"), Input: &web.BulkInput{Type: "select", Options: databases}},
		{Key: "delete", Label: t.Get("Delete"), Confirm: t.Get("Permanently delete the selected users? This cannot be undone."), Danger: true},
	}
}

// handleDatabasesBulk replays each database through the same routes as its row buttons
func handleDatabasesBulk(a *appctx.App, mux http.Handler, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	actions := databasesBulkActions(web.RequestTranslator(a, r), mysqlUserOptions(r.Context(), userContext))
	web.ServeBulkDispatch(a, mux, w, r, actions, func(action, value, db string) (*web.BulkCall, web.BulkResult) {
		return mysqlBulkCall(action, value, db, value)
	})
}

// handleUsersBulk is the same for /mysql/users, where the item is the user and the value the database
func handleUsersBulk(a *appctx.App, mux http.Handler, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	actions := usersBulkActions(web.RequestTranslator(a, r), mysqlDatabaseOptions(r.Context(), userContext))
	web.ServeBulkDispatch(a, mux, w, r, actions, func(action, value, user string) (*web.BulkCall, web.BulkResult) {
		switch action {
		case "password":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/mysql/change_user_password", Form: url.Values{"db_user": {user}, "new_password": {value}}})
		case "delete":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/delete_db_user", Form: url.Values{"db_user": {user}}})
		}
		return mysqlBulkCall(action, value, value, user)
	})
}

func mysqlBulkCall(action, value, db, user string) (*web.BulkCall, web.BulkResult) {
	switch action {
	case "export":
		return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/mysql/export", Form: url.Values{
			"database_name": {db}, "export_destination": {"files"}, "export_format": {"gzip"}, "local_path": {value},
		}})
	case "optimize", "repair":
		return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/mysql/" + action + "/" + db})
	case "assign":
		return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/mysql/assign", Form: url.Values{"db_user": {user}, "database_name": {db}, "privileges": {"ALL PRIVILEGES"}}})
	case "unassign":
		return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/mysql/remove_user_from_db", Form: url.Values{"db_user": {user}, "database_name": {db}}})
	case "delete":
		return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/mysql/delete", Form: url.Values{"database_name": {db}}})
	}
	return web.Skip("Unknown bulk action.")
}

func remoteAccessBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "remove", Label: t.Get("Remove access"), Confirm: t.Get("Remove remote access for the selected user and host pairs?"), Danger: true},
	}
}
