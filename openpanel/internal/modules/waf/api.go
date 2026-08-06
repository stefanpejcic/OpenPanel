package waf

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterAPI registers the WAF API routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "waf", "GET /api/waf", func(w http.ResponseWriter, r *http.Request) { apiWAFList(a, w, r) })
	apiregistry.Handle(mux, a, "waf", "GET /api/waf/{domain}", func(w http.ResponseWriter, r *http.Request) { apiWAFDomainGet(a, w, r) })
	apiregistry.Handle(mux, a, "waf", "POST /api/waf/{domain}", func(w http.ResponseWriter, r *http.Request) { apiWAFDomainToggle(a, w, r) })
	apiregistry.Handle(mux, a, "waf", "PUT /api/waf/{domain}/rules", func(w http.ResponseWriter, r *http.Request) { apiWAFDomainRules(a, w, r) })
	apiregistry.Handle(mux, a, "waf", "GET /api/waf/log/{domain}", func(w http.ResponseWriter, r *http.Request) { apiWAFLog(a, w, r) })
	apiregistry.Handle(mux, a, "waf", "GET /api/waf/stats/{domain}", func(w http.ResponseWriter, r *http.Request) { apiWAFStats(a, w, r) })
	apiregistry.Handle(mux, a, "waf", "GET /api/waf/ids/{id_type}", func(w http.ResponseWriter, r *http.Request) { apiWAFIDs(a, w, r) })
}

func apiOwnOr403(a *appctx.App, w http.ResponseWriter, r *http.Request, userID int, domain string) bool {
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		writeJSONError(w, http.StatusForbidden, "You do not own this domain.")
		return false
	}
	return true
}

// apiWAFList returns the WAF on/off status for every domain the current user owns.
func apiWAFList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domains, _ := a.AllDomainsForUser(r.Context(), userID)
	result := make(map[string]string, len(domains))
	for _, d := range domains {
		result[d.DomainURL] = wafStatusForDomain(d.DomainURL)
	}
	writeJSON(w, http.StatusOK, map[string]any{"domains": result})
}

// apiWAFDomainGet returns a single domain's WAF status plus its removed rule IDs/tags.
func apiWAFDomainGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := firstPathSegment(r.PathValue("domain"))
	if !apiOwnOr403(a, w, r, userID, domain) {
		return
	}

	status, removedRules, removedTags := readWAFStatus(domain)
	if removedRules == nil {
		removedRules = []string{}
	}
	if removedTags == nil {
		removedTags = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"domain": domain, "status": status, "removed_rules": removedRules, "removed_tags": removedTags,
	})
}

// readWAFStatus shares handleWAFDomain's GET-branch logic, factored out
// for the API's single-domain GET route which needs the values directly
// rather than a rendered page.
func readWAFStatus(domain string) (status string, removedRules, removedTags []string) {
	content, readErr := os.ReadFile(domainConfigPath(domain))
	if readErr != nil {
		return "Not Found", nil, nil
	}
	contentStr := string(content)
	switch {
	case strings.Contains(contentStr, "SecRuleEngine On"):
		status = "On"
	case strings.Contains(contentStr, "SecRuleEngine Off"):
		status = "Off"
	default:
		status = "Unknown"
	}

	foundRule, foundTag := false, false
	for _, line := range strings.Split(contentStr, "\n") {
		line = strings.TrimSpace(line)
		if !foundRule && strings.HasPrefix(line, "SecRuleRemoveById") {
			ids := strings.Fields(line)[1:]
			if len(ids) > 0 && ids[0] == excludedRuleID {
				ids = ids[1:]
			}
			removedRules = append(removedRules, ids...)
			foundRule = true
		} else if !foundTag && strings.HasPrefix(line, "SecRuleRemoveByTag") {
			tags := strings.Fields(line)[1:]
			if len(tags) > 0 && strings.EqualFold(tags[0], excludedTag) {
				tags = tags[1:]
			}
			removedTags = append(removedTags, tags...)
			foundTag = true
		}
		if foundRule && foundTag {
			break
		}
	}
	return status, removedRules, removedTags
}

