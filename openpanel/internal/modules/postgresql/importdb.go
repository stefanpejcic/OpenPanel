package postgresql

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
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

var importSecureFilenameRE = regexp.MustCompile(`[^A-Za-z0-9_.-]`)

// secureFilename strips directory components and anything but ASCII
// letters/digits/dot/dash/underscore, so an uploaded filename can't be used
// to escape the target directory or inject odd characters into a path.
func secureFilename(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = importSecureFilenameRE.ReplaceAllString(name, "_")
	name = strings.Trim(name, "._")
	return name
}

var importDBNameRE = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// handlePostgresImportDB imports an uploaded .sql/.sql.gz dump into a
// PostgreSQL database. urlDBName carries the optional
// /postgresql/import/<dbname> URL segment.
func handlePostgresImportDB(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	urlDBName := r.PathValue("dbname")

	if !docker.IsServiceRunning(ctx, userContext, "postgres") {
		flashSess(a, w, r, "warning", "postgres container is not running. Please allow a few moments for initialization..")
		docker.StartComposeServiceIfNotRunning(ctx, userContext, "postgres")
	}

	if r.Method == http.MethodPost {
		if mpErr := r.ParseMultipartForm(1 << 30); mpErr != nil {
			_ = r.ParseForm()
		}
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

			targetDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_mysql_dumps/_data/"
			tempFilePath := filepath.Join(targetDir, filename)

			imported, importErr := importPostgresDump(ctx, userContext, dbName, targetDir, tempFilePath, filename, file)
			if _, statErr := os.Stat(tempFilePath); statErr == nil {
				_ = os.Remove(tempFilePath)
			}

			if imported {
				ipAddress := reqip.ClientIP(r)
				_ = logger.RecordUserAction(a.Config, currentUsername, "imported "+filename+" into PostgreSQL database "+dbName, ipAddress)
				flashSess(a, w, r, "success", "Successfully imported "+filename+" into database: "+dbName)
				renderImportPage(a, w, r, "", http.StatusOK)
				return
			}
			flashSess(a, w, r, "error", "Error importing "+filename+" into database "+dbName+": "+importErr)
			renderImportPage(a, w, r, dbName, http.StatusOK)
			return
		}

		flashSess(a, w, r, "error", "No database file uploaded!")
		renderImportPage(a, w, r, "", http.StatusOK)
		return
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Import feature is enabled for PostgreSQL."})
		return
	}
	renderImportPage(a, w, r, urlDBName, http.StatusOK)
}

// importPostgresDump saves the upload, chowns it to the account's UID,
// `podman cp`s it into the postgres container, then `psql -f`s it into the
// target database.
func importPostgresDump(ctx context.Context, userContext, dbName, targetDir, tempFilePath, filename string, uploaded io.Reader) (bool, string) {
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

	containerFilePath := "/tmp/" + filename
	copyArgv := podmanmanager.PodmanArgv(userContext, "cp", tempFilePath, "postgres:"+containerFilePath)
	if runErr := podmanmanager.Command(ctx, userContext, copyArgv).Run(); runErr != nil {
		return false, runErr.Error()
	}

	importArgv := podmanmanager.PodmanArgv(userContext, "exec", "-i", "postgres",
		"psql", "-U", "postgres", "-d", dbName, "-f", containerFilePath)
	cmd := podmanmanager.Command(ctx, userContext, importArgv)
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
