package mysql

import (
	"fmt"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// handleDatabasesUsers lists MySQL users for this account.
func handleDatabasesUsers(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := GetMySQLVersion(ctx, a, userContext)
	showAll := r.URL.Query().Get("show_all") != ""

	status := docker.GetContainerStatus(ctx, userContext, mysqlVersion)
	if status.State == "not_found" {
		docker.StartOrStopContainer(ctx, userContext, mysqlVersion, "activate", "detached")
	}
	status.Health = resolveContainerHealth(ctx, userContext, status)

	var usersOutput []string
	if status.Health == "healthy" {
		query := "SELECT User FROM mysql.user WHERE User NOT IN (" + restricted.usersSQL + ")"
		if showAll {
			query = "SELECT User FROM mysql.user"
		}
		rows, execErr := mysqlmanager.Exec(ctx, userContext, query, "")
		if execErr != nil {
			flashSess(a, w, r, "error", "Error fetching users: "+execErr.Error())
		} else {
			for _, row := range rows {
				usersOutput = append(usersOutput, toStringCell(row[0]))
			}
		}
	} else {
		flashSess(a, w, r, "warning", mysqlWarningFlashMessage(mysqlVersion, status.State, status.Health))
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"container_state": status.State, "health_status": status.Health,
			"users": usersOutput, "show_all": showAll,
		})
		return
	}

	renderUsersPage(a, w, r, status, mysqlVersion, usersOutput, showAll)
}

// handleDatabasesUser creates a new database user.
func handleDatabasesUser(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := GetMySQLVersion(ctx, a, userContext)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		dbUser := r.Form.Get("db_user")
		dbHost := r.Form.Get("db_host")
		if dbHost == "" {
			dbHost = "%"
		}
		password := r.Form.Get("password")

		switch {
		case dbUser == "":
			flashAndRedirect(a, w, r, "error", "User name is required.", "/mysql/user")
			return
		case !validators.IsValidIdentifier(dbUser):
			flashAndRedirect(a, w, r, "error", fmt.Sprintf("Name %s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", dbUser), "/mysql/user")
			return
		case isRestrictedUser(dbUser):
			flashAndRedirect(a, w, r, "error", "This username is not allowed.", "/mysql/user")
			return
		case !validators.IsValidHost(dbHost):
			flashAndRedirect(a, w, r, "error", "Invalid host format.", "/mysql/user")
			return
		case !validators.IsPasswordStrongEnough(password, validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)):
			flashAndRedirect(a, w, r, "error", "Password does not meet the required strength.", "/mysql/user")
			return
		}

		escapedPassword := escapeMySQLString(password)
		if _, execErr := mysqlmanager.Exec(ctx, userContext, "CREATE USER '"+dbUser+"'@'"+dbHost+"' IDENTIFIED BY '"+escapedPassword+"'", ""); execErr != nil {
			flashSess(a, w, r, "error", execErr.Error())
		} else {
			invalidateDatabasesInfo(ctx, a, userContext)
			ipAddress := reqip.ClientIP(r)
			_ = logger.RecordUserAction(a.Config, currentUsername, "created a MySQL database user "+dbUser, ipAddress)
			flashSess(a, w, r, "success", "Successfully created a database user "+dbUser)
		}
	}

	renderCreateUserPage(a, w, r, mysqlVersion)
}

// handleDatabasesPassword renders the change-password form for one existing user.
func handleDatabasesPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	dbUser := r.PathValue("db_user")
	if isRestrictedUser(dbUser) {
		flashAndRedirect(a, w, r, "error", "This is a system username that can not be edited.", "/mysql/users")
		return
	}
	renderChangePasswordPage(a, w, r, dbUser)
}

// handleDeleteDBUser drops a database user.
func handleDeleteDBUser(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	dbHost := r.Form.Get("db_host")
	if dbHost == "" {
		dbHost = "%"
	}

	var dbUser string
	if r.Method == http.MethodPost {
		dbUser = r.Form.Get("db_user")
	} else {
		dbUser = r.URL.Query().Get("db_user")
	}

	switch {
	case dbUser == "":
		flashAndRedirect(a, w, r, "error", "User name is required.", "/delete_db_user")
		return
	case !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", fmt.Sprintf("Name %s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", dbUser), "/delete_db_user")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that can not be deleted.", "/mysql/users")
		return
	case !validators.IsValidHost(dbHost):
		flashAndRedirect(a, w, r, "error", "Invalid host format.", "/mysql/users")
		return
	}

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "DROP USER IF EXISTS '"+dbUser+"'@'"+dbHost+"'", ""); execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error deleting user: "+execErr.Error(), "/mysql/users")
		return
	}
	invalidateDatabasesInfo(ctx, a, userContext)

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted a MySQL database user "+dbUser, ipAddress)
	flashSess(a, w, r, "success", "Successfully deleted user "+dbUser)
	http.Redirect(w, r, "/mysql/users", http.StatusFound)
}

// handleChangeMySQLUserPassword sets a new password for an existing database user.
func handleChangeMySQLUserPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	dbUser := r.Form.Get("db_user")
	dbHost := r.Form.Get("db_host")
	if dbHost == "" {
		dbHost = "%"
	}
	newPassword := r.Form.Get("new_password")

	switch {
	case dbUser == "":
		flashAndRedirect(a, w, r, "error", "User name is required.", "/mysql/change_user_password")
		return
	case !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", fmt.Sprintf("Name %s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", dbUser), "/mysql/change_user_password")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that can not be edited.", "/mysql/users")
		return
	case !validators.IsValidHost(dbHost):
		flashAndRedirect(a, w, r, "error", "Invalid host format.", "/mysql/users")
		return
	case !validators.IsPasswordStrongEnough(newPassword, validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)):
		flashAndRedirect(a, w, r, "error", "Password does not meet the required strength.", "/mysql/change_user_password")
		return
	}

	escapedPassword := escapeMySQLString(newPassword)
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "ALTER USER '"+dbUser+"'@'"+dbHost+"' IDENTIFIED BY '"+escapedPassword+"'", ""); execErr != nil {
		flashAndRedirect(a, w, r, "error", "Error changing password: "+execErr.Error(), "/mysql/users")
		return
	}
	mysqlmanager.InvalidatePool(userContext)

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "changed password for MySQL database user "+dbUser, ipAddress)
	flashSess(a, w, r, "success", "Successfully changed password for user "+dbUser)
	http.Redirect(w, r, "/mysql/users", http.StatusFound)
}
