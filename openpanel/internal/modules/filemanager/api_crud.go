package filemanager

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// apiCreateFile is handleCreateFile (crud.go) with a JSON body/response
// instead of a form post + flash/redirect.
func apiCreateFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path     string `json:"path"`
		Filename string `json:"filename"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Path = r.Form.Get("path_param")
		body.Filename = r.Form.Get("filename")
	}
	filename := strings.TrimSpace(body.Filename)
	pathParam := strings.TrimPrefix(body.Path, "/")

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if filename == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Filename is missing"})
		return
	}

	filePath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, filename), false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}
	if _, statErr := os.Stat(filePath); statErr == nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Error creating file: already exists."})
		return
	}

	f, createErr := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0o644)
	if createErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error creating file! Check permissions."})
		return
	}
	_ = f.Close()

	warning := ""
	if chownErr := chownToUser(ctx, a, filePath, user.Context); chownErr != nil {
		warning = "Failed to set ownership for the new file."
	}

	_ = logger.RecordUserAction(a.Config, user.Username, "created a new file "+filename+" via File Manager API", reqip.ClientIP(r))
	writeJSON(w, http.StatusCreated, map[string]string{"message": "File created successfully.", "path": filepath.Join(pathParam, filename), "warning": warning})
}

// apiCreateFolder is handleCreateFolder (crud.go) with a JSON body/response.
func apiCreateFolder(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path       string `json:"path"`
		Foldername string `json:"foldername"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Path = r.Form.Get("path_param")
		body.Foldername = r.Form.Get("foldername")
	}
	folderName := strings.TrimSpace(body.Foldername)
	pathParam := strings.TrimPrefix(body.Path, "/")

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if folderName == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Foldername is missing"})
		return
	}

	folderPath, perr := paths.SecureUserPath("HOME", user.Context, filepath.Join(pathParam, folderName), false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}

	if mkErr := os.MkdirAll(folderPath, 0o755); mkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error creating directory! Check permissions."})
		return
	}

	warning := ""
	if chownErr := chownToUser(ctx, a, folderPath, user.Context); chownErr != nil {
		warning = "Failed to set ownership for the new folder."
	}

	_ = logger.RecordUserAction(a.Config, user.Username, "created a new folder "+folderName+" via File Manager API", reqip.ClientIP(r))
	writeJSON(w, http.StatusCreated, map[string]string{"message": "Folder created successfully.", "path": filepath.Join(pathParam, folderName), "warning": warning})
}

// apiRenameFile is handleRenameFile (crud.go) with a JSON body/response.
func apiRenameFile(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path    string `json:"path"`
		OldName string `json:"old_name"`
		NewName string `json:"new_name"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Path = r.Form.Get("path_param")
		body.OldName = r.Form.Get("old_name")
		body.NewName = r.Form.Get("new_name")
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	oldRelPath := filepath.Join(body.Path, body.OldName)
	newRelPath := filepath.Join(body.Path, body.NewName)

	oldPath, perr := paths.SecureUserPath("HOME", user.Context, oldRelPath, true)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}
	newPath, perr := paths.SecureUserPath("HOME", user.Context, newRelPath, false)
	if perr != nil {
		status, msg := pathErrorStatus(perr)
		writeJSON(w, status, map[string]string{"error": msg})
		return
	}

	if renameErr := os.Rename(oldPath, newPath); renameErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error renaming item! Check permissions or if the new name already exists."})
		return
	}

	_ = logger.RecordUserAction(a.Config, user.Username,
		"renamed file /var/www/html/"+oldRelPath+" to /var/www/html/"+newRelPath+" using File Manager API", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "File renamed successfully.", "path": newRelPath})
}

// apiChangePermissions is handleChangePermissions (crud.go) with a JSON
// body/response.
func apiChangePermissions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path        string   `json:"path"`
		Filenames   []string `json:"filenames"`
		Permissions string   `json:"permissions"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Path = r.Form.Get("path_param")
		body.Filenames = r.Form["filename"]
		body.Permissions = r.Form.Get("permissions")
	}

	if !permissionsRE.MatchString(body.Permissions) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid permissions format."})
		return
	}

	ctx := r.Context()
	user, err := currentUser(ctx, a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	mode, _ := strconv.ParseUint(body.Permissions, 8, 32)

	changed := []string{}
	var errored []string
	for _, filename := range body.Filenames {
		relPath := filepath.Join(body.Path, filename)
		filePath, perr := paths.SecureUserPath("HOME", user.Context, relPath, true)
		if perr != nil {
			errored = append(errored, filename)
			continue
		}
		if chmodErr := os.Chmod(filePath, os.FileMode(mode)); chmodErr != nil {
			errored = append(errored, filename)
			continue
		}
		_ = logger.RecordUserAction(a.Config, user.Username, "changed permissions to "+body.Permissions+" for "+filePath+" using File Manager API", reqip.ClientIP(r))
		changed = append(changed, filename)
	}

	writeJSON(w, http.StatusOK, map[string]any{"changed": changed, "errored": errored})
}
