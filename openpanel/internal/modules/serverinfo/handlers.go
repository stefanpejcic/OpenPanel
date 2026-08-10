package serverinfo

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
)

// handleServerInfo serves a static page shell, entirely filled in
// client-side via the /json/system/hosting/* fetches below.
func handleServerInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	renderServerInfoPage(a, w, r)
}

// handleContainerUsage serves the current resource-usage snapshot page.
func handleContainerUsage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)
	userContext, _ := data["context"].(string)

	stats, statsErr := a.GetResourceUsage(r.Context(), username, userContext)

	sess, _ := a.Sessions.Get(r, session.CookieName)
	if statsErr != nil {
		flash.Add(sess, "danger", "An error occurred while fetching resource usage.")
	} else if len(stats) == 0 {
		flash.Add(sess, "warning", "No resource usage data available yet. Please try again later")
	}
	_ = a.Sessions.Save(r, w, sess)

	renderStatsPage(a, w, r, stats)
}

// handleUsageHistory serves the paginated resource-usage history page.
func handleUsageHistory(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := data["context"].(string)
	usageFilePath := "/home/" + userContext + "/resource_usage.txt"
	chartsMode := a.Config.Get("resource_usage_charts_mode", "one")

	var usageData []ResourceUsageLine
	if content, readErr := os.ReadFile(usageFilePath); readErr == nil {
		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var entry ResourceUsageLine
			if unmarshalErr := json.Unmarshal([]byte(line), &entry); unmarshalErr == nil {
				usageData = append(usageData, entry)
			}
		}
	}

	totalLines := len(usageData)
	showAll := r.URL.Query().Get("show_all") == "true"

	itemsPerPage, _ := strconv.Atoi(a.Config.Get("resource_usage_items_per_page", "25"))
	if itemsPerPage <= 0 {
		itemsPerPage = 25
	}
	totalPages := 1
	if showAll {
		itemsPerPage = totalLines
		totalPages = 1
	} else if itemsPerPage > 0 {
		totalPages = (totalLines + itemsPerPage - 1) / itemsPerPage
	}
	if totalPages < 1 {
		totalPages = 1
	}

	currentPage, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if currentPage < 1 {
		currentPage = 1
	}
	startIdx := (currentPage - 1) * itemsPerPage
	endIdx := startIdx + itemsPerPage
	if startIdx > totalLines {
		startIdx = totalLines
	}
	if endIdx > totalLines {
		endIdx = totalLines
	}
	var paginated []ResourceUsageLine
	if startIdx < endIdx {
		paginated = usageData[startIdx:endIdx]
	}

	renderUsageHistoryPage(a, w, r, chartsMode, showAll, itemsPerPage, totalPages, totalLines, currentPage, paginated)
}

// handleSystemHostingInfo returns platform/uname info, uptime, load
// average, and the server's public IP.
func handleSystemHostingInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	uptime, loadAvg := getUptimeAndLoad()
	platform := getPlatformInfo()

	writeJSON(w, http.StatusOK, map[string]any{
		"system":    platform.System,
		"node":      platform.Node,
		"release":   platform.Release,
		"version":   platform.Version,
		"machine":   platform.Machine,
		"processor": platform.Processor,
		"ip":        a.GetCachedIPForUserOrPublicIPv4(r.Context(), username),
		"uptime":    uptime,
		"load_avg":  loadAvg,
	})
}

// handleSystemHostingPlan returns the user's hosting plan limits plus the
// webserver/mysql type and nameservers configured for their context.
func handleSystemHostingPlan(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := data["context"].(string)
	planID, _ := data["hosting_plan"].(int)

	plan := appctx.PlanDetails{
		DomainsLimit: "0", WebsitesLimit: "0", DBLimit: "0", CPU: "0", RAM: "0",
		EmailLimit: "0", FTPLimit: "0", DiskLimit: "0", InodesLimit: "0",
		Bandwidth: "0", MaxEmailQuota: "0",
	}
	if planID > 0 {
		if details, planErr := a.QueryPlanDetailsByID(r.Context(), planID); planErr == nil {
			plan = details
		}
	}

	var ns1, ns2, ns3, ns4 string
	if v := a.Config.Get("ns1", ""); v != "" {
		ns1 = v
		ns2 = a.Config.Get("ns2", "")
		ns3 = a.Config.Get("ns3", "")
		ns4 = a.Config.Get("ns4", "")
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"context":              userContext,
		"plan_webserver":       webserver.GetEnvFileValue(userContext, "WEB_SERVER"),
		"plan_mysql":           webserver.GetEnvFileValue(userContext, "MYSQL_TYPE"),
		"plan_domains_limit":   plan.DomainsLimit,
		"plan_db_limit":        plan.DBLimit,
		"plan_cpu_limit":       plan.CPU,
		"plan_ram_limit":       plan.RAM,
		"plan_websites_limit":  plan.WebsitesLimit,
		"plan_ftp_limit":       plan.FTPLimit,
		"plan_email_limit":     plan.EmailLimit,
		"plan_max_email_quota": plan.MaxEmailQuota,
		"plan_description":     plan.Description,
		"plan_disk_limit":      plan.DiskLimit,
		"plan_inodes_limit":    plan.InodesLimit,
		"plan_bandwidth":       plan.Bandwidth,
		"ns1":                  ns1, "ns2": ns2, "ns3": ns3, "ns4": ns4,
	})
}

// getEnvPort reads /home/<context>/.env and, for host:port style values,
// returns just the port segment.
func getEnvPort(context, key string) string {
	content, err := os.ReadFile("/home/" + context + "/.env")
	if err != nil {
		return ""
	}
	prefix := key + "="
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimRight(line, "\r")
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		raw := strings.Trim(strings.TrimSpace(line[len(prefix):]), `'"`)
		parts := strings.Split(raw, ":")
		switch len(parts) {
		case 3:
			return parts[1]
		case 2:
			return parts[0]
		default:
			return raw
		}
	}
	return ""
}

// handleSystemHostingPorts returns the host-exposed ports for the user's
// MySQL and Postgres containers.
func handleSystemHostingPorts(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	writeJSON(w, http.StatusOK, map[string]any{
		"remote_mysql_port":    getEnvPort(username, "MYSQL_PORT"),
		"remote_postgres_port": getEnvPort(username, "POSTGRES_PORT"),
	})
}
