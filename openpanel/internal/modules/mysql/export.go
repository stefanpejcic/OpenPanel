package mysql

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
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

	filenameBase := databaseName + ".sql"
	argv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, dumpCmd, "-u", "root", databaseName)
	dumpOutput, dumpErr := podmanmanager.Command(ctx, userContext, argv).Output()
	if dumpErr != nil {
		flashAndRedirect(a, w, r, "error", "Failed to export database "+databaseName+".", "/mysql")
		return
	}

	switch exportDestination {
	case "browser":
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "exported MYSQL database "+databaseName+" to browser", ipAddress)

		if exportFormat == "gzip" {
			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)
			_, _ = gz.Write(dumpOutput)
			_ = gz.Close()

			w.Header().Set("Content-Type", "application/gzip")
			w.Header().Set("Content-Disposition", `attachment; filename="`+filenameBase+`.gz"`)
			_, _ = w.Write(buf.Bytes())
			return
		}

		w.Header().Set("Content-Type", "application/sql")
		w.Header().Set("Content-Disposition", `attachment; filename="`+filenameBase+`"`)
		_, _ = w.Write(dumpOutput)
		return

	case "files":
		localPath := strings.TrimSpace(r.Form.Get("local_path"))
		if localPath == "" {
			flashAndRedirect(a, w, r, "error", "Destination path is required for folder export.", "/mysql")
			return
		}

		realPath := filepath.Clean(localPath)
		allowedBase := "/var/www/html"
		if realPath != allowedBase && !strings.HasPrefix(realPath, allowedBase+string(filepath.Separator)) {
			flashAndRedirect(a, w, r, "error", "Invalid export path. Must be inside /var/www/html/", "/mysql")
			return
		}

		newBase := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
		relativePart, relErr := filepath.Rel(allowedBase, realPath)
		if relErr != nil {
			flashAndRedirect(a, w, r, "error", "Invalid export path. Must be inside /var/www/html/", "/mysql")
			return
		}
		finalPath := filepath.Join(newBase, relativePart)

		if info, statErr := os.Stat(finalPath); statErr != nil || !info.IsDir() {
			flashAndRedirect(a, w, r, "error", "Export path "+localPath+" does not exist or is not a directory.", "/mysql")
			return
		}

		timestamp := time.Now().Format("2006-01-02_15-04-05")
		baseName := databaseName + "_" + timestamp
		fileName := baseName + ".sql"
		if exportFormat == "gzip" {
			fileName = baseName + ".sql.gz"
		}
		destFile := filepath.Join(finalPath, fileName)
		displayFile := filepath.Join(localPath, fileName)

		writeErr := writeExportFile(destFile, dumpOutput, exportFormat == "gzip")
		if writeErr == nil {
			if uid, uidErr := a.GetUID(ctx, userContext); uidErr == nil {
				uidStr := strconv.Itoa(uid)
				_ = exec.CommandContext(ctx, "chown", uidStr+":"+uidStr, destFile).Run()
			}
		}
		if writeErr != nil {
			if _, statErr := os.Stat(destFile); statErr == nil {
				_ = os.Remove(destFile)
			}
			flashAndRedirect(a, w, r, "error", "Failed to write export to "+displayFile+" - try export to browser instead.", "/mysql")
			return
		}

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "exported MYSQL database "+databaseName+" to folder "+localPath, ipAddress)
		flashSess(a, w, r, "success", "Database '"+databaseName+"' exported to "+displayFile)
		http.Redirect(w, r, "/mysql", http.StatusFound)
		return

	default:
		flashAndRedirect(a, w, r, "error", "Invalid export destination.", "/mysql")
		return
	}
}

func writeExportFile(path string, data []byte, asGzip bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if !asGzip {
		_, err = f.Write(data)
		return err
	}
	gz := gzip.NewWriter(f)
	if _, err := gz.Write(data); err != nil {
		_ = gz.Close()
		return err
	}
	return gz.Close()
}
