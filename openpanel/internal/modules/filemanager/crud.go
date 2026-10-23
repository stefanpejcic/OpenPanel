package filemanager

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path"
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
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleCreateFile creates an empty file at the given path and optionally redirects straight into the editor for it
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
		web.FlashRedirect(a, w, r, "error", "Filename is missing", filesRedirectPath(pathParam))
		return
	}

	root, rel, perr := openUserPath(user, filepath.Join(pathParam, filename), false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}
	defer root.Close()

	f, createErr := root.OpenFile(rel, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if createErr != nil {
		if os.IsExist(createErr) {
			web.FlashRedirect(a, w, r, "error", "Error creating file: already exists.", filesRedirectPath(pathParam))
			return
		}
		web.FlashRedirect(a, w, r, "error", "Error creating file! Check permissions.", filesRedirectPath(pathParam))
		return
	}
	_ = f.Close()

	sess, _ := a.Sessions.Get(r, session.CookieName)
	if chownErr := chownToUser(ctx, a, root, rel, user.Context); chownErr != nil {
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
		web.FlashRedirect(a, w, r, "error", "Foldername is missing", filesRedirectPath(pathParam))
		return
	}

	root, rel, perr := openUserPath(user, filepath.Join(pathParam, folderName), false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}
	defer root.Close()

	if mkErr := root.MkdirAll(rel, 0o755); mkErr != nil {
		web.FlashRedirect(a, w, r, "error", "Error creating directory! Check permissions.", filesRedirectPath(pathParam))
		return
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)
	if chownErr := chownToUser(ctx, a, root, rel, user.Context); chownErr != nil {
		flash.Add(sess, "error", "Failed to set ownership for the new folder.")
	}
	flash.Add(sess, "success", "Folder created successfully.")
	_ = a.Sessions.Save(r, w, sess)

	_ = logger.RecordUserAction(a.Config, user.Username, "created a new folder "+folderName+" via File Manager", reqip.ClientIP(r))
	http.Redirect(w, r, filesRedirectPath(pathParam), http.StatusFound)
}

// handleRenameFile renames or moves a file/directory within the same listing by renaming it in place
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

	root, oldPath, perr := openUserPath(user, oldRelPath, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}
	defer root.Close()
	newPath, perr := root.Rel(newRelPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}

	if renameErr := renameNoReplace(root.Root, oldPath, newPath); renameErr != nil {
		if errors.Is(renameErr, os.ErrExist) {
			web.FlashRedirect(a, w, r, "error", "Error renaming item: the new name already exists.", filesRedirectPath(pathParam))
			return
		}
		web.FlashRedirect(a, w, r, "error", "Error renaming item! Check permissions or if the new name already exists.", filesRedirectPath(pathParam))
		return
	}

	_ = logger.RecordUserAction(a.Config, user.Username,
		"renamed file /var/www/html/"+oldRelPath+" to /var/www/html/"+newRelPath+" using File Manager", reqip.ClientIP(r))
	web.FlashRedirect(a, w, r, "success", "File renamed successfully.", filesRedirectPath(pathParam))
}

// handleDeleteFile deletes a file or directory, either permanently or by moving it to the trash depending on the "mode" query parameter
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

	// symlinks just get unlinked, never followed and never trashed
	if _, linkErr := paths.RemoveUserSymlink("HOME", user.Context, targetRelPath); !errors.Is(linkErr, paths.ErrNotSymlink) {
		if linkErr != nil {
			status, msg := pathErrorStatus(linkErr)
			web.WriteJSON(w, status, map[string]any{"success": false, "error": msg})
			return
		}
		_ = logger.RecordUserAction(a.Config, user.Username, "deleted symlink "+targetRelPath+" using File Manager", reqip.ClientIP(r))
		web.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Symlink deleted"})
		return
	}

	root, itemPath, perr := openUserPath(user, targetRelPath, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]any{"success": false, "error": msg})
		return
	}
	defer root.Close()
	displayPath := filepath.Join(pathParam, itemName)

	if mode == "trash" {
		if _, trashErr := moveItemToTrash(ctx, a, root, itemPath, itemName, user.Context); trashErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Delete failed: " + trashErr.Error()})
			return
		}
		_ = logger.RecordUserAction(a.Config, user.Username, "moved "+itemType+" "+displayPath+" to Trash using File Manager", reqip.ClientIP(r))
		web.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Moved to Trash"})
		return
	}

	var rmErr error
	if itemType == "directory" {
		rmErr = root.RemoveAll(itemPath)
	} else {
		rmErr = root.Remove(itemPath)
	}
	if rmErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "message": "Delete failed: " + rmErr.Error()})
		return
	}
	_ = logger.RecordUserAction(a.Config, user.Username, "deleted "+itemType+" "+displayPath+" using File Manager", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": "Deleted permanently"})
}

var permissionsRE = regexp.MustCompile(`^[0-7]{3,4}$`)

// chmodPath applies mode to rel, and to everything under it when recursive is set and rel is a directory, symlinks inside are skipped rather than followed
func chmodPath(root *paths.UserRoot, rel string, mode os.FileMode, recursive bool) error {
	if !recursive {
		return root.Chmod(rel, mode)
	}
	return fs.WalkDir(root.FS(), rel, func(itemPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		return root.Chmod(itemPath, mode)
	})
}

