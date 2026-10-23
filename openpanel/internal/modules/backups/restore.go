package backups

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// restoreDump recreates dbName and pipes the dump into the engine's own client inside the user's container, so a big dump never sits in memory and COPY/DELIMITER blocks restore fine
func restoreDump(ctx context.Context, userContext, engine, dbName string, dump io.Reader) error {
	var argv []string
	if engine == "postgres" {
		safeDB := strings.ReplaceAll(dbName, `"`, `""`)
		if _, err := postgresmanager.Exec(ctx, userContext, `DROP DATABASE IF EXISTS "`+safeDB+`"`, ""); err != nil {
			return err
		}
		if _, err := postgresmanager.Exec(ctx, userContext, `CREATE DATABASE "`+safeDB+`"`, ""); err != nil {
			return err
		}
		argv = podmanmanager.PodmanArgv(userContext, "exec", "-i", "postgres", "psql", "-U", "postgres", "-X", "-q", "-d", dbName)
	} else {
		safeDB := strings.ReplaceAll(dbName, "`", "``")
		if _, err := mysqlmanager.Exec(ctx, userContext, "DROP DATABASE IF EXISTS `"+safeDB+"`", ""); err != nil {
			return err
		}
		if _, err := mysqlmanager.Exec(ctx, userContext, "CREATE DATABASE `"+safeDB+"`", ""); err != nil {
			return err
		}
		// container and client binary share the name, mysql or mariadb
		client := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
		if client == "" {
			client = "mysql"
		}
		argv = podmanmanager.PodmanArgv(userContext, "exec", "-i", client, client, dbName)
	}

	cmd := podmanmanager.Command(ctx, userContext, argv)
	cmd.Stdin = dump
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			lines := strings.Split(msg, "\n")
			return fmt.Errorf("restoring database %s: %s", dbName, lines[len(lines)-1])
		}
		return fmt.Errorf("restoring database %s: %w", dbName, err)
	}
	return nil
}

// dumpDBName turns backup/<engine>/<db>.sql(.gz) into the database name, minus the timestamp suffix some dumps carry
func dumpDBName(member string) (string, bool) {
	base := filepath.Base(member)
	name := strings.TrimSuffix(strings.TrimSuffix(base, ".sql.gz"), ".sql")
	if name == base {
		return "", false
	}
	name = dbTimestampSuffixRE.ReplaceAllString(name, "")
	return name, name != ""
}

// dumpEngine returns the engine of a dump under backup/mysql/ or backup/postgres/ (canonical names)
func dumpEngine(name string) (string, bool) {
	for _, engine := range []string{"mysql", "postgres"} {
		if strings.HasPrefix(name, "backup/"+engine+"/") {
			return engine, true
		}
	}
	return "", false
}

// restoreDumpInto is swapped in tests, the real one needs the account's db container
var restoreDumpInto = restoreDump

// restoreDatabasesFromTar restores every dump in the archive (or only the one named only) straight off the stream, returns the restored database names
// with only set it stops right after that dump or once the dump dirs are past, on archives with 01- dump dirs that's before html/
func restoreDatabasesFromTar(ctx context.Context, userContext string, archive io.Reader, only string) ([]string, error) {
	gz, err := gzip.NewReader(archive)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var restored []string
	layout := newArchiveLayout()
	dumpDirs := []string{"backup/mysql", "backup/postgres"}
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return restored, err
		}

		raw := strings.Trim(header.Name, "/")
		name := layout.observe(raw)
		if only != "" && layout.pastAll(raw, dumpDirs) {
			break
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		engine, ok := dumpEngine(name)
		if !ok {
			continue
		}
		dbName, ok := dumpDBName(name)
		if !ok || (only != "" && dbName != only) {
			continue
		}

		var dump io.Reader = tr
		if strings.HasSuffix(name, ".gz") {
			zr, err := gzip.NewReader(tr)
			if err != nil {
				return restored, err
			}
			dump = zr
		}
		if err := restoreDumpInto(ctx, userContext, engine, dbName, dump); err != nil {
			return restored, err
		}
		restored = append(restored, dbName)
		if only != "" {
			break
		}
	}
	return restored, nil
}

