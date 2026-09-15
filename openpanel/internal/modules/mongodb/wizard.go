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
)

// handleDatabasesWizard creates a database, a user, and grants that user a role on the database in one step, all server-side
func handleDatabasesWizard(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	status := docker.GetContainerStatus(ctx, userContext, "mongodb")
	if status.State != "running" {
		flashAndRedirect(a, w, r, "warning", "MongoDB service is not ready yet. Please wait for the installation to finish before creating a database.", "/mongodb")
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
			flashAndRedirect(a, w, r, "error", "Database name is required.", "/mongodb/wizard")
			return
		case !validators.IsValidIdentifier(databaseName):
			flashAndRedirect(a, w, r, "error", "Name "+databaseName+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/mongodb/wizard")
			return
		case isRestrictedDatabase(databaseName):
			flashAndRedirect(a, w, r, "error", "This is a system database that can not be used.", "/mongodb/wizard")
			return
		case dbUser == "":
			flashAndRedirect(a, w, r, "error", "User name is required.", "/mongodb/wizard")
			return
		case !validators.IsValidIdentifier(dbUser):
			flashAndRedirect(a, w, r, "error", "Name "+dbUser+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "/mongodb/wizard")
			return
		case isRestrictedUser(dbUser):
			flashAndRedirect(a, w, r, "error", "This username is not allowed.", "/mongodb/wizard")
			return
		case !validators.IsPasswordStrongEnough(password, validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)):
			flashAndRedirect(a, w, r, "error", "Password does not meet the required strength.", "/mongodb/wizard")
			return
		case !isValidRole(role):
			flashAndRedirect(a, w, r, "error", "Invalid role.", "/mongodb/wizard")
			return
		}

		injectedData, _ := a.InjectData(ctx, userID)
		planID, _ := injectedData["hosting_plan"].(int)
		plan, _ := a.QueryPlanDetailsByID(ctx, planID)
		dbLimit := 100
		if v := atoiDefault(plan.DBLimit, 0); v != 0 {
			dbLimit = v
		} else {
			dbLimit = 1000000
		}

		dbUsage := 0
		if dbs, countErr := mongomanager.ListDatabases(ctx, userContext); countErr == nil {
			dbUsage = len(dbs)
		}
		if dbUsage >= dbLimit {
			flashAndRedirect(a, w, r, "error", "You have reached the maximum number of databases allowed."+plan.UpgradeMessage(), "/mongodb/wizard")
			return
		}

		if createDBErr := mongomanager.CreateDatabase(ctx, userContext, databaseName); createDBErr != nil {
			flashAndRedirect(a, w, r, "error", "Failed to create database: "+createDBErr.Error(), "/mongodb/wizard")
			return
		}

		if createUserErr := mongomanager.CreateUser(ctx, userContext, dbUser, password); createUserErr != nil {
			flashAndRedirect(a, w, r, "error", "Failed to create user: "+createUserErr.Error(), "/mongodb/wizard")
			return
		}

		if grantErr := mongomanager.GrantRole(ctx, userContext, dbUser, role, databaseName); grantErr != nil {
			flashAndRedirect(a, w, r, "error", "Database and user created, but failed to grant role: "+grantErr.Error(), "/mongodb/wizard")
			return
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername,
			"used wizard to create MongoDB database "+databaseName+" and user "+dbUser, ipAddress)
		flashSess(a, w, r, "success", "Successfully created database "+databaseName+", user "+dbUser+", and granted "+role+".")
		http.Redirect(w, r, "/mongodb", http.StatusFound)
		return
	}

	renderWizardPage(a, w, r)
}