// handleChangePermissions applies an octal permissions string to one or more selected files, recursing into
// selected directories when the "recursive" checkbox is set
func handleChangePermissions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	filenames := r.Form["filename"]
	permissions := r.Form.Get("permissions")
	pathParam := r.Form.Get("path_param")
	recursive := r.Form.Get("recursive") == "on"

	if !permissionsRE.MatchString(permissions) {
		web.FlashRedirect(a, w, r, "error", "Invalid permissions format.", filesRedirectPath(pathParam))
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	mode, _ := strconv.ParseUint(permissions, 8, 32)

	root, perr := paths.OpenUserRoot("HOME", user.Context)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		http.Error(w, msg, status)
		return
	}
	defer root.Close()

	changed := 0
	var errored []string
	for _, filename := range filenames {
		relPath := filepath.Join(pathParam, filename)
		rel, perr := root.Rel(relPath, true)
		if perr != nil {
			errored = append(errored, filename)
			continue
		}
		info, statErr := root.Lstat(rel)
		if statErr != nil {
			errored = append(errored, filename)
			continue
		}
		if chmodErr := chmodPath(root, rel, os.FileMode(mode), recursive && info.IsDir()); chmodErr != nil {
			errored = append(errored, filename)
			continue
		}
		action := "changed permissions to " + permissions
		if recursive && info.IsDir() {
			action += " recursively"
		}
		_ = logger.RecordUserAction(a.Config, user.Username, action+" for "+root.Abs(rel)+" using File Manager", reqip.ClientIP(r))
		changed++
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)
	if changed == 1 {
		flash.Add(sess, "success", "Permissions changed.")
	} else if changed > 1 {
		flash.Add(sess, "success", web.Tr(a, r, "%(changed)s items updated.", "changed", strconv.Itoa(changed)))
	}
	if len(errored) > 0 {
		flash.Add(sess, "error", web.Tr(a, r, "Error changing permissions for: %(errored)s. Check ownership.", "errored", strings.Join(errored, ", ")))
	}
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, filesRedirectPath(pathParam), http.StatusFound)
}

// handleCopyItem copies a file or directory to a destination path, failing if the destination already exists
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
		web.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "File name is empty"})
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	root, src, perr := openUserPath(user, filepath.Join(pathParam, itemName), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]any{"success": false, "error": msg})
		return
	}
	defer root.Close()
	dst, perr := root.Rel(destinationPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]any{"success": false, "error": msg})
		return
	}
	if info, statErr := root.Stat(dst); statErr == nil && info.IsDir() {
		dst = path.Join(dst, itemName)
	}

	var copyErr error
	if itemType == "directory" {
		copyErr = copyTree(root.Root, src, dst)
	} else {
		copyErr = copyFile(root.Root, src, dst)
	}
	if copyErr != nil {
		if os.IsExist(copyErr) {
			web.WriteJSON(w, http.StatusConflict, map[string]any{"success": false, "error": "Destination already exists"})
			return
		}
		if os.IsPermission(copyErr) {
			web.WriteJSON(w, http.StatusForbidden, map[string]any{"success": false, "error": "Permission denied"})
			return
		}
		web.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "Error copying item: " + copyErr.Error()})
		return
	}

	chownRecursive(ctx, a, root, dst, user.Context)
	_ = logger.RecordUserAction(a.Config, user.Username,
		"copied "+itemType+" "+filepath.Join(pathParam, itemName)+" to "+destinationPath+" using File Manager", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleMoveItem moves a file or directory to a destination path.
func handleMoveItem(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	itemName := q.Get("item_name")
	pathParam := q.Get("path_param")
	destinationPath := strings.TrimPrefix(q.Get("destination_path"), "/")

	if itemName == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "File name is empty"})
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	root, src, perr := openUserPath(user, filepath.Join(pathParam, itemName), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]any{"success": false, "error": msg})
		return
	}
	defer root.Close()
	dst, perr := root.Rel(destinationPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]any{"success": false, "error": msg})
		return
	}
	if info, statErr := root.Stat(dst); statErr == nil && info.IsDir() {
		dst = path.Join(dst, itemName)
	}

	if moveErr := renameNoReplace(root.Root, src, dst); moveErr != nil {
		if errors.Is(moveErr, os.ErrExist) {
			web.WriteJSON(w, http.StatusConflict, map[string]any{"success": false, "error": "Destination already exists"})
			return
		}
		web.WriteJSON(w, http.StatusInternalServerError, map[string]any{"success": false, "error": "Move failed: " + moveErr.Error()})
		return
	}

	if info, statErr := root.Lstat(dst); statErr == nil && info.IsDir() {
		chownRecursive(ctx, a, root, dst, user.Context)
	} else {
		_ = chownToUser(ctx, a, root, dst, user.Context)
	}

	_ = logger.RecordUserAction(a.Config, user.Username,
		"moved "+filepath.Join(pathParam, itemName)+" to "+destinationPath+" using File Manager", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}
