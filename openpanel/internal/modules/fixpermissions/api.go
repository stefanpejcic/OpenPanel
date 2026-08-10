package fixpermissions

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterAPI wires the /api/fix-permissions routes onto mux, gated
// behind the "fix_permissions" feature flag.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "fix_permissions", "GET /api/fix-permissions", func(w http.ResponseWriter, r *http.Request) { apiFixPermissionsList(a, w, r) })
	apiregistry.Handle(mux, a, "fix_permissions", "POST /api/fix-permissions", func(w http.ResponseWriter, r *http.Request) { apiFixPermissionsRun(a, w, r) })
}

func writeAPIFixPermsJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiFixPermissionsList lists every directory under /var/www/html/ (the
// same `find`-backed listing the web page's directory picker uses).
func apiFixPermissionsList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)
	volume := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"

	out, _ := exec.CommandContext(ctx, "find", volume, "-type", "d").Output()
	directories := []string{}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		directories = append(directories, strings.Replace(line, volume, "/var/www/html/", 1))
	}

	writeAPIFixPermsJSON(w, http.StatusOK, map[string]any{"directories": directories})
}

// apiFixPermissionsRun is handleFixPermissions's POST path (fixpermissions.go)
// with a JSON body/response instead of a form post + flash.
func apiFixPermissionsRun(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		Directory string `json:"directory"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Directory = r.Form.Get("directory")
	}

	baseDirectory := "/var/www/html"
	raw := strings.TrimSpace(body.Directory)
	args := []string{"files-fix_permissions", username}

	if raw != "" && raw != "/" {
		fixDirectory := filepath.Clean(raw)
		if !isRelativeTo(fixDirectory, baseDirectory) {
			writeAPIFixPermsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid directory"})
			return
		}
		relativePath := strings.TrimPrefix(strings.TrimPrefix(fixDirectory, baseDirectory), "/")
		if relativePath != "" && relativePath != "." {
			args = append(args, filepath.Join(volume, relativePath))
		}
	}

	if runErr := exec.CommandContext(ctx, "opencli", args...).Run(); runErr != nil {
		writeAPIFixPermsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fix permissions."})
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "fixed permissions via API", reqip.ClientIP(r))
	writeAPIFixPermsJSON(w, http.StatusOK, map[string]string{"message": "Permissions fixed."})
}
