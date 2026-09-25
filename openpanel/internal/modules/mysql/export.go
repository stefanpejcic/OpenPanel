package mysql

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

// handleExportDatabase streams a database dump (sql or gzip) for download.
func handleExportDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	databaseName := strings.TrimSpace(r.Form.Get("database_name"))
	exportDestination := r.Form.Get("export_destination")
	if exportDestination == "" {
		exportDestination = "browser"
	}
	exportFormat := r.Form.Get("export_format")
	if exportFormat == "" {
		exportFormat = "sql"
	}
	mysqlVersion := GetMySQLVersion(ctx, a, userContext)

	switch {
	case databaseName == "":
		flashAndRedirect(a, w, r, "error", "Database name is required.", "/mysql")
		return
	case !validators.IsValidIdentifier(databaseName):
		flashAndRedirect(a, w, r, "error", "Name "+databaseName+" is not allowed. Please use alphanumeric characters and '_' - [a-zA-Z0-9_]+", "/mysql")
		return
	case isRestrictedDatabase(databaseName):
		flashAndRedirect(a, w, r, "error", "Database '"+databaseName+"' is restricted and cannot be exported. Contact Administrator", "/mysql")
		return
	case exportFormat != "sql" && exportFormat != "gzip":
		flashAndRedirect(a, w, r, "error", "Invalid export format provided, select SQL or GZIP.", "/mysql")
		return
	}

	var dumpCmd string
	switch mysqlVersion {
	case "mysql":
		dumpCmd = "mysqldump"
	case "mariadb":
		dumpCmd = "mariadb-dump"
	default:
		flashAndRedirect(a, w, r, "error", "Unsupported database engine: "+mysqlVersion, "/mysql")
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, dumpCmd, "-u", "root", databaseName)
	dumpOutput, dumpErr := podmanmanager.Command(ctx, userContext, argv).Output()
	if dumpErr != nil {
		flashAndRedirect(a, w, r, "error", "Failed to export database "+databaseName+".", "/mysql")
		return
	}

	switch exportDestination {
	case "browser":
		_ = logger.RecordUserAction(a.Config, currentUsername, "exported MYSQL database "+databaseName+" to browser", reqip.ClientIP(r))
		dbexport.Send(w, databaseName+".sql", "application/sql", dumpOutput, exportFormat == "gzip")
		return

	case "files":
		localPath := strings.TrimSpace(r.Form.Get("local_path"))
		displayFile, saveErr := dbexport.SaveToFiles(userContext, localPath, databaseName, ".sql", dumpOutput, exportFormat == "gzip")
		if saveErr != nil {
			flashAndRedirect(a, w, r, "error", saveErr.Error(), "/mysql")
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "exported MYSQL database "+databaseName+" to folder "+localPath, reqip.ClientIP(r))
		flashAndRedirect(a, w, r, "success", "Database '"+databaseName+"' exported to "+displayFile, "/mysql")
		return

	default:
		flashAndRedirect(a, w, r, "error", "Invalid export destination.", "/mysql")
		return
	}
}
