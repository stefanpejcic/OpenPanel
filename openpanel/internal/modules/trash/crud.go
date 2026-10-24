package trash

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

type jsonResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// handleRestoreFile moves one trashed item back to its original path, recorded in .trash_restore
func handleRestoreFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	itemName := r.URL.Query().Get("filename")
	if itemName == "" {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Missing filename."})
		return
	}

	_, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	trashDir := "/home/" + userContext + "/.local/share/Trash"

	itemName = filepath.Base(itemName)
	if itemName == "" || itemName == "." || itemName == "/" {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Invalid filename."})
		return
	}

	trashedPath := filepath.Join(trashDir, itemName)
	trashRestoreFile := filepath.Join(trashDir, ".trash_restore")

	if _, err := tenantfs.Stat(trashedPath); err != nil {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "File or directory does not exist in Trash."})
		return
	}

	restoreInfo, err := tenantfs.ReadFile(trashRestoreFile)
	if err != nil {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: ".trash_restore file is missing."})
		return
	}

	var restoreDestination string
	var linesToKeep []string
	for _, line := range strings.Split(string(restoreInfo), "\n") {
		trimmed := strings.TrimSpace(line)
		key, meta, ok := splitRestoreLine(trimmed)
		if !ok {
			continue
		}
		if key == itemName {
			restoreDestination = meta.originalPath
			continue
		}
		linesToKeep = append(linesToKeep, trimmed)
	}

	if restoreDestination == "" {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Original path not found in .trash_restore."})
		return
	}

	restoreDestination = absPath(restoreDestination)
	volumeAbs := absPath(volume)

	if !isWithin(resolveOrSelf(restoreDestination), resolveOrSelf(volumeAbs)) {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Restore path is outside your permitted data directory."})
		return
	}

	if err := restoreItem(trashedPath, restoreDestination); err != nil {
		if errors.Is(err, errRestoreExists) {
			web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Can't restore: an item with the same name is already in the original folder."})
			return
		}
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: err.Error()})
		return
	}

	_ = tenantfs.WriteFile(trashRestoreFile, []byte(joinLines(linesToKeep)), 0o644)

	_ = logger.RecordUserAction(a.Config, username, "restored "+restoreDestination+" from Trash using File Manager", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, jsonResult{Success: true, Message: "Restored from Trash"})
}

// handleDeleteTrash permanently deletes one item from the trash.
func handleDeleteTrash(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	itemName := r.URL.Query().Get("filename")
	if itemName == ".trash_restore" {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Deletion of system file .trash_restore is not allowed."})
		return
	}

	itemName = filepath.Base(itemName)
	if itemName == "" || itemName == "." || itemName == "/" {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Invalid filename."})
		return
	}

	_, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	trashDir := "/home/" + userContext + "/.local/share/Trash"
	trashedPath := filepath.Join(trashDir, itemName)
	trashRestoreFile := filepath.Join(trashDir, ".trash_restore")

	info, err := tenantfs.Lstat(trashedPath)
	if err != nil {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "File or directory does not exist in Trash."})
		return
	}

	var detectedType string
	switch {
	case info.IsDir():
		if err := tenantfs.RemoveAll(trashedPath); err != nil {
			web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: err.Error()})
			return
		}
		detectedType = "directory"
	case info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0:
		if err := tenantfs.Remove(trashedPath); err != nil {
			web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: err.Error()})
			return
		}
		detectedType = "file"
	default:
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Unsupported item type."})
		return
	}

	if content, err := tenantfs.ReadFile(trashRestoreFile); err == nil {
		_ = tenantfs.WriteFile(trashRestoreFile, []byte(joinLines(dropRestoreLine(string(content), itemName))), 0o644)
	}

	_ = logger.RecordUserAction(a.Config, username, "permanently deleted "+detectedType+" "+trashedPath+" using File Manager", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, jsonResult{Success: true, Message: "Deleted permanently"})
}

// handleDeleteAll empties the whole Trash, keeping .trash_restore itself (truncated) rather than removing it outright
func handleDeleteAll(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	trashDir := "/home/" + userContext + "/.local/share/Trash"

	failed, emptyErr := emptyTrash(trashDir)
	if emptyErr != nil {
		flashAndRedirectToTrash(a, w, r, "error", "Unexpected error")
		return
	}
	if len(failed) > 0 {
		flashAndRedirectToTrash(a, w, r, "error", web.Tr(a, r, "Could not delete %(items)s from the trash.", "items", strings.Join(failed, ", ")))
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "Emptied trash using File Manager", reqip.ClientIP(r))
	flashAndRedirectToTrash(a, w, r, "success", "Trash emptied successfully")
}

