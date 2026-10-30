package web

import (
	"slices"
	"sort"
)

// DBEngine is what the shared db/*.html pages need to know about the database engine they're for
type DBEngine struct {
	Slug        string // url prefix, "mysql", "postgresql" or "mongodb"
	Name        string
	Host        string // what apps inside the account connect to
	Port        string
	NewIntro    string
	ConfigTitle string
	ConfigIntro string
}

var (
	MySQLEngine = DBEngine{
		Slug: "mysql", Name: "MySQL", Host: "localhost", Port: "3306",
		NewIntro:    "Create a new database to store and organize data, providing a foundational structure for applications to manage and retrieve information efficiently.",
		ConfigTitle: "MySQL Configuration",
		ConfigIntro: "View and modify the MySQL service configuration. Any changes will restart the service.",
	}
	PostgreSQLEngine = DBEngine{
		Slug: "postgresql", Name: "PostgreSQL", Host: "postgres", Port: "5432",
		NewIntro:    "Create a new PostgreSQL database to store and organize data, providing a foundational structure for applications to manage and retrieve information efficiently.",
		ConfigTitle: "PostgreSQL Configuration",
		ConfigIntro: "View and modify the PostgreSQL service configuration. Any changes will restart the service.",
	}
	MongoDBEngine = DBEngine{
		Slug: "mongodb", Name: "MongoDB", Host: "mongodb", Port: "27017",
		NewIntro: "Create a new MongoDB database to store and organize data, providing a foundational structure for applications to manage and retrieve information efficiently.",
	}
)

// DBNames is the account's databases and users, the create pages list them and check new names against them
type DBNames struct {
	Databases []string
	Users     []string
	Assigned  []DBAssignment // nil where the engine doesn't list it

	// names the server refuses, so the forms can say so while typing; users are lowercase, the server compares them that way
	ReservedDatabases []string
	ReservedUsers     []string
}

// SortedKeys lists a set's names in order, for the reserved lists above
func SortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// DBAssignment is one database and the users that can use it, comma separated
type DBAssignment struct {
	Database string `json:"database"`
	Users    string `json:"users"`
}

func (n DBNames) HasDatabase(name string) bool { return slices.Contains(n.Databases, name) }

func (n DBNames) HasUser(name string) bool { return slices.Contains(n.Users, name) }
