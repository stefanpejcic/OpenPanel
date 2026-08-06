package waf

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

func flashAndRedirectLog(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, "/server/waf/log", http.StatusFound)
}

func configIntOrDefault(a *appctx.App, key string, def int) int {
	v, err := strconv.Atoi(a.Config.Get(key, strconv.Itoa(def)))
	if err != nil || v <= 0 {
		return def
	}
	return v
}

// handleWAFLog mirrors waf.py's view_coraza_waf_log(): with no domain,
// a domain picker; with one, its paginated Coraza JSON log.
func handleWAFLog(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domainName := r.PathValue("domain_name")

	if domainName == "" {
		domains, _ := a.AllDomainsForUser(r.Context(), userID)
		renderWAFLogSelectPage(a, w, r, domains)
		return
	}

	if !a.CheckDomainBelongsToUser(r.Context(), userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	logFilePath := wafLogPath(domainName)
	info, statErr := os.Stat(logFilePath)
	if statErr != nil {
		flashAndRedirectLog(a, w, r, "error", "Log file not found for domain "+domainName+".")
		return
	}
	if info.Size() == 0 {
		flashAndRedirectLog(a, w, r, "info", "Log file for domain "+domainName+" is empty.")
		return
	}

	content, readErr := os.ReadFile(logFilePath)
	if readErr != nil {
		flashAndRedirectLog(a, w, r, "danger", "Error reading log file: "+readErr.Error())
		return
	}

	var jsonLogs []json.RawMessage
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		jsonLogs = append(jsonLogs, json.RawMessage(line))
	}
	for i, j := 0, len(jsonLogs)-1; i < j; i, j = i+1, j-1 {
		jsonLogs[i], jsonLogs[j] = jsonLogs[j], jsonLogs[i]
	}

	totalLogs := len(jsonLogs)
	showAll := r.URL.Query().Get("show_all") == "true"

	itemsPerPage := configIntOrDefault(a, "domain_log_per_page", 1000)
	totalPages := 1
	if showAll {
		itemsPerPage = totalLogs
		if itemsPerPage <= 0 {
			itemsPerPage = 1
		}
		totalPages = 1
	} else {
		totalPages = totalLogs / itemsPerPage
		if totalLogs%itemsPerPage != 0 {
			totalPages++
		}
		if totalPages < 1 {
			totalPages = 1
		}
	}

	totalAllowedForShowAll := configIntOrDefault(a, "domain_log_max_for_show_all", 10000)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
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
	var paginated []json.RawMessage
	if startIdx < endIdx {
		paginated = jsonLogs[startIdx:endIdx]
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"domain_name": domainName, "json_logs": paginated, "current_page": page,
			"items_per_page": itemsPerPage, "total_pages": totalPages, "total_lines": totalLogs,
			"total_allowed_lines_for_show_all": totalAllowedForShowAll,
		})
		return
	}

	renderWAFLogPage(a, w, r, domainName, paginated, showAll, page, itemsPerPage, totalPages, totalLogs, totalAllowedForShowAll)
}
