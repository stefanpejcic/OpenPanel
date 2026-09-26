package postgresql

import (
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/dbexport"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func nameOptions(names []string) []web.BulkOption {
	options := make([]web.BulkOption, 0, len(names))
	for _, n := range names {
		options = append(options, web.BulkOption{Value: n, Label: n})
	}
	return options
}

func databasesBulkActions(t i18n.Translator, users []string) []web.BulkAction {
	return []web.BulkAction{
		{Key: "export", Label: t.Get("Export"), Confirm: t.Get("Export the selected databases as .sql.gz files to this folder:"),
			Input: &web.BulkInput{Type: "text", Default: dbexport.WebRoot, Placeholder: dbexport.WebRoot}},
		{Key: "assign", Label: t.Get("Assign user"), Confirm: t.Get("Give this user all privileges on the selected databases:"), Input: &web.BulkInput{Type: "select", Options: nameOptions(users)}},
		{Key: "unassign", Label: t.Get("Remove user"), Confirm: t.Get("Remove this user from the selected databases:"), Input: &web.BulkInput{Type: "select", Options: nameOptions(users)}},
		{Key: "delete", Label: t.Get("Delete"), Confirm: t.Get("Permanently delete the selected databases? This cannot be undone."), Danger: true},
	}
}

func usersBulkActions(t i18n.Translator, databases []string) []web.BulkAction {
	return []web.BulkAction{
		{Key: "password", Label: t.Get("Change password"), Confirm: t.Get("Set this password for the selected users:"), Input: &web.BulkInput{Type: "password", Placeholder: t.Get("New password")}},
		{Key: "assign", Label: t.Get("Add to database"), Confirm: t.Get("Give the selected users all privileges on:"), Input: &web.BulkInput{Type: "select", Options: nameOptions(databases)}},
		{Key: "unassign", Label: t.Get("Remove from database"), Confirm: t.Get("Remove the selected users from:"), Input: &web.BulkInput{Type: "select", Options: nameOptions(databases)}},
		{Key: "delete", Label: t.Get("Delete"), Confirm: t.Get("Permanently delete the selected users? This cannot be undone."), Danger: true},
	}
}

// handleBulk serves both /postgresql/bulk (items are databases) and /postgresql/users/bulk (items are users)
func handleBulk(a *appctx.App, mux http.Handler, w http.ResponseWriter, r *http.Request, usersPage bool) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	databases, users, _ := ComputeDatabaseAndUserNames(r.Context(), userContext)
	t := web.RequestTranslator(a, r)
	actions := databasesBulkActions(t, users)
	if usersPage {
		actions = usersBulkActions(t, databases)
	}
	web.ServeBulkDispatch(a, mux, w, r, actions, func(action, value, item string) (*web.BulkCall, web.BulkResult) {
		db, user := item, value
		if usersPage {
			db, user = value, item
		}
		switch {
		case action == "export":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/postgresql/export", Form: url.Values{
				"database_name": {db}, "export_destination": {"files"}, "export_format": {"gzip"}, "local_path": {value},
			}})
		case action == "assign":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/postgresql/assign", Form: url.Values{"db_user": {user}, "database_name": {db}}})
		case action == "unassign":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/postgresql/remove_user", Form: url.Values{"db_user": {user}, "database_name": {db}}})
		case action == "password":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/postgresql/change_user_password", Form: url.Values{"db_user": {item}, "new_password": {value}}})
		case action == "delete" && usersPage:
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/delete_postgres_user", Form: url.Values{"db_user": {item}}})
		case action == "delete":
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/postgresql/delete", Form: url.Values{"database_name": {item}}})
		}
		return web.Skip("Unknown bulk action.")
	})
}
