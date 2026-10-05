package cmsapp

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
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// same backups/<domain>/<timestamp>/{database.sql,files.tar.gz} layout as wordpress/backups.go

var (
	backupFolderRE = regexp.MustCompile(`^20\d{2}-`)
	backupDateRE   = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`)
)

type backupDateInfo struct {
	Date           string `json:"date"`
	HasDbBackup    bool   `json:"hasDbBackup"`
	HasFilesBackup bool   `json:"hasFilesBackup"`
}

type filesBackupDateInfo struct {
	Date           string `json:"date"`
	HasFilesBackup bool   `json:"hasFilesBackup"`
}

func htmlVolume(userContext string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
}

// ownsSelectedDomain checks the domain part of "domain.tld/sub" belongs to the user
func ownsSelectedDomain(a *appctx.App, r *http.Request, userID int, selectedDomain string) bool {
	domain := selectedDomain
	if idx := strings.Index(selectedDomain, "/"); idx != -1 {
		domain = selectedDomain[:idx]
	}
	return a.CheckDomainBelongsToUser(r.Context(), userID, domain)
}

func phpContainerFor(a *appctx.App, r *http.Request, userContext, selectedDomain string) string {
	domain := selectedDomain
	if idx := strings.Index(selectedDomain, "/"); idx != -1 {
		domain = selectedDomain[:idx]
	}
	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		return webServer
	}
	return "php-fpm-" + php.GetPHPVForDomain(r.Context(), a, userContext, domain)
}

func (app *App) HandleGetBackupDates(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	selectedDomain := r.PathValue("selected_domain")

	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	backupsPath := htmlVolume(userContext) + "backups/" + selectedDomain
	if _, statErr := os.Stat(backupsPath); statErr != nil {
		if mkErr := os.MkdirAll(backupsPath, 0o755); mkErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
			return
		}
	}

	entries, readErr := os.ReadDir(backupsPath)
	if readErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}

	var dates []backupDateInfo
	for _, entry := range entries {
		if !entry.IsDir() || !backupFolderRE.MatchString(entry.Name()) {
			continue
		}
		files, _ := os.ReadDir(filepath.Join(backupsPath, entry.Name()))
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

	// files-only apps never had the db flag in their response
	if app.DBMode == DBNone {
		var filesOnly []filesBackupDateInfo
		for _, d := range dates {
			filesOnly = append(filesOnly, filesBackupDateInfo{Date: d.Date, HasFilesBackup: d.HasFilesBackup})
		}
		web.WriteJSON(w, http.StatusOK, filesOnly)
		return
	}
	web.WriteJSON(w, http.StatusOK, dates)
}

// matchingTables lists the site's tables, prefix-filtered unless the app dumps every table
func (app *App) matchingTables(r *http.Request, userContext, dbName, tablePrefix string) ([]string, error) {
	query := "SHOW TABLES IN `" + dbName + "`"
	if app.DBMode == DBPrefixed || tablePrefix != "" {
		query += " LIKE '" + tablePrefix + "%'"
	}
	rows, err := mysqlmanager.Exec(r.Context(), userContext, query, "")
	if err != nil {
		return nil, err
	}
	var tables []string
	for _, row := range rows {
		tables = append(tables, mysqlmanager.ToString(row[0]))
	}
	return tables, nil
}

func (app *App) HandleRestoreBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	selectedDomain := r.PathValue("selected_domain")
	backupDate := r.URL.Query().Get("backup_date")
	docroot := r.URL.Query().Get("docroot")

	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !ownsSelectedDomain(a, r, userID, selectedDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if app.FilesDir == nil {
		if docroot == "" {
			docroot = "/var/www/html/" + selectedDomain
		}
		if !appkit.SafeDocroot(docroot) {
			web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid docroot"})
			return
		}
	}

	if !backupDateRE.MatchString(backupDate) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup date."})
		return
	}

	filesDir := docroot
	if app.FilesDir != nil {
		filesDir = app.FilesDir(selectedDomain)
	}
	filesHostPath := htmlVolume(userContext) + strings.TrimPrefix(filesDir, "/var/www/html/") + "/"
	backupsPathOnHostOS := htmlVolume(userContext) + "backups/" + selectedDomain + "/"
	backupsPathInContainer := "/var/www/html/backups/" + selectedDomain + "/"

	backupDatePathOnHostOS := filepath.Join(backupsPathOnHostOS, backupDate)
	backupDatePathInContainer := filepath.Join(backupsPathInContainer, backupDate)
	targzPathOnHostOS := filepath.Join(backupDatePathOnHostOS, "files.tar.gz")
	ipAddress := reqip.ClientIP(r)

	if app.DBMode == DBNone {
		if _, statErr := os.Stat(targzPathOnHostOS); statErr != nil {
			_, _ = w.Write([]byte("No files to restore, expected file: " + backupDatePathInContainer + "/files.tar.gz."))
			return
		}
		if runErr := exec.CommandContext(ctx, "tar", "-xzf", targzPathOnHostOS, "-C", filesHostPath).Run(); runErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "restored "+app.Name+" files backup from "+backupDatePathInContainer+" on "+selectedDomain, ipAddress)
		_, _ = w.Write([]byte("Backup restored successfully: files."))
		return
	}

	databaseSQLPathOnHostOS := filepath.Join(backupDatePathOnHostOS, "database.sql")
	databaseSQLPathInContainer := filepath.Join(backupDatePathInContainer, "database.sql")
	var restoredItems []string

	if _, statErr := os.Stat(targzPathOnHostOS); statErr == nil {
		if runErr := exec.CommandContext(ctx, "tar", "-xzf", targzPathOnHostOS, "-C", filesHostPath).Run(); runErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
			return
		}
		if app.ChownAfterRestore {
			if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
				_ = exec.CommandContext(ctx, "chown", "-R", strconv.Itoa(uid)+":"+strconv.Itoa(uid), filesHostPath).Run()
			}
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "restored "+app.Name+" "+app.filesLabel()+" backup from "+backupDatePathInContainer+" on "+selectedDomain, ipAddress)
		restoredItems = append(restoredItems, "files")
	}

	if _, statErr := os.Stat(databaseSQLPathOnHostOS); statErr == nil {
		dbInfo := app.DBInfo(userContext, docroot, selectedDomain)
		dbName := dbInfo["database_name"]
		tablePrefix := dbInfo["database_prefix"]

		// clear the old tables first so ones added after the backup don't linger
		dropFirst := dbName != "" && (app.DBMode == DBOptionalPrefix || (app.DBMode == DBPrefixed && tablePrefix != ""))
		if dropFirst {
			if tables, execErr := app.matchingTables(r, userContext, dbName, tablePrefix); execErr == nil && len(tables) > 0 {
				var quoted []string
				for _, t := range tables {
					quoted = append(quoted, "`"+t+"`")
				}
				_, _ = mysqlmanager.Exec(ctx, userContext, "DROP TABLE "+strings.Join(quoted, ", "), dbName)
			}
		}

		if dbName != "" {
			mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
			importArgv := podmanmanager.PodmanArgv(userContext, "exec", "-i", mysqlVersion, mysqlVersion, dbName)
			f, openErr := os.Open(databaseSQLPathOnHostOS)
			if openErr != nil {
				web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": openErr.Error()})
				return
			}
			cmd := exec.CommandContext(ctx, importArgv[0], importArgv[1:]...)
			cmd.Stdin = f
			cmd.Env = podmanmanager.PodmanEnv(userContext)
			runErr := cmd.Run()
			f.Close()
			if runErr != nil {
				web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Database import failed: " + runErr.Error()})
				return
			}
			_ = logger.RecordUserAction(a.Config, currentUsername, "restored "+app.Name+" database backup from "+databaseSQLPathInContainer+" on "+selectedDomain, ipAddress)
			restoredItems = append(restoredItems, "database")
		}
	}

	if len(restoredItems) > 0 {
		_, _ = w.Write([]byte("Backup restored successfully: " + strings.Join(restoredItems, " and ") + "."))
		return
	}
	_, _ = w.Write([]byte("No files to restore, expected files: " + backupDatePathInContainer + "/files.tar.gz " + databaseSQLPathInContainer + "."))
}

func (app *App) HandleRunBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	selectedDomain := r.PathValue("selected_domain")
	docroot := r.URL.Query().Get("docroot")
	if app.FilesDir == nil && (docroot == "" || !appkit.SafeDocroot(docroot)) {
		http.Error(w, "Document root is not provided or invalid.", http.StatusInternalServerError)
		return
	}

	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !ownsSelectedDomain(a, r, userID, selectedDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	backupDatabase := app.DBMode != DBNone && r.URL.Query().Get("backup_database") == "true"
	backupFiles := r.URL.Query().Get("backup_files") == "true"
	if !backupDatabase && !backupFiles {
		http.Error(w, "No backup options selected.", http.StatusBadRequest)
		return
	}

	htmlVol := htmlVolume(userContext)
	mysqlDumpVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_mysql_dumps/_data/"
	_ = os.MkdirAll(htmlVol, 0o755)
	if app.DBMode != DBNone {
		_ = os.MkdirAll(mysqlDumpVolume, 0o755)
	}

	uid, uidErr := podmanmanager.GetUID(userContext)
	if uidErr == nil {
		_ = os.Chown(htmlVol, uid, uid)
		if app.DBMode != DBNone {
			_ = os.Chown(mysqlDumpVolume, uid, uid)
		}
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupDirectory := filepath.Join(htmlVol, "backups", selectedDomain, timestamp)
	inPHPBackupDirectory := filepath.Join("/var/www/html/backups", selectedDomain, timestamp)
	if mkErr := os.MkdirAll(backupDirectory, 0o755); mkErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
		return
	}
	if uidErr == nil {
		_ = exec.CommandContext(ctx, "chown", strconv.Itoa(uid)+":"+strconv.Itoa(uid), backupDirectory).Run()
	}

	phpContainer := phpContainerFor(a, r, userContext, selectedDomain)

	if backupDatabase {
		dbInfo := app.DBInfo(userContext, docroot, selectedDomain)
		dbName, tablePrefix := dbInfo["database_name"], dbInfo["database_prefix"]
		if app.DBMode == DBPrefixed && (dbName == "" || tablePrefix == "") {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to determine database name/prefix from " + app.ConfigFile + "."})
			return
		}
		if dbName == "" {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to determine database name from " + app.ConfigFile + "."})
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
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unsupported MYSQL_TYPE: " + mysqlVersion})
			return
		}

		dumpArgv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, dumpCmd, "-u", "root", "--single-transaction", dbName)
		if app.DBMode != DBWhole {
			tables, execErr := app.matchingTables(r, userContext, dbName, tablePrefix)
			if execErr != nil {
				web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": execErr.Error()})
				return
			}
			if len(tables) == 0 {
				web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "No matching tables found"})
				return
			}
			dumpArgv = append(dumpArgv, tables...)
		}
		dumpArgv = append(dumpArgv, "--result-file=/tmp/dumps/database.sql")
		if _, runErr := podmanmanager.Command(ctx, userContext, dumpArgv).Output(); runErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": dumpCmd + " failed: " + runErr.Error()})
			return
		}

		mysqlDumpPath := filepath.Join(mysqlDumpVolume, "database.sql")
		if renameErr := os.Rename(mysqlDumpPath, filepath.Join(backupDirectory, "database.sql")); renameErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": renameErr.Error()})
			return
		}
		if uidErr == nil {
			_ = exec.CommandContext(ctx, "chown", strconv.Itoa(uid)+":"+strconv.Itoa(uid), backupDirectory).Run()
		}
	}

	if backupFiles {
		filesDir := docroot
		if app.FilesDir != nil {
			filesDir = app.FilesDir(selectedDomain)
		}
		tarArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "bash", "-c", "cd "+filesDir+" && tar -czf "+inPHPBackupDirectory+"/files.tar.gz .")
		if runErr := podmanmanager.Command(ctx, userContext, tarArgv).Run(); runErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
			return
		}
	}

	var logAction string
	switch {
	case backupDatabase && backupFiles:
		logAction = "generated full backup for " + app.Name + " website " + selectedDomain
	case backupDatabase:
		logAction = "generated database backup for " + app.Name + " website " + selectedDomain
	case backupFiles:
		logAction = "generated " + app.filesLabel() + " backup for " + app.Name + " website " + selectedDomain
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, logAction, reqip.ClientIP(r))
	_, _ = w.Write([]byte("Backup completed successfully!"))
}
