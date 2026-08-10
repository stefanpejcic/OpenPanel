package mysql

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// apiMySQLImportDatabase imports an uploaded .sql/.sql.gz dump into a
// database (multipart field: "file"). This is the API equivalent of
// POST /mysql/import/{dbname}, reusing the same podman-stdin pipeline
// (importDatabaseDump, defined in importdb.go) that the web upload form
// uses. Wired from api.go's apiMySQLDatabasesPostDispatch, so it shares
// the "mysql" feature gate rather than the web route's separate
// "mysql_import" gate - see apiMySQLDatabasesPostDispatch's doc comment
// for why (Go's ServeMux only allows one registration of the
// "POST /api/mysql/databases/{rest...}" wildcard).
func apiMySQLImportDatabase(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dbName := r.PathValue("db_name")

	if !validators.IsValidIdentifier(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid database name."})
		return
	}
	if isRestrictedDatabase(dbName) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "This database cannot be imported into."})
		return
	}

	maxBytes := int64(atoiDefault(mysqlImportMaxSizeGB, 1))*1024*1024*1024 + (1 << 20)
	if mpErr := r.ParseMultipartForm(maxBytes); mpErr != nil {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "SQL file required (multipart field: 'file')."})
		return
	}
	file, header, fileErr := r.FormFile("file")
	if fileErr != nil || header.Filename == "" {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "SQL file required (multipart field: 'file')."})
		return
	}
	defer file.Close()

	filename := secureFilename(header.Filename)
	if !strings.HasSuffix(filename, ".sql") && !strings.HasSuffix(filename, ".sql.gz") {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "File must be a .sql or .sql.gz dump."})
		return
	}
	maxImportBytes := int64(atoiDefault(mysqlImportMaxSizeGB, 1)) * 1024 * 1024 * 1024
	if header.Size > maxImportBytes {
		writeAPIMySQLJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "Uploaded file exceeds " + mysqlImportMaxSizeGB + " GB limit."})
		return
	}

	mysqlVersion := GetMySQLVersion(ctx, a, userContext)
	docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")

	targetDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_mysql_dumps/_data/"
	tempFilePath := filepath.Join(targetDir, filename)

	imported, errDetail := importDatabaseDump(ctx, userContext, mysqlVersion, dbName, targetDir, tempFilePath, file)
	if _, statErr := os.Stat(tempFilePath); statErr == nil {
		_ = os.Remove(tempFilePath)
	}
	if !imported {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": "Import failed: " + errDetail})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "imported "+header.Filename+" into MySQL database "+dbName+" via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"database": dbName, "file": header.Filename, "imported": true})
}
