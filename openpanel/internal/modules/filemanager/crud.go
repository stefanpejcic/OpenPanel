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
		http.Redirect(w, r, "/file-manager/edit-file/"+escapeURLPath(pathForRedirect)+"?new=true", http.StatusFound)
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

	if mkErr := mkdirNew(root.Root, rel, folderName); mkErr != nil {
		msg := "Error creating directory! Check permissions."
		switch {
		case errors.Is(mkErr, errFolderNameSlash):
			msg = "Folder name can't contain a slash."
		case errors.Is(mkErr, os.ErrExist):
			msg = "Error creating directory: already exists."
		}
		web.FlashRedirect(a, w, r, "error", msg, filesRedirectPath(pathParam))
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

// deleteItem removes one item and returns what it did, "symlink" for an unlinked symlink, "trash" or "permanent" otherwise
func deleteItem(a *appctx.App, r *http.Request, user injectedUser, pathParam, itemName, itemType, mode string) (string, int, error) {
	targetRelPath := filepath.Join(pathParam, itemName)

	// symlinks just get unlinked, never followed and never trashed
	if _, linkErr := paths.RemoveUserSymlink("HOME", user.Context, targetRelPath); !errors.Is(linkErr, paths.ErrNotSymlink) {
		if linkErr != nil {
			status, msg := pathErrorStatus(linkErr)
			return "", status, errors.New(msg)
		}
		return "symlink", http.StatusOK, nil
	}

	root, itemPath, perr := openUserPath(user, targetRelPath, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		return "", status, errors.New(msg)
	}
	defer root.Close()

	if mode == "trash" {
		if _, trashErr := moveItemToTrash(r.Context(), a, root, itemPath, itemName, user.Context); trashErr != nil {
			return "", http.StatusInternalServerError, errors.New("Delete failed: " + trashErr.Error())
		}
		return "trash", http.StatusOK, nil
	}

	var rmErr error
	if itemType == "directory" {
		rmErr = root.RemoveAll(itemPath)
	} else {
		rmErr = root.Remove(itemPath)
	}
	if rmErr != nil {
		return "", http.StatusInternalServerError, errors.New("Delete failed: " + rmErr.Error())
	}
	return "permanent", http.StatusOK, nil
}

// deleteLogLine is the activity line for one kind of delete, a single item keeps the old per-file wording
func deleteLogLine(kind, itemType string, names []string) string {
	if len(names) == 1 {
		switch kind {
		case "symlink":
			return "deleted symlink " + names[0] + " using File Manager"
		case "trash":
			return "moved " + itemType + " " + names[0] + " to Trash using File Manager"
		}
		return "deleted " + itemType + " " + names[0] + " using File Manager"
	}
	joined := logger.JoinNames(names)
	count := strconv.Itoa(len(names))
	switch kind {
	case "symlink":
		return "deleted " + count + " symlinks using File Manager: " + joined
	case "trash":
		return "moved " + count + " items to Trash using File Manager: " + joined
	}
	return "deleted " + count + " items using File Manager: " + joined
}

// itemFailure is one item a batch request couldn't handle
type itemFailure struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

// handleDeleteFile deletes one or more items (repeated filename/item_type params), permanently or to trash per "mode", and logs one activity line per kind instead of one per file
func handleDeleteFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	names := q["filename"]
	types := q["item_type"]
	pathParam := q.Get("path_param")
	mode := q.Get("mode")
	if mode == "" {
		mode = "permanent"
	}
	if len(names) == 0 {
		web.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "File name is empty"})
		return
	}

	user, err := currentUser(r.Context(), a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	done := map[string][]string{}
	doneType := map[string]string{}
	var failed []itemFailure
	failStatus := http.StatusOK
	for i, name := range names {
		itemType := "file"
		if i < len(types) && types[i] != "" {
			itemType = types[i]
		}
		kind, status, derr := deleteItem(a, r, user, pathParam, name, itemType, mode)
		if derr != nil {
			failed = append(failed, itemFailure{Name: name, Error: derr.Error()})
			if failStatus == http.StatusOK {
				failStatus = status
			}
			continue
		}
		done[kind] = append(done[kind], filepath.Join(pathParam, name))
		doneType[kind] = itemType
	}

	for _, kind := range []string{"trash", "permanent", "symlink"} {
		if len(done[kind]) > 0 {
			_ = logger.RecordUserAction(a.Config, user.Username, deleteLogLine(kind, doneType[kind], done[kind]), reqip.ClientIP(r))
		}
	}

	if len(failed) > 0 {
		web.WriteJSON(w, failStatus, map[string]any{"success": false, "error": failed[0].Error, "failed": failed})
		return
	}
	msg := "Deleted permanently"
	if len(done["trash"]) > 0 {
		msg = "Moved to Trash"
	} else if len(done["symlink"]) > 0 && len(done["permanent"]) == 0 {
		msg = "Symlink deleted"
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"success": true, "message": msg})
}

