package fixpermissions

import (
	"context"
	"encoding/json"
	"mime"
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

// RegisterAPI wires the /api/fix-permissions route onto mux, gated behind the "fix_permissions" feature flag
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "fix_permissions", "POST /api/fix-permissions", func(w http.ResponseWriter, r *http.Request) { apiFixPermissionsRun(a, w, r) })
}

func writeAPIFixPermsJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiFixPermissionsRun is handleFixPermissions's POST path (fixpermissions.go) with a JSON body/response instead of a form post + flash
func apiFixPermissionsRun(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)

	baseDirectory := "/var/www/html"
	raw := strings.TrimSpace(apiDirectory(r))
	args := []string{"files-fix_permissions", username}

	if raw != "" && raw != "/" {
		fixDirectory := filepath.Clean(raw)
		if !isRelativeTo(fixDirectory, baseDirectory) {
			writeAPIFixPermsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid directory"})
			return
		}
		relativePath := strings.TrimPrefix(strings.TrimPrefix(fixDirectory, baseDirectory), "/")
		if relativePath != "" && relativePath != "." {
			// script does its own /var/www/html -> volume translation, so pass the relative path, not a resolved one
			args = append(args, relativePath)
		}
	}

	// a client that times out mustn't kill the fix halfway
	if runErr := exec.CommandContext(context.WithoutCancel(ctx), "opencli", args...).Run(); runErr != nil {
		writeAPIFixPermsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fix permissions."})
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "fixed permissions via API", reqip.ClientIP(r))
	writeAPIFixPermsJSON(w, http.StatusOK, map[string]string{"message": "Permissions fixed."})
}

// apiDirectory reads directory from a form post or a JSON body, trying JSON first drained form bodies so the whole account got fixed
func apiDirectory(r *http.Request) string {
	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct == "application/x-www-form-urlencoded" || ct == "multipart/form-data" {
		return r.FormValue("directory")
	}
	var body struct {
		Directory string `json:"directory"`
	}
	if json.NewDecoder(r.Body).Decode(&body) == nil {
		return body.Directory
	}
	// no JSON body, a ?directory= query still works
	return r.FormValue("directory")
}
