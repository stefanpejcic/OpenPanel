package web

import (
	"net/http"
	"net/url"

	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// DBCreated drives partials/_db_created.html, a blank Name means show the form
type DBCreated struct {
	Name      string
	Slug      string
	CanImport bool
}

// CreatedDBURL is where the create form lands, so a refresh can't resubmit the POST
func CreatedDBURL(slug, name string) string {
	return "/" + slug + "/new?created=" + url.QueryEscape(name)
}

// CreatedDBUserURL is the same thing for the create user form
func CreatedDBUserURL(slug, name string) string {
	return "/" + slug + "/user?created=" + url.QueryEscape(name)
}

// CreatedDBFromRequest reads the ?created= name set by the URLs above, users pass "" for importFeature
func CreatedDBFromRequest(r *http.Request, layout LayoutData, slug, importFeature string) DBCreated {
	name := r.URL.Query().Get("created")
	if !validators.IsValidIdentifier(name) {
		return DBCreated{}
	}
	return DBCreated{Name: name, Slug: slug, CanImport: layout.UserAllowed[importFeature]}
}

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
