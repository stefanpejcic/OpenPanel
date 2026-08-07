// Package diskusage is a `find -maxdepth 1 -type d -exec du -sh`
// per-directory disk usage browser.
package diskusage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

const cacheTTL = 5 * time.Second

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// Register wires the disk-usage route onto mux, gated behind the
// "disk_usage" feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "disk_usage")(h)
	}
	// "{directory...}" already covers the bare "/disk-usage/" case too
	// (directory resolves to "" for that exact path).
	mux.Handle("GET /disk-usage/{directory...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleDiskUsage(a, w, r)
	}))
}

// resolveUnderHome resolves directory (a user-supplied relative or
// absolute-looking path) under /home/<context>/, rejecting any escape.
// Unlike filemanager's secure_user_path, the volume root here is the raw
// account home directory, not the docroot volume - a deliberately simpler
// check, since this browser only ever needs to stay within the home dir.
func resolveUnderHome(userContext, directory string) (string, error) {
	volume := "/home/" + userContext + "/"
	target := filepath.Join(volume, strings.TrimPrefix(directory, "/"))
	target = target + "/" // keep the prefix comparison below insensitive to a missing trailing slash
	cleanVolume := filepath.Clean(volume)
	cleanTarget := filepath.Clean(target)
	if cleanTarget != cleanVolume && !strings.HasPrefix(cleanTarget, cleanVolume+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid directory")
	}
	return cleanTarget, nil
}

func handleDiskUsage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)

	directory := r.PathValue("directory")
	if directory == "" {
		directory = "/"
	}

	actualFolder, resolveErr := resolveUnderHome(userContext, directory)
	if resolveErr != nil {
		http.Error(w, "Invalid directory", http.StatusForbidden)
		return
	}

	output := diskUsageOutput(ctx, a, userContext, actualFolder)

	if r.URL.Query().Get("json") != "" {
		writeJSON(w, map[string]any{"total_du_output": output})
		return
	}

	renderDiskUsagePage(a, w, r, output)
}

// diskUsageOutput runs `find . -maxdepth 1 -type d ! -name . -exec du -sh
// {} \;` under actualFolder, cached 5s.
func diskUsageOutput(ctx context.Context, a *appctx.App, userContext, actualFolder string) string {
	out, _ := cache.Memoize(ctx, a.Cache, "disk_usage:"+userContext+":"+actualFolder, cacheTTL, func() (string, error) {
		cmd := exec.CommandContext(ctx, "find", ".", "-maxdepth", "1", "-type", "d", "!", "-name", ".", "-exec", "du", "-sh", "{}", ";")
		cmd.Dir = actualFolder
		out, err := cmd.Output()
		if err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				return "", nil //nolint:nilerr // command failure -> empty output, not an error response
			}
			// find/du exit non-zero on permission-denied for individual
			// subdirectories but still write the rest of their output to
			// stdout, so keep using it instead of discarding everything.
		}
		return strings.ReplaceAll(string(out), "./", ""), nil
	})
	return out
}
