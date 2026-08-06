package postgresql

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// handleDatabasesAssign grants a Postgres user access to a database.
// Unlike MySQL's equivalent, there is no privilege checklist here -
// Postgres assignment is always the fixed GRANT ALL PRIVILEGES + USAGE +
// CREATE trio.
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

		switch {
		case dbUser == "" || !validators.IsValidIdentifier(dbUser):
			flashAndRedirect(a, w, r, "error", "Invalid or missing user name.", "/postgresql/assign")
			return
		case databaseName == "" || !validators.IsValidIdentifier(databaseName):
			flashAndRedirect(a, w, r, "error", "Invalid or missing database name.", "/postgresql/assign")
			return
		case isRestrictedDatabase(databaseName):
			flashAndRedirect(a, w, r, "error", "This is a system database that cannot be used.", "/postgresql")
			return
		}

		grantSQL := `GRANT ALL PRIVILEGES ON DATABASE "` + databaseName + `" TO "` + dbUser + `"; ` +
			`GRANT USAGE ON SCHEMA public TO "` + dbUser + `"; ` +
			`GRANT CREATE ON SCHEMA public TO "` + dbUser + `";`

		if _, execErr := postgresmanager.Exec(ctx, userContext, grantSQL, databaseName); execErr != nil {
			flashSess(a, w, r, "error", "Failed to assign user to database.")
		} else {
			ipAddress := reqip.ClientIP(r)
			_ = logger.RecordUserAction(a.Config, currentUsername, "assigned all privileges to PostgreSQL user "+dbUser+" on database "+databaseName, ipAddress)
			flashSess(a, w, r, "success", "Successfully added a user "+dbUser+" to PostgreSQL database")
		}

		http.Redirect(w, r, "/postgresql", http.StatusFound)
		return
	}

	renderAssignPage(a, w, r)
}

// handleDatabasesRemove renders the page for revoking a user's access to a
// PostgreSQL database.
func handleDatabasesRemove(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	renderRemovePage(a, w, r)
}

// handleRemovePostgresUserFromDB revokes a PostgreSQL user's privileges on
// a database.
func handleRemovePostgresUserFromDB(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	dbUser := r.Form.Get("db_user")
	databaseName := r.Form.Get("database_name")

	switch {
	case dbUser == "" || !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", "Invalid or missing user name.", "/postgresql/remove")
		return
	case databaseName == "" || !validators.IsValidIdentifier(databaseName):
		flashAndRedirect(a, w, r, "error", "Invalid or missing database name.", "/postgresql/remove")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that cannot be edited.", "/postgresql/remove")
		return
	case isRestrictedDatabase(databaseName):
		flashAndRedirect(a, w, r, "error", "This is a system database that cannot be edited.", "/postgresql/remove")
		return
	}

	if _, execErr := postgresmanager.Exec(ctx, userContext, `REVOKE ALL PRIVILEGES ON DATABASE "`+databaseName+`" FROM "`+dbUser+`"`, "postgres"); execErr != nil {
		flashSess(a, w, r, "error", "Failed to revoke privileges for user "+dbUser+" from PostgreSQL database "+databaseName)
	} else {
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "revoked all privileges for PostgreSQL user "+dbUser+" from database "+databaseName, ipAddress)
		flashSess(a, w, r, "success", "Successfully revoked all privileges for user "+dbUser+" from PostgreSQL database "+databaseName)
	}

	http.Redirect(w, r, "/postgresql", http.StatusFound)
}
