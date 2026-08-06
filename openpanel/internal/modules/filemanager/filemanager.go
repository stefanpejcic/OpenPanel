// Package filemanager implements the file browser (list/upload/create/
// delete/rename/move/copy/compress/extract/permissions/edit), built on top
// of internal/core/paths' SecureUserPath guard for every user-supplied
// path.
package filemanager

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/paths"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// Register wires the file manager's routes onto mux, gated behind the
// "filemanager" feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "filemanager")(h)
	}

	mux.Handle("GET /files", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFiles(a, w, r, "") }))
	// "GET /files/{path_param...}" already covers the bare "/files/" case
	// too (path_param resolves to "" for that exact path), so no separate
	// registration is needed for it - net/http.ServeMux would refuse to
	// register both anyway (they'd be flagged as an ambiguous conflict).
	mux.Handle("GET /files/{path_param...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleFiles(a, w, r, r.PathValue("path_param"))
	}))

	mux.Handle("POST /file-manager/new-file", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCreateFile(a, w, r) }))
	mux.Handle("POST /file-manager/new-directory", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCreateFolder(a, w, r) }))
	mux.Handle("/file-manager/upload", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleUploadFiles(a, w, r) }))
	mux.Handle("POST /file-manager/extract-archive", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleExtractFiles(a, w, r) }))
	mux.Handle("POST /file-manager/create-archive", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCompressFiles(a, w, r) }))
	mux.Handle("POST /file-manager/rename", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleRenameFile(a, w, r) }))
	mux.Handle("DELETE /file-manager/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleDeleteFile(a, w, r) }))
	mux.Handle("POST /file-manager/permissions", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleChangePermissions(a, w, r) }))
	mux.Handle("POST /file-manager/copy", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleCopyItem(a, w, r) }))
	mux.Handle("POST /file-manager/move", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMoveItem(a, w, r) }))
	mux.Handle("/file-manager/edit-file/{file_path...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleEditFile(a, w, r, r.PathValue("file_path"))
	}))
	mux.Handle("GET /file-manager/download-file/{filename...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleDownloadFile(a, w, r, r.PathValue("filename"))
	}))
	mux.Handle("GET /file-manager/view-file/{filename...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleViewFile(a, w, r, r.PathValue("filename"))
	}))
	mux.Handle("POST /file-manager/wget", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleWgetFiles(a, w, r) }))
	mux.Handle("GET /file-manager/wget/status/{download_id}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleWgetStatus(a, w, r)
	}))

	mux.Handle("GET /json/folders/{path_param...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleFolders(a, w, r, r.PathValue("path_param"))
	}))
}

// stripQuotes trims a leading/trailing matching pair of single or double
// quotes, used for config values that may be quoted in openpanel.config.
func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
			(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// injectedUser is the small subset of InjectData() every filemanager route
// needs.
type injectedUser struct {
	Username string
	Context  string
}

func currentUser(ctx context.Context, a *appctx.App, r *http.Request) (injectedUser, error) {
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		return injectedUser{}, err
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)
	return injectedUser{Username: username, Context: userContext}, nil
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

func filesRedirectPath(pathParam string) string {
	if pathParam == "" {
		return "/files"
	}
	return "/files/" + pathParam
}

// pathErrorStatus extracts the HTTP status from a *paths.Error, defaulting
// to 500 for anything else.
func pathErrorStatus(err error) (int, string) {
	if perr, ok := err.(*paths.Error); ok {
		return perr.Code, perr.Message
	}
	return http.StatusInternalServerError, err.Error()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// chownToUser sets ownership of path to the given user's uid. It's
// best-effort: the error is returned but not acted on here, leaving each
// caller to decide whether to flash a warning or ignore it.
func chownToUser(ctx context.Context, a *appctx.App, path, userContext string) error {
	uid, err := a.GetUID(ctx, userContext)
	if err != nil || uid <= 0 {
		return err
	}
	return os.Chown(path, uid, uid)
}

// chownRecursive sets ownership of root and everything beneath it,
// used after extract/copy/move so the destination ends up owned by the
// target user rather than the panel process.
func chownRecursive(ctx context.Context, a *appctx.App, root, userContext string) {
	uid, err := a.GetUID(ctx, userContext)
	if err != nil || uid <= 0 {
		return
	}
	_ = os.Chown(root, uid, uid)
	_ = filepath.Walk(root, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil //nolint:nilerr // best-effort: a walk error here shouldn't abort chowning the rest of the tree
		}
		_ = os.Chown(p, uid, uid)
		return nil
	})
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return n
}

// runCommand runs a command and returns its error, for the simple
// fire-and-check call sites in this package that don't need output.
func runCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Run()
}
