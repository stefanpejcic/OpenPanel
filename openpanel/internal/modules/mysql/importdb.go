package mysql

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

var mysqlImportSecureFilenameRE = regexp.MustCompile(`[^A-Za-z0-9_.-]`)

// secureFilename strips directory components and unsafe characters from a
// user-supplied upload filename before it's used as a path segment.
func secureFilename(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = mysqlImportSecureFilenameRE.ReplaceAllString(name, "_")
	name = strings.Trim(name, "._")
	return name
}

var importDBNameRE = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// handleMySQLImportDB imports an uploaded .sql/.sql.gz dump into a database.
// urlDBName carries the optional /mysql/import/{dbname} URL segment.
func handleMySQLImportDB(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
	urlDBName := r.PathValue("dbname")

	if !docker.IsServiceRunning(ctx, userContext, mysqlVersion) {
		flashSess(a, w, r, "warning", mysqlVersion+" container is not running. Please allow a few moments for the initialization..")
		docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")
	}

	if r.Method == http.MethodPost {
		maxBytes := int64(atoiDefault(mysqlImportMaxSizeGB, 1))*1024*1024*1024 + (1 << 20)
		_ = r.ParseMultipartForm(maxBytes)

		file, fileHeader, fileErr := r.FormFile("db_file")
		dbName := r.FormValue("database_name")

		if fileErr == nil && dbName != "" {
			defer file.Close()

			filename := secureFilename(fileHeader.Filename)
			if !strings.HasSuffix(filename, ".sql") && !strings.HasSuffix(filename, ".sql.gz") {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			if !importDBNameRE.MatchString(dbName) {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			maxImportBytes := int64(atoiDefault(mysqlImportMaxSizeGB, 1)) * 1024 * 1024 * 1024
			if fileHeader.Size > maxImportBytes {
				flashSess(a, w, r, "error", "Uploaded file exceeds "+mysqlImportMaxSizeGB+" GB limit.")
				renderImportPage(a, w, r, mysqlVersion, "", http.StatusRequestEntityTooLarge)
				return
			}

			targetDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_mysql_dumps/_data/"
			tempFilePath := filepath.Join(targetDir, filename)

			imported, errDetail := importDatabaseDump(ctx, userContext, mysqlVersion, dbName, targetDir, tempFilePath, file)
			if _, statErr := os.Stat(tempFilePath); statErr == nil {
				_ = os.Remove(tempFilePath)
			}

			if imported {
				ipAddress := reqip.ClientIP(r)
				_ = logger.RecordUserAction(a.Config, currentUsername, "imported "+fileHeader.Filename+" into MySQL database "+dbName, ipAddress)
				flashSess(a, w, r, "success", "Successfully imported from "+fileHeader.Filename+" file to database: "+dbName)
				renderImportPage(a, w, r, mysqlVersion, "", http.StatusOK)
				return
			}
			flashSess(a, w, r, "error", "Import into '"+dbName+"' failed: "+errDetail)
			// falls through to the shared bottom render below, using the
			// URL's dbname (not the form's dbName) - the failure page
			// should reflect what page the admin was already on.
		} else {
			flashSess(a, w, r, "error", "No database file uploaded!")
			renderImportPage(a, w, r, mysqlVersion, urlDBName, http.StatusOK)
			return
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Import feature is enabled for MySQL/MariaDB."})
		return
	}
	renderImportPage(a, w, r, mysqlVersion, urlDBName, http.StatusOK)
}

// importDatabaseDump saves the uploaded dump, chowns it to the account's
// UID, then streams it into the target database via `podman exec -i
// <mysql> <mysql> <dbname>` (mysql/mariadb CLI reads a plain .sql dump on
// stdin). Returns (ok, stderr-detail-on-failure).
func importDatabaseDump(ctx context.Context, userContext, mysqlVersion, dbName, targetDir, tempFilePath string, uploaded io.Reader) (bool, string) {
	if mkErr := os.MkdirAll(targetDir, 0o755); mkErr != nil {
		return false, mkErr.Error()
	}

	out, createErr := os.Create(tempFilePath)
	if createErr != nil {
		return false, createErr.Error()
	}
	_, copyErr := io.Copy(out, uploaded)
	_ = out.Close()
	if copyErr != nil {
		return false, copyErr.Error()
	}

	if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
		uidStr := strconv.Itoa(uid)
		_ = exec.Command("chown", "-R", uidStr+":"+uidStr, targetDir).Run()
	}

	inFile, openErr := os.Open(tempFilePath)
	if openErr != nil {
		return false, openErr.Error()
	}
	defer inFile.Close()

	importArgv := podmanmanager.PodmanArgv(userContext, "exec", "-i", mysqlVersion, mysqlVersion, dbName)
	cmd := podmanmanager.Command(ctx, userContext, importArgv)
	cmd.Stdin = inFile
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if runErr := cmd.Run(); runErr != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = "Unknown error"
		}
		return false, detail
	}
	return true, ""
}
