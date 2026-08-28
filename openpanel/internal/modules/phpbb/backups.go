package phpbb

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// This file mirrors flarum/backups.go's directory layout, naming and
// restore/run logic exactly (same backups/<domain>/<timestamp>/{database.sql,
// files.tar.gz} structure) - only the DB name lookup differs, since
// phpBB's config.php uses a plain $dbname = '...'; assignment rather than
// Flarum's PHP-array 'database' => '...' key.

func toStringCell(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case uint64:
		return strconv.FormatUint(t, 10)
	case int:
		return strconv.Itoa(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return ""
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

// extractPhpbbDatabaseInfoForBackup reads config.php straight off the
// host filesystem, reusing phpbbDBNameRE (defined in manage.go).
func extractPhpbbDatabaseInfoForBackup(userContext, docroot string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(docroot, wwwPrefix) {
		return map[string]string{"error": "invalid docroot"}
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "config.php"))
	if err != nil {
		return map[string]string{"error": "config.php not found"}
	}
	text := string(content)

	nameMatch := phpbbDBNameRE.FindStringSubmatch(text)
	if nameMatch == nil {
		return map[string]string{"error": "No database information found in config.php"}
	}
	return map[string]string{"database_name": nameMatch[1]}
}

var phpbbBackupFolderRE = regexp.MustCompile(`^20\d{2}-`)

type phpbbBackupDateInfo struct {
	Date           string `json:"date"`
	HasDbBackup    bool   `json:"hasDbBackup"`
	HasFilesBackup bool   `json:"hasFilesBackup"`
}

func handlePhpbbGetBackupDates(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	selectedDomain := r.PathValue("selected_domain")

	_, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	backupsPath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/backups/" + selectedDomain
	if _, statErr := os.Stat(backupsPath); statErr != nil {
		if mkErr := os.MkdirAll(backupsPath, 0o755); mkErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
			return
		}
	}

	entries, readErr := os.ReadDir(backupsPath)
	if readErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}

	var dates []phpbbBackupDateInfo
	for _, entry := range entries {
		if !entry.IsDir() || !phpbbBackupFolderRE.MatchString(entry.Name()) {
			continue
		}
		folderPath := filepath.Join(backupsPath, entry.Name())
		files, _ := os.ReadDir(folderPath)
		info := phpbbBackupDateInfo{Date: entry.Name()}
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".sql") {
				info.HasDbBackup = true
			}
			if strings.HasSuffix(f.Name(), ".tar.gz") {
				info.HasFilesBackup = true
			}
		}
		dates = append(dates, info)
	}

	writeJSON(w, http.StatusOK, dates)
}

var phpbbBackupDateRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`)

func handlePhpbbRestoreBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	selectedDomain := r.PathValue("selected_domain")
	backupDate := r.URL.Query().Get("backup_date")
	docroot := r.URL.Query().Get("docroot")

	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := selectedDomain
	if idx := strings.Index(selectedDomain, "/"); idx != -1 {
		domain = selectedDomain[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if docroot == "" {
		docroot = "/var/www/html/" + selectedDomain
	}
	if !strings.HasPrefix(docroot, "/var/www/html/") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid docroot"})
		return
	}

	if !phpbbBackupDateRE.MatchString(backupDate) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup date."})
		return
	}

	docrootOnHostOS := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, "/var/www/html/") + "/"
	backupsPathOnHostOS := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/backups/" + selectedDomain + "/"
	backupsPathInContainer := "/var/www/html/backups/" + selectedDomain + "/"

	backupDatePathOnHostOS := filepath.Join(backupsPathOnHostOS, backupDate)
	backupDatePathInContainer := filepath.Join(backupsPathInContainer, backupDate)
	databaseSQLPathOnHostOS := filepath.Join(backupDatePathOnHostOS, "database.sql")
	databaseSQLPathInContainer := filepath.Join(backupDatePathInContainer, "database.sql")
	targzPathOnHostOS := filepath.Join(backupDatePathOnHostOS, "files.tar.gz")

	ipAddress := reqip.ClientIP(r)
	var restoredItems []string

	if _, statErr := os.Stat(targzPathOnHostOS); statErr == nil {
		if runErr := exec.CommandContext(ctx, "tar", "-xzf", targzPathOnHostOS, "-C", docrootOnHostOS).Run(); runErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
			return
		}
		// tar (run here as a plain host process, not inside the
		// container) restores each entry's exact archived numeric
		// ownership - since files.tar.gz was created via `tar -czf`
		// *inside* the container, that's the container-internal uid
		// (0/root), not the host uid the container's own rootless podman
		// mapping expects. Left unfixed, every restored file/dir shows up
		// as unmapped ("nobody") inside the container - confirmed live:
		// phpBB's own lazily-created cache/production/ dir came back
		// owned by nobody:nogroup after a restore, and every request
		// then fataled with "Unable to write to the cache directory".
		// Re-chowning to the account's real uid after extraction (same
		// as install.go's own post-extract step) fixes it.
		if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
			_ = exec.CommandContext(ctx, "chown", "-R", itoa(uid)+":"+itoa(uid), docrootOnHostOS).Run()
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "restored phpBB files backup from "+backupDatePathInContainer+" on "+selectedDomain, ipAddress)
		restoredItems = append(restoredItems, "files")
	}

	if _, statErr := os.Stat(databaseSQLPathOnHostOS); statErr == nil {
		dbInfo := extractPhpbbDatabaseInfoForBackup(userContext, docroot)
		dbName := dbInfo["database_name"]

		if dbName != "" {
			tableQuery := "SHOW TABLES IN `" + dbName + "`"
			if rows, execErr := mysqlmanager.Exec(ctx, userContext, tableQuery, ""); execErr == nil {
				var tables []string
				for _, row := range rows {
					tables = append(tables, toStringCell(row[0]))
				}
				if len(tables) > 0 {
					var quoted []string
					for _, t := range tables {
						quoted = append(quoted, "`"+t+"`")
					}
					_, _ = mysqlmanager.Exec(ctx, userContext, "DROP TABLE "+strings.Join(quoted, ", "), dbName)
				}
			}
		}

		if dbName != "" {
			mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
			importArgv := podmanmanager.PodmanArgv(userContext, "exec", "-i", mysqlVersion, mysqlVersion, dbName)
			f, openErr := os.Open(databaseSQLPathOnHostOS)
			if openErr != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": openErr.Error()})
				return
			}
			cmd := exec.CommandContext(ctx, importArgv[0], importArgv[1:]...)
			cmd.Stdin = f
			cmd.Env = podmanmanager.PodmanEnv(userContext)
			runErr := cmd.Run()
			f.Close()
			if runErr != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Database import failed: " + runErr.Error()})
				return
			}
			_ = logger.RecordUserAction(a.Config, currentUsername, "restored phpBB database backup from "+databaseSQLPathInContainer+" on "+selectedDomain, ipAddress)
			restoredItems = append(restoredItems, "database")
		}
	}

	if len(restoredItems) > 0 {
		_, _ = w.Write([]byte("Backup restored successfully: " + strings.Join(restoredItems, " and ") + "."))
		return
	}
	_, _ = w.Write([]byte("No files to restore, expected files: " + backupDatePathInContainer + "/files.tar.gz " + databaseSQLPathInContainer + "."))
}

func handlePhpbbRunBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	selectedDomain := r.PathValue("selected_domain")
	docroot := r.URL.Query().Get("docroot")
	if docroot == "" {
		http.Error(w, "Document root is not provided or invalid.", http.StatusInternalServerError)
		return
	}

	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := selectedDomain
	if idx := strings.Index(selectedDomain, "/"); idx != -1 {
		domain = selectedDomain[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	backupDatabase := r.URL.Query().Get("backup_database") == "true"
	backupFiles := r.URL.Query().Get("backup_files") == "true"
	if !backupDatabase && !backupFiles {
		http.Error(w, "No backup options selected.", http.StatusBadRequest)
		return
	}

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	mysqlDumpVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_mysql_dumps/_data/"
	_ = os.MkdirAll(htmlVolume, 0o755)
	_ = os.MkdirAll(mysqlDumpVolume, 0o755)

	uid, uidErr := podmanmanager.GetUID(userContext)
	if uidErr == nil {
		_ = os.Chown(htmlVolume, uid, uid)
		_ = os.Chown(mysqlDumpVolume, uid, uid)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupDirectory := filepath.Join(htmlVolume, "backups", selectedDomain, timestamp)
	inPHPBackupDirectory := filepath.Join("/var/www/html/backups", selectedDomain, timestamp)
	if mkErr := os.MkdirAll(backupDirectory, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
		return
	}
	if uidErr == nil {
		_ = exec.CommandContext(ctx, "chown", itoa(uid)+":"+itoa(uid), backupDirectory).Run()
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	phpContainer := webServer
	if !strings.Contains(strings.ToLower(webServer), "litespeed") {
		phpVersion := php.GetPHPVForDomain(ctx, a, userContext, domain)
		phpContainer = "php-fpm-" + phpVersion
	}

	if backupDatabase {
		dbInfo := extractPhpbbDatabaseInfoForBackup(userContext, docroot)
		dbName := dbInfo["database_name"]
		if dbName == "" {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to determine database name from config.php."})
			return
		}

		mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
		var dumpCmd string
		switch mysqlVersion {
		case "mysql":
			dumpCmd = "mysqldump"
		case "mariadb":
			dumpCmd = "mariadb-dump"
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unsupported MYSQL_TYPE: " + mysqlVersion})
			return
		}

		rows, execErr := mysqlmanager.Exec(ctx, userContext, "SHOW TABLES IN `"+dbName+"`", "")
		if execErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
			return
		}
		var tables []string
		for _, row := range rows {
			tables = append(tables, toStringCell(row[0]))
		}
		if len(tables) == 0 {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "No matching tables found"})
			return
		}

		dumpArgv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, dumpCmd, "-u", "root", dbName)
		dumpArgv = append(dumpArgv, tables...)
		dumpArgv = append(dumpArgv, "--result-file=/tmp/dumps/database.sql")
		if _, runErr := podmanmanager.Command(ctx, userContext, dumpArgv).Output(); runErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": dumpCmd + " failed: " + runErr.Error()})
			return
		}

		mysqlDumpPath := filepath.Join(mysqlDumpVolume, "database.sql")
		if renameErr := os.Rename(mysqlDumpPath, filepath.Join(backupDirectory, "database.sql")); renameErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": renameErr.Error()})
			return
		}
		if uidErr == nil {
			_ = exec.CommandContext(ctx, "chown", itoa(uid)+":"+itoa(uid), backupDirectory).Run()
		}
	}

	if backupFiles {
		tarArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "bash", "-c", "cd "+docroot+" && tar -czf "+inPHPBackupDirectory+"/files.tar.gz .")
		if runErr := podmanmanager.Command(ctx, userContext, tarArgv).Run(); runErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
			return
		}
	}

	var logAction string
	switch {
	case backupDatabase && backupFiles:
		logAction = "generated full backup for phpBB website " + selectedDomain
	case backupDatabase:
		logAction = "generated database backup for phpBB website " + selectedDomain
	case backupFiles:
		logAction = "generated files backup for phpBB website " + selectedDomain
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, logAction, reqip.ClientIP(r))
	_, _ = w.Write([]byte("Backup completed successfully!"))
}
