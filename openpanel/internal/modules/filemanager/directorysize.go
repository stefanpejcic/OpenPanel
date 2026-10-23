package filemanager

import (
	"net/http"
	"os/exec"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// HandleDirectorySize serves `/json/directory-size`, used by both the file manager table's "Calculate" links and various app single-page views - feature "helpers" is unconditionally granted to every user (see baselineFeatures), so like /docker/tags and /json/check_if_file_exists this is registered unconditionally rather than gated behind "filemanager"
func HandleDirectorySize(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	user, err := currentUser(r.Context(), a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	root, rel, perr := openUserPath(user, sizeFolder(r), true)
	if perr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Invalid folder path."})
		return
	}
	defer root.Close()

	out, runErr := runOnDir(r.Context(), root, rel, "du", "-sh")
	if runErr != nil {
		errMsg := runErr.Error()
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			errMsg = strings.TrimSpace(string(exitErr.Stderr))
		}
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": errMsg})
		return
	}

	size := strings.SplitN(string(out), "\t", 2)[0]
	web.WriteJSON(w, http.StatusOK, map[string]string{"size": size})
}

// sizeFolder is the folder query param as a path under the user's files, Query() already decoded it so it isn't decoded again
func sizeFolder(r *http.Request) string {
	folder := strings.TrimPrefix(r.URL.Query().Get("folder"), "/")
	return strings.TrimPrefix(folder, "var/www/html/")
}

// RegisterDirectorySize wires up the always-on directory-size route.
func RegisterDirectorySize(mux *http.ServeMux, a *appctx.App) {
	requireLogin := auth.RequireLogin(a, "helpers")
	mux.Handle("GET /json/directory-size", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleDirectorySize(a, w, r)
	})))
}
