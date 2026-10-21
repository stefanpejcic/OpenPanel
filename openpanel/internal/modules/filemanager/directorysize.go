package filemanager

import (
	"net/http"
	"net/url"
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

	folder := r.URL.Query().Get("folder")
	if decoded, decErr := url.QueryUnescape(folder); decErr == nil {
		folder = decoded
	}
	folder = strings.TrimPrefix(folder, "/")
	folder = strings.TrimPrefix(folder, "var/www/html/")

	root, rel, perr := openUserPath(user, folder, true)
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

// RegisterDirectorySize wires up the always-on directory-size route.
func RegisterDirectorySize(mux *http.ServeMux, a *appctx.App) {
	requireLogin := auth.RequireLogin(a, "helpers")
	mux.Handle("GET /json/directory-size", requireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HandleDirectorySize(a, w, r)
	})))
}
