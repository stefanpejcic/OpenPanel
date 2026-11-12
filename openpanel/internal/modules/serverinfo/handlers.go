package serverinfo

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dashboard"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleServerInfo serves a static page shell, entirely filled in client-side via the /json/system/hosting/* fetches below
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

	renderStatsPage(a, w, r, stats, usageHistory(readUsageLines(userContext)))
}

// readUsageLines parses every JSON line of /home/<context>/resource_usage.txt, skipping bad ones
func readUsageLines(userContext string) []ResourceUsageLine {
	content, err := tenantfs.ReadFile("/home/" + userContext + "/resource_usage.txt")
	if err != nil {
		return nil
	}
	var lines []ResourceUsageLine
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry ResourceUsageLine
		if json.Unmarshal([]byte(line), &entry) == nil {
			lines = append(lines, entry)
		}
	}
	return lines
}

// handleUsageHistory serves the hourly usage history, the page picks the range and filters client-side
func handleUsageHistory(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := data["context"].(string)
	samples, limits := historySamples(readUsageLines(userContext))

	// collect_stats keeps 100 lines when the setting is missing
	retention, convErr := strconv.Atoi(a.Config.Get("resource_usage_retention", "100"))
	if convErr != nil || retention < 0 {
		retention = 100
	}
	rowLimit, _ := strconv.Atoi(a.Config.Get("resource_usage_items_per_page", "25"))
	if rowLimit <= 0 {
		rowLimit = 25
	}
	renderUsageHistoryPage(a, w, r, UsageHistoryPageData{
		Samples: samples, Limits: limits, Retention: retention, Ranges: historyRanges(retention), RowLimit: rowLimit,
	})
}

// buildHostingInfo gathers platform/uname info, uptime, load average, and the server's public IP - the data behind both /json/system/hosting/info and /api/hosting/info
func buildHostingInfo(a *appctx.App, r *http.Request, username string) map[string]any {
	uptime, loadAvg := getUptimeAndLoad()
	platform := getPlatformInfo()

	return map[string]any{
		"system":    platform.System,
		"node":      platform.Node,
		"release":   platform.Release,
		"version":   platform.Version,
		"machine":   platform.Machine,
		"processor": platform.Processor,
		"ip":        a.GetCachedIPForUserOrPublicIPv4(r.Context(), username),
		"uptime":    uptime,
		"load_avg":  loadAvg,
	}
}

// handleSystemHostingInfo returns platform/uname info, uptime, load average, and the server's public IP
func handleSystemHostingInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	web.WriteJSON(w, http.StatusOK, buildHostingInfo(a, r, username))
}

// buildHostingPlan gathers the user's hosting plan limits plus the webserver/mysql type and nameservers configured for their context - the data behind both /json/system/hosting/plan and /api/hosting/plan
func buildHostingPlan(a *appctx.App, r *http.Request, userContext string, planID int) map[string]any {
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

	return map[string]any{
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
	}
}

// handleSystemHostingPlan returns the user's hosting plan limits plus the webserver/mysql type and nameservers configured for their context
func handleSystemHostingPlan(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := data["context"].(string)
	planID, _ := data["hosting_plan"].(int)

	plan := buildHostingPlan(a, r, userContext, planID)
	if userID, ok := auth.UserID(r); ok {
		plan["usage"] = dashboard.PlanUsage(a, r.Context(), userID, data)
	}
	web.WriteJSON(w, http.StatusOK, plan)
}

// getEnvPort reads /home/<context>/.env and, for host:port style values, returns just the port segment
func getEnvPort(context, key string) string {
	content, err := tenantfs.ReadFile("/home/" + context + "/.env")
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

// buildHostingPorts gathers the host-exposed ports for the user's MySQL and Postgres containers - the data behind both /json/system/hosting/ports and /api/hosting/ports
func buildHostingPorts(username string) map[string]any {
	return map[string]any{
		"remote_mysql_port":    getEnvPort(username, "MYSQL_PORT"),
		"remote_postgres_port": getEnvPort(username, "POSTGRES_PORT"),
	}
}

// handleSystemHostingPorts returns the host-exposed ports for the user's MySQL and Postgres containers
func handleSystemHostingPorts(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	data, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	web.WriteJSON(w, http.StatusOK, buildHostingPorts(username))
}

// featuresFile is the features list the admin panel ships, the Panel tab falls back to GitHub without it
const featuresFile = "/etc/openpanel/openadmin/config/features.json"

// readFeatures returns the file as-is when it holds a JSON array
func readFeatures(path string) ([]byte, bool) {
	raw, err := tenantfs.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var list []json.RawMessage
	if json.Unmarshal(raw, &list) != nil {
		return nil, false
	}
	return raw, true
}

// handleSystemFeatures serves the local features list, 404 tells the page to use the GitHub copy
func handleSystemFeatures(w http.ResponseWriter, _ *http.Request) {
	raw, ok := readFeatures(featuresFile)
	if !ok {
		web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "features list not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	_, _ = w.Write(raw)
}