// a leading special digit (1777, 2755, 4755) used to pass and then get dropped by os.FileMode, the UI only sends three digits
var permissionsRE = regexp.MustCompile(`^0?[0-7]{3}$`)

var specialPermissionsRE = regexp.MustCompile(`^[1-7][0-7]{3}$`)

// permissionsError is "" for a valid mode, setuid/setgid/sticky are refused on purpose since they're risky in a web root
func permissionsError(permissions string) string {
	switch {
	case permissionsRE.MatchString(permissions):
		return ""
	case specialPermissionsRE.MatchString(permissions):
		return "Special permission bits (setuid, setgid, sticky) are not supported, use three digits like 755."
	}
	return "Invalid permissions format."
}

// chmodPath applies mode to rel, and when recursive is set and rel is a directory, dirMode to it and every folder under it and mode to the files; symlinks inside are skipped rather than followed
func chmodPath(root *paths.UserRoot, rel string, mode, dirMode os.FileMode, recursive bool) error {
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
		if d.IsDir() {
			return root.Chmod(itemPath, dirMode)
		}
		return root.Chmod(itemPath, mode)
	})
}

// recursiveModes parses the file mode and the optional folder mode, folders get the file mode when no folder mode is sent
func recursiveModes(permissions, folderPermissions string) (mode, dirMode os.FileMode, msg string) {
	if msg := permissionsError(permissions); msg != "" {
		return 0, 0, msg
	}
	if folderPermissions == "" {
		folderPermissions = permissions
	} else if msg := permissionsError(folderPermissions); msg != "" {
		return 0, 0, msg
	}
	m, _ := strconv.ParseUint(permissions, 8, 32)
	d, _ := strconv.ParseUint(folderPermissions, 8, 32)
	return os.FileMode(m), os.FileMode(d), ""
}

// chmodAction is the activity log line, it names both modes when they differ
func chmodAction(permissions, folderPermissions string, recursive bool) string {
	if !recursive {
		return "changed permissions to " + permissions
	}
	if folderPermissions != "" && folderPermissions != permissions {
		return "changed permissions to " + permissions + " for files and " + folderPermissions + " for folders recursively"
	}
	return "changed permissions to " + permissions + " recursively"
}

// handleChangePermissions applies an octal permissions string to one or more selected files, recursing into
// selected directories when the "recursive" checkbox is set
func handleChangePermissions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	filenames := r.Form["filename"]
	permissions := r.Form.Get("permissions")
	// only used with recursive, folders get the file mode when it's empty
	folderPermissions := r.Form.Get("folder_permissions")
	pathParam := r.Form.Get("path_param")
	recursive := r.Form.Get("recursive") == "on"
	// the /files permissions editor posts with fetch and updates the rows itself instead of reloading
	wantsJSON := strings.Contains(r.Header.Get("Accept"), "application/json")

	mode, dirMode, msg := recursiveModes(permissions, folderPermissions)
	if msg != "" {
		if wantsJSON {
			web.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": web.Tr(a, r, msg)})
			return
		}
		web.FlashRedirect(a, w, r, "error", msg, filesRedirectPath(pathParam))
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	root, perr := paths.OpenUserRoot("HOME", user.Context)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		if wantsJSON {
			web.WriteJSON(w, status, map[string]any{"success": false, "error": msg})
			return
		}
		http.Error(w, msg, status)
		return
	}
	defer root.Close()

	changed := []string{}
	errored := []string{}
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
		if chmodErr := chmodPath(root, rel, mode, dirMode, recursive && info.IsDir()); chmodErr != nil {
			errored = append(errored, filename)
			continue
		}
		action := chmodAction(permissions, folderPermissions, recursive && info.IsDir())
		_ = logger.RecordUserAction(a.Config, user.Username, action+" for "+root.Abs(rel)+" using File Manager", reqip.ClientIP(r))
		changed = append(changed, filename)
	}

	if wantsJSON {
		res := map[string]any{"success": len(errored) == 0, "changed": changed, "errored": errored}
		if len(errored) > 0 {
			res["error"] = web.Tr(a, r, "Error changing permissions for: %(errored)s. Check ownership.", "errored", strings.Join(errored, ", "))
		}
		web.WriteJSON(w, http.StatusOK, res)
		return
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)
	if len(changed) == 1 {
		flash.Add(sess, "success", "Permissions changed.")
	} else if len(changed) > 1 {
		flash.Add(sess, "success", web.Tr(a, r, "%(changed)s items updated.", "changed", strconv.Itoa(len(changed))))
	}
	if len(errored) > 0 {
		flash.Add(sess, "error", web.Tr(a, r, "Error changing permissions for: %(errored)s. Check ownership.", "errored", strings.Join(errored, ", ")))
	}
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, filesRedirectPath(pathParam), http.StatusFound)
}

