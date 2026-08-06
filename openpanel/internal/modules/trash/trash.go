// Package trash implements the standalone /files.trash list/restore/
// empty-trash page. The minimal subset needed by filemanager's
// delete-to-trash action (moveItemToTrash / uniqueTrashName) already
// lives in internal/modules/filemanager/trash.go - that helper pair is
// duplicated here rather than shared, to keep the two features
// independent.
package trash

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// Register wires the trash routes onto mux, gated behind the "trash"
// feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "trash")(h)
	}

	mux.Handle("GET /files.trash", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFilesInTrash(a, w, r, "") }))
	// "GET /files.trash/{path_param...}" already covers the bare
	// "/files.trash/" case (path_param resolves to ""), so no separate
	// registration is needed for it - see filemanager.go's identical note.
	mux.Handle("GET /files.trash/{path_param...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleFilesInTrash(a, w, r, r.PathValue("path_param"))
	}))

	mux.Handle("POST /file-manager/restoreTrash", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRestoreFile(a, w, r) }))
	mux.Handle("DELETE /file-manager/deleteTrash", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteTrash(a, w, r) }))
	mux.Handle("POST /files.trash/deleteall", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteAll(a, w, r) }))
	mux.Handle("POST /files.trash/restoreall", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRestoreAll(a, w, r) }))
}

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

func flashAndRedirectToTrash(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, "/files.trash", http.StatusFound)
}

// handleFilesInTrash lists the contents of the user's trash directory (or
// a subdirectory within it), parsed from `ls -l`/`ls -la` output.
func handleFilesInTrash(a *appctx.App, w http.ResponseWriter, r *http.Request, pathParam string) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	trashBase := filepath.Clean(rawTrashBase(userContext))

	title := pathParam
	if title == "" {
		title = "Trash"
	}

	var directory string
	if pathParam != "" {
		candidate := filepath.Clean(filepath.Join(trashBase, pathParam))
		if !isWithin(candidate, trashBase) {
			flashAndRedirectToTrash(a, w, r, "error", "Invalid directory.")
			return
		}
		directory = candidate
	} else {
		directory = trashBase
	}

	info, statErr := os.Stat(directory)
	if statErr != nil || !info.IsDir() {
		if directory == trashBase {
			if mkdirErr := os.MkdirAll(trashBase, 0o755); mkdirErr != nil {
				renderTrashPage(a, w, r, title, pathParam, nil)
				return
			}
		} else {
			flashAndRedirectToTrash(a, w, r, "error", "Directory does not exist.")
			return
		}
	}

	flag := "-la"
	if r.URL.Query().Get("hidden_files") == "false" {
		flag = "-l"
	}

	out, cmdErr := exec.CommandContext(r.Context(), "ls", flag, directory).Output()
	if cmdErr != nil {
		renderTrashPage(a, w, r, title, pathParam, nil)
		return
	}

	trashInfoContent, _ := os.ReadFile(rawTrashInfoPath(userContext))
	filesInfo := parseLsOutputTrash(string(out), string(trashInfoContent))

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, filesInfo)
		return
	}

	renderTrashPage(a, w, r, title, pathParam, filesInfo)
}

func rawTrashBase(userContext string) string {
	return "/home/" + userContext + "/.local/share/Trash/"
}

func rawTrashInfoPath(userContext string) string {
	return "/home/" + userContext + "/.local/share/Trash/.trash_restore"
}

// isWithin reports whether candidate is base itself or a descendant of it.
func isWithin(candidate, base string) bool {
	rel, err := filepath.Rel(base, candidate)
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && rel != "..")
}
