package postgresql

import (
	"net/http"

	"github.com/lib/pq"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleDatabasesWizard creates a database, a user, and grants that user access to the database in one step, all server-side - it used to do this via three sequential client-side fetch() calls, but those handlers always redirect and fetch() treats a redirect as success, so the wizard always reported success even when nothing was created
func handleDatabasesWizard(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	status := docker.GetContainerStatus(ctx, userContext, "postgres")
	if status.State != "running" {
		flashAndRedirect(a, w, r, "warning", "Postgres service is not ready yet. Please wait for the installation to finish before creating a database.", "/postgresql")
		return
	}

	if !checkPostgresInsideContainer(ctx, userContext) {
		http.Redirect(w, r, "/postgresql", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		databaseName := r.Form.Get("database_name")
		dbUser := r.Form.Get("db_user")
		password := r.Form.Get("password")

		switch {
		case databaseName == "":
			flashAndRedirect(a, w, r, "error", "Database name is required.", "/postgresql/wizard")
			return
		case !validators.IsValidIdentifier(databaseName):
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Name %(database_name)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "database_name", databaseName), "/postgresql/wizard")
			return
		case isRestrictedDatabase(databaseName):
			flashAndRedirect(a, w, r, "error", "This is a system database that can not be used.", "/postgresql/wizard")
			return
		case dbUser == "":
			flashAndRedirect(a, w, r, "error", "User name is required.", "/postgresql/wizard")
			return
		case !validators.IsValidIdentifier(dbUser):
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Name %(db_user)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+ ", "db_user", dbUser), "/postgresql/wizard")
			return
		case isRestrictedUser(dbUser):
			flashAndRedirect(a, w, r, "error", "This username is not allowed.", "/postgresql/wizard")
			return
		case !validators.IsPasswordStrongEnough(password, validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)):
			flashAndRedirect(a, w, r, "error", "Password does not meet the required strength.", "/postgresql/wizard")
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
		if rows, countErr := postgresmanager.Exec(ctx, userContext,
			"SELECT COUNT(*) FROM pg_database WHERE datname NOT IN ('postgres', 'template0', 'template1')", "postgres"); countErr == nil && len(rows) > 0 {
			dbUsage = postgresmanager.ToInt(rows[0][0])
		}
		if dbUsage >= dbLimit {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "You have reached the maximum number of databases allowed.%(upgrade_message)s", "upgrade_message", plan.UpgradeMessage()), "/postgresql/wizard")
			return
		}

		if _, execErr := postgresmanager.Exec(ctx, userContext, `CREATE DATABASE "`+databaseName+`"`, "postgres"); execErr != nil {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Failed to create database: %(error)s", "error", execErr.Error()), "/postgresql/wizard")
			return
		}

		if _, execErr := postgresmanager.Exec(ctx, userContext, `CREATE USER "`+dbUser+`" WITH PASSWORD `+pq.QuoteLiteral(password), "postgres"); execErr != nil {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Failed to create user: %(error)s", "error", execErr.Error()), "/postgresql/wizard")
			return
		}

		grantSQL := `GRANT ALL PRIVILEGES ON DATABASE "` + databaseName + `" TO "` + dbUser + `"; ` +
			`GRANT USAGE ON SCHEMA public TO "` + dbUser + `"; ` +
			`GRANT CREATE ON SCHEMA public TO "` + dbUser + `";`
		if _, execErr := postgresmanager.Exec(ctx, userContext, grantSQL, databaseName); execErr != nil {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Database and user created, but failed to grant privileges: %(error)s", "error", execErr.Error()), "/postgresql/wizard")
			return
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername,
			"used wizard to create PostgreSQL database "+databaseName+" and user "+dbUser, ipAddress)
		flashSess(a, w, r, "success", web.Tr(a, r, "Successfully created database %(database_name)s, user %(db_user)s, and granted privileges.", "database_name", databaseName, "db_user", dbUser))
		http.Redirect(w, r, "/postgresql", http.StatusFound)
		return
	}

	renderWizardPage(a, w, r)
}