// transferItem copies or moves one item into destinationPath, failing if the destination already exists
func transferItem(a *appctx.App, r *http.Request, user injectedUser, action, pathParam, itemName, itemType, destinationPath string) (int, error) {
	if itemName == "" {
		return http.StatusBadRequest, errors.New("File name is empty")
	}
	ctx := r.Context()
	root, src, perr := openUserPath(user, filepath.Join(pathParam, itemName), true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		return status, errors.New(msg)
	}
	defer root.Close()
	dst, perr := root.Rel(destinationPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		return status, errors.New(msg)
	}
	if info, statErr := root.Stat(dst); statErr == nil && info.IsDir() {
		dst = path.Join(dst, itemName)
	}

	if action == "move" {
		if moveErr := renameNoReplace(root.Root, src, dst); moveErr != nil {
			if errors.Is(moveErr, os.ErrExist) {
				return http.StatusConflict, errors.New("Destination already exists")
			}
			return http.StatusInternalServerError, errors.New("Move failed: " + moveErr.Error())
		}
		if info, statErr := root.Lstat(dst); statErr == nil && info.IsDir() {
			chownRecursive(ctx, a, root, dst, user.Context)
		} else {
			_ = chownToUser(ctx, a, root, dst, user.Context)
		}
		return http.StatusOK, nil
	}

	var copyErr error
	if itemType == "directory" {
		copyErr = copyTree(root.Root, src, dst)
	} else {
		copyErr = copyFile(root.Root, src, dst)
	}
	if copyErr != nil {
		if os.IsExist(copyErr) {
			return http.StatusConflict, errors.New("Destination already exists")
		}
		if os.IsPermission(copyErr) {
			return http.StatusForbidden, errors.New("Permission denied")
		}
		return http.StatusInternalServerError, errors.New("Error copying item: " + copyErr.Error())
	}
	chownRecursive(ctx, a, root, dst, user.Context)
	return http.StatusOK, nil
}

// transferLogLine is the activity line for a copy or move, a single item keeps the old wording
func transferLogLine(action, itemType string, names []string, destinationPath string) string {
	verb := "copied "
	if action == "move" {
		verb = "moved "
	}
	if len(names) == 1 {
		if action == "copy" {
			verb += itemType + " "
		}
		return verb + names[0] + " to " + destinationPath + " using File Manager"
	}
	return verb + strconv.Itoa(len(names)) + " items to " + destinationPath + " using File Manager: " + logger.JoinNames(names)
}

// handleTransfer copies or moves one or more items (repeated item_name/item_type params) and logs a single activity line for the batch
func handleTransfer(a *appctx.App, w http.ResponseWriter, r *http.Request, action string) {
	q := r.URL.Query()
	names := q["item_name"]
	types := q["item_type"]
	pathParam := q.Get("path_param")
	destinationPath := q.Get("destination_path")
	if destinationPath == "" && action == "copy" {
		destinationPath = pathParam
	}
	destinationPath = strings.TrimPrefix(destinationPath, "/")

	if len(names) == 0 {
		web.WriteJSON(w, http.StatusBadRequest, map[string]any{"success": false, "error": "File name is empty"})
		return
	}

	user, err := currentUser(r.Context(), a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var done []string
	var failed []itemFailure
	failStatus := http.StatusOK
	itemType := "file"
	for i, name := range names {
		t := "file"
		if i < len(types) && types[i] != "" {
			t = types[i]
		}
		status, terr := transferItem(a, r, user, action, pathParam, name, t, destinationPath)
		if terr != nil {
			failed = append(failed, itemFailure{Name: name, Error: terr.Error()})
			if failStatus == http.StatusOK {
				failStatus = status
			}
			continue
		}
		done = append(done, filepath.Join(pathParam, name))
		itemType = t
	}

	if len(done) > 0 {
		_ = logger.RecordUserAction(a.Config, user.Username, transferLogLine(action, itemType, done, destinationPath), reqip.ClientIP(r))
	}
	if len(failed) > 0 {
		web.WriteJSON(w, failStatus, map[string]any{"success": false, "error": failed[0].Error, "failed": failed})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"success": true})
}

// handleCopyItem copies items to a destination path, failing for any that already exist there
func handleCopyItem(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	handleTransfer(a, w, r, "copy")
}

// handleMoveItem moves items to a destination path
func handleMoveItem(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	handleTransfer(a, w, r, "move")
}
