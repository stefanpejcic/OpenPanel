// Package backupwizard ports modules/files/backup_wizard.py: the
// single-click "back up my whole account" flow, backed by `opencli
// user-backup` writing .tar.gz archives into the user's docroot volume's
// _backups/ folder.
package backupwizard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// Register wires modules/files/backup_wizard.py's routes onto mux, gated
// behind the "backup_wizard" feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "backup_wizard")(h)
	}

	mux.Handle("GET /backup-wizard", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBackupWizard(a, w, r) }))
	mux.Handle("GET /backup-wizard/status", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBackupWizardStatus(a, w, r) }))
	mux.Handle("POST /backup-wizard/create", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleBackupWizardCreate(a, w, r) }))
	mux.Handle("GET /backup-wizard/download/{filename...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleBackupWizardDownload(a, w, r, r.PathValue("filename"))
	}))
}

func backupDir(userContext string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/_backups"
}

// isBackupInProgress mirrors _is_backup_in_progress().
func isBackupInProgress(ctx context.Context, currentUsername string) bool {
	err := exec.CommandContext(ctx, "pgrep", "-f", "user-backup.*--account.*"+currentUsername).Run()
	return err == nil
}

// formatSize mirrors _format_size().
func formatSize(numBytes float64) string {
	for _, unit := range []string{"B", "KB", "MB", "GB"} {
		if numBytes < 1024.0 {
			return fmt.Sprintf("%.1f %s", numBytes, unit)
		}
		numBytes /= 1024.0
	}
	return fmt.Sprintf("%.1f TB", numBytes)
}

var inProgressLogNameRE = regexp.MustCompile(`_backup_(\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2})\.log$`)

// inProgressInfo mirrors _in_progress_info(): (started, currentSize,
// inProgressFilename) for the currently-running backup, best-effort from
// the newest matching log file and the newest .tar.gz on disk.
func inProgressInfo(currentUsername, userContext string) (started, currentSize, inProgressFile string) {
	logDir := "/var/log/openpanel/admin/backups"
	if entries, err := os.ReadDir(logDir); err == nil {
		type logEntry struct {
			name    string
			modTime time.Time
		}
		var logs []logEntry
		prefix := currentUsername + "_backup_"
		for _, e := range entries {
			if e.IsDir() || !strings.HasPrefix(e.Name(), prefix) || !strings.HasSuffix(e.Name(), ".log") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			logs = append(logs, logEntry{name: e.Name(), modTime: info.ModTime()})
		}
		sort.Slice(logs, func(i, j int) bool { return logs[i].modTime.After(logs[j].modTime) })
		if len(logs) > 0 {
			if m := inProgressLogNameRE.FindStringSubmatch(logs[0].name); m != nil {
				datePart, timePart, _ := strings.Cut(m[1], "_")
				started = datePart + " " + strings.ReplaceAll(timePart, "-", ":")
			}
		}
	}

	dir := backupDir(userContext)
	if entries, err := os.ReadDir(dir); err == nil {
		type tarEntry struct {
			name    string
			modTime time.Time
			size    int64
		}
		var tars []tarEntry
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".tar.gz") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			tars = append(tars, tarEntry{name: e.Name(), modTime: info.ModTime(), size: info.Size()})
		}
		sort.Slice(tars, func(i, j int) bool { return tars[i].modTime.After(tars[j].modTime) })
		if len(tars) > 0 {
			inProgressFile = tars[0].name
			currentSize = formatSize(float64(tars[0].size))
		}
	}

	return started, currentSize, inProgressFile
}

// BackupFile is one entry of the "Existing Backups" table.
type BackupFile struct {
	Name       string `json:"name"`
	SizeRaw    int64  `json:"size_raw"`
	Size       string `json:"size"`
	Mtime      string `json:"mtime"`
	InProgress bool   `json:"in_progress"`
}

// listBackups mirrors _list_backups().
func listBackups(userContext, inProgressFile string) []BackupFile {
	dir := backupDir(userContext)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	type fileInfo struct {
		entry   os.DirEntry
		modTime time.Time
	}
	var files []fileInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".tar.gz") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{entry: e, modTime: info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].modTime.After(files[j].modTime) })

	result := make([]BackupFile, 0, len(files))
	for _, f := range files {
		info, err := f.entry.Info()
		if err != nil {
			continue
		}
		result = append(result, BackupFile{
			Name: f.entry.Name(), SizeRaw: info.Size(), Size: formatSize(float64(info.Size())),
			Mtime: info.ModTime().Format("2006-01-02 15:04:05"), InProgress: f.entry.Name() == inProgressFile,
		})
	}
	return result
}

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

