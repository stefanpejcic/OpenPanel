package serverinfo

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterUsageAPI wires the resource-usage JSON API routes onto mux.
func RegisterUsageAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "usage", "GET /api/usage", func(w http.ResponseWriter, r *http.Request) { apiUsageCurrent(a, w, r) })
	apiregistry.Handle(mux, a, "usage", "GET /api/usage/history", func(w http.ResponseWriter, r *http.Request) { apiUsageHistory(a, w, r) })
}

func writeAPIUsageJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiUsageCurrent returns the most recent resource-usage snapshot.
func apiUsageCurrent(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)
	userContext, _ := data["context"].(string)

	stats, statsErr := a.GetResourceUsage(r.Context(), username, userContext)
	if statsErr != nil || len(stats) == 0 {
		writeAPIUsageJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "No usage data available yet"})
		return
	}
	writeAPIUsageJSON(w, http.StatusOK, stats)
}

type apiUsageEntry struct {
	Timestamp         string  `json:"timestamp"`
	CPUPercent        float64 `json:"cpu_percent"`
	MemPercent        float64 `json:"mem_percent"`
	MemUsed           string  `json:"mem_used"`
	MemTotal          string  `json:"mem_total"`
	BandwidthSent     string  `json:"bandwidth_sent"`
	BandwidthLimit    string  `json:"bandwidth_limit"`
	BandwidthUsagePct float64 `json:"bandwidth_usage_pct"`
	Warning           string  `json:"warning,omitempty"`
}

// apiUsageHistory returns paginated resource-usage history entries.
func apiUsageHistory(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := data["context"].(string)
	usageFile := "/home/" + userContext + "/resource_usage.txt"

	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage <= 0 {
		perPage, _ = strconv.Atoi(a.Config.Get("resource_usage_items_per_page", "25"))
		if perPage <= 0 {
			perPage = 25
		}
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	showAll := strings.ToLower(r.URL.Query().Get("show_all")) == "true"

	content, readErr := os.ReadFile(usageFile)
	if readErr != nil {
		writeAPIUsageJSON(w, http.StatusOK, map[string]any{"entries": []apiUsageEntry{}, "total": 0, "page": page, "per_page": perPage, "total_pages": 0})
		return
	}

	var entries []apiUsageEntry
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var d ResourceUsageLine
		if json.Unmarshal([]byte(line), &d) != nil {
			continue
		}
		entries = append(entries, apiUsageEntry{
			Timestamp: d.Timestamp, CPUPercent: d.CPU.Usage.Pct, MemPercent: d.Memory.UsagePct,
			MemUsed: d.Memory.Used.Human, MemTotal: d.Memory.Total.Human,
			BandwidthSent: d.Bandwidth.TotalSent.Human, BandwidthLimit: d.Bandwidth.Limit.Human,
			BandwidthUsagePct: d.Bandwidth.UsagePct, Warning: d.Warning,
		})
	}

	total := len(entries)
	var paginated []apiUsageEntry
	totalPages := 0
	if showAll {
		paginated = entries
		perPage = total
		totalPages = 1
	} else {
		if perPage > 0 {
			totalPages = (total + perPage - 1) / perPage
		}
		start := (page - 1) * perPage
		if start < 0 {
			start = 0
		}
		if start > total {
			start = total
		}
		end := start + perPage
		if end > total {
			end = total
		}
		paginated = entries[start:end]
	}
	if paginated == nil {
		paginated = []apiUsageEntry{}
	}

	writeAPIUsageJSON(w, http.StatusOK, map[string]any{
		"entries": paginated, "total": total, "page": page, "per_page": perPage, "total_pages": totalPages,
	})
}
