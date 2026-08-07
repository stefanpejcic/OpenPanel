package postgresql

import (
	"net/http"

	"github.com/lib/pq"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// handleDatabasesUsers lists the PostgreSQL roles, starting the container
// in the background if it isn't running yet.
func handleDatabasesUsers(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	showAll := r.URL.Query().Get("show_all") != ""

	if !checkPostgresInsideContainer(ctx, userContext) {
		http.Redirect(w, r, "/postgresql", http.StatusFound)
		return
	}

	status := docker.GetContainerStatus(ctx, userContext, "postgres")
	var usersList []string

	switch {
	case status.State == "not_found":
		flashSess(a, w, r, "warning", "Postgres service is not yet installed. Starting it in the background..")
		docker.StartOrStopContainer(ctx, userContext, "postgres", "activate", "detached")
	case status.State != "running":
		flashSess(a, w, r, "warning", "Postgres container is not running. Please allow a few moments for the initialization..")
	default:
		query := "SELECT rolname FROM pg_roles ORDER BY rolname"
		if !showAll {
			query = "SELECT rolname FROM pg_roles WHERE rolname NOT IN (" + filteredRolesSQL() + ") ORDER BY rolname"
		}
		rows, execErr := postgresmanager.Exec(ctx, userContext, query, "postgres")
		if execErr != nil {
			flashSess(a, w, r, "error", "Error fetching users: "+execErr.Error())
		} else {
			for _, row := range rows {
				usersList = append(usersList, toStringCell(row[0]))
			}
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"users": usersList, "show_all": showAll,
			"container_state": status.State, "health_status": status.Health,
		})
		return
	}

	userRows := make([]UserRow, len(usersList))
	for i, name := range usersList {
		userRows[i] = UserRow{Name: name, IsSystem: isSystemUser(name)}
	}
	renderUsersPage(a, w, r, status, userRows, showAll)
}

// handleDatabasesUser creates a new PostgreSQL database user.
func handleDatabasesUser(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		dbUser := r.Form.Get("db_user")
		password := r.Form.Get("password")

		switch {
		case dbUser == "":
			flashAndRedirect(a, w, r, "error", "User name is required.", "/postgresql/user")
			return
		case !validators.IsValidIdentifier(dbUser):
			flashAndRedirect(a, w, r, "error", "Name "+dbUser+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/postgresql/user")
			return
		case isRestrictedUser(dbUser):
			flashAndRedirect(a, w, r, "error", "This username is not allowed.", "/postgresql/user")
			return
		case !validators.IsPasswordStrongEnough(password, validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)):
			flashAndRedirect(a, w, r, "error", "Password does not meet the required strength.", "/postgresql/user")
			return
		}

		if _, execErr := postgresmanager.Exec(ctx, userContext, `CREATE USER "`+dbUser+`" WITH PASSWORD `+pq.QuoteLiteral(password), "postgres"); execErr != nil {
			flashSess(a, w, r, "error", "Failed to create user: "+execErr.Error())
		} else {
			ipAddress := reqip.ClientIP(r)
			_ = logger.RecordUserAction(a.Config, currentUsername, "created a PostgreSQL user "+dbUser, ipAddress)
			flashSess(a, w, r, "success", "Successfully created a PostgreSQL user "+dbUser)
		}

		http.Redirect(w, r, "/postgresql/user", http.StatusFound)
		return
	}

	renderCreateUserPage(a, w, r)
}

// handleDatabasesPassword renders the change-password form for one
// existing user.
func handleDatabasesPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	dbUser := r.PathValue("db_user")
	if isRestrictedUser(dbUser) {
		flashAndRedirect(a, w, r, "error", "This is a system username that can not be edited.", "/postgresql/user")
		return
	}
	renderChangePasswordPage(a, w, r, dbUser)
}

// handleDeletePostgresUser revokes a user's privileges on every database
// and drops the role.
func handleDeletePostgresUser(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var dbUser string
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		dbUser = r.Form.Get("db_user")
	} else {
		dbUser = r.URL.Query().Get("db_user")
	}

	switch {
	case dbUser == "":
		flashAndRedirect(a, w, r, "error", "User name is required.", "/delete_postgres_user")
		return
	case !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", "Name "+dbUser+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+", "/delete_postgres_user")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that cannot be deleted.", "/postgresql/users")
		return
	}

	dbList := []string{}
	if rows, execErr := postgresmanager.Exec(ctx, userContext, "SELECT datname FROM pg_database WHERE datistemplate = false", "postgres"); execErr == nil {
		for _, row := range rows {
			dbList = append(dbList, toStringCell(row[0]))
		}
	}

	for _, db := range dbList {
		_, _ = postgresmanager.Exec(ctx, userContext, `REVOKE ALL PRIVILEGES ON DATABASE "`+db+`" FROM "`+dbUser+`"`, "postgres")
		_, _ = postgresmanager.Exec(ctx, userContext, `REVOKE ALL PRIVILEGES ON SCHEMA public FROM "`+dbUser+`"`, db)
	}

	if _, execErr := postgresmanager.Exec(ctx, userContext, `DROP ROLE IF EXISTS "`+dbUser+`"`, "postgres"); execErr != nil {
		flashSess(a, w, r, "error", "Error deleting user "+dbUser+": "+execErr.Error())
	} else {
		flashSess(a, w, r, "success", "Successfully deleted user "+dbUser)
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted a PostgreSQL database user "+dbUser, ipAddress)

	http.Redirect(w, r, "/postgresql/users", http.StatusFound)
}

// handleChangePostgresUserPassword updates an existing user's password.
func handleChangePostgresUserPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	dbUser := r.Form.Get("db_user")
	newPassword := r.Form.Get("new_password")

	switch {
	case dbUser == "":
		flashAndRedirect(a, w, r, "error", "User name is required.", "/postgresql/change_user_password")
		return
	case !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", "Name "+dbUser+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/postgresql/change_user_password")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that can not be edited.", "/postgresql/users")
		return
	case !validators.IsPasswordStrongEnough(newPassword, validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)):
		flashAndRedirect(a, w, r, "error", "Password does not meet the required strength.", "/postgresql/change_user_password")
		return
	}

	if _, execErr := postgresmanager.Exec(ctx, userContext, `ALTER USER "`+dbUser+`" WITH PASSWORD `+pq.QuoteLiteral(newPassword), "postgres"); execErr != nil {
		flashSess(a, w, r, "error", "Error changing password for user "+dbUser+": "+execErr.Error())
	} else {
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "changed password for PostgreSQL user "+dbUser, ipAddress)
		flashSess(a, w, r, "success", "Successfully changed password for user "+dbUser)
	}

	http.Redirect(w, r, "/postgresql/users", http.StatusFound)
}
