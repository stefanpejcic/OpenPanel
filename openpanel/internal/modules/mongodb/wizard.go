package mongodb

import (
	"net/http"
	"strings"

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
		// right after it's started MongoDB doesn't answer yet, say so instead of bouncing back without a word
		web.FlashRedirect(a, w, r, "warning", "MongoDB is still starting, please try again in a moment.", "/mongodb")
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		databaseName := strings.TrimSpace(r.Form.Get("database_name"))
		dbUser := strings.TrimSpace(r.Form.Get("db_user"))
		password := r.Form.Get("password")
		role := r.Form.Get("role")
		if role == "" {
			role = "readWrite"
		}

		// failures keep what was typed in the form
		fail := func(message string) {
			web.Flash(a, w, r, "error", message)
			renderWizardPage(a, w, r, databaseName, dbUser, role)
		}
		names := loadNames(ctx, userContext)
		threshold := validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)
		if msg := checkNewDatabase(databaseName, names); msg != "" {
			fail(web.Tr(a, r, msg))
			return
		}
		if msg := checkNewUser(dbUser, password, threshold, names); msg != "" {
			fail(web.Tr(a, r, msg))
			return
		}
		if !isValidRole(role) {
			fail(web.Tr(a, r, "Invalid role."))
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
		if len(names.Databases) >= dbLimit {
			fail(web.Tr(a, r, "You have reached the maximum number of databases allowed.%(upgrade_message)s", "upgrade_message", plan.UpgradeMessage()))
			return
		}

		if createDBErr := mongomanager.CreateDatabase(ctx, userContext, databaseName); createDBErr != nil {
			fail(web.Tr(a, r, "Failed to create database: %(error)s", "error", createDBErr.Error()))
			return
		}

		if createUserErr := mongomanager.CreateUser(ctx, userContext, dbUser, password); createUserErr != nil {
			fail(web.Tr(a, r, "Failed to create user: %(error)s", "error", createUserErr.Error()))
			return
		}

		if grantErr := mongomanager.GrantRole(ctx, userContext, dbUser, role, databaseName); grantErr != nil {
			fail(web.Tr(a, r, "Database and user created, but failed to grant role: %(error)s", "error", grantErr.Error()))
			return
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername,
			"used wizard to create MongoDB database "+databaseName+" and user "+dbUser, ipAddress)
		renderWizardSuccess(a, w, r, databaseName, dbUser, password, role)
		return
	}

	renderWizardPage(a, w, r, "", "", "")
}
