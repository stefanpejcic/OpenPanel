package mysql

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	mysqldriver "github.com/go-sql-driver/mysql"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// handleDatabasesAssign grants a user privileges on a database. On success
// or a mid-transaction failure it falls through to re-rendering the same
// page (pre-filled with the just-submitted user/database) rather than
// redirecting - only the early field-validation failures redirect elsewhere.
func handleDatabasesAssign(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var dbUser, databaseName string

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		dbUser = r.Form.Get("db_user")
		dbHost := strings.TrimSpace(r.Form.Get("db_host"))
		if dbHost == "" {
			dbHost = "%"
		}
		databaseName = r.Form.Get("database_name")
		selectedPrivileges := r.Form["privileges"]

		switch {
		case dbUser == "":
			flashAndRedirect(a, w, r, "error", "User name is required.", "/mysql/assign")
			return
		case !validators.IsValidIdentifier(dbUser):
			flashAndRedirect(a, w, r, "error", "Name "+dbUser+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/mysql/assign")
			return
		case databaseName == "":
			flashAndRedirect(a, w, r, "error", "Database name is required.", "/mysql/assign")
			return
		case !validators.IsValidIdentifier(databaseName):
			flashAndRedirect(a, w, r, "error", "Name "+databaseName+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/mysql/assign")
			return
		case len(selectedPrivileges) == 0:
			target := "/mysql/assign?" + url.Values{"database_name": {databaseName}, "database_user": {dbUser}}.Encode()
			flashAndRedirect(a, w, r, "error", "At least one privilege must be selected.", target)
			return
		case isRestrictedDatabase(databaseName):
			flashAndRedirect(a, w, r, "error", "This is a system database that can not be used.", "/mysql")
			return
		case !validators.IsValidHost(dbHost):
			flashAndRedirect(a, w, r, "error", "Invalid host format.", "/mysql")
			return
		}

		privilegesSQL := strings.Join(selectedPrivileges, ", ")
		for _, p := range selectedPrivileges {
			if p == "ALL PRIVILEGES" {
				privilegesSQL = "ALL PRIVILEGES"
				break
			}
		}

		if _, revokeErr := mysqlmanager.Exec(ctx, userContext, "REVOKE ALL PRIVILEGES ON `"+databaseName+"`.* FROM '"+dbUser+"'@'"+dbHost+"'", ""); revokeErr != nil {
			var mysqlErr *mysqldriver.MySQLError
			// 1141: no existing grant on this database for this user - safe to ignore, matches the source.
			if !errors.As(revokeErr, &mysqlErr) || mysqlErr.Number != 1141 {
				flashSess(a, w, r, "error", revokeErr.Error())
				renderAssignPage(a, w, r, dbUser, databaseName)
				return
			}
		}

		if _, grantErr := mysqlmanager.Exec(ctx, userContext, "GRANT "+privilegesSQL+" ON `"+databaseName+"`.* TO '"+dbUser+"'@'"+dbHost+"'", ""); grantErr != nil {
			flashSess(a, w, r, "error", grantErr.Error())
			renderAssignPage(a, w, r, dbUser, databaseName)
			return
		}
		if _, flushErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); flushErr != nil {
			flashSess(a, w, r, "error", flushErr.Error())
			renderAssignPage(a, w, r, dbUser, databaseName)
			return
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername,
			"assigned privileges "+privilegesSQL+" to MySQL user "+dbUser+" on database "+databaseName, ipAddress)
		flashSess(a, w, r, "success", "Privileges granted successfully for user '"+dbUser+"' on database '"+databaseName+"'")
	}

	renderAssignPage(a, w, r, dbUser, databaseName)
}

// handleMySQLUserPrivileges returns a user's privileges on one database.
func handleMySQLUserPrivileges(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbUser := r.PathValue("db_user")
	databaseName := r.PathValue("database_name")
	const dbHost = "%"

	if !validators.IsValidIdentifier(dbUser) || !validators.IsValidIdentifier(databaseName) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid database or username"})
		return
	}

	rows, execErr := mysqlmanager.Exec(ctx, userContext, "SHOW GRANTS FOR '"+dbUser+"'@'"+dbHost+"'", "")
	if execErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch privileges", "details": execErr.Error()})
		return
	}

	privilegesSet := map[string]bool{}
	needle := "`" + databaseName + "`."
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
	writeJSON(w, http.StatusOK, map[string]any{"privileges": privileges})
}

// handleDatabasesRemove renders the revoke-access form.
func handleDatabasesRemove(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	renderRemovePage(a, w, r)
}

// handleRemoveUserFromDB revokes a user's privileges on a database.
func handleRemoveUserFromDB(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	dbUser := r.Form.Get("db_user")
	dbHost := strings.TrimSpace(r.Form.Get("db_host"))
	if dbHost == "" {
		dbHost = "%"
	}
	databaseName := r.Form.Get("database_name")

	switch {
	case dbUser == "":
		flashAndRedirect(a, w, r, "error", "User name is required.", "/remove_user_from_db")
		return
	case !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", "Name "+dbUser+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/remove_user_from_db")
		return
	case databaseName == "":
		flashAndRedirect(a, w, r, "error", "Database name is required.", "/remove_user_from_db")
		return
	case !validators.IsValidIdentifier(databaseName):
		flashAndRedirect(a, w, r, "error", "Name "+databaseName+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/remove_user_from_db")
		return
	case isRestrictedDatabase(strings.ToLower(databaseName)):
		flashAndRedirect(a, w, r, "error", "This is a system database that can not be edited.", "/mysql/users")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that can not be edited.", "/mysql/users")
		return
	case !validators.IsValidHost(dbHost):
		flashAndRedirect(a, w, r, "error", "Invalid host format.", "/mysql/users")
		return
	}

	if _, revokeErr := mysqlmanager.Exec(ctx, userContext, "REVOKE ALL PRIVILEGES ON `"+databaseName+"`.* FROM '"+dbUser+"'@'"+dbHost+"'", ""); revokeErr != nil {
		flashAndRedirect(a, w, r, "error", "Failed to revoke privileges: "+revokeErr.Error(), "/mysql/remove")
		return
	}
	if _, flushErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); flushErr != nil {
		flashAndRedirect(a, w, r, "error", "Failed to revoke privileges: "+flushErr.Error(), "/mysql/remove")
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "revoked all privileges for MySQL user "+dbUser+" from database "+databaseName, ipAddress)
	flashSess(a, w, r, "success", "Successfully revoked all privileges for user "+dbUser+" from database "+databaseName)
	http.Redirect(w, r, "/mysql", http.StatusFound)
}
