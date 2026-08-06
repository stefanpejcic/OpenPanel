package wordpress

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

// toStringCell converts one mysqlmanager.Exec() result cell to a string,
// matching how Python's DB-API cursor results stringify implicitly
// wherever the original code did f"{row[i]}" or str(row[i]). The MySQL
// driver scans INTEGER columns (e.g. wp_users.ID) into native int64/uint64,
// not []byte/string - a missing case here isn't a compile error, it's a
// silent empty string, which the WP autologin flow found the hard way:
// wp_users.ID being read as "" -> strconv.Atoi returns 0 -> a
// user_id: 0 in the serialized PHP session data -> PHP's empty(0) is
// true -> the mu-plugin rejects the token as "invalid" even though every
// other part of the flow (token, hashing, DB write, redirect) was correct.
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

var backupFolderRE = regexp.MustCompile(`^20\d{2}-`)

type backupDateInfo struct {
	Date           string `json:"date"`
	HasDbBackup    bool   `json:"hasDbBackup"`
	HasFilesBackup bool   `json:"hasFilesBackup"`
}

// handleGetBackupDates mirrors get_backup_dates().
func handleGetBackupDates(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	var dates []backupDateInfo
	for _, entry := range entries {
		if !entry.IsDir() || !backupFolderRE.MatchString(entry.Name()) {
			continue
		}
		folderPath := filepath.Join(backupsPath, entry.Name())
		files, _ := os.ReadDir(folderPath)
		info := backupDateInfo{Date: entry.Name()}
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

var backupDateRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`)

// handleRestoreBackup mirrors restore_backup().
func handleRestoreBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	selectedDomain := r.PathValue("selected_domain")
	backupDate := r.URL.Query().Get("backup_date")
	docroot := r.URL.Query().Get("docroot")
	phpVersion := r.URL.Query().Get("php_version")

	_, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := selectedDomain
	var subdirectory string
	hasSubdir := false
	if idx := strings.Index(selectedDomain, "/"); idx != -1 {
		domain = selectedDomain[:idx]
		subdirectory = selectedDomain[idx+1:]
		hasSubdir = true
	}

	if phpVersion == "" || docroot == "" {
		dom, found, dbErr := lookupDomainByURL(ctx, a, domain)
		if dbErr != nil {
			writeJSON(w, http.StatusOK, map[string]string{"error": "An error occurred: " + dbErr.Error()})
			return
		}
		if !found {
			writeJSON(w, http.StatusOK, map[string]string{"error": "Domain not found."})
			return
		}
		domain = dom.DomainURL
		docroot = strings.TrimPrefix(dom.Docroot.String, "/var/www/html/")
		docroot = strings.TrimPrefix(docroot, "/")
		phpVersion = dom.PHPVersion.String
	}

	docrootOnHostOS := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + docroot + "/"
	docrootInContainer := "/var/www/html/" + docroot + "/"
	backupsPathOnHostOS := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/backups/" + docroot + "/"
	backupsPathInContainer := "/var/www/html/backups/" + docroot + "/"

	if hasSubdir {
		docrootOnHostOS += subdirectory
		docrootInContainer += subdirectory
		backupsPathOnHostOS += subdirectory
		backupsPathInContainer += subdirectory
	}

	if !backupDateRE.MatchString(backupDate) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup date."})
		return
	}

	backupDatePathOnHostOS := filepath.Join(backupsPathOnHostOS, backupDate)
	backupDatePathInContainer := filepath.Join(backupsPathInContainer, backupDate)
	databaseSQLPathOnHostOS := filepath.Join(backupDatePathOnHostOS, "database.sql")
	databaseSQLPathInContainer := filepath.Join(backupDatePathInContainer, "database.sql")
	targzPathOnHostOS := filepath.Join(backupDatePathOnHostOS, "files.tar.gz")

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	ipAddress := reqip.ClientIP(r)

	var restoredItems []string

	if _, statErr := os.Stat(targzPathOnHostOS); statErr == nil {
		if runErr := exec.CommandContext(ctx, "tar", "-xzf", targzPathOnHostOS, "-C", docrootOnHostOS).Run(); runErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "restored WordPress files backup from "+backupDatePathInContainer+" on "+selectedDomain, ipAddress)
		restoredItems = append(restoredItems, "files")
	}

	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	if isLitespeed {
		if _, statErr := os.Stat(databaseSQLPathOnHostOS); statErr == nil {
			wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, webServer)
			resetArgv := append(append([]string{}, wpBase...), "db", "reset", "--yes", "--path="+docrootInContainer, "--allow-root", "--skip-themes")
			_ = podmanmanager.Command(ctx, userContext, resetArgv).Run()
			importArgv := append(append([]string{}, wpBase...), "db", "import", databaseSQLPathInContainer, "--path="+docrootInContainer, "--allow-root", "--skip-themes")
			if runErr := podmanmanager.Command(ctx, userContext, importArgv).Run(); runErr == nil {
				_ = logger.RecordUserAction(a.Config, currentUsername, "restored WordPress database backup from "+databaseSQLPathInContainer+" on "+selectedDomain, ipAddress)
				restoredItems = append(restoredItems, "database")
			}
		}
	} else if _, statErr := os.Stat(databaseSQLPathOnHostOS); statErr == nil {
		wpConfigFile := filepath.Join(docrootOnHostOS, "wp-config.php")
		content, readErr := os.ReadFile(wpConfigFile)
		if readErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": wpConfigFile + " does not exist - failed to retrieve db_name and table prefix."})
			return
		}
		dbNameMatch := dbNameRE.FindStringSubmatch(string(content))
		tablePrefixMatch := tablePrefixValueRE.FindStringSubmatch(string(content))
		var dbName, tablePrefix string
		if dbNameMatch != nil {
			dbName = dbNameMatch[1]
		}
		if tablePrefixMatch != nil {
			tablePrefix = tablePrefixMatch[1]
		}

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
			_ = logger.RecordUserAction(a.Config, currentUsername, "restored WordPress database backup from "+databaseSQLPathInContainer+" on "+selectedDomain, ipAddress)
			restoredItems = append(restoredItems, "database")
		}
	}

	if len(restoredItems) > 0 {
		_, _ = w.Write([]byte("Backup restored successfully: " + strings.Join(restoredItems, " and ") + "."))
		return
	}
	_, _ = w.Write([]byte("No files to restore, expected files: " + backupDatePathInContainer + "/files.tar.gz " + databaseSQLPathInContainer + "."))
}

var (
	dbNameRE           = regexp.MustCompile(`define\(\s*['"]DB_NAME['"]\s*,\s*['"]([^'"]+)['"]`)
	tablePrefixValueRE = regexp.MustCompile(`\$table_prefix\s*=\s*['"]([^'"]+)['"]`)
)

// handleRunBackup mirrors run_wordpress_backup().
func handleRunBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	selectedDomain := r.PathValue("selected_domain")
	docroot := r.URL.Query().Get("docroot")
	if docroot == "" {
		http.Error(w, "Document root is not provided or invalid.", http.StatusInternalServerError)
		return
	}

	_, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
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
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpContainer := webServer
	if !isLitespeed {
		domainOnly := strings.Split(selectedDomain, "/")[0]
		phpVersion := php.GetPHPVForDomain(ctx, a, userContext, domainOnly)
		phpContainer = "php-fpm-" + phpVersion
	}

	if backupDatabase {
		wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)

		prefixArgv := append(append([]string{}, wpBase...), "config", "get", "table_prefix", "--path="+docroot, "--skip-themes", "--allow-root")
		prefixOut, prefixErr := podmanmanager.Command(ctx, userContext, prefixArgv).Output()
		tablePrefix := strings.TrimSpace(string(prefixOut))
		if prefixErr != nil || tablePrefix == "" {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve table prefix."})
			return
		}

		dbNameArgv := append(append([]string{}, wpBase...), "config", "get", "DB_NAME", "--path="+docroot, "--skip-themes", "--allow-root")
		dbNameOut, dbNameErr := podmanmanager.Command(ctx, userContext, dbNameArgv).Output()
		dbName := strings.TrimSpace(string(dbNameOut))
		if dbNameErr != nil || dbName == "" {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve db_name from wp-cli."})
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
		logAction = "generated full backup for WordPress website " + selectedDomain
	case backupDatabase:
		logAction = "generated database backup for WordPress website " + selectedDomain
	case backupFiles:
		logAction = "generated files backup for WordPress website " + selectedDomain
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, logAction, reqip.ClientIP(r))
	_, _ = w.Write([]byte("Backup completed successfully!"))
}