// chownAncestors chowns dir and every parent directory up to (but not including) root - root is already correctly owned, this just fixes up new subdirectories os.MkdirAll created as real host root beneath it
func chownAncestors(dir, root string, uid int) {
	for dir != root && dir != "." && dir != string(filepath.Separator) && strings.HasPrefix(dir, root) {
		_ = tenantfs.Lchown(dir, uid, uid)
		dir = filepath.Dir(dir)
	}
}

var restoreFilesPathMap = func(userContext string) map[string]string {
	userHome := "/home/" + userContext
	return map[string]string{
		"backup/html":      userHome + "/docker-data/volumes/" + userContext + "_html_data/_data",
		"backup/vhosts":    userHome + "/docker-data/volumes/" + userContext + "_webserver_data/_data",
		"backup/mail":      userHome + "/docker-data/volumes/" + userContext + "_mail_data/_data",
		"backup/crons.ini": userHome + "/crons.ini",
	}
}

// restoreSections are the restore targets that bring back one part of the archive on its own
var restoreSections = map[string]struct {
	prefixes []string
	label    string
}{
	"html":   {[]string{"backup/html"}, "Website files"},
	"vhosts": {[]string{"backup/vhosts", "backup/dns"}, "Domains"},
	"mail":   {[]string{"backup/emails", "backup/mail"}, "Emails"},
	"crons":  {[]string{"backup/crons.ini"}, "Cron jobs"},
	"dns":    {[]string{"backup/dns"}, "DNS zones"},
}

// selectiveRestoreTargets restore single items picked from the lists saved at reindex
var selectiveRestoreTargets = map[string]bool{"paths": true, "mailboxes": true, "domains": true, "dns_zones": true}

// restoreFilesFromTar writes every file of the archive (or only those under the "only" prefixes) to where the plan says, chowned to the account's uid so its rootless containers can read them
// with only set it stops reading once it's past those prefixes
func restoreFilesFromTar(archive io.Reader, plan *restorePlan, uid int, only []string) ([]string, error) {
	gz, err := gzip.NewReader(archive)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var extracted []string
	layout := newArchiveLayout()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return extracted, err
		}
		raw := strings.Trim(header.Name, "/")
		name := layout.observe(raw)
		if len(only) > 0 && layout.pastAll(raw, only) {
			break
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		if len(only) > 0 && !slices.ContainsFunc(only, func(prefix string) bool { return insidePath(name, prefix) }) {
			continue
		}
		if dest, ok := plan.extractEntry(tr, name, uid); ok {
			extracted = append(extracted, dest)
		}
	}
	return extracted, nil
}

// handleListBackupsFromDestination serves the list of remote backups (from the cached index file) and, on POST, kicks off a background reindex against the SSH destination
func handleListBackupsFromDestination(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	config, userHome := readBackupEnv(userContext)
	jsonFile := filepath.Join(userHome, "available_backups.json")
	lockFile := filepath.Join(userHome, "reindex.lock")

	if r.Method == http.MethodPost {
		took, lockErr := takeReindexLock(lockFile)
		if lockErr != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !took {
			web.WriteJSON(w, http.StatusConflict, map[string]string{"message": "Reindex already in progress."})
			return
		}
		go doReindex(userHome, config, backupNamePrefix(config, userContext), jsonFile, lockFile)

		if r.URL.Query().Get("output") == "json" {
			web.WriteJSON(w, http.StatusAccepted, map[string]string{"message": "Reindex started."})
			return
		}
		http.Redirect(w, r, "/backups/list", http.StatusFound)
		return
	}

	reindexing := reindexRunning(lockFile)

	var backups []BackupInfo
	var reindexErr string
	if content, readErr := tenantfs.ReadFile(jsonFile); readErr == nil {
		var errPayload struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(content, &errPayload) == nil && errPayload.Error != "" {
			reindexErr = errPayload.Error
		} else {
			_ = json.Unmarshal(content, &backups)
		}
	}
	// an index written before the prefix filter existed can still hold other accounts' archives
	if prefix := backupNamePrefix(config, userContext); prefix != "" {
		own := backups[:0]
		for _, b := range backups {
			if strings.HasPrefix(b.BackupFile, prefix) {
				own = append(own, b)
			}
		}
		backups = own
	}

	if r.URL.Query().Get("output") == "json" {
		// status=1 is the polling shape (onboarding wizard's destination test), which needs "reindexing" to tell an in-progress run apart from a finished one - kept opt-in so the existing Array.isArray() poller keeps working unchanged
		if r.URL.Query().Get("status") == "1" {
			web.WriteJSON(w, http.StatusOK, map[string]any{"reindexing": reindexing, "error": reindexErr, "count": len(backups)})
			return
		}
		if reindexErr != "" {
			// this used to fall through to the HTML error page even for output=json callers - return JSON instead
			web.WriteJSON(w, http.StatusOK, map[string]string{"error": reindexErr})
			return
		}
		web.WriteJSON(w, http.StatusOK, backups)
		return
	}

	renderBackupRestorePage(a, w, r, backups, reindexing, reindexErr)
}

