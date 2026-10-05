package mongodb

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleDatabasesWizard creates a database, a user, and grants that user a role on the database in one step, all server-side
func handleDatabasesWizard(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	status := docker.GetContainerStatus(ctx, userContext, "mongodb")
	if status.State != "running" {
		web.FlashRedirect(a, w, r, "warning", "MongoDB service is not ready yet. Please wait for the installation to finish before creating a database.", "/mongodb")
		return
	}

	if !checkMongoInsideContainer(ctx, userContext) {
		http.Redirect(w, r, "/mongodb", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		databaseName := r.Form.Get("database_name")
		dbUser := r.Form.Get("db_user")
		password := r.Form.Get("password")
		role := r.Form.Get("role")
		if role == "" {
			role = "readWrite"
		}

		switch {
		case databaseName == "":
			web.FlashRedirect(a, w, r, "error", "Database name is required.", "/mongodb/wizard")
			return
		case !validators.IsValidIdentifier(databaseName):
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Name %(database_name)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "database_name", databaseName), "/mongodb/wizard")
			return
		case isRestrictedDatabase(databaseName):
			web.FlashRedirect(a, w, r, "error", "This is a system database that can not be used.", "/mongodb/wizard")
			return
		case dbUser == "":
			web.FlashRedirect(a, w, r, "error", "User name is required.", "/mongodb/wizard")
			return
		case !validators.IsValidIdentifier(dbUser):
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Name %(db_user)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "db_user", dbUser), "/mongodb/wizard")
			return
		case isRestrictedUser(dbUser):
			web.FlashRedirect(a, w, r, "error", "This username is not allowed.", "/mongodb/wizard")
			return
		case !validators.IsPasswordStrongEnough(password, validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)):
			web.FlashRedirect(a, w, r, "error", "Password does not meet the required strength.", "/mongodb/wizard")
			return
		case !isValidRole(role):
			web.FlashRedirect(a, w, r, "error", "Invalid role.", "/mongodb/wizard")
			return
		}

		injectedData, _ := a.InjectData(ctx, userID)
		planID, _ := injectedData["hosting_plan"].(int)
		plan, _ := a.QueryPlanDetailsByID(ctx, planID)
		dbLimit := 100
		if v := web.AtoiDefault(plan.DBLimit, 0); v != 0 {
			dbLimit = v
		} else {
			dbLimit = 1000000
		}

		dbUsage := 0
		if dbs, countErr := mongomanager.ListDatabases(ctx, userContext); countErr == nil {
			dbUsage = len(dbs)
		}
		if dbUsage >= dbLimit {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "You have reached the maximum number of databases allowed.%(upgrade_message)s", "upgrade_message", plan.UpgradeMessage()), "/mongodb/wizard")
			return
		}

		if createDBErr := mongomanager.CreateDatabase(ctx, userContext, databaseName); createDBErr != nil {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Failed to create database: %(error)s", "error", createDBErr.Error()), "/mongodb/wizard")
			return
		}

		if createUserErr := mongomanager.CreateUser(ctx, userContext, dbUser, password); createUserErr != nil {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Failed to create user: %(error)s", "error", createUserErr.Error()), "/mongodb/wizard")
			return
		}

		if grantErr := mongomanager.GrantRole(ctx, userContext, dbUser, role, databaseName); grantErr != nil {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Database and user created, but failed to grant role: %(error)s", "error", grantErr.Error()), "/mongodb/wizard")
			return
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername,
			"used wizard to create MongoDB database "+databaseName+" and user "+dbUser, ipAddress)
		web.Flash(a, w, r, "success", web.Tr(a, r, "Successfully created database %(database_name)s, user %(db_user)s, and granted %(role)s.", "database_name", databaseName, "db_user", dbUser, "role", role))
		http.Redirect(w, r, "/mongodb", http.StatusFound)
		return
	}

	renderWizardPage(a, w, r)
}
