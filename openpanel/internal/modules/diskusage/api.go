package diskusage

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterAPI wires the disk-usage API route onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "disk_usage", "GET /api/disk-usage/{directory...}", func(w http.ResponseWriter, r *http.Request) {
		apiHandleDiskUsage(a, w, r)
	})
}

type diskUsageEntry struct {
	Size string `json:"size"`
	Path string `json:"path"`
}

// apiHandleDiskUsage returns per-subdirectory disk usage for the requested
// directory as JSON.
func apiHandleDiskUsage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		writeJSON(w, map[string]string{"error": "Path traversal detected"})
		return
	}

	output := diskUsageOutput(ctx, a, userContext, actualFolder)

	var entries []diskUsageEntry
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 {
			entries = append(entries, diskUsageEntry{Size: parts[0], Path: parts[1]})
		}
	}
	if entries == nil {
		entries = []diskUsageEntry{}
	}

	writeJSON(w, map[string]any{"directory": directory, "entries": entries})
}
