package web

// DBEngine is what the shared db/*.html pages need to know about the database engine they're for
type DBEngine struct {
	Slug        string // url prefix, "mysql" or "postgresql"
	NewIntro    string
	ConfigTitle string
	ConfigIntro string
}

var (
	MySQLEngine = DBEngine{
		Slug:        "mysql",
		NewIntro:    "Create a new database to store and organize data, providing a foundational structure for applications to manage and retrieve information efficiently.",
		ConfigTitle: "MySQL Configuration",
		ConfigIntro: "View and modify the MySQL service configuration. Any changes will restart the service.",
	}
	PostgreSQLEngine = DBEngine{
		Slug:        "postgresql",
		NewIntro:    "Create a new PostgreSQL database to store and organize data, providing a foundational structure for applications to manage and retrieve information efficiently.",
		ConfigTitle: "PostgreSQL Configuration",
		ConfigIntro: "View and modify the PostgreSQL service configuration. Any changes will restart the service.",
	}
)
