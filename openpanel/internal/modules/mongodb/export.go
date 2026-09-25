package mongodb

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/dbexport"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// mongoToolAuth runs a mongo tool inside the container with the root credentials the container already has, so the password never goes through the panel
const mongoToolAuth = `exec "$0" --username "$MONGO_INITDB_ROOT_USERNAME" --password "$MONGO_INITDB_ROOT_PASSWORD" --authenticationDatabase admin "$@"`

// handleExportDatabase dumps one database with mongodump and sends it to the browser or into /var/www/html
func handleExportDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	databaseName := strings.TrimSpace(r.Form.Get("database_name"))
	destination := r.Form.Get("export_destination")
	format := r.Form.Get("export_format")
	if format == "" {
		format = "archive"
	}

	switch {
	case databaseName == "":
		flashAndRedirect(a, w, r, "error", "Database name is required.", "/mongodb")
		return
	case !validators.IsValidIdentifier(databaseName):
		flashAndRedirect(a, w, r, "error", "Name "+databaseName+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+", "/mongodb")
		return
	case isRestrictedDatabase(databaseName):
		flashAndRedirect(a, w, r, "error", "Database '"+databaseName+"' is a system database and cannot be exported.", "/mongodb")
		return
	case format != "archive" && format != "gzip":
		flashAndRedirect(a, w, r, "error", "Invalid export format provided, select Archive or GZIP.", "/mongodb")
		return
	case destination != "browser" && destination != "files":
		flashAndRedirect(a, w, r, "error", "Invalid export destination.", "/mongodb")
		return
	}

	// mongorestore needs --gzip archives compressed per collection by mongodump itself, so no gzipping on our side
	args := []string{"exec", "mongodb", "sh", "-c", mongoToolAuth, "mongodump", "--quiet", "--archive", "--db", databaseName}
	ext := ".archive"
	if format == "gzip" {
		args = append(args, "--gzip")
		ext = ".archive.gz"
	}
	dump, dumpErr := podmanmanager.Command(r.Context(), userContext, podmanmanager.PodmanArgv(userContext, args...)).Output()
	if dumpErr != nil {
		flashAndRedirect(a, w, r, "error", "Failed to export database "+databaseName+".", "/mongodb")
		return
	}

	if destination == "browser" {
		_ = logger.RecordUserAction(a.Config, currentUsername, "exported MongoDB database "+databaseName+" to browser", reqip.ClientIP(r))
		dbexport.Send(w, databaseName+ext, "application/octet-stream", dump, false)
		return
	}

	localPath := strings.TrimSpace(r.Form.Get("local_path"))
	displayFile, saveErr := dbexport.SaveToFiles(userContext, localPath, databaseName, ext, dump, false)
	if saveErr != nil {
		flashAndRedirect(a, w, r, "error", saveErr.Error(), "/mongodb")
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "exported MongoDB database "+databaseName+" to folder "+localPath, reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", "Database '"+databaseName+"' exported to "+displayFile, "/mongodb")
}
