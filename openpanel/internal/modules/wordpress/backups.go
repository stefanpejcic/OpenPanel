package wordpress

import (
	"context"
	"errors"
	"net/http"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var backupFolderRE = regexp.MustCompile(`^20\d{2}-`)

// validSiteName checks a domain or domain/folder taken from the URL, since it ends up in host paths and wp-cli/tar args
func validSiteName(site string) bool {
	domain, folder, _ := strings.Cut(site, "/")
	if !validateDomain(domain) {
		return false
	}
	_, err := paths.CleanSubdir(folder)
	return err == nil
}

type backupDateInfo struct {
	Date           string `json:"date"`
	HasDbBackup    bool   `json:"hasDbBackup"`
	HasFilesBackup bool   `json:"hasFilesBackup"`
	Size           int64  `json:"size"`
	Auto           bool   `json:"auto"`
}

// handleGetBackupDates mirrors get_backup_dates().
func handleGetBackupDates(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	selectedDomain := r.PathValue("selected_domain")

	userID, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !a.CheckDomainBelongsToUser(r.Context(), userID, strings.Split(selectedDomain, "/")[0]) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	if !validSiteName(selectedDomain) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid domain."})
		return
	}

	appkit.ProtectBackupsDir("/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data")
	backupsPath := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/backups/" + selectedDomain
	if _, statErr := tenantfs.Stat(backupsPath); statErr != nil {
		if mkErr := tenantfs.MkdirAll(backupsPath, 0o755); mkErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": mkErr.Error()})
			return
		}
	}

	entries, readErr := tenantfs.ReadDir(backupsPath)
	if readErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		return
	}

	var dates []backupDateInfo
	for _, entry := range entries {
		if !entry.IsDir() || !backupFolderRE.MatchString(entry.Name()) {
			continue
		}
		folderPath := filepath.Join(backupsPath, entry.Name())
		files, _ := tenantfs.ReadDir(folderPath)
		info := backupDateInfo{Date: entry.Name()}
		for _, f := range files {
			if fi, infoErr := f.Info(); infoErr == nil && fi.Mode().IsRegular() {
				info.Size += fi.Size()
			}
			if strings.HasSuffix(f.Name(), ".sql") {
				info.HasDbBackup = true
			}
			if strings.HasSuffix(f.Name(), ".tar.gz") {
				info.HasFilesBackup = true
			}
			if f.Name() == autoUpdateMarker {
				info.Auto = true
			}
		}
		dates = append(dates, info)
	}

	web.WriteJSON(w, http.StatusOK, dates)
}

