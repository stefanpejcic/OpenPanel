package postgresql

import (
	"net/http"
	"strings"

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
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	status := docker.GetContainerStatus(ctx, userContext, "postgres")
	if status.State != "running" {
		web.FlashRedirect(a, w, r, "warning", "Postgres service is not ready yet. Please wait for the installation to finish before creating a database.", "/postgresql")
		return
	}

	if !checkPostgresInsideContainer(ctx, userContext) {
		http.Redirect(w, r, "/postgresql", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		databaseName := strings.TrimSpace(r.Form.Get("database_name"))
		dbUser := strings.TrimSpace(r.Form.Get("db_user"))
		password := r.Form.Get("password")

		// failures keep what was typed in the form
		fail := func(message string) {
			web.Flash(a, w, r, "error", message)
			renderWizardPage(a, w, r, databaseName, dbUser)
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
		if rows, countErr := postgresmanager.Exec(ctx, userContext,
			"SELECT COUNT(*) FROM pg_database WHERE datname NOT IN ('postgres', 'template0', 'template1')", "postgres"); countErr == nil && len(rows) > 0 {
			dbUsage = postgresmanager.ToInt(rows[0][0])
		}
		if dbUsage >= dbLimit {
			fail(web.Tr(a, r, "You have reached the maximum number of databases allowed.%(upgrade_message)s", "upgrade_message", plan.UpgradeMessage()))
			return
		}

		if _, execErr := postgresmanager.Exec(ctx, userContext, `CREATE DATABASE "`+databaseName+`"`, "postgres"); execErr != nil {
			fail(web.Tr(a, r, "Failed to create database: %(error)s", "error", execErr.Error()))
			return
		}

		if _, execErr := postgresmanager.Exec(ctx, userContext, `CREATE USER "`+dbUser+`" WITH PASSWORD `+pq.QuoteLiteral(password), "postgres"); execErr != nil {
			fail(web.Tr(a, r, "Failed to create user: %(error)s", "error", execErr.Error()))
			return
		}

		grantSQL := `GRANT ALL PRIVILEGES ON DATABASE "` + databaseName + `" TO "` + dbUser + `"; ` +
			`GRANT USAGE ON SCHEMA public TO "` + dbUser + `"; ` +
			`GRANT CREATE ON SCHEMA public TO "` + dbUser + `";`
		if _, execErr := postgresmanager.Exec(ctx, userContext, grantSQL, databaseName); execErr != nil {
			fail(web.Tr(a, r, "Database and user created, but failed to grant privileges: %(error)s", "error", execErr.Error()))
			return
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername,
			"used wizard to create PostgreSQL database "+databaseName+" and user "+dbUser, ipAddress)
		renderWizardSuccess(a, w, r, databaseName, dbUser, password)
		return
	}

	renderWizardPage(a, w, r, "", "")
}
