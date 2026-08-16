package matomo

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// This file mirrors wordpress/backups.go's (and every other CMS module's)
// directory layout, naming and restore/run logic exactly - same
// backups/<domain>/<timestamp>/{database.sql,files.tar.gz} structure under
// the user's html_data volume. Only the DB name/prefix lookup differs:
// extractMatomoDatabaseInfoForBackup reads config/config.ini.php directly.

var matomoBackupFolderRE = regexp.MustCompile(`^20\d{2}-`)

type matomoBackupDateInfo struct {
	Date           string `json:"date"`
	HasDbBackup    bool   `json:"hasDbBackup"`
	HasFilesBackup bool   `json:"hasFilesBackup"`
}

// handleMatomoGetBackupDates mirrors wordpress/backups.go's
// handleGetBackupDates.
func handleMatomoGetBackupDates(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	var dates []matomoBackupDateInfo
	for _, entry := range entries {
		if !entry.IsDir() || !matomoBackupFolderRE.MatchString(entry.Name()) {
			continue
		}
		folderPath := filepath.Join(backupsPath, entry.Name())
		files, _ := os.ReadDir(folderPath)
		info := matomoBackupDateInfo{Date: entry.Name()}
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

var matomoBackupDateRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`)

// handleMatomoRestoreBackup mirrors wordpress/backups.go's
// handleRestoreBackup.
func handleMatomoRestoreBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	if !matomoBackupDateRE.MatchString(backupDate) {
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
		_ = logger.RecordUserAction(a.Config, currentUsername, "restored Matomo files backup from "+backupDatePathInContainer+" on "+selectedDomain, ipAddress)
		restoredItems = append(restoredItems, "files")
	}

	if _, statErr := os.Stat(databaseSQLPathOnHostOS); statErr == nil {
		dbInfo := extractMatomoDatabaseInfoForBackup(userContext, docroot)
		dbName := dbInfo["database_name"]
		tablePrefix := dbInfo["database_prefix"]

		if dbName != "" && tablePrefix != "" {
			if rows, execErr := mysqlmanager.Exec(ctx, userContext, "SHOW TABLES IN `"+dbName+"` LIKE '"+tablePrefix+"%'", ""); execErr == nil {
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
			_ = logger.RecordUserAction(a.Config, currentUsername, "restored Matomo database backup from "+databaseSQLPathInContainer+" on "+selectedDomain, ipAddress)
			restoredItems = append(restoredItems, "database")
		}
	}

	if len(restoredItems) > 0 {
		_, _ = w.Write([]byte("Backup restored successfully: " + strings.Join(restoredItems, " and ") + "."))
		return
	}
	_, _ = w.Write([]byte("No files to restore, expected files: " + backupDatePathInContainer + "/files.tar.gz " + databaseSQLPathInContainer + "."))
}

// handleMatomoRunBackup mirrors wordpress/backups.go's handleRunBackup.
func handleMatomoRunBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		dbInfo := extractMatomoDatabaseInfoForBackup(userContext, docroot)
		dbName, tablePrefix := dbInfo["database_name"], dbInfo["database_prefix"]
		if dbName == "" || tablePrefix == "" {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to determine database name/prefix from config/config.ini.php."})
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

		rows, execErr := mysqlmanager.Exec(ctx, userContext, "SHOW TABLES IN `"+dbName+"` LIKE '"+tablePrefix+"%'", "")
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
		logAction = "generated full backup for Matomo website " + selectedDomain
	case backupDatabase:
		logAction = "generated database backup for Matomo website " + selectedDomain
	case backupFiles:
		logAction = "generated files backup for Matomo website " + selectedDomain
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, logAction, reqip.ClientIP(r))
	_, _ = w.Write([]byte("Backup completed successfully!"))
}
