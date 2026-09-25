package mongodb

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleDatabasesAssign grants a Mongo user a role scoped to a database.
func handleDatabasesAssign(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		dbUser := r.Form.Get("db_user")
		databaseName := r.Form.Get("database_name")
		role := r.Form.Get("role")
		if role == "" {
			role = "readWrite"
		}

		switch {
		case dbUser == "" || !validators.IsValidIdentifier(dbUser):
			flashAndRedirect(a, w, r, "error", "Invalid or missing user name.", "/mongodb/assign")
			return
		case databaseName == "" || !validators.IsValidIdentifier(databaseName):
			flashAndRedirect(a, w, r, "error", "Invalid or missing database name.", "/mongodb/assign")
			return
		case isRestrictedDatabase(databaseName):
			flashAndRedirect(a, w, r, "error", "This is a system database that cannot be used.", "/mongodb")
			return
		case !isValidRole(role):
			flashAndRedirect(a, w, r, "error", "Invalid role.", "/mongodb/assign")
			return
		}

		if grantErr := mongomanager.GrantRole(ctx, userContext, dbUser, role, databaseName); grantErr != nil {
			flashSess(a, w, r, "error", web.Tr(a, r, "Failed to assign user to database: %(error)s", "error", grantErr.Error()))
		} else {
			ipAddress := reqip.ClientIP(r)
			_ = logger.RecordUserAction(a.Config, currentUsername, "granted "+role+" to MongoDB user "+dbUser+" on database "+databaseName, ipAddress)
			flashSess(a, w, r, "success", web.Tr(a, r, "Successfully added a user %(db_user)s to MongoDB database", "db_user", dbUser))
		}

		http.Redirect(w, r, "/mongodb", http.StatusFound)
		return
	}

	renderAssignPage(a, w, r)
}

// handleDatabasesRemove renders the page for revoking a user's role on a MongoDB database
func handleDatabasesRemove(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	renderRemovePage(a, w, r)
}

// handleRemoveMongoUserFromDB revokes a MongoDB user's role on a database
func handleRemoveMongoUserFromDB(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	dbUser := r.Form.Get("db_user")
	databaseName := r.Form.Get("database_name")
	role := r.Form.Get("role")
	if role == "" {
		role = "readWrite"
	}

	switch {
	case dbUser == "" || !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", "Invalid or missing user name.", "/mongodb/remove")
		return
	case databaseName == "" || !validators.IsValidIdentifier(databaseName):
		flashAndRedirect(a, w, r, "error", "Invalid or missing database name.", "/mongodb/remove")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that cannot be edited.", "/mongodb/remove")
		return
	case isRestrictedDatabase(databaseName):
		flashAndRedirect(a, w, r, "error", "This is a system database that cannot be edited.", "/mongodb/remove")
		return
	}

	if revokeErr := mongomanager.RevokeRole(ctx, userContext, dbUser, role, databaseName); revokeErr != nil {
		flashSess(a, w, r, "error", web.Tr(a, r, "Failed to revoke role for user %(db_user)s from MongoDB database %(database_name)s: %(error)s", "db_user", dbUser, "database_name", databaseName, "error", revokeErr.Error()))
	} else {
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "revoked "+role+" for MongoDB user "+dbUser+" from database "+databaseName, ipAddress)
		flashSess(a, w, r, "success", web.Tr(a, r, "Successfully revoked access for user %(db_user)s from MongoDB database %(database_name)s", "db_user", dbUser, "database_name", databaseName))
	}

	http.Redirect(w, r, "/mongodb", http.StatusFound)
}
