package postgresql

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

// handleExportDatabase dumps one database with pg_dump and sends it to the browser or into /var/www/html
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
		format = "sql"
	}

	switch {
	case databaseName == "":
		flashAndRedirect(a, w, r, "error", "Database name is required.", "/postgresql")
		return
	case !validators.IsValidIdentifier(databaseName):
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Name %(database_name)s is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+", "database_name", databaseName), "/postgresql")
		return
	case isSystemDatabase(databaseName):
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Database '%(database_name)s' is a system database and cannot be exported.", "database_name", databaseName), "/postgresql")
		return
	case format != "sql" && format != "gzip":
		flashAndRedirect(a, w, r, "error", "Invalid export format provided, select SQL or GZIP.", "/postgresql")
		return
	case destination != "browser" && destination != "files":
		flashAndRedirect(a, w, r, "error", "Invalid export destination.", "/postgresql")
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", "postgres", "pg_dump", "-U", "postgres", "--no-owner", "--no-privileges", databaseName)
	dump := podmanmanager.Command(r.Context(), userContext, argv)
	failedMsg := web.Tr(a, r, "Failed to export database %(database_name)s.", "database_name", databaseName)

	if destination == "browser" {
		if sendErr := dbexport.Send(w, dump, databaseName+".sql", "application/sql", format == "gzip"); sendErr != nil {
			flashAndRedirect(a, w, r, "error", failedMsg, "/postgresql")
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "exported PostgreSQL database "+databaseName+" to browser", reqip.ClientIP(r))
		return
	}

	localPath := strings.TrimSpace(r.Form.Get("local_path"))
	displayFile, saveErr := dbexport.SaveToFiles(dump, userContext, localPath, databaseName, ".sql", format == "gzip")
	if errors.Is(saveErr, dbexport.ErrDumpFailed) {
		flashAndRedirect(a, w, r, "error", failedMsg, "/postgresql")
		return
	} else if saveErr != nil {
		flashAndRedirect(a, w, r, "error", saveErr.Error(), "/postgresql")
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "exported PostgreSQL database "+databaseName+" to folder "+localPath, reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", web.Tr(a, r, "Database '%(database_name)s' exported to %(display_file)s", "database_name", databaseName, "display_file", displayFile), "/postgresql")
}
