package backups

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// restoreMySQLFromSQL drops/recreates the database, then executes the dump in ";\n"-delimited chunks, tolerating failures on statements starting with SET, /*, or -- (harmless lines a dump may contain that can fail depending on server config)
func restoreMySQLFromSQL(ctx context.Context, userContext, dbName, sqlContent string) error {
	safeDB := strings.ReplaceAll(dbName, "`", "``")
	if _, err := mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+safeDB+"`", ""); err != nil {
		return err
	}
	if _, err := mysqlmanager.Exec(ctx, userContext, "CREATE DATABASE `"+safeDB+"`", ""); err != nil {
		return err
	}

	for _, stmt := range strings.Split(sqlContent, ";\n") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := mysqlmanager.Exec(ctx, userContext, stmt, dbName); err != nil {
			upper := strings.ToUpper(stmt)
			if !strings.HasPrefix(upper, "SET") && !strings.HasPrefix(stmt, "/*") && !strings.HasPrefix(stmt, "--") {
				return err
			}
		}
	}
	return nil
}

// restorePostgresFromSQL drops/recreates the database, then executes the dump the same tolerant way restoreMySQLFromSQL does - pg_dump uses the same "statement;\n" shape and also emits harmless SET/comment lines that can fail depending on server config
func restorePostgresFromSQL(ctx context.Context, userContext, dbName, sqlContent string) error {
	safeDB := strings.ReplaceAll(dbName, `"`, `""`)
	if _, err := postgresmanager.Exec(ctx, userContext, `DROP DATABASE IF EXISTS "`+safeDB+`"`, ""); err != nil {
		return err
	}
	if _, err := postgresmanager.Exec(ctx, userContext, `CREATE DATABASE "`+safeDB+`"`, ""); err != nil {
		return err
	}

	for _, stmt := range strings.Split(sqlContent, ";\n") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := postgresmanager.Exec(ctx, userContext, stmt, dbName); err != nil {
			upper := strings.ToUpper(stmt)
			if !strings.HasPrefix(upper, "SET") && !strings.HasPrefix(stmt, "/*") && !strings.HasPrefix(stmt, "--") {
				return err
			}
		}
	}
	return nil
}

// restoreSQLMember dispatches a tarSQLMember to the right engine based on which backup/<engine>/ folder it came from - an archive can contain both MySQL and PostgreSQL dumps side by side, and running one through the wrong manager would fail outright or silently corrupt the wrong engine's database of the same name
func restoreSQLMember(ctx context.Context, userContext string, member tarSQLMember, dbName string) error {
	if member.Engine == "postgres" {
		return restorePostgresFromSQL(ctx, userContext, dbName, member.Content)
	}
	return restoreMySQLFromSQL(ctx, userContext, dbName, member.Content)
}

// chownAncestors chowns dir and every parent directory up to (but not including) root - root is already correctly owned, this just fixes up new subdirectories os.MkdirAll created as real host root beneath it
func chownAncestors(dir, root string, uid int) {
	for dir != root && dir != "." && dir != string(filepath.Separator) && strings.HasPrefix(dir, root) {
		_ = os.Chown(dir, uid, uid)
		dir = filepath.Dir(dir)
	}
}

var restoreFilesPathMap = func(userContext string) map[string]string {
	userHome := "/home/" + userContext
	return map[string]string{
		"backup/html":      userHome + "/docker-data/volumes/" + userContext + "_html_data/_data",
		"backup/vhosts":    userHome + "/docker-data/volumes/" + userContext + "_vhosts_data/_data",
		"backup/mail":      userHome + "/docker-data/volumes/" + userContext + "_mail_data/_data",
		"backup/crons.ini": userHome + "/crons.ini",
	}
}