// handleRestoreAll restores every item listed in .trash_restore back to its original path - the containment check here uses plain abspath (no symlink resolution), unlike handleRestoreFile's resolveOrSelf, an intentional asymmetry not an oversight
func handleRestoreAll(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	trashDir := "/home/" + userContext + "/.local/share/Trash"
	trashRestorePath := filepath.Join(trashDir, ".trash_restore")

	content, err := tenantfs.ReadFile(trashRestorePath)
	if err != nil {
		flashAndRedirectToTrash(a, w, r, "info", "Nothing to restore")
		return
	}

	allowedBase := absPath(volume)

	var restoreErrors []string
	var linesToKeep []string

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		key, meta, ok := splitRestoreLine(trimmed)
		if !ok {
			continue
		}

		restorePath := absPath(meta.originalPath)

		if !isWithin(restorePath, allowedBase) {
			restoreErrors = append(restoreErrors, key+": invalid restore path")
			linesToKeep = append(linesToKeep, trimmed)
			continue
		}

		trashedItemPath := filepath.Join(trashDir, key)
		if _, statErr := tenantfs.Stat(trashedItemPath); statErr != nil {
			restoreErrors = append(restoreErrors, key+": not found in trash")
			linesToKeep = append(linesToKeep, trimmed)
			continue
		}

		if err := restoreItem(trashedItemPath, restorePath); err != nil {
			restoreErrors = append(restoreErrors, key+": "+err.Error())
			linesToKeep = append(linesToKeep, trimmed)
			continue
		}
		_ = logger.RecordUserAction(a.Config, username, "restored "+restorePath+" from Trash using Restore All", reqip.ClientIP(r))
	}

	if len(restoreErrors) > 0 {
		_ = tenantfs.WriteFile(trashRestorePath, []byte(joinLines(linesToKeep)), 0o644)
		flashAndRedirectToTrash(a, w, r, "warning", "Some items could not be restored.")
		return
	}

	_ = tenantfs.WriteFile(trashRestorePath, nil, 0o644)
	flashAndRedirectToTrash(a, w, r, "success", "All items restored successfully")
}

var errRestoreExists = errors.New("an item with the same name is already in the original folder")

// restoreItem moves a trashed item back to dest, recreating dest's folder if it was deleted but never replacing what's there now
func restoreItem(trashedPath, dest string) error {
	if _, err := tenantfs.Lstat(dest); err == nil {
		return errRestoreExists
	}
	info, err := tenantfs.Lstat(trashedPath)
	if err != nil {
		return err
	}
	missing := missingDirs(filepath.Dir(dest))
	if err := tenantfs.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	// we run as root, so hand recreated folders to whoever owns the item
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		for _, dir := range missing {
			_ = tenantfs.Lchown(dir, int(st.Uid), int(st.Gid))
		}
	}
	return tenantfs.Rename(trashedPath, dest)
}

// missingDirs lists dir and any of its parents that don't exist yet, outermost first
func missingDirs(dir string) []string {
	var missing []string
	for d := dir; d != filepath.Dir(d); d = filepath.Dir(d) {
		if _, err := tenantfs.Lstat(d); !errors.Is(err, fs.ErrNotExist) {
			break
		}
		missing = append([]string{d}, missing...)
	}
	return missing
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

// resolveOrSelf resolves symlinks, falling back to the cleaned absolute path if the target doesn't exist yet
func resolveOrSelf(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return p
}

// handleTrashSize is the trash page's "Calculate" link, /json/directory-size only reaches the website files volume
func handleTrashSize(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	itemName := filepath.Base(r.URL.Query().Get("name"))
	if itemName == "" || itemName == "." || itemName == "/" || itemName == ".." {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid folder name."})
		return
	}

	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	target := filepath.Join("/home/"+userContext+"/.local/share/Trash", itemName)
	if info, statErr := tenantfs.Lstat(target); statErr != nil || !info.IsDir() {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Folder does not exist in Trash."})
		return
	}

	out, runErr := exec.CommandContext(r.Context(), "du", "-sh", target).Output()
	if runErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not calculate the folder size."})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]string{"size": strings.SplitN(string(out), "\t", 2)[0]})
}

// removeTrashEntry is swapped in tests, a delete as root almost never fails
var removeTrashEntry = tenantfs.RemoveAll

// emptyTrash deletes everything in trashDir and keeps the .trash_restore lines of whatever is left, stopping at the first failure used to leave lines for items already deleted
func emptyTrash(trashDir string) (failed []string, err error) {
	entries, err := tenantfs.ReadDir(trashDir)
	if err != nil {
		return nil, err
	}
	left := map[string]bool{}
	for _, entry := range entries {
		if entry.Name() == ".trash_restore" {
			continue
		}
		if rmErr := removeTrashEntry(filepath.Join(trashDir, entry.Name())); rmErr != nil {
			failed = append(failed, entry.Name())
			left[entry.Name()] = true
		}
	}

	restorePath := filepath.Join(trashDir, ".trash_restore")
	var keep []string
	if len(left) > 0 {
		if content, readErr := tenantfs.ReadFile(restorePath); readErr == nil {
			for _, line := range strings.Split(string(content), "\n") {
				if name, _, ok := splitRestoreLine(strings.TrimSpace(line)); ok && left[name] {
					keep = append(keep, line)
				}
			}
		}
	}
	if len(keep) == 0 {
		return failed, tenantfs.WriteFile(restorePath, nil, 0o644)
	}
	return failed, tenantfs.WriteFile(restorePath, []byte(joinLines(keep)), 0o644)
}