func flashAndRedirectToWizard(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, "/backup-wizard", http.StatusFound)
}

// handleBackupWizard mirrors backup_wizard().
func handleBackupWizard(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	inProgress := isBackupInProgress(r.Context(), currentUsername)
	var started, size, inProgressFile string
	if inProgress {
		started, size, inProgressFile = inProgressInfo(currentUsername, userContext)
	}
	backups := listBackups(userContext, inProgressFile)

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, statusPayload{InProgress: inProgress, Backups: backups})
		return
	}

	renderBackupWizardPage(a, w, r, inProgress, started, size, backups)
}

// handleBackupWizardStatus mirrors backup_wizard_status().
func handleBackupWizardStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	inProgress := isBackupInProgress(r.Context(), currentUsername)
	var started, size, inProgressFile string
	if inProgress {
		started, size, inProgressFile = inProgressInfo(currentUsername, userContext)
	}
	backups := listBackups(userContext, inProgressFile)

	writeJSON(w, http.StatusOK, statusPayload{
		InProgress: inProgress, InProgressStarted: started, InProgressSize: size, Backups: backups,
	})
}

type statusPayload struct {
	InProgress        bool         `json:"in_progress"`
	InProgressStarted string       `json:"in_progress_started,omitempty"`
	InProgressSize    string       `json:"in_progress_size,omitempty"`
	Backups           []BackupFile `json:"backups"`
}

// handleBackupWizardCreate mirrors backup_wizard_create(): fires
// `opencli user-backup` in the background (matches Python's
// subprocess.Popen + start_new_session=True fire-and-forget).
func handleBackupWizardCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if isBackupInProgress(r.Context(), currentUsername) {
		flashAndRedirectToWizard(a, w, r, "error", "A backup is already in progress. Please wait for it to finish.")
		return
	}

	if err := os.MkdirAll(backupDir(userContext), 0o755); err != nil {
		flashAndRedirectToWizard(a, w, r, "error", "Failed to start backup.")
		return
	}

	// nil Stdout/Stderr discard the child's output, matching Python's
	// subprocess.DEVNULL; a bare exec.Command (no context) plus the
	// detached goroutine below matches start_new_session=True's
	// fire-and-forget semantics (the backup must outlive this request).
	cmd := exec.Command("opencli", "user-backup", "--account", currentUsername)
	if err := cmd.Start(); err != nil {
		flashAndRedirectToWizard(a, w, r, "error", "Failed to start backup.")
		return
	}
	go func() { _ = cmd.Wait() }()

	_ = logger.RecordUserAction(a.Config, currentUsername, "started a backup", reqip.ClientIP(r))
	flashAndRedirectToWizard(a, w, r, "success", "Backup started. It will appear in the list below once complete.")
}

// handleBackupWizardDownload mirrors backup_wizard_download().
func handleBackupWizardDownload(a *appctx.App, w http.ResponseWriter, r *http.Request, filename string) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	safeName := filepath.Base(filename)
	if safeName != filename {
		flashAndRedirectToWizard(a, w, r, "error", "Invalid filename.")
		return
	}

	dir := backupDir(userContext)
	filePath := filepath.Join(dir, safeName)

	resolvedDir, dirErr := filepath.Abs(dir)
	resolvedFile, fileErr := filepath.Abs(filePath)
	if dirErr != nil || fileErr != nil || !isWithin(resolvedFile, resolvedDir) {
		flashAndRedirectToWizard(a, w, r, "error", "Invalid file path.")
		return
	}

	info, statErr := os.Stat(filePath)
	if statErr != nil {
		flashAndRedirectToWizard(a, w, r, "error", "Backup file not found.")
		return
	}

	if isBackupInProgress(r.Context(), currentUsername) {
		_, _, inProgressFile := inProgressInfo(currentUsername, userContext)
		if inProgressFile == safeName {
			flashAndRedirectToWizard(a, w, r, "error", "This backup is still being created. Please wait until it completes before downloading.")
			return
		}
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "downloaded backup "+safeName, reqip.ClientIP(r))

	f, openErr := os.Open(filePath)
	if openErr != nil {
		flashAndRedirectToWizard(a, w, r, "error", "Backup file not found.")
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+safeName+"\"")
	http.ServeContent(w, r, safeName, info.ModTime(), f)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func isWithin(candidate, base string) bool {
	rel, err := filepath.Rel(base, candidate)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && rel != "..")
}