var backupDateRE = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}$`)

// handleRestoreBackup mirrors restore_backup().
func handleRestoreBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	selectedDomain := r.PathValue("selected_domain")
	backupDate := r.URL.Query().Get("backup_date")
	docroot := r.URL.Query().Get("docroot")
	phpVersion := r.URL.Query().Get("php_version")

	userID, currentUsername, userContext, err := auth.Injected(a, r)
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
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	if !validSiteName(selectedDomain) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid domain."})
		return
	}
	// a caller-supplied docroot is only used together with php_version, it gets joined into host paths so it must stay inside the html volume
	if phpVersion != "" && docroot != "" {
		if !appkit.SafeDocroot("/var/www/html/" + strings.TrimPrefix(strings.TrimPrefix(docroot, "/var/www/html/"), "/")) {
			web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid docroot"})
			return
		}
		docroot = strings.TrimPrefix(strings.TrimPrefix(docroot, "/var/www/html/"), "/")
	}

	if phpVersion == "" || docroot == "" {
		dom, found, dbErr := lookupDomainByURL(ctx, a, domain)
		if dbErr != nil {
			web.WriteJSON(w, http.StatusOK, map[string]string{"error": "An error occurred: " + dbErr.Error()})
			return
		}
		if !found {
			web.WriteJSON(w, http.StatusOK, map[string]string{"error": "Domain not found."})
			return
		}
		domain = dom.DomainURL
		docroot = strings.TrimPrefix(dom.Docroot.String, "/var/www/html/")
		docroot = strings.TrimPrefix(docroot, "/")
		phpVersion = dom.PHPVersion.String
	}

	docrootOnHostOS := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + docroot + "/"
	docrootInContainer := "/var/www/html/" + docroot + "/"
	// run and get_dates keep backups under the site name, going by the docroot folder missed them whenever it wasn't named after the domain
	backupsPathOnHostOS := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/backups/" + selectedDomain + "/"
	backupsPathInContainer := "/var/www/html/backups/" + selectedDomain + "/"

	if hasSubdir {
		docrootOnHostOS += subdirectory
		docrootInContainer += subdirectory
	}

	if !backupDateRE.MatchString(backupDate) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup date."})
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

	if _, statErr := tenantfs.Stat(targzPathOnHostOS); statErr == nil {
		if runErr := tenantfs.ExtractTar(ctx, targzPathOnHostOS, docrootOnHostOS, 0); runErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": runErr.Error()})
			return
		}
		// files the restore recreates come out root-owned, and WordPress can't update or delete those later
		if uid, uidErr := podmanmanager.GetUID(userContext); uidErr == nil {
			_ = tenantfs.ChownR(docrootOnHostOS, uid, uid)
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "restored WordPress files backup from "+backupDatePathInContainer+" on "+selectedDomain, ipAddress)
		restoredItems = append(restoredItems, "files")
	}

	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	if isLitespeed {
		if _, statErr := tenantfs.Stat(databaseSQLPathOnHostOS); statErr == nil {
			wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, webServer)
			resetArgv := append(append([]string{}, wpBase...), "db", "reset", "--yes", "--path="+docrootInContainer, "--allow-root", "--skip-themes")
			_ = podmanmanager.Command(ctx, userContext, resetArgv).Run()
			importArgv := append(append([]string{}, wpBase...), "db", "import", databaseSQLPathInContainer, "--path="+docrootInContainer, "--allow-root", "--skip-themes")
			if runErr := podmanmanager.Command(ctx, userContext, importArgv).Run(); runErr == nil {
				_ = logger.RecordUserAction(a.Config, currentUsername, "restored WordPress database backup from "+databaseSQLPathInContainer+" on "+selectedDomain, ipAddress)
				restoredItems = append(restoredItems, "database")
			}
		}
	} else if _, statErr := tenantfs.Stat(databaseSQLPathOnHostOS); statErr == nil {
		wpConfigFile := filepath.Join(docrootOnHostOS, "wp-config.php")
		content, readErr := tenantfs.ReadFile(wpConfigFile)
		if readErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": wpConfigFile + " does not exist - failed to retrieve db_name and table prefix."})
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
		// both come from wp-config.php and go into SQL run with the account's mysql root login
		if (dbName != "" && !validWPConfigName(dbName)) || (tablePrefix != "" && !tablePrefixSafeRE.MatchString(tablePrefix)) {
			web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "wp-config.php has an unsupported DB_NAME or table prefix."})
			return
		}

		if dbName != "" && tablePrefix != "" {
			if rows, execErr := mysqlmanager.Exec(ctx, userContext, "SHOW TABLES IN `"+dbName+"` LIKE '"+tablePrefix+"%'", ""); execErr == nil {
				var tables []string
				for _, row := range rows {
					tables = append(tables, mysqlmanager.ToString(row[0]))
				}
				if len(tables) > 0 {
					var quoted []string
					for _, t := range tables {
						quoted = append(quoted, "`"+strings.ReplaceAll(t, "`", "``")+"`")
					}
					_, _ = mysqlmanager.Exec(ctx, userContext, "DROP TABLE "+strings.Join(quoted, ", "), dbName)
				}
			}
		}

		if dbName != "" {
			mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
			importArgv := podmanmanager.PodmanArgv(userContext, "exec", "-i", mysqlVersion, mysqlVersion, dbName)
			f, openErr := tenantfs.Open(databaseSQLPathOnHostOS)
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

// filesBackupArgv tars docroot without a shell, the site name in the backup path used to go through bash -c
func filesBackupArgv(userContext, phpContainer, docroot, backupDir string) []string {
	return podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "tar", "-czf", backupDir+"/files.tar.gz", "-C", docroot, ".")
}

