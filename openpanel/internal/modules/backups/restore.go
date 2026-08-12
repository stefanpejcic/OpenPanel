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

	"github.com/pkg/sftp"
	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// fetchBackupViaSSH downloads one backup archive from the SSH/SFTP
// destination into a local temp file. The caller must call cleanup() once
// done with the local copy.
func fetchBackupViaSSH(config map[string]string, backupFilename string) (localPath string, cleanup func(), err error) {
	client, err := dialSSH(config)
	if err != nil {
		return "", nil, err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return "", nil, err
	}
	defer sftpClient.Close()

	remotePath := strings.TrimRight(config["SSH_REMOTE_PATH"], "/")
	remoteFilePath := remotePath + "/" + backupFilename

	tmpDir, err := os.MkdirTemp("", "backup_restore_")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(tmpDir) }

	remoteFile, err := sftpClient.Open(remoteFilePath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	defer remoteFile.Close()

	localPath = filepath.Join(tmpDir, backupFilename)
	localFile, err := os.Create(localPath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	defer localFile.Close()

	if _, err := io.Copy(localFile, remoteFile); err != nil {
		cleanup()
		return "", nil, err
	}

	return localPath, cleanup, nil
}

// restoreMySQLFromSQL drops/recreates the database, then executes the
// dump in ";\n"-delimited chunks, tolerating failures on statements that
// start with SET, /*, or -- (harmless SET/comment statements a dump may
// contain that can fail depending on server config).
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

// restorePostgresFromSQL drops/recreates the database, then executes the
// dump the same tolerant way restoreMySQLFromSQL does (pg_dump output uses
// the same "statement;\n" shape, and also emits harmless SET/comment lines
// that can fail depending on server config/extensions installed).
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

// restoreSQLMember dispatches a tarSQLMember to the right engine based on
// which backup/<engine>/ folder it came from - a backup archive can contain
// both MySQL and PostgreSQL dumps side by side (see the "backup" service's
// volume mounts in docker-compose.yml: mysql_dumps -> backup/mysql,
// pg_data -> backup/postgres), and running a Postgres dump through
// mysqlmanager (or vice versa) would fail outright or, worse, silently
// corrupt the wrong engine's database of the same name.
func restoreSQLMember(ctx context.Context, userContext string, member tarSQLMember, dbName string) error {
	if member.Engine == "postgres" {
		return restorePostgresFromSQL(ctx, userContext, dbName, member.Content)
	}
	return restoreMySQLFromSQL(ctx, userContext, dbName, member.Content)
}

// chownAncestors chowns dir and every parent directory up to (but not
// including) root - root already belongs to the account's bind-mounted
// volume and is already correctly owned; this only needs to fix up any new
// subdirectories os.MkdirAll just created (as real host root) beneath it.
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

// restoreFilesFromTar extracts html/vhosts/mail/crons entries from the
// archive into their host paths. uid is the account's rootless-container
// UID (from appctx.App.GetUID); every path this writes is chowned to it
// afterward, since the writing process itself runs as real host root - a
// file left root-owned on a bind-mounted volume is unreadable/undeletable
// from inside the account's own (rootless, UID-remapped) containers, same
// as internal/modules/filemanager's chownRecursive has to do after writing
// into these volumes as root.
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

// tarSQLMember is one *.sql entry found while scanning a downloaded
// backup archive. Engine is "mysql" or "postgres", inferred from which
// backup/<engine>/ top-level folder the entry lives under (see
// restoreSQLMember).
type tarSQLMember struct {
	Name    string
	Content string
	Engine  string
}

// scanSQLMembers collects every *.sql entry from a downloaded archive in a
// single pass, since archive/tar.Reader can't seek backward.
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

// handleListBackupsFromDestination serves the list of remote backups
// (from the cached index file) and, on POST, kicks off a background
// reindex against the SSH destination.
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
	if content, readErr := os.ReadFile(jsonFile); readErr == nil {
		var errPayload struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(content, &errPayload) == nil && errPayload.Error != "" {
			renderBackupRestorePage(a, w, r, nil, reindexing, errPayload.Error)
			return
		}
		_ = json.Unmarshal(content, &backups)
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, backups)
		return
	}

	renderBackupRestorePage(a, w, r, backups, reindexing, "")
}

// handleRestoreFromBackup downloads a remote backup archive and restores
// it - all files, a single database, or files only, per restore_target.
func handleRestoreFromBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// The JS client sends a FormData body (multipart/form-data), not
	// application/x-www-form-urlencoded - ParseForm() alone doesn't read
	// multipart bodies.
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

	localPath, cleanup, fetchErr := fetchBackupViaSSH(config, safeName)
	if fetchErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fetchErr.Error()})
		return
	}
	defer cleanup()

	ctx := r.Context()
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

// handleDownloadBackup downloads a remote backup archive and streams it
// back to the client.
func handleDownloadBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// The JS client sends a FormData body (multipart/form-data), not
	// application/x-www-form-urlencoded - ParseForm() alone doesn't read
	// multipart bodies.
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

	localPath, cleanup, fetchErr := fetchBackupViaSSH(config, safeName)
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