// handleRestoreFromBackup restores from a remote backup archive per restore_target: everything, a single database, all files, one section (html/vhosts/mail/crons/dns), or single items (paths/mailboxes/domains/dns_zones) streamed without downloading the whole archive
// restoreRequest is one restore, from the panel's queue or the API, with the form values the restore page posts
type restoreRequest struct {
	UserID      int        `json:"user_id"`
	Username    string     `json:"username"`
	UserContext string     `json:"user_context"`
	IP          string     `json:"ip"`
	Form        url.Values `json:"form"`
}

func handleRestoreFromBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// the JS client sends a FormData body, not urlencoded - ParseForm() alone doesn't read multipart bodies
	_ = r.ParseMultipartForm(1 << 20)
	// direct restores (the API) wait for the account's queue, only one restore runs at a time
	q := queueFor(userContext)
	q.restore.Lock()
	status, body := performRestore(r.Context(), a, restoreRequest{UserID: userID, Username: currentUsername, UserContext: userContext, IP: reqip.ClientIP(r), Form: r.Form})
	q.restore.Unlock()
	web.WriteJSON(w, status, body)
}

// performRestore runs one restore and returns the http status and json body for it
func performRestore(ctx context.Context, a *appctx.App, req restoreRequest) (int, map[string]any) {
	backupFile := strings.TrimSpace(req.Form.Get("backup_file"))
	restoreTarget := req.Form.Get("restore_target")
	if restoreTarget == "" {
		restoreTarget = "all"
	}
	database := strings.TrimSpace(req.Form.Get("database"))
	if _, ok := restoreSections[restoreTarget]; !ok && !selectiveRestoreTargets[restoreTarget] && restoreTarget != "all" && restoreTarget != "database" && restoreTarget != "files" {
		return http.StatusBadRequest, map[string]any{"error": "Unknown restore target: " + restoreTarget}
	}

	if backupFile == "" {
		return http.StatusBadRequest, map[string]any{"error": "No backup file specified."}
	}
	safeName := filepath.Base(backupFile)
	if safeName != backupFile {
		return http.StatusBadRequest, map[string]any{"error": "Invalid backup filename."}
	}

	config, _ := readBackupEnv(req.UserContext)
	store, storeErr := newRemoteStore(config)
	if storeErr != nil {
		return http.StatusBadRequest, map[string]any{"error": storeErr.Error()}
	}

	plan, err := newRestorePlan(ctx, a, req.UserID, req.UserContext)
	if err != nil {
		return http.StatusInternalServerError, map[string]any{"error": err.Error()}
	}

	if selectiveRestoreTargets[restoreTarget] {
		return restoreSelectedItems(ctx, a, req, store, safeName, restoreTarget, plan)
	}

	uid, _ := a.GetUID(ctx, req.UserContext)

	// a single database or section is streamed, reading stops once it's restored and everything but html/ comes before the website files
	if restoreTarget != "all" && restoreTarget != "files" {
		if restoreTarget == "database" && database == "" {
			return http.StatusBadRequest, map[string]any{"error": "No database specified."}
		}
		rc, err := store.Open(ctx, safeName)
		if err != nil {
			return http.StatusInternalServerError, map[string]any{"error": err.Error()}
		}
		defer rc.Close()

		if restoreTarget == "database" {
			restored, err := restoreDatabasesFromTar(ctx, req.UserContext, rc, database)
			if err != nil {
				return http.StatusInternalServerError, map[string]any{"error": err.Error()}
			}
			if len(restored) == 0 {
				return http.StatusNotFound, map[string]any{"error": "Database '" + database + "' not found in this backup."}
			}
			_ = logger.RecordUserAction(a.Config, req.Username, "restored database '"+database+"' from backup "+safeName, req.IP)
			return http.StatusOK, map[string]any{"success": true, "message": "Database '" + database + "' restored successfully."}
		}

		section := restoreSections[restoreTarget]
		extracted, err := restoreFilesFromTar(rc, plan, uid, section.prefixes)
		if err != nil {
			return http.StatusInternalServerError, map[string]any{"error": err.Error()}
		}
		if len(extracted) == 0 {
			return http.StatusNotFound, map[string]any{"error": section.label + " not found in this backup."}
		}
		applyRestored(ctx, req.UserContext, extracted, plan)
		_ = logger.RecordUserAction(a.Config, req.Username, "restored "+strings.ToLower(section.label)+" from backup "+safeName, req.IP)
		return http.StatusOK, map[string]any{"success": true, "message": section.label + " restored (" + strconv.Itoa(len(extracted)) + " items)."}
	}

	// full and all-files restores download first, a dropped connection then can't leave the site half restored
	localPath, cleanup, fetchErr := fetchArchive(ctx, store, safeName)
	if fetchErr != nil {
		return http.StatusInternalServerError, map[string]any{"error": fetchErr.Error()}
	}
	defer cleanup()
	f, err := tenantfs.Open(localPath)
	if err != nil {
		return http.StatusInternalServerError, map[string]any{"error": err.Error()}
	}
	defer f.Close()

	extracted, err := restoreFilesFromTar(f, plan, uid, nil)
	if err != nil {
		return http.StatusInternalServerError, map[string]any{"error": err.Error()}
	}
	applyRestored(ctx, req.UserContext, extracted, plan)
	if restoreTarget == "files" {
		_ = logger.RecordUserAction(a.Config, req.Username, "restored files from backup "+safeName, req.IP)
		return http.StatusOK, map[string]any{"success": true, "message": "Files restored (" + strconv.Itoa(len(extracted)) + " items)."}
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return http.StatusInternalServerError, map[string]any{"error": err.Error()}
	}
	if _, err := restoreDatabasesFromTar(ctx, req.UserContext, f, ""); err != nil {
		return http.StatusInternalServerError, map[string]any{"error": err.Error()}
	}
	_ = logger.RecordUserAction(a.Config, req.Username, "restored full backup from "+safeName, req.IP)
	return http.StatusOK, map[string]any{"success": true, "message": "Full restore completed."}
}

// handleDownloadBackup downloads a remote backup archive and streams it back to the client
func handleDownloadBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// the JS client sends a FormData body, not urlencoded - ParseForm() alone doesn't read multipart bodies
	_ = r.ParseMultipartForm(1 << 20)
	backupFile := strings.TrimSpace(r.Form.Get("backup_file"))
	if backupFile == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "No backup file specified."})
		return
	}
	safeName := filepath.Base(backupFile)
	if safeName != backupFile {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid backup filename."})
		return
	}

	config, _ := readBackupEnv(userContext)
	store, storeErr := newRemoteStore(config)
	if storeErr != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": storeErr.Error()})
		return
	}

	// streamed straight through, no temp copy of the archive
	rc, openErr := store.Open(r.Context(), safeName)
	if openErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": openErr.Error()})
		return
	}
	defer rc.Close()

	_ = logger.RecordUserAction(a.Config, currentUsername, "downloaded backup "+safeName, reqip.ClientIP(r))

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+safeName+"\"")
	if _, copyErr := io.Copy(w, rc); copyErr != nil {
		// no Content-Length here, so drop the connection or the browser saves the cut off archive as complete
		panic(http.ErrAbortHandler)
	}
}
