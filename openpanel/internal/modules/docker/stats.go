package docker

import (
	"encoding/json"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// containerStatEntry is the shape containers.html's JS and base.html's
// per-service status widget both expect from /json/services (see
// RegisterServicesJSON in docker.go for how that route gets wired up).
type containerStatEntry struct {
	ID       string `json:"ID"`
	Name     string `json:"Name"`
	CPUPerc  string `json:"CPUPerc"`
	MemUsage string `json:"MemUsage"`
	MemPerc  string `json:"MemPerc"`
	NetIO    string `json:"NetIO"`
	BlockIO  string `json:"BlockIO"`
	PIDs     string `json:"PIDs"`
}

// podmanStatsRow is one element of `podman stats --format json`'s output -
// podman's own CLI already computes and formats every field this route
// needs, so there's no need to hit the lower-level /libpod/containers/stats
// endpoint and format the numbers by hand. Using the CLI here keeps this
// consistent with the rest of the package (podmanmanager.Command), rather
// than adding a REST client dependency just for this one route.
type podmanStatsRow struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CPUPercent string `json:"cpu_percent"`
	MemUsage   string `json:"mem_usage"`
	MemPercent string `json:"mem_percent"`
	NetIO      string `json:"net_io"`
	BlockIO    string `json:"block_io"`
	PIDs       string `json:"pids"`
}

// handleServicesStats serves GET /json/services[?name=], returning live
// per-container CPU/memory/network/PID stats.
func handleServicesStats(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)
	containerName := r.URL.Query().Get("name")

	argv := podmanmanager.PodmanArgv(userContext, "stats", "--no-stream", "--format", "json")
	if containerName != "" {
		argv = append(argv, containerName)
	}
	out, err := podmanmanager.Command(ctx, userContext, argv).Output()
	if err != nil {
		if containerName != "" {
			writeJSON(w, map[string]any{"container_stat": nil})
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Could not retrieve container stats")
		return
	}

	var rows []podmanStatsRow
	if json.Unmarshal(out, &rows) != nil {
		writeJSONError(w, http.StatusInternalServerError, "Could not retrieve container stats")
		return
	}

	entries := make([]containerStatEntry, len(rows))
	for i, row := range rows {
		entries[i] = containerStatEntry{
			ID: row.ID, Name: row.Name, CPUPerc: row.CPUPercent, MemUsage: row.MemUsage,
			MemPerc: row.MemPercent, NetIO: row.NetIO, BlockIO: row.BlockIO, PIDs: row.PIDs,
		}
	}

	if containerName != "" {
		if len(entries) == 0 {
			writeJSON(w, map[string]any{"container_stat": nil})
			return
		}
		writeJSON(w, map[string]any{"container_stat": entries[0]})
		return
	}
	writeJSON(w, map[string]any{"container_stats": entries})
}