// restoreFilesFromTar extracts html/vhosts/mail/crons entries from the archive into their host paths. uid is the account's rootless-container UID; every path written is chowned to it afterward, since this process runs as real host root and a root-owned file would be unreadable from inside the account's rootless containers - same as filemanager's chownRecursive has to do.
func restoreFilesFromTar(localPath, userContext string, uid int) ([]string, error) {
	pathMap := restoreFilesPathMap(userContext)

	f, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var extracted []string
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return extracted, err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		name := strings.Trim(header.Name, "/")

		var dest, root string
		for prefix, hostPath := range pathMap {
			if strings.HasPrefix(name, prefix+"/") || name == prefix {
				relative := strings.TrimPrefix(strings.TrimPrefix(name, prefix), "/")
				root = hostPath
				if info, statErr := os.Stat(hostPath); statErr == nil && info.IsDir() {
					dest = filepath.Join(hostPath, relative)
				} else if name == prefix {
					dest = hostPath
				}
				break
			}
		}

		if dest == "" {
			continue
		}
		destDir := filepath.Dir(dest)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			continue
		}
		if uid > 0 {
			chownAncestors(destDir, root, uid)
		}
		out, err := os.Create(dest)
		if err != nil {
			continue
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			continue
		}
		out.Close()
		if uid > 0 {
			_ = os.Chown(dest, uid, uid)
		}
		extracted = append(extracted, dest)
	}

	return extracted, nil
}

// tarSQLMember is one *.sql entry found while scanning a downloaded backup archive. Engine is "mysql" or "postgres", inferred from which backup/<engine>/ folder the entry lives under (see restoreSQLMember).
type tarSQLMember struct {
	Name    string
	Content string
	Engine  string
}

// scanSQLMembers collects every *.sql entry from a downloaded archive in a single pass, since archive/tar.Reader can't seek backward
func scanSQLMembers(localPath string) ([]tarSQLMember, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var members []tarSQLMember
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return members, err
		}
		if header.Typeflag != tar.TypeReg || !strings.HasSuffix(header.Name, ".sql") {
			continue
		}
		content, err := io.ReadAll(tr)
		if err != nil {
			return members, err
		}
		engine := "mysql"
		name := strings.Trim(header.Name, "/")
		if name == "backup/postgres" || strings.HasPrefix(name, "backup/postgres/") {
			engine = "postgres"
		}
		members = append(members, tarSQLMember{Name: header.Name, Content: string(content), Engine: engine})
	}
	return members, nil
}

// handleListBackupsFromDestination serves the list of remote backups (from the cached index file) and, on POST, kicks off a background reindex against the SSH destination
func handleListBackupsFromDestination(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	config, userHome := readBackupEnv(userContext)
	jsonFile := filepath.Join(userHome, "available_backups.json")
	lockFile := filepath.Join(userHome, "reindex.lock")

	if r.Method == http.MethodPost {
		if _, statErr := os.Stat(lockFile); statErr == nil {
			writeJSON(w, http.StatusConflict, map[string]string{"message": "Reindex already in progress."})
			return
		}

		if err := os.WriteFile(lockFile, []byte("running"), 0o644); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		go doReindex(userHome, config, jsonFile, lockFile)

		if r.URL.Query().Get("output") == "json" {
			writeJSON(w, http.StatusAccepted, map[string]string{"message": "Reindex started."})
			return
		}
		http.Redirect(w, r, "/backups/list", http.StatusFound)
		return
	}

	reindexing := false
	if _, statErr := os.Stat(lockFile); statErr == nil {
		reindexing = true
	}

	var backups []BackupInfo
	var reindexErr string
	if content, readErr := os.ReadFile(jsonFile); readErr == nil {
		var errPayload struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(content, &errPayload) == nil && errPayload.Error != "" {
			reindexErr = errPayload.Error
		} else {
			_ = json.Unmarshal(content, &backups)
		}
	}

	if r.URL.Query().Get("output") == "json" {
		// status=1 is the polling shape (onboarding wizard's destination test), which needs "reindexing" to tell an in-progress run apart from a finished one - kept opt-in so the existing Array.isArray() poller keeps working unchanged
		if r.URL.Query().Get("status") == "1" {
			writeJSON(w, http.StatusOK, map[string]any{"reindexing": reindexing, "error": reindexErr, "count": len(backups)})
			return
		}
		if reindexErr != "" {
			// this used to fall through to the HTML error page even for output=json callers - return JSON instead
			writeJSON(w, http.StatusOK, map[string]string{"error": reindexErr})
			return
		}
		writeJSON(w, http.StatusOK, backups)
		return
	}

	renderBackupRestorePage(a, w, r, backups, reindexing, reindexErr)
}

