package mongodb

import (
	"errors"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/dbexport"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
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
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Name %(database_name)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+", "database_name", databaseName), "/mongodb")
		return
	case isRestrictedDatabase(databaseName):
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Database '%(database_name)s' is a system database and cannot be exported.", "database_name", databaseName), "/mongodb")
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
	dump := podmanmanager.Command(r.Context(), userContext, podmanmanager.PodmanArgv(userContext, args...))
	failedMsg := web.Tr(a, r, "Failed to export database %(database_name)s.", "database_name", databaseName)

	if destination == "browser" {
		if sendErr := dbexport.Send(w, dump, databaseName+ext, "application/octet-stream", false); sendErr != nil {
			flashAndRedirect(a, w, r, "error", failedMsg, "/mongodb")
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "exported MongoDB database "+databaseName+" to browser", reqip.ClientIP(r))
		return
	}

	localPath := strings.TrimSpace(r.Form.Get("local_path"))
	displayFile, saveErr := dbexport.SaveToFiles(dump, userContext, localPath, databaseName, ext, false)
	if errors.Is(saveErr, dbexport.ErrDumpFailed) {
		flashAndRedirect(a, w, r, "error", failedMsg, "/mongodb")
		return
	} else if saveErr != nil {
		flashAndRedirect(a, w, r, "error", saveErr.Error(), "/mongodb")
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "exported MongoDB database "+databaseName+" to folder "+localPath, reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", web.Tr(a, r, "Database '%(database_name)s' exported to %(display_file)s", "database_name", databaseName, "display_file", displayFile), "/mongodb")
}
