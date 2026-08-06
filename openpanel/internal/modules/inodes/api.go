package inodes

import (
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterAPI wires the inodes API route onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "inodes", "GET /api/inodes/{directory...}", func(w http.ResponseWriter, r *http.Request) {
		apiHandleInodes(a, w, r)
	})
}

type inodesEntry struct {
	Folder     string `json:"folder"`
	InodeCount int    `json:"inode_count"`
}

// apiHandleInodes returns the inode count for each subdirectory of the
// requested directory, one level deep, as JSON.
func apiHandleInodes(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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

	output := inodesOutput(ctx, a, userContext, actualFolder)

	var entries []inodesEntry
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		count, _ := strconv.Atoi(fields[0])
		folder := strings.TrimLeft(strings.TrimPrefix(line, fields[0]), " \t")
		entries = append(entries, inodesEntry{Folder: folder, InodeCount: count})
	}
	if entries == nil {
		entries = []inodesEntry{}
	}

	writeJSON(w, map[string]any{"directory": directory, "entries": entries})
}
