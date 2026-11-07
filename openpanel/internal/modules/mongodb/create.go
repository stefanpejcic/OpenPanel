package mongodb

import (
	"context"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// mongo database names have to stay under 64 bytes, users get the same cap so the forms read alike
const maxNameLen = 63

// loadNames lists the account's databases and users, empty when MongoDB doesn't answer
func loadNames(ctx context.Context, userContext string) web.DBNames {
	names := web.DBNames{ReservedDatabases: mongomanager.SystemDatabases(), ReservedUsers: web.SortedKeys(restrictedUsers)}
	databases, users, err := ComputeDatabaseAndUserNames(ctx, userContext)
	if err != nil {
		return names
	}
	names.Databases, names.Users = databases, users
	return names
}

// checkNewDatabase is the server side of the name checks on the new database and wizard forms, "" when the name is fine
func checkNewDatabase(name string, names web.DBNames) string {
	switch {
	case name == "":
		return "Database name is required."
	case !validators.IsValidIdentifier(name), len(name) > maxNameLen:
		return "Only letters, numbers, and underscores are allowed (1-63 characters)."
	case isRestrictedDatabase(name):
		return "This is a system database that can not be used."
	case names.HasDatabase(name):
		return "A database with this name already exists."
	}
	return ""
}

// checkNewUser does the same for a new user and its password
func checkNewUser(user, password string, threshold int, names web.DBNames) string {
	switch {
	case user == "":
		return "User name is required."
	case !validators.IsValidIdentifier(user), len(user) > maxNameLen:
		return "Only letters, numbers, and underscores are allowed (1-63 characters)."
	case isRestrictedUser(user):
		return "This username is not allowed."
	case len(password) < 8:
		return "Password must be at least 8 characters long"
	case !validators.IsPasswordStrongEnough(password, threshold):
		return "Password does not meet the required strength."
	case names.HasUser(user):
		return "A user with this name already exists."
	}
	return ""
}

// planDBLimit is the plan's database limit, 0 when unlimited or unknown
func planDBLimit(ctx context.Context, a *appctx.App, injectedData map[string]any) int {
	planID, _ := injectedData["hosting_plan"].(int)
	plan, err := a.QueryPlanDetailsByID(ctx, planID)
	if err != nil {
		return 0
	}
	return web.AtoiDefault(plan.DBLimit, 0)
}
