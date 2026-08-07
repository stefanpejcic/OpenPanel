// Package inodes is a per-top-level-directory inode count browser
// (`find . -printf '%h\n'`, tallied by first path segment).
package inodes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

// Register wires the inodes-explorer route onto mux, gated behind the
// "inodes" feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "inodes")(h)
	}
	// "{directory...}" already covers the bare "/inodes-explorer/" case
	// too (directory resolves to "" for that exact path).
	mux.Handle("GET /inodes-explorer/{directory...}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleInodesExplorer(a, w, r)
	}))
}

// resolveUnderHome is the same inline path-traversal guard used by
// internal/modules/diskusage (see that package's comment) - not shared
// cross-package since it's a handful of lines duplicated across two
// near-identical, independently-gated modules.
func resolveUnderHome(userContext, directory string) (string, error) {
	volume := "/home/" + userContext + "/"
	target := filepath.Join(volume, strings.TrimPrefix(directory, "/")) + "/"
	cleanVolume := filepath.Clean(volume)
	cleanTarget := filepath.Clean(target)
	if cleanTarget != cleanVolume && !strings.HasPrefix(cleanTarget, cleanVolume+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid directory")
	}
	return cleanTarget, nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func handleInodesExplorer(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	output := inodesOutput(ctx, a, userContext, actualFolder)

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, output)
		return
	}

	renderInodesPage(a, w, r, output)
}

type countedFolder struct {
	Folder string
	Count  int
}

// inodesOutput runs `find . -printf '%h\n'` under actualFolder and tallies
// the results per top-level folder, cached 10s.
//
// Requires GNU find (the `-printf` action is a GNU findutils extension,
// not supported by BusyBox find); the runtime image installs the
// `findutils` package for this.
func inodesOutput(ctx context.Context, a *appctx.App, userContext, actualFolder string) string {
	out, _ := cache.Memoize(ctx, a.Cache, "inodes_explorer:"+userContext+":"+actualFolder, 10*time.Second, func() (string, error) {
		cmd := exec.CommandContext(ctx, "find", ".", "-printf", "%h\n")
		cmd.Dir = actualFolder
		raw, err := cmd.Output()
		if err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				return "", nil //nolint:nilerr // command failure -> empty output, not an error response
			}
			// find exits non-zero on permission-denied for individual
			// subdirectories but still writes the rest of its output to
			// stdout, so keep using it instead of discarding everything.
		}

		counts := map[string]int{}
		var order []string
		for _, line := range strings.Split(string(raw), "\n") {
			if line == "" {
				continue
			}
			line = strings.TrimPrefix(line, "./")
			topLevel, _, _ := strings.Cut(line, "/")
			if _, ok := counts[topLevel]; !ok {
				order = append(order, topLevel)
			}
			counts[topLevel]++
		}

		folders := make([]countedFolder, 0, len(order))
		for _, f := range order {
			folders = append(folders, countedFolder{Folder: f, Count: counts[f]})
		}
		sort.SliceStable(folders, func(i, j int) bool { return folders[i].Count > folders[j].Count })

		lines := make([]string, len(folders))
		for i, f := range folders {
			lines[i] = fmt.Sprintf("%d %s", f.Count, f.Folder)
		}
		return strings.Join(lines, "\n"), nil
	})
	return out
}
