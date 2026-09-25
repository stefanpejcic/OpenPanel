package mongodb

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleDatabasesUsers lists the MongoDB users, starting the container in the background if it isn't running yet
func handleDatabasesUsers(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !checkMongoInsideContainer(ctx, userContext) {
		http.Redirect(w, r, "/mongodb", http.StatusFound)
		return
	}

	status := docker.GetContainerStatus(ctx, userContext, "mongodb")
	var userRows []UserRow

	switch {
	case status.State == "not_found":
		flashSess(a, w, r, "warning", "MongoDB service is not yet installed. Starting it in the background..")
		docker.StartOrStopContainer(ctx, userContext, "mongodb", "activate", "detached")
	case status.State != "running":
		flashSess(a, w, r, "warning", "MongoDB container is not running. Please allow a few moments for the initialization..")
	default:
		users, listErr := mongomanager.ListUsers(ctx, userContext)
		if listErr != nil {
			flashSess(a, w, r, "error", web.Tr(a, r, "Error fetching users: %(error)s", "error", listErr.Error()))
		} else {
			for _, u := range users {
				userRows = append(userRows, UserRow{Name: u.Username, Roles: rolesDisplay(u.Roles), IsSystem: isRestrictedUser(u.Username)})
			}
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"users":           userRows,
			"container_state": status.State, "health_status": status.Health,
		})
		return
	}

	renderUsersPage(a, w, r, status, userRows)
}

// handleDatabasesUser creates a new MongoDB user with no roles yet - access is granted separately via /mongodb/assign.
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
			flashAndRedirect(a, w, r, "error", "User name is required.", "/mongodb/user")
			return
		case !validators.IsValidIdentifier(dbUser):
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Name %(db_user)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "db_user", dbUser), "/mongodb/user")
			return
		case isRestrictedUser(dbUser):
			flashAndRedirect(a, w, r, "error", "This username is not allowed.", "/mongodb/user")
			return
		case !validators.IsPasswordStrongEnough(password, validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)):
			flashAndRedirect(a, w, r, "error", "Password does not meet the required strength.", "/mongodb/user")
			return
		}

		if createErr := mongomanager.CreateUser(ctx, userContext, dbUser, password); createErr != nil {
			flashSess(a, w, r, "error", web.Tr(a, r, "Failed to create user: %(error)s", "error", createErr.Error()))
		} else {
			ipAddress := reqip.ClientIP(r)
			_ = logger.RecordUserAction(a.Config, currentUsername, "created a MongoDB user "+dbUser, ipAddress)
			flashSess(a, w, r, "success", web.Tr(a, r, "Successfully created a MongoDB user %(db_user)s", "db_user", dbUser))
		}

		http.Redirect(w, r, "/mongodb/user", http.StatusFound)
		return
	}

	renderCreateUserPage(a, w, r)
}

// handleDatabasesPassword renders the change-password form for one existing user
func handleDatabasesPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	dbUser := r.PathValue("db_user")
	if isRestrictedUser(dbUser) {
		flashAndRedirect(a, w, r, "error", "This is a system username that can not be edited.", "/mongodb/user")
		return
	}
	renderChangePasswordPage(a, w, r, dbUser)
}

// handleDeleteMongoUser drops a MongoDB user entirely.
func handleDeleteMongoUser(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		flashAndRedirect(a, w, r, "error", "User name is required.", "/delete_mongodb_user")
		return
	case !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Name %(db_user)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+", "db_user", dbUser), "/delete_mongodb_user")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that cannot be deleted.", "/mongodb/users")
		return
	}

	if dropErr := mongomanager.DropUser(ctx, userContext, dbUser); dropErr != nil {
		flashSess(a, w, r, "error", web.Tr(a, r, "Error deleting user %(db_user)s: %(error)s", "db_user", dbUser, "error", dropErr.Error()))
	} else {
		flashSess(a, w, r, "success", web.Tr(a, r, "Successfully deleted user %(db_user)s", "db_user", dbUser))
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted a MongoDB database user "+dbUser, ipAddress)

	http.Redirect(w, r, "/mongodb/users", http.StatusFound)
}

// handleChangeMongoUserPassword updates an existing user's password.
func handleChangeMongoUserPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		flashAndRedirect(a, w, r, "error", "User name is required.", "/mongodb/change_user_password")
		return
	case !validators.IsValidIdentifier(dbUser):
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Name %(db_user)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "db_user", dbUser), "/mongodb/change_user_password")
		return
	case isRestrictedUser(dbUser):
		flashAndRedirect(a, w, r, "error", "This is a system username that can not be edited.", "/mongodb/users")
		return
	case !validators.IsPasswordStrongEnough(newPassword, validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)):
		flashAndRedirect(a, w, r, "error", "Password does not meet the required strength.", "/mongodb/change_user_password")
		return
	}

	if changeErr := mongomanager.ChangeUserPassword(ctx, userContext, dbUser, newPassword); changeErr != nil {
		flashSess(a, w, r, "error", web.Tr(a, r, "Error changing password for user %(db_user)s: %(error)s", "db_user", dbUser, "error", changeErr.Error()))
	} else {
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "changed password for MongoDB user "+dbUser, ipAddress)
		flashSess(a, w, r, "success", web.Tr(a, r, "Successfully changed password for user %(db_user)s", "db_user", dbUser))
	}

	http.Redirect(w, r, "/mongodb/users", http.StatusFound)
}
