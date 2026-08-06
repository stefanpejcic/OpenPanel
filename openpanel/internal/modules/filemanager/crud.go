package filemanager

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// handleCreateFile creates an empty file at the given path and optionally
// redirects straight into the editor for it.
func handleCreateFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	filename := strings.TrimSpace(r.Form.Get("filename"))
	pathParam := strings.TrimPrefix(r.Form.Get("path_param"), "/")

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if filename == "" {
		flashAndRedirect(a, w, r, "error", "Filename is missing", filesRedirectPath(pathParam))
		return
	}

	filePath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, filename), false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}
	if _, statErr := os.Stat(filePath); statErr == nil {
		flashAndRedirect(a, w, r, "error", "Error creating file: already exists.", filesRedirectPath(pathParam))
		return
	}

	f, createErr := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0o644)
	if createErr != nil {
		flashAndRedirect(a, w, r, "error", "Error creating file! Check permissions.", filesRedirectPath(pathParam))
		return
	}
	_ = f.Close()

	sess, _ := a.Sessions.Get(r, session.CookieName)
	if chownErr := chownToUser(ctx, a, filePath, user.Context); chownErr != nil {
		flash.Add(sess, "error", "Failed to set ownership for the new file.")
	}

	flash.Add(sess, "success", "File created successfully.")
	_ = a.Sessions.Save(r, w, sess)
	_ = logger.RecordUserAction(a.Config, user.Username, "created a new file "+filename+" via File Manager", reqip.ClientIP(r))

	if r.Form.Get("open") == "true" {
		pathForRedirect := filepath.Join(pathParam, filename)
		http.Redirect(w, r, "/file-manager/edit-file/"+pathForRedirect+"?new=true", http.StatusFound)
		return
	}

	http.Redirect(w, r, filesRedirectPath(pathParam), http.StatusFound)
}

// handleCreateFolder creates a new directory at the given path.
func handleCreateFolder(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	folderName := strings.TrimSpace(r.Form.Get("foldername"))
	pathParam := strings.TrimPrefix(r.Form.Get("path_param"), "/")

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if folderName == "" {
		flashAndRedirect(a, w, r, "error", "Foldername is missing", filesRedirectPath(pathParam))
		return
	}

	folderPath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, folderName), false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}

	if mkErr := os.MkdirAll(folderPath, 0o755); mkErr != nil {
		flashAndRedirect(a, w, r, "error", "Error creating directory! Check permissions.", filesRedirectPath(pathParam))
		return
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)
	if chownErr := chownToUser(ctx, a, folderPath, user.Context); chownErr != nil {
		flash.Add(sess, "error", "Failed to set ownership for the new folder.")
	}
	flash.Add(sess, "success", "Folder created successfully.")
	_ = a.Sessions.Save(r, w, sess)

	_ = logger.RecordUserAction(a.Config, user.Username, "created a new folder "+folderName+" via File Manager", reqip.ClientIP(r))
	http.Redirect(w, r, filesRedirectPath(pathParam), http.StatusFound)
}

// handleRenameFile renames or moves a file/directory within the same
// listing by renaming it in place.
func handleRenameFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	oldName := r.Form.Get("old_name")
	newName := r.Form.Get("new_name")
	pathParam := r.Form.Get("path_param")

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	oldRelPath := filepath.Join(pathParam, oldName)
	newRelPath := filepath.Join(pathParam, newName)

	oldPath, perr := paths.SecureUserPath("HOME", user.Context, oldRelPath, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}
	newPath, perr := paths.SecureUserPath("HOME", user.Context, newRelPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}

	if renameErr := os.Rename(oldPath, newPath); renameErr != nil {
		flashAndRedirect(a, w, r, "error", "Error renaming item! Check permissions or if the new name already exists.", filesRedirectPath(pathParam))
		return
	}

	_ = logger.RecordUserAction(a.Config, user.Username,
		"renamed file /var/www/html/"+oldRelPath+" to /var/www/html/"+newRelPath+" using File Manager", reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", "File renamed successfully.", filesRedirectPath(pathParam))
}

// handleDeleteFile deletes a file or directory, either permanently or by
// moving it to the trash depending on the "mode" query parameter.
func handleDeleteFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	itemName := r.URL.Query().Get("filename")
	pathParam := r.URL.Query().Get("path_param")
	itemType := r.URL.Query().Get("item_type")
	if itemType == "" {
		itemType = "file"
	}
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "permanent"
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	targetRelPath := filepath.Join(pathParam, itemName)
	itemPath, perr := paths.SecureUserPath("HOME", user.Context, targetRelPath, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]any{"success": false, "error": msg})
		return
	}
	displayPath := filepath.Join(pathParam, itemName)

	if mode == "trash" {
		if _, trashErr := moveItemToTrash(ctx, a, itemPath, itemName, user.Context); trashErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Delete failed: " + trashErr.Error()})
			return
		}
		_ = logger.RecordUserAction(a.Config, user.Username, "moved "+itemType+" "+displayPath+" to Trash using File Manager", reqip.ClientIP(r))
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Moved to Trash"})
		return
	}

	var rmErr error
	if itemType == "directory" {
		rmErr = os.RemoveAll(itemPath)
	} else {
		rmErr = os.Remove(itemPath)
	}
	if rmErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Delete failed: " + rmErr.Error()})
		return
	}
	_ = logger.RecordUserAction(a.Config, user.Username, "deleted "+itemType+" "+displayPath+" using File Manager", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Deleted permanently"})
}

