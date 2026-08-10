package trash

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterAPI wires the /api/trash routes onto mux, gated behind the
// "trash" feature flag. handleRestoreFile and handleDeleteTrash already
// speak pure JSON with no session/flash dependency and are reused here
// unmodified; handleFilesInTrash already supports "?output=json" and is
// invoked with that forced on, matching filemanager's own
// forceJSONOutput pattern. Restore-all/empty-trash needed a JSON-response
// variant since the web handlers only flash+redirect.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "trash", "GET /api/trash", func(w http.ResponseWriter, r *http.Request) {
		handleFilesInTrash(a, w, forceJSONOutput(r), "")
	})
	apiregistry.Handle(mux, a, "trash", "GET /api/trash/{path_param...}", func(w http.ResponseWriter, r *http.Request) {
		handleFilesInTrash(a, w, forceJSONOutput(r), r.PathValue("path_param"))
	})
	apiregistry.Handle(mux, a, "trash", "POST /api/trash/restore", func(w http.ResponseWriter, r *http.Request) { handleRestoreFile(a, w, r) })
	apiregistry.Handle(mux, a, "trash", "DELETE /api/trash/delete", func(w http.ResponseWriter, r *http.Request) { handleDeleteTrash(a, w, r) })
	apiregistry.Handle(mux, a, "trash", "POST /api/trash/restore-all", func(w http.ResponseWriter, r *http.Request) { apiRestoreAll(a, w, r) })
	apiregistry.Handle(mux, a, "trash", "POST /api/trash/delete-all", func(w http.ResponseWriter, r *http.Request) { apiDeleteAll(a, w, r) })
}

// forceJSONOutput clones r with output=json set on the query string, so
// handleFilesInTrash's existing "?output=json" branch can be reused
// unmodified for a dedicated /api/ route.
func forceJSONOutput(r *http.Request) *http.Request {
	q := r.URL.Query()
	q.Set("output", "json")
	r2 := r.Clone(r.Context())
	r2.URL.RawQuery = q.Encode()
	return r2
}

// apiDeleteAll is handleDeleteAll (crud.go) with a JSON response instead
// of flash+redirect.
func apiDeleteAll(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	trashDir := "/home/" + userContext + "/.local/share/Trash"

	entries, err := os.ReadDir(trashDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, jsonResult{Success: false, Error: "Unexpected error"})
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
		writeJSON(w, http.StatusInternalServerError, jsonResult{Success: false, Error: "Unexpected error"})
		return
	}

	if err := os.WriteFile(filepath.Join(trashDir, ".trash_restore"), nil, 0o644); err != nil {
		writeJSON(w, http.StatusInternalServerError, jsonResult{Success: false, Error: "Unexpected error"})
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "Emptied trash via API", reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, jsonResult{Success: true, Message: "Trash emptied successfully"})
}

// apiRestoreAll is handleRestoreAll (crud.go) with a JSON response
// instead of flash+redirect.
func apiRestoreAll(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		writeJSON(w, http.StatusOK, map[string]any{"restored": []string{}, "errors": []string{}, "message": "Nothing to restore"})
		return
	}

	allowedBase := absPath(volume)

	var restored []string
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
		_ = logger.RecordUserAction(a.Config, username, "restored "+restorePath+" from Trash using Restore All (API)", reqip.ClientIP(r))
		restored = append(restored, key)
	}

	if restored == nil {
		restored = []string{}
	}
	if restoreErrors == nil {
		restoreErrors = []string{}
	}

	if len(linesToKeep) > 0 {
		_ = os.WriteFile(trashRestorePath, []byte(joinLines(linesToKeep)), 0o644)
	} else {
		_ = os.WriteFile(trashRestorePath, nil, 0o644)
	}

	writeJSON(w, http.StatusOK, map[string]any{"restored": restored, "errors": restoreErrors})
}
