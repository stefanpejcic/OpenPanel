package trash

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// itemFailure is one item a batch request couldn't handle
type itemFailure struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

type batchResult struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Error   string        `json:"error,omitempty"`
	Failed  []itemFailure `json:"failed,omitempty"`
}

// cleanTrashName keeps only the base name so a request can't reach outside the Trash folder
func cleanTrashName(itemName string) (string, error) {
	itemName = filepath.Base(itemName)
	if itemName == "" || itemName == "." || itemName == "/" {
		return "", errors.New("Invalid filename.")
	}
	return itemName, nil
}

// restoreOne moves one trashed item back to its original path from .trash_restore and returns that path
func restoreOne(userContext, itemName string) (string, error) {
	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
	trashDir := "/home/" + userContext + "/.local/share/Trash"

	itemName, err := cleanTrashName(itemName)
	if err != nil {
		return "", err
	}

	trashedPath := filepath.Join(trashDir, itemName)
	trashRestoreFile := filepath.Join(trashDir, ".trash_restore")

	if _, err := tenantfs.Stat(trashedPath); err != nil {
		return "", errors.New("File or directory does not exist in Trash.")
	}

	restoreInfo, err := tenantfs.ReadFile(trashRestoreFile)
	if err != nil {
		return "", errors.New(".trash_restore file is missing.")
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
		return "", errors.New("Original path not found in .trash_restore.")
	}

	restoreDestination = absPath(restoreDestination)
	if !isWithin(resolveOrSelf(restoreDestination), resolveOrSelf(absPath(volume))) {
		return "", errors.New("Restore path is outside your permitted data directory.")
	}

	if err := restoreItem(trashedPath, restoreDestination); err != nil {
		if errors.Is(err, errRestoreExists) {
			return "", errors.New("Can't restore: an item with the same name is already in the original folder.")
		}
		return "", err
	}

	_ = tenantfs.WriteFile(trashRestoreFile, []byte(joinLines(linesToKeep)), 0o644)
	return restoreDestination, nil
}

// deleteTrashOne permanently removes one item from the trash and returns its type and path
func deleteTrashOne(userContext, itemName string) (string, string, error) {
	if itemName == ".trash_restore" {
		return "", "", errors.New("Deletion of system file .trash_restore is not allowed.")
	}
	itemName, err := cleanTrashName(itemName)
	if err != nil {
		return "", "", err
	}

	trashDir := "/home/" + userContext + "/.local/share/Trash"
	trashedPath := filepath.Join(trashDir, itemName)
	trashRestoreFile := filepath.Join(trashDir, ".trash_restore")

	info, err := tenantfs.Lstat(trashedPath)
	if err != nil {
		return "", "", errors.New("File or directory does not exist in Trash.")
	}

	var detectedType string
	switch {
	case info.IsDir():
		if err := tenantfs.RemoveAll(trashedPath); err != nil {
			return "", "", err
		}
		detectedType = "directory"
	case info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0:
		if err := tenantfs.Remove(trashedPath); err != nil {
			return "", "", err
		}
		detectedType = "file"
	default:
		return "", "", errors.New("Unsupported item type.")
	}

	if content, err := tenantfs.ReadFile(trashRestoreFile); err == nil {
		_ = tenantfs.WriteFile(trashRestoreFile, []byte(joinLines(dropRestoreLine(string(content), itemName))), 0o644)
	}
	return detectedType, trashedPath, nil
}

// restoreLogLine is the activity line for restored items, a single item keeps the old wording
func restoreLogLine(paths []string, via string) string {
	if len(paths) == 1 {
		return "restored " + paths[0] + " from Trash using " + via
	}
	return "restored " + strconv.Itoa(len(paths)) + " items from Trash using " + via + ": " + logger.JoinNames(paths)
}

// trashDeleteLogLine is the activity line for items deleted from the trash, a single item keeps the old wording
func trashDeleteLogLine(itemType string, paths []string) string {
	if len(paths) == 1 {
		return "permanently deleted " + itemType + " " + paths[0] + " using File Manager"
	}
	return "permanently deleted " + strconv.Itoa(len(paths)) + " items from Trash using File Manager: " + logger.JoinNames(paths)
}

func writeBatch(w http.ResponseWriter, failed []itemFailure, okMsg string) {
	if len(failed) > 0 {
		web.WriteJSON(w, http.StatusOK, batchResult{Success: false, Error: failed[0].Error, Failed: failed})
		return
	}
	web.WriteJSON(w, http.StatusOK, batchResult{Success: true, Message: okMsg})
}

// handleRestoreFile moves one or more trashed items (repeated filename params) back to their original paths and logs one line for the batch
func handleRestoreFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	names := r.URL.Query()["filename"]
	if len(names) == 0 || names[0] == "" {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Missing filename."})
		return
	}

	_, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var restored []string
	var failed []itemFailure
	for _, name := range names {
		dest, rerr := restoreOne(userContext, name)
		if rerr != nil {
			failed = append(failed, itemFailure{Name: name, Error: rerr.Error()})
			continue
		}
		restored = append(restored, dest)
	}

	if len(restored) > 0 {
		_ = logger.RecordUserAction(a.Config, username, restoreLogLine(restored, "File Manager"), reqip.ClientIP(r))
	}
	writeBatch(w, failed, "Restored from Trash")
}

// handleDeleteTrash permanently deletes one or more items (repeated filename params) from the trash and logs one line for the batch
func handleDeleteTrash(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	names := r.URL.Query()["filename"]
	if len(names) == 0 {
		web.WriteJSON(w, http.StatusOK, jsonResult{Success: false, Error: "Invalid filename."})
		return
	}

	_, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var deleted []string
	var failed []itemFailure
	itemType := "file"
	for _, name := range names {
		t, p, derr := deleteTrashOne(userContext, name)
		if derr != nil {
			failed = append(failed, itemFailure{Name: name, Error: derr.Error()})
			continue
		}
		deleted = append(deleted, p)
		itemType = t
	}

	if len(deleted) > 0 {
		_ = logger.RecordUserAction(a.Config, username, trashDeleteLogLine(itemType, deleted), reqip.ClientIP(r))
	}
	writeBatch(w, failed, "Deleted permanently")
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
	var restored []string

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
		restored = append(restored, restorePath)
	}
	if len(restored) > 0 {
		_ = logger.RecordUserAction(a.Config, username, restoreLogLine(restored, "Restore All"), reqip.ClientIP(r))
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