// handleRestoreFromBackup downloads a remote backup archive and restores it - all files, a single database, or files only, per restore_target
func handleRestoreFromBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// the JS client sends a FormData body, not urlencoded - ParseForm() alone doesn't read multipart bodies
	_ = r.ParseMultipartForm(1 << 20)
	backupFile := strings.TrimSpace(r.Form.Get("backup_file"))
	restoreTarget := r.Form.Get("restore_target")
	if restoreTarget == "" {
		restoreTarget = "all"
	}
	database := strings.TrimSpace(r.Form.Get("database"))

	if backupFile == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No backup file specified."})
		return
	}
	safeName := filepath.Base(backupFile)
	if safeName != backupFile {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup filename."})
		return
	}

	config, _ := readBackupEnv(userContext)
	store, storeErr := newRemoteStore(config)
	if storeErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": storeErr.Error()})
		return
	}

	ctx := r.Context()

	localPath, cleanup, fetchErr := store.Fetch(ctx, safeName)
	if fetchErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fetchErr.Error()})
		return
	}
	defer cleanup()

	uid, _ := a.GetUID(ctx, userContext)

	switch restoreTarget {
	case "all":
		if _, err := restoreFilesFromTar(localPath, userContext, uid); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		sqlMembers, err := scanSQLMembers(localPath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		for _, member := range sqlMembers {
			dbName := strings.TrimSuffix(filepath.Base(member.Name), ".sql")
			if err := restoreSQLMember(ctx, userContext, member, dbName); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "restored full backup from "+safeName, reqip.ClientIP(r))
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Full restore completed."})

	case "database":
		if database == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No database specified."})
			return
		}
		sqlMembers, err := scanSQLMembers(localPath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		var found *tarSQLMember
		for i := range sqlMembers {
			if strings.HasSuffix(sqlMembers[i].Name, database+".sql") {
				found = &sqlMembers[i]
				break
			}
		}
		if found == nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Database '" + database + "' not found in this backup."})
			return
		}
		if err := restoreSQLMember(ctx, userContext, *found, database); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "restored database '"+database+"' from backup "+safeName, reqip.ClientIP(r))
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Database '" + database + "' restored successfully."})

	case "files":
		extracted, err := restoreFilesFromTar(localPath, userContext, uid)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "restored files from backup "+safeName, reqip.ClientIP(r))
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Files restored (" + strconv.Itoa(len(extracted)) + " items)."})

	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unknown restore target: " + restoreTarget})
	}
}

// handleDownloadBackup downloads a remote backup archive and streams it back to the client
func handleDownloadBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// the JS client sends a FormData body, not urlencoded - ParseForm() alone doesn't read multipart bodies
	_ = r.ParseMultipartForm(1 << 20)
	backupFile := strings.TrimSpace(r.Form.Get("backup_file"))
	if backupFile == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No backup file specified."})
		return
	}
	safeName := filepath.Base(backupFile)
	if safeName != backupFile {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup filename."})
		return
	}

	config, _ := readBackupEnv(userContext)
	store, storeErr := newRemoteStore(config)
	if storeErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": storeErr.Error()})
		return
	}

	localPath, cleanup, fetchErr := store.Fetch(r.Context(), safeName)
	if fetchErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fetchErr.Error()})
		return
	}
	defer cleanup()

	f, openErr := os.Open(localPath)
	if openErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": openErr.Error()})
		return
	}
	defer f.Close()

	_ = logger.RecordUserAction(a.Config, currentUsername, "downloaded backup "+safeName, reqip.ClientIP(r))

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+safeName+"\"")
	_, _ = io.Copy(w, f)
}
