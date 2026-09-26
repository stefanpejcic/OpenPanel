package mysql

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleDatabasesWizard creates a database and a user for it in one step - unlike most other handlers in this package, validation/creation failures here re-render the same form (preserving the typed database/user names) rather than redirecting, so the admin doesn't have to retype everything
func handleDatabasesWizard(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := GetMySQLVersion(ctx, a, userContext)

	status := docker.GetContainerStatus(ctx, userContext, mysqlVersion)
	if status.State != "running" {
		flashAndRedirect(a, w, r, "warning", web.Tr(a, r, "%(mysql_version)s service is not ready yet. Please wait for the installation to finish before creating a database.", "mysql_version", mysqlVersion), "/mysql")
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		databaseName := strings.TrimSpace(r.Form.Get("database_name"))
		dbUser := strings.TrimSpace(r.Form.Get("db_user"))
		dbHost := strings.TrimSpace(r.Form.Get("db_host"))
		if dbHost == "" {
			dbHost = "%"
		}
		password := r.Form.Get("password")
		selectedPrivs := r.Form["privileges"]

		flashAndRerender := func(category, message string) {
			flashSess(a, w, r, category, message)
			renderWizardForm(a, w, r, mysqlVersion, databaseName, dbUser)
		}

		switch {
		case databaseName == "":
			flashAndRerender("error", "Database name is required.")
			return
		case !validators.IsValidIdentifier(databaseName):
			flashAndRerender("error", web.Tr(a, r, "Name %(database_name)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+", "database_name", databaseName))
			return
		case isRestrictedDatabase(databaseName):
			flashAndRerender("error", "This is a system database that can not be used.")
			return
		case dbUser == "":
			flashAndRerender("error", "User name is required.")
			return
		case !validators.IsValidIdentifier(dbUser):
			flashAndRerender("error", web.Tr(a, r, "Name %(db_user)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+", "db_user", dbUser))
			return
		case !validators.IsValidHost(dbHost):
			flashAndRerender("error", "Invalid host format.")
			return
		case len(selectedPrivs) == 0:
			flashAndRerender("error", "At least one privilege must be selected.")
			return
		}
		privilegesSQL, ok := buildPrivilegesSQL(selectedPrivs)
		if !ok {
			flashAndRerender("error", "Invalid privilege selected.")
			return
		}

		injectedData, _ := a.InjectData(ctx, userID)
		planID, _ := injectedData["hosting_plan"].(int)
		plan, _ := a.QueryPlanDetailsByID(ctx, planID)
		dbLimit := atoiDefault(plan.DBLimit, 0)
		invalidateDatabaseCount(ctx, a, currentUsername)
		dbUsage := getDatabaseCount(ctx, a, currentUsername, userContext)
		if dbLimit != 0 && dbUsage >= dbLimit {
			flashAndRerender("error", web.Tr(a, r, "Error creating database: '%(database_name)s' - You have reached the maximum number of databases allowed.%(upgrade_message)s", "database_name", databaseName, "upgrade_message", plan.UpgradeMessage()))
			return
		}

		escapedPassword := escapeMySQLString(password)

		docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")

		queries := []string{
			"CREATE DATABASE IF NOT EXISTS `" + databaseName + "`",
			"CREATE USER IF NOT EXISTS '" + dbUser + "'@'" + dbHost + "' IDENTIFIED BY '" + escapedPassword + "'",
			"GRANT " + privilegesSQL + " ON `" + databaseName + "`.* TO '" + dbUser + "'@'" + dbHost + "'",
			"FLUSH PRIVILEGES",
		}
		for _, q := range queries {
			if _, execErr := mysqlmanager.Exec(ctx, userContext, q, ""); execErr != nil {
				flashAndRerender("error", execErr.Error())
				return
			}
		}
		invalidateDatabasesInfo(ctx, a, userContext)
		invalidateDatabaseCount(ctx, a, currentUsername)

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername,
			"used wizard to create MySQL database "+databaseName+", user "+dbUser+", privileges "+privilegesSQL, ipAddress)

		renderWizardSuccess(a, w, r, mysqlVersion, databaseName, dbUser, password, dbHost, privilegesSQL)
		return
	}

	renderWizardForm(a, w, r, mysqlVersion, "", "")
}