var permissionsRE = regexp.MustCompile(`^[0-7]{3,4}$`)

// handleChangePermissions applies an octal permissions string to one or
// more selected files.
func handleChangePermissions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	filenames := r.Form["filename"]
	permissions := r.Form.Get("permissions")
	pathParam := r.Form.Get("path_param")

	if !permissionsRE.MatchString(permissions) {
		flashAndRedirect(a, w, r, "error", "Invalid permissions format.", filesRedirectPath(pathParam))
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	mode, _ := strconv.ParseUint(permissions, 8, 32)

	changed := 0
	var errored []string
	for _, filename := range filenames {
		relPath := filepath.Join(pathParam, filename)
		filePath, perr := paths.SecureUserPath("HOME", user.Context, relPath, true)
		if perr != nil {
			errored = append(errored, filename)
			continue
		}
		if chmodErr := os.Chmod(filePath, os.FileMode(mode)); chmodErr != nil {
			errored = append(errored, filename)
			continue
		}
		_ = logger.RecordUserAction(a.Config, user.Username, "changed permissions to "+permissions+" for "+filePath+" using File Manager", reqip.ClientIP(r))
		changed++
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)
	if changed == 1 {
		flash.Add(sess, "success", "Permissions changed.")
	} else if changed > 1 {
		flash.Add(sess, "success", strconv.Itoa(changed)+" items updated.")
	}
	if len(errored) > 0 {
		flash.Add(sess, "error", "Error changing permissions for: "+strings.Join(errored, ", ")+". Check ownership.")
	}
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, filesRedirectPath(pathParam), http.StatusFound)
}

// handleCopyItem copies a file or directory to a destination path, failing
// if the destination already exists.
func handleCopyItem(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	itemName := q.Get("item_name")
	itemType := q.Get("item_type")
	pathParam := q.Get("path_param")
	destinationPath := q.Get("destination_path")
	if destinationPath == "" {
		destinationPath = pathParam
	}
	destinationPath = strings.TrimPrefix(destinationPath, "/")

	if itemName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "File name is empty"})
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	src, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, itemName), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]any{"success": false, "error": msg})
		return
	}
	dst, perr := paths.SecureUserPath("HOME", user.Context, destinationPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]any{"success": false, "error": msg})
		return
	}
	if info, statErr := os.Stat(dst); statErr == nil && info.IsDir() {
		dst = filepath.Join(dst, itemName)
	}

	var copyErr error
	if itemType == "directory" {
		copyErr = copyTree(src, dst)
	} else {
		copyErr = copyFile(src, dst)
	}
	if copyErr != nil {
		if os.IsExist(copyErr) {
			writeJSON(w, http.StatusConflict, map[string]any{"success": false, "error": "Destination already exists"})
			return
		}
		if os.IsPermission(copyErr) {
			writeJSON(w, http.StatusForbidden, map[string]any{"success": false, "error": "Permission denied"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "Error copying item: " + copyErr.Error()})
		return
	}

	chownRecursive(ctx, a, dst, user.Context)
	_ = logger.RecordUserAction(a.Config, user.Username,
		"copied "+itemType+" "+filepath.Join(pathParam, itemName)+" to "+destinationPath+" using File Manager", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleMoveItem moves a file or directory to a destination path.
func handleMoveItem(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	itemName := q.Get("item_name")
	pathParam := q.Get("path_param")
	destinationPath := strings.TrimPrefix(q.Get("destination_path"), "/")

	if itemName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "File name is empty"})
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	src, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, itemName), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]any{"success": false, "error": msg})
		return
	}
	dst, perr := paths.SecureUserPath("HOME", user.Context, destinationPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]any{"success": false, "error": msg})
		return
	}
	if info, statErr := os.Stat(dst); statErr == nil && info.IsDir() {
		dst = filepath.Join(dst, itemName)
	}

	if moveErr := os.Rename(src, dst); moveErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "Move failed: " + moveErr.Error()})
		return
	}

	if info, statErr := os.Stat(dst); statErr == nil && info.IsDir() {
		chownRecursive(ctx, a, dst, user.Context)
	} else {
		_ = chownToUser(ctx, a, dst, user.Context)
	}

	_ = logger.RecordUserAction(a.Config, user.Username,
		"moved "+filepath.Join(pathParam, itemName)+" to "+destinationPath+" using File Manager", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}
