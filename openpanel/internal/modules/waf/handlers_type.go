package waf

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

// handleWAFType returns a Redis-cached (300s) list of valid rule IDs or
// tags, fed to the autocomplete on the WAF domain page.
func handleWAFType(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	typ := r.PathValue("type")
	if typ != "tags" && typ != "ids" {
		writeJSONError(w, http.StatusBadRequest, "Invalid type. Must be 'tags' or 'ids'.")
		return
	}

	data, _ := cache.Memoize(r.Context(), a.Cache, "waf_"+typ+"_cache", 300*time.Second, func() ([]string, error) {
		return extractIDs(r.Context(), typ), nil
	})
	if data == nil {
		data = []string{}
	}
	writeJSON(w, http.StatusOK, data)
}

// extractIDs runs `opencli waf <type>` and pulls the ID/tag off the end of
// each "label: value" output line. Any failure (including the binary not
// existing) yields an empty list rather than an error.
func extractIDs(ctx context.Context, typ string) []string {
	out, err := exec.CommandContext(ctx, "opencli", "waf", typ).Output()
	if err != nil {
		return []string{}
	}
	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.Split(line, ":")
		ids = append(ids, parts[len(parts)-1])
	}
	return ids
}
