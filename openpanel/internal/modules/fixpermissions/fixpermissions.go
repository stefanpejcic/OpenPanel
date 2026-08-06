// Package fixpermissions runs `opencli files-fix_permissions` over a
// user-chosen directory under /var/www/html/.
package fixpermissions

import (
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// Register wires the fix-permissions route onto mux, gated behind the
// "fix_permissions" feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "fix_permissions")(h)
	}
	mux.Handle("/fix-permissions", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleFixPermissions(a, w, r)
	}))
}

func handleFixPermissions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		baseDirectory := "/var/www/html"
		raw := strings.TrimSpace(r.Form.Get("directory"))

		args := []string{"files-fix_permissions", username}

		if raw != "" && raw != "/" {
			fixDirectory := filepath.Clean(raw)
			if !isRelativeTo(fixDirectory, baseDirectory) {
				http.Error(w, "Invalid fix_directory", http.StatusBadRequest)
				return
			}
			relativePath := strings.TrimPrefix(strings.TrimPrefix(fixDirectory, baseDirectory), "/")
			if relativePath != "" && relativePath != "." {
				args = append(args, filepath.Join(volume, relativePath))
			}
		}

		_ = exec.CommandContext(ctx, "opencli", args...).Run()

		sess, _ := a.Sessions.Get(r, session.CookieName)
		flash.Add(sess, "success", "Success.")
		_ = a.Sessions.Save(r, w, sess)
	}

	out, _ := exec.CommandContext(ctx, "find", volume, "-type", "d").Output()
	var directories []string
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		directories = append(directories, strings.Replace(line, volume, "/var/www/html/", 1))
	}

	renderFixPermissionsPage(a, w, r, directories)
}

// isRelativeTo reports whether path equals base or is nested under it.
func isRelativeTo(path, base string) bool {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return false
	}
	return rel == "." || !strings.HasPrefix(rel, "..")
}