// apiWAFDomainToggle enables or disables the WAF for a single domain and reloads Caddy.
func apiWAFDomainToggle(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := firstPathSegment(r.PathValue("domain"))
	if !apiOwnOr403(a, w, r, userID, domain) {
		return
	}
	currentUsername, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Status = r.Form.Get("status")
	}
	newStatus := strings.TrimSpace(body.Status)
	if newStatus != "On" && newStatus != "Off" {
		writeJSONError(w, http.StatusBadRequest, "status must be On or Off")
		return
	}

	path := domainConfigPath(domain)
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeJSONError(w, http.StatusNotFound, "Config file not found for "+domain)
		return
	}
	contentStr := string(content)
	if newStatus == "On" {
		contentStr = strings.ReplaceAll(contentStr, "SecRuleEngine Off", "SecRuleEngine On")
	} else {
		contentStr = strings.ReplaceAll(contentStr, "SecRuleEngine On", "SecRuleEngine Off")
	}

	if writeErr := os.WriteFile(path, []byte(contentStr), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	if reloadErr := reloadCaddy(ctx); reloadErr != nil {
		writeJSON(w, http.StatusMultiStatus, map[string]string{"warning": "Config written but Caddy reload failed"})
		return
	}

	actionWord := "disabled"
	if newStatus == "On" {
		actionWord = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, actionWord+" WAF for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"domain": domain, "status": newStatus})
}

var apiRuleIDRE = regexp.MustCompile(`^\d+$`)

// apiWAFDomainRules replaces the set of excluded rule IDs/tags for a domain and reloads Caddy.
func apiWAFDomainRules(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := firstPathSegment(r.PathValue("domain"))
	if !apiOwnOr403(a, w, r, userID, domain) {
		return
	}
	currentUsername, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		RemovedRules []string `json:"removed_rules"`
		RemovedTags  []string `json:"removed_tags"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	for _, ruleID := range body.RemovedRules {
		if !apiRuleIDRE.MatchString(ruleID) {
			writeJSONError(w, http.StatusBadRequest, "removed_rules must contain only numeric IDs")
			return
		}
	}
	// removed_tags has no meaningful format restriction - any non-empty
	// string is accepted as a tag.
	for _, tag := range body.RemovedTags {
		if tag == "" {
			writeJSONError(w, http.StatusBadRequest, "removed_tags contains invalid characters")
			return
		}
	}

	removedRules := append([]string{excludedRuleID}, filterOut(body.RemovedRules, excludedRuleID)...)
	removedTags := append([]string{excludedTag}, filterOutFold(body.RemovedTags, excludedTag)...)

	path := domainConfigPath(domain)
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeJSONError(w, http.StatusNotFound, "Config file not found for "+domain)
		return
	}

	newContent := rewriteDirectivesBlock(string(content), removedRules, removedTags)
	if writeErr := os.WriteFile(path, []byte(newContent), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	if reloadErr := reloadCaddy(ctx); reloadErr != nil {
		writeJSON(w, http.StatusMultiStatus, map[string]string{"warning": "Rules written but Caddy reload failed"})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated WAF rules for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]any{
		"domain": domain, "removed_rules": removedRules[1:], "removed_tags": removedTags[1:],
	})
}

func filterOutFold(items []string, exclude string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !strings.EqualFold(item, exclude) {
			out = append(out, item)
		}
	}
	return out
}

// apiWAFLog returns a paginated, newest-first page of Coraza log entries for a domain.
func apiWAFLog(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := firstPathSegment(r.PathValue("domain"))
	if !apiOwnOr403(a, w, r, userID, domain) {
		return
	}

	logPath := wafLogPath(domain)
	content, readErr := os.ReadFile(logPath)
	if readErr != nil {
		writeJSON(w, http.StatusOK, map[string]any{"domain": domain, "entries": []json.RawMessage{}, "total": 0})
		return
	}

	var allLogs []json.RawMessage
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		allLogs = append(allLogs, json.RawMessage(line))
	}
	for i, j := 0, len(allLogs)-1; i < j; i, j = i+1, j-1 {
		allLogs[i], allLogs[j] = allLogs[j], allLogs[i]
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 100
	}

	total := len(allLogs)
	start := (page - 1) * perPage
	end := start + perPage
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	entries := allLogs[start:end]
	if entries == nil {
		entries = []json.RawMessage{}
	}

	totalPages := (total + perPage - 1) / perPage
	writeJSON(w, http.StatusOK, map[string]any{
		"domain": domain, "entries": entries, "total": total,
		"page": page, "per_page": perPage, "total_pages": totalPages,
	})
}

// apiWAFStats returns check/block counts for a domain over a recent time window.
func apiWAFStats(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	domain := firstPathSegment(r.PathValue("domain"))
	if !apiOwnOr403(a, w, r, userID, domain) {
		return
	}

	seconds, _ := strconv.Atoi(r.URL.Query().Get("seconds"))
	if seconds <= 0 {
		seconds = 60
	}
	stats := readWAFLogs(wafLogPath(domain), seconds)
	writeJSON(w, http.StatusOK, map[string]any{
		"domain": domain, "seconds": seconds, "checks": stats.Checks, "blocks": stats.Blocks,
	})
}

// apiWAFIDs returns the cached list of valid rule IDs or tags for the given type.
func apiWAFIDs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	idType := r.PathValue("id_type")
	if idType != "tags" && idType != "ids" {
		writeJSONError(w, http.StatusBadRequest, "type must be 'tags' or 'ids'")
		return
	}

	data, _ := cache.Memoize(r.Context(), a.Cache, "waf_"+idType+"_cache", 300*time.Second, func() ([]string, error) {
		return extractIDs(r.Context(), idType), nil
	})
	if data == nil {
		data = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{idType: data})
}
