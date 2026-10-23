package filemanager

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// apiCreateFile is handleCreateFile (crud.go) with a JSON body/response instead of a form post + flash/redirect.
func apiCreateFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path     string `json:"path"`
		Filename string `json:"filename"`
	}
	decodeAPIBody(r, &body, func(form url.Values) {
		body.Path = form.Get("path_param")
		body.Filename = form.Get("filename")
	})
	filename := strings.TrimSpace(body.Filename)
	pathParam := strings.TrimPrefix(body.Path, "/")

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if filename == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Filename is missing"})
		return
	}

	root, rel, perr := openUserPath(user, filepath.Join(pathParam, filename), false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	defer root.Close()

	f, createErr := root.OpenFile(rel, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if createErr != nil {
		if os.IsExist(createErr) {
			web.WriteJSON(w, http.StatusConflict, map[string]string{"error": "Error creating file: already exists."})
			return
		}
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error creating file! Check permissions."})
		return
	}
	_ = f.Close()

	warning := ""
	if chownErr := chownToUser(ctx, a, root, rel, user.Context); chownErr != nil {
		warning = "Failed to set ownership for the new file."
	}

	_ = logger.RecordUserAction(a.Config, user.Username, "created a new file "+filename+" via File Manager API", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusCreated, map[string]string{"message": "File created successfully.", "path": filepath.Join(pathParam, filename), "warning": warning})
}

// apiCreateFolder is handleCreateFolder (crud.go) with a JSON body/response.
func apiCreateFolder(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path       string `json:"path"`
		Foldername string `json:"foldername"`
	}
	decodeAPIBody(r, &body, func(form url.Values) {
		body.Path = form.Get("path_param")
		body.Foldername = form.Get("foldername")
	})
	folderName := strings.TrimSpace(body.Foldername)
	pathParam := strings.TrimPrefix(body.Path, "/")

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if folderName == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Foldername is missing"})
		return
	}

	root, rel, perr := openUserPath(user, filepath.Join(pathParam, folderName), false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	defer root.Close()

	if mkErr := root.MkdirAll(rel, 0o755); mkErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error creating directory! Check permissions."})
		return
	}

	warning := ""
	if chownErr := chownToUser(ctx, a, root, rel, user.Context); chownErr != nil {
		warning = "Failed to set ownership for the new folder."
	}

	_ = logger.RecordUserAction(a.Config, user.Username, "created a new folder "+folderName+" via File Manager API", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusCreated, map[string]string{"message": "Folder created successfully.", "path": filepath.Join(pathParam, folderName), "warning": warning})
}

// apiRenameFile is handleRenameFile (crud.go) with a JSON body/response.
func apiRenameFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path    string `json:"path"`
		OldName string `json:"old_name"`
		NewName string `json:"new_name"`
	}
	decodeAPIBody(r, &body, func(form url.Values) {
		body.Path = form.Get("path_param")
		body.OldName = form.Get("old_name")
		body.NewName = form.Get("new_name")
	})

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	oldRelPath := filepath.Join(body.Path, body.OldName)
	newRelPath := filepath.Join(body.Path, body.NewName)

	root, oldPath, perr := openUserPath(user, oldRelPath, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	defer root.Close()
	newPath, perr := root.Rel(newRelPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}

	if renameErr := renameNoReplace(root.Root, oldPath, newPath); renameErr != nil {
		if errors.Is(renameErr, os.ErrExist) {
			web.WriteJSON(w, http.StatusConflict, map[string]string{"error": "Error renaming item: the new name already exists."})
			return
		}
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error renaming item! Check permissions or if the new name already exists."})
		return
	}

	_ = logger.RecordUserAction(a.Config, user.Username,
		"renamed file /var/www/html/"+oldRelPath+" to /var/www/html/"+newRelPath+" using File Manager API", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "File renamed successfully.", "path": newRelPath})
}

// apiChangePermissions is handleChangePermissions (crud.go) with a JSON body/response
func apiChangePermissions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path        string   `json:"path"`
		Filenames   []string `json:"filenames"`
		Permissions string   `json:"permissions"`
		Recursive   bool     `json:"recursive"`
	}
	decodeAPIBody(r, &body, func(form url.Values) {
		body.Path = form.Get("path_param")
		body.Filenames = form["filename"]
		body.Permissions = form.Get("permissions")
		body.Recursive = form.Get("recursive") == "on"
	})

	if !permissionsRE.MatchString(body.Permissions) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid permissions format."})
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	mode, _ := strconv.ParseUint(body.Permissions, 8, 32)

	root, perr := paths.OpenUserRoot("HOME", user.Context)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		web.WriteJSON(w, status, map[string]string{"error": msg})
		return
	}
	defer root.Close()

	changed := []string{}
	var errored []string
	for _, filename := range body.Filenames {
		relPath := filepath.Join(body.Path, filename)
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
		if chmodErr := chmodPath(root, rel, os.FileMode(mode), body.Recursive && info.IsDir()); chmodErr != nil {
			errored = append(errored, filename)
			continue
		}
		action := "changed permissions to " + body.Permissions
		if body.Recursive && info.IsDir() {
			action += " recursively"
		}
		_ = logger.RecordUserAction(a.Config, user.Username, action+" for "+root.Abs(rel)+" using File Manager API", reqip.ClientIP(r))
		changed = append(changed, filename)
	}

	web.WriteJSON(w, http.StatusOK, map[string]any{"changed": changed, "errored": errored})
}