// handleRunBackup mirrors run_wordpress_backup().
func handleRunBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	selectedDomain := r.PathValue("selected_domain")
	docroot := r.URL.Query().Get("docroot")
	if docroot == "" || !appkit.SafeDocroot(docroot) {
		http.Error(w, "Document root is not provided or invalid.", http.StatusInternalServerError)
		return
	}

	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, strings.Split(selectedDomain, "/")[0]) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	if !validSiteName(selectedDomain) {
		http.Error(w, "Invalid domain.", http.StatusBadRequest)
		return
	}

	backupDatabase := r.URL.Query().Get("backup_database") == "true"
	backupFiles := r.URL.Query().Get("backup_files") == "true"

	if !backupDatabase && !backupFiles {
		http.Error(w, "No backup options selected.", http.StatusBadRequest)
		return
	}

	if _, runErr := runWPBackup(ctx, a, userContext, selectedDomain, docroot, backupDatabase, backupFiles); runErr != nil {
		web.WriteJSON(w, backupErrStatus(runErr), map[string]string{"error": runErr.Error()})
		return
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

// wpBackupError keeps the status code handleRunBackup has always answered a failed step with
type wpBackupError struct {
	status int
	msg    string
}

func (e *wpBackupError) Error() string { return e.msg }

func backupFail(status int, msg string) error { return &wpBackupError{status: status, msg: msg} }

func backupErrStatus(err error) int {
	var be *wpBackupError
	if errors.As(err, &be) {
		return be.status
	}
	return http.StatusInternalServerError
}

// wpBackupSlots allows one backup per account at a time, every dump goes through the same /tmp/dumps/database.sql
var wpBackupSlots sync.Map

func lockWPBackups(ctx context.Context, userContext string) (func(), error) {
	v, _ := wpBackupSlots.LoadOrStore(userContext, make(chan struct{}, 1))
	slot := v.(chan struct{})
	select {
	case slot <- struct{}{}:
		return func() { <-slot }, nil
	case <-ctx.Done():
		return nil, backupFail(http.StatusServiceUnavailable, "Another backup is still running, try again in a moment.")
	}
}

// wpBackupsRoot is the host folder that holds site's dated backups
func wpBackupsRoot(userContext, site string) string {
	return filepath.Join("/home", userContext, "docker-data", "volumes", userContext+"_html_data", "_data", "backups", site)
}

// runWPBackup backs up site's database and/or files into backups/<site>/<timestamp> and returns that folder's host path, also when a later step failed after it was made
func runWPBackup(ctx context.Context, a *appctx.App, userContext, site, docroot string, backupDatabase, backupFiles bool) (string, error) {
	if !backupDatabase && !backupFiles {
		return "", backupFail(http.StatusBadRequest, "No backup options selected.")
	}
	unlock, lockErr := lockWPBackups(ctx, userContext)
	if lockErr != nil {
		return "", lockErr
	}
	defer unlock()

	htmlVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	mysqlDumpVolume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_mysql_dumps/_data/"
	_ = tenantfs.MkdirAll(htmlVolume, 0o755)
	_ = tenantfs.MkdirAll(mysqlDumpVolume, 0o755)
	appkit.ProtectBackupsDir(htmlVolume)

	uid, uidErr := podmanmanager.GetUID(userContext)
	if uidErr == nil {
		_ = tenantfs.Lchown(htmlVolume, uid, uid)
		_ = tenantfs.Lchown(mysqlDumpVolume, uid, uid)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupDirectory := filepath.Join(htmlVolume, "backups", site, timestamp)
	inPHPBackupDirectory := filepath.Join("/var/www/html/backups", site, timestamp)
	if mkErr := tenantfs.MkdirAll(backupDirectory, 0o755); mkErr != nil {
		return "", backupFail(http.StatusInternalServerError, mkErr.Error())
	}
	if uidErr == nil {
		_ = tenantfs.Lchown(backupDirectory, uid, uid)
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	phpContainer := webServer
	if !isLitespeed {
		domainOnly := strings.Split(site, "/")[0]
		phpVersion := php.GetPHPVForDomain(ctx, a, userContext, domainOnly)
		phpContainer = "php-fpm-" + phpVersion
	}

	if backupDatabase {
		wpBase := podmanmanager.BuildWPCLIBaseCommand(userContext, phpContainer)

		prefixArgv := append(append([]string{}, wpBase...), "config", "get", "table_prefix", "--path="+docroot, "--skip-themes", "--allow-root")
		prefixOut, prefixErr := podmanmanager.Command(ctx, userContext, prefixArgv).Output()
		tablePrefix := strings.TrimSpace(string(prefixOut))
		if prefixErr != nil || tablePrefix == "" {
			return backupDirectory, backupFail(http.StatusInternalServerError, "Failed to retrieve table prefix.")
		}

		dbNameArgv := append(append([]string{}, wpBase...), "config", "get", "DB_NAME", "--path="+docroot, "--skip-themes", "--allow-root")
		dbNameOut, dbNameErr := podmanmanager.Command(ctx, userContext, dbNameArgv).Output()
		dbName := strings.TrimSpace(string(dbNameOut))
		if dbNameErr != nil || dbName == "" {
			return backupDirectory, backupFail(http.StatusInternalServerError, "Failed to retrieve db_name from wp-cli.")
		}
		if !validWPConfigName(dbName) || !tablePrefixSafeRE.MatchString(tablePrefix) {
			return backupDirectory, backupFail(http.StatusBadRequest, "wp-config.php has an unsupported DB_NAME or table prefix.")
		}

		mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
		var dumpCmd string
		switch mysqlVersion {
		case "mysql":
			dumpCmd = "mysqldump"
		case "mariadb":
			dumpCmd = "mariadb-dump"
		default:
			return backupDirectory, backupFail(http.StatusInternalServerError, "Unsupported MYSQL_TYPE: "+mysqlVersion)
		}

		rows, execErr := mysqlmanager.Exec(ctx, userContext, "SHOW TABLES IN `"+dbName+"` LIKE '"+tablePrefix+"%'", "")
		if execErr != nil {
			return backupDirectory, backupFail(http.StatusInternalServerError, execErr.Error())
		}
		var tables []string
		for _, row := range rows {
			tables = append(tables, mysqlmanager.ToString(row[0]))
		}
		if len(tables) == 0 {
			return backupDirectory, backupFail(http.StatusInternalServerError, "No matching tables found")
		}

		dumpArgv := podmanmanager.PodmanArgv(userContext, "exec", mysqlVersion, dumpCmd, "-u", "root", "--single-transaction", dbName)
		dumpArgv = append(dumpArgv, tables...)
		dumpArgv = append(dumpArgv, "--result-file=/tmp/dumps/database.sql")
		if _, runErr := podmanmanager.Command(ctx, userContext, dumpArgv).Output(); runErr != nil {
			return backupDirectory, backupFail(http.StatusInternalServerError, dumpCmd+" failed: "+runErr.Error())
		}

		mysqlDumpPath := filepath.Join(mysqlDumpVolume, "database.sql")
		if renameErr := tenantfs.Rename(mysqlDumpPath, filepath.Join(backupDirectory, "database.sql")); renameErr != nil {
			return backupDirectory, backupFail(http.StatusInternalServerError, renameErr.Error())
		}
		if uidErr == nil {
			_ = tenantfs.Lchown(backupDirectory, uid, uid)
		}
	}

	if backupFiles {
		tarArgv := filesBackupArgv(userContext, phpContainer, docroot, inPHPBackupDirectory)
		if runErr := podmanmanager.Command(ctx, userContext, tarArgv).Run(); runErr != nil {
			return backupDirectory, backupFail(http.StatusInternalServerError, runErr.Error())
		}
	}
	return backupDirectory, nil
}
