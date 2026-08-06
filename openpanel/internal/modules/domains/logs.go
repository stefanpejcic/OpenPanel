package domains

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// AccessLogRequest mirrors one Caddy JSON access-log line's nested
// "request" object.
type AccessLogRequest struct {
	ClientIP string `json:"client_ip"`
	Method   string `json:"method"`
	URI      string `json:"uri"`
}

// AccessLogEntry mirrors one Caddy JSON access-log line.
type AccessLogEntry struct {
	TS       float64          `json:"ts"`
	Level    string           `json:"level"`
	Logger   string           `json:"logger"`
	Msg      string           `json:"msg"`
	Request  AccessLogRequest `json:"request"`
	Status   int              `json:"status"`
	Size     int              `json:"size"`
	Duration float64          `json:"duration"`
}

// handleViewDomainAccessLog shows recent access-log entries for a domain.
func handleViewDomainAccessLog(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	domainName := r.PathValue("domain_name")
	if domainName == "" {
		domainsList, _ := a.AllDomainsForUser(ctx, userID)
		renderDomainLogsSelectPage(a, w, r, domainsList)
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	logFilePath := "/var/log/caddy/domlogs/" + domainName + "/access.log"
	info, statErr := os.Stat(logFilePath)
	if statErr != nil {
		flashAndRedirect(a, w, r, "error", "Log file not found for domain "+domainName+".", "/domains/log")
		return
	}
	if info.Size() == 0 {
		flashAndRedirect(a, w, r, "info", "Log file for domain "+domainName+" is empty.", "/domains/log")
		return
	}

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		flashAndRedirect(a, w, r, "danger", "Error reading log file: "+err.Error(), "/domains/log")
		return
	}

	var entries []AccessLogEntry
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry AccessLogEntry
		if json.Unmarshal([]byte(line), &entry) == nil {
			entries = append(entries, entry)
		}
	}
	// reverse so the most recent entries come first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	totalLogs := len(entries)
	showAll := r.URL.Query().Get("show_all") == "true"

	var itemsPerPage, totalPages int
	if showAll {
		itemsPerPage = totalLogs
		totalPages = 1
	} else {
		itemsPerPage = atoiDefault(a.Config.Get("domain_log_per_page", "1000"), 1000)
		if itemsPerPage < 1 {
			itemsPerPage = 1000
		}
		totalPages = totalLogs / itemsPerPage
		if totalLogs%itemsPerPage != 0 {
			totalPages++
		}
	}
	totalAllowedForShowAll := atoiDefault(a.Config.Get("domain_log_max_for_show_all", "10000"), 10000)

	page := atoiDefault(r.URL.Query().Get("page"), 1)
	if page < 1 {
		page = 1
	}
	startIdx := (page - 1) * itemsPerPage
	endIdx := startIdx + itemsPerPage
	if startIdx > totalLogs {
		startIdx = totalLogs
	}
	if endIdx > totalLogs {
		endIdx = totalLogs
	}
	var paginated []AccessLogEntry
	if itemsPerPage > 0 {
		paginated = entries[startIdx:endIdx]
	}

	renderDomainLogsPage(a, w, r, domainName, paginated, showAll, page, itemsPerPage, totalPages, totalLogs, totalAllowedForShowAll)
}
