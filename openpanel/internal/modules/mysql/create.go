package mysql

import (
	"context"

	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// loadNames lists the account's databases and users for the create pages, empty when MySQL doesn't answer
func loadNames(ctx context.Context, userContext string) web.DBNames {
	names := web.DBNames{ReservedDatabases: restricted.databases, ReservedUsers: restricted.users}
	info, err := ComputeDatabasesInfo(ctx, userContext)
	if err != nil {
		return names
	}
	names.Databases, names.Users = info.Databases, info.Users
	for _, row := range info.AssignedDatabases {
		names.Assigned = append(names.Assigned, web.DBAssignment{Database: row.Database, Users: row.Users})
	}
	return names
}

// checkNewDatabase is the server side of the name checks on the new database and wizard forms, "" when the (prefixed) name is fine
func checkNewDatabase(name string, names web.DBNames) string {
	switch {
	case name == "":
		return "Database name is required."
	case !validators.IsValidIdentifier(name):
		return "Only letters, numbers, and underscores are allowed (1-63 characters)."
	case len(name) > 64:
		return "Database name is too long. MySQL identifiers are limited to 64 characters."
	case isRestrictedDatabase(name):
		return "This is a system database that can not be used."
	case names.HasDatabase(name):
		return "A database with this name already exists."
	}
	return ""
}

// checkNewUser does the same for a new user, its host and password
func checkNewUser(user, host, password string, threshold int, names web.DBNames) string {
	switch {
	case user == "":
		return "User name is required."
	case len(user) > maxUserLen:
		return "User name is too long. MySQL user names are limited to 32 characters."
	case !validators.IsValidIdentifier(user):
		return "Only letters, numbers, and underscores are allowed (1-31 characters)."
	case isRestrictedUser(user):
		return "This username is not allowed."
	case !validators.IsValidHost(host):
		return "Invalid host format."
	case len(password) < 8:
		return "Password must be at least 8 characters long"
	case !validators.IsPasswordStrongEnough(password, threshold):
		return "Password does not meet the required strength."
	case names.HasUser(user):
		return "A user with this name already exists."
	}
	return ""
}
