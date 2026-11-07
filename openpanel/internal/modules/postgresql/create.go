package postgresql

import (
	"context"

	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// postgres cuts longer identifiers to NAMEDATALEN-1 without saying so
const maxNameLen = 63

// loadNames lists the account's databases, users and who can use which database, empty when PostgreSQL doesn't answer
func loadNames(ctx context.Context, userContext string) web.DBNames {
	names := web.DBNames{ReservedDatabases: web.SortedKeys(restrictedDatabases), ReservedUsers: web.SortedKeys(restrictedUsers)}
	databases, users, err := ComputeDatabaseAndUserNames(ctx, userContext)
	if err != nil {
		return names
	}
	names.Databases, names.Users = databases, users
	// CREATE rather than CONNECT, PUBLIC can connect to every database by default
	rows, err := postgresmanager.Exec(ctx, userContext, `
		SELECT d.datname, STRING_AGG(r.rolname, ', ' ORDER BY r.rolname)
		FROM pg_database d
		JOIN pg_roles r ON has_database_privilege(r.rolname, d.datname, 'CREATE')
		WHERE d.datname NOT IN ('postgres', 'template0', 'template1')
		  AND r.rolname NOT LIKE 'pg_%'
		  AND r.rolname NOT IN (`+filteredRolesSQL()+`)
		GROUP BY d.datname
		ORDER BY d.datname`, "postgres")
	if err == nil {
		for _, row := range rows {
			names.Assigned = append(names.Assigned, web.DBAssignment{Database: toStringCell(row[0]), Users: toStringCell(row[1])})
		}
	}
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
