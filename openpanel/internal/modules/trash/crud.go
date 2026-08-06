package trash

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

type jsonResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// handleRestoreFile moves one trashed item back to its original path,
// recorded in .trash_restore.
func handleRestoreFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	itemName := r.URL.Query().Get("filename")
	if itemName == "" {
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Missing filename."})
		return
	}

	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	trashDir := "/home/" + userContext + "/.local/share/Trash"

	itemName = filepath.Base(itemName)
	if itemName == "" || itemName == "." || itemName == "/" {
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Invalid filename."})
		return
	}

	trashedPath := filepath.Join(trashDir, itemName)
	trashRestoreFile := filepath.Join(trashDir, ".trash_restore")

	if _, err := os.Stat(trashedPath); err != nil {
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: "File or directory does not exist in Trash."})
		return
	}

	restoreInfo, err := os.ReadFile(trashRestoreFile)
	if err != nil {
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: ".trash_restore file is missing."})
		return
	}

	var restoreDestination string
	var linesToKeep []string
	for _, line := range strings.Split(string(restoreInfo), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.Contains(trimmed, "=") {
			continue
		}
		key, value, _ := strings.Cut(trimmed, "=")
		if key == itemName {
			if idx := strings.Index(value, "|deletion_date="); idx != -1 {
				restoreDestination = strings.TrimSpace(value[:idx])
			} else {
				restoreDestination = strings.TrimSpace(value)
			}
			continue
		}
		linesToKeep = append(linesToKeep, trimmed)
	}

	if restoreDestination == "" {
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Original path not found in .trash_restore."})
		return
	}

	restoreDestination = absPath(restoreDestination)
	volumeAbs := absPath(volume)

	if !isWithin(resolveOrSelf(restoreDestination), resolveOrSelf(volumeAbs)) {
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Restore path is outside your permitted data directory."})
		return
	}

	if err := os.Rename(trashedPath, restoreDestination); err != nil {
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: err.Error()})
		return
	}

	_ = os.WriteFile(trashRestoreFile, []byte(joinLines(linesToKeep)), 0o644)

	_ = logger.RecordUserAction(a.Config, username, "restored "+restoreDestination+" from Trash using File Manager", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, jsonResult{Success: true, Message: "Restored from Trash"})
}

// handleDeleteTrash permanently deletes one item from the trash.
func handleDeleteTrash(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	itemName := r.URL.Query().Get("filename")
	if itemName == ".trash_restore" {
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Deletion of system file .trash_restore is not allowed."})
		return
	}

	itemName = filepath.Base(itemName)
	if itemName == "" || itemName == "." || itemName == "/" {
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Invalid filename."})
		return
	}

	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	trashDir := "/home/" + userContext + "/.local/share/Trash"
	trashedPath := filepath.Join(trashDir, itemName)
	trashRestoreFile := filepath.Join(trashDir, ".trash_restore")

	info, err := os.Lstat(trashedPath)
	if err != nil {
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: "File or directory does not exist in Trash."})
		return
	}

	var detectedType string
	switch {
	case info.IsDir():
		if err := os.RemoveAll(trashedPath); err != nil {
			writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: err.Error()})
			return
		}
		detectedType = "directory"
	case info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0:
		if err := os.Remove(trashedPath); err != nil {
			writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: err.Error()})
			return
		}
		detectedType = "file"
	default:
		writeJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Unsupported item type."})
		return
	}

	if content, err := os.ReadFile(trashRestoreFile); err == nil {
		var keep []string
		for _, line := range strings.Split(string(content), "\n") {
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, itemName+"=") {
				keep = append(keep, line)
			}
		}
		_ = os.WriteFile(trashRestoreFile, []byte(joinLines(keep)), 0o644)
	}

	_ = logger.RecordUserAction(a.Config, username, "permanently deleted "+detectedType+" "+trashedPath+" using File Manager", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, jsonResult{Success: true, Message: "Deleted permanently"})
}

// handleDeleteAll empties the whole Trash, keeping .trash_restore itself
// (truncated) rather than removing it outright.
func handleDeleteAll(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	trashDir := "/home/" + userContext + "/.local/share/Trash"

	entries, err := os.ReadDir(trashDir)
	if err != nil {
		flashAndRedirectToTrash(a, w, r, "error", "Unexpected error")
		return
	}

	var clearErr error
	for _, entry := range entries {
		if entry.Name() == ".trash_restore" {
			continue
		}
		if clearErr = os.RemoveAll(filepath.Join(trashDir, entry.Name())); clearErr != nil {
			break
		}
	}

	if clearErr != nil {
		flashAndRedirectToTrash(a, w, r, "error", "Unexpected error")
		return
	}

	if err := os.WriteFile(filepath.Join(trashDir, ".trash_restore"), nil, 0o644); err != nil {
		flashAndRedirectToTrash(a, w, r, "error", "Unexpected error")
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "Emptied trash using File Manager", reqip.ClientIP(r))
	flashAndRedirectToTrash(a, w, r, "success", "Trash emptied successfully")
}

// handleRestoreAll restores every item listed in .trash_restore back to
// its original path. Note the containment check here uses plain abspath
// (no symlink resolution), unlike handleRestoreFile's resolveOrSelf - an
// intentional asymmetry, not an oversight.
func handleRestoreAll(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	trashDir := "/home/" + userContext + "/.local/share/Trash"
	trashRestorePath := filepath.Join(trashDir, ".trash_restore")

	content, err := os.ReadFile(trashRestorePath)
	if err != nil {
		flashAndRedirectToTrash(a, w, r, "info", "Nothing to restore")
		return
	}

	allowedBase := absPath(volume)

	var restoreErrors []string
	var linesToKeep []string

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.Contains(trimmed, "=") {
			continue
		}
		key, value, _ := strings.Cut(trimmed, "=")

		var originalPath string
		if idx := strings.Index(value, "|deletion_date="); idx != -1 {
			originalPath = strings.TrimSpace(value[:idx])
		} else {
			originalPath = strings.TrimSpace(value)
		}

		restorePath := absPath(originalPath)

		if !isWithin(restorePath, allowedBase) {
			restoreErrors = append(restoreErrors, key+": invalid restore path")
			linesToKeep = append(linesToKeep, trimmed)
			continue
		}

		trashedItemPath := filepath.Join(trashDir, key)
		if _, statErr := os.Stat(trashedItemPath); statErr != nil {
			restoreErrors = append(restoreErrors, key+": not found in trash")
			linesToKeep = append(linesToKeep, trimmed)
			continue
		}

		if err := os.Rename(trashedItemPath, restorePath); err != nil {
			restoreErrors = append(restoreErrors, key+": "+err.Error())
			linesToKeep = append(linesToKeep, trimmed)
			continue
		}
		_ = logger.RecordUserAction(a.Config, username, "restored "+restorePath+" from Trash using Restore All", reqip.ClientIP(r))
	}

	if len(restoreErrors) > 0 {
		_ = os.WriteFile(trashRestorePath, []byte(joinLines(linesToKeep)), 0o644)
		flashAndRedirectToTrash(a, w, r, "warning", "Some items could not be restored.")
		return
	}

	_ = os.WriteFile(trashRestorePath, nil, 0o644)
	flashAndRedirectToTrash(a, w, r, "success", "All items restored successfully")
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func absPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return filepath.Clean(abs)
}

// resolveOrSelf resolves symlinks, falling back to the cleaned absolute
// path if the target doesn't exist yet.
func resolveOrSelf(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return p
}
