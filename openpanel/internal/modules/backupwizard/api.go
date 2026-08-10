package backupwizard

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterAPI wires the /api/backup-wizard routes onto mux, gated behind
// the "backup_wizard" feature flag.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "backup_wizard", "GET /api/backup-wizard/status", func(w http.ResponseWriter, r *http.Request) { apiBackupWizardStatus(a, w, r) })
	apiregistry.Handle(mux, a, "backup_wizard", "POST /api/backup-wizard/create", func(w http.ResponseWriter, r *http.Request) { apiBackupWizardCreate(a, w, r) })
	apiregistry.Handle(mux, a, "backup_wizard", "GET /api/backup-wizard/download/{filename...}", func(w http.ResponseWriter, r *http.Request) {
		apiBackupWizardDownload(a, w, r, r.PathValue("filename"))
	})
}

// apiBackupWizardStatus reuses the exact same statusPayload shape as
// handleBackupWizardStatus (backupwizard.go).
func apiBackupWizardStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

// apiBackupWizardCreate is handleBackupWizardCreate (backupwizard.go) with
// a JSON response instead of a flash+redirect.
func apiBackupWizardCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if isBackupInProgress(r.Context(), currentUsername) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "A backup is already in progress. Please wait for it to finish."})
		return
	}

	if mkdirErr := os.MkdirAll(backupDir(userContext), 0o755); mkdirErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to start backup."})
		return
	}

	cmd := exec.Command("opencli", "user-backup", "--account", currentUsername)
	if startErr := cmd.Start(); startErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to start backup."})
		return
	}
	go func() { _ = cmd.Wait() }()

	_ = logger.RecordUserAction(a.Config, currentUsername, "started a backup via API", reqip.ClientIP(r))
	writeJSON(w, http.StatusAccepted, map[string]string{"message": "Backup started. Poll GET /api/backup-wizard/status for progress."})
}

// apiBackupWizardDownload is handleBackupWizardDownload (backupwizard.go)
// with JSON error responses instead of a flash+redirect.
func apiBackupWizardDownload(a *appctx.App, w http.ResponseWriter, r *http.Request, filename string) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	safeName := filepath.Base(filename)
	if safeName != filename {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid filename."})
		return
	}

	dir := backupDir(userContext)
	filePath := filepath.Join(dir, safeName)

	resolvedDir, dirErr := filepath.Abs(dir)
	resolvedFile, fileErr := filepath.Abs(filePath)
	if dirErr != nil || fileErr != nil || !isWithin(resolvedFile, resolvedDir) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid file path."})
		return
	}

	info, statErr := os.Stat(filePath)
	if statErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Backup file not found."})
		return
	}

	if isBackupInProgress(r.Context(), currentUsername) {
		_, _, inProgressFile := inProgressInfo(currentUsername, userContext)
		if inProgressFile == safeName {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "This backup is still being created. Please wait until it completes before downloading."})
			return
		}
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "downloaded backup "+safeName+" via API", reqip.ClientIP(r))

	f, openErr := os.Open(filePath)
	if openErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Backup file not found."})
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+safeName+"\"")
	http.ServeContent(w, r, safeName, info.ModTime(), f)
}
