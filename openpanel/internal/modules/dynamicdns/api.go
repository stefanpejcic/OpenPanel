package dynamicdns

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterAPI wires the dynamic DNS JSON API routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "dynamic_dns", "GET /api/dynamic-dns", func(w http.ResponseWriter, r *http.Request) { apiDynamicDNSList(a, w, r) })
	apiregistry.Handle(mux, a, "dynamic_dns", "POST /api/dynamic-dns", func(w http.ResponseWriter, r *http.Request) { apiDynamicDNSCreate(a, w, r) })
	apiregistry.Handle(mux, a, "dynamic_dns", "PUT /api/dynamic-dns", func(w http.ResponseWriter, r *http.Request) { apiDynamicDNSUpdate(a, w, r) })
	apiregistry.Handle(mux, a, "dynamic_dns", "DELETE /api/dynamic-dns", func(w http.ResponseWriter, r *http.Request) { apiDynamicDNSDelete(a, w, r) })
}

// apiDynamicDNSList returns every dynamic DNS entry across all of the
// user's domains, grouped by domain.
func apiDynamicDNSList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	userDomains, err := a.AllDomainsForUser(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	result := map[string][]DynDNSEntry{}
	for _, d := range userDomains {
		entries := parseDynamicDNSFromZone(d.DomainURL)
		if len(entries) > 0 {
			result[d.DomainURL] = entries
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"domains": result})
}

type dynamicDNSAPIBody struct {
	Domain     string `json:"domain"`
	Subdomain  string `json:"subdomain"`
	IP         string `json:"ip"`
	LineNumber any    `json:"line_number"`
	Token      string `json:"token"`
}

func lineNumberFromBody(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case string:
		i, err := strconv.Atoi(n)
		return i, err == nil
	default:
		return 0, false
	}
}

// apiDynamicDNSCreate creates a new dynamic DNS entry and returns it along
// with its generated token.
func apiDynamicDNSCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body dynamicDNSAPIBody
	_ = json.NewDecoder(r.Body).Decode(&body)
	domain := strings.TrimSpace(body.Domain)
	subdomain := strings.TrimSpace(body.Subdomain)
	ip := strings.TrimSpace(body.IP)
	if ip == "" {
		ip = "0.0.0.0"
	}

	if domain == "" || subdomain == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and subdomain are required"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	token, ok := addDynamicDNSEntry(domain, subdomain, "A", ip)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create entry"})
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "created dynamic DNS entry "+subdomain+"."+domain, reqip.ClientIP(r))

	entries := parseDynamicDNSFromZone(domain)
	var entry *DynDNSEntry
	for i := range entries {
		if entries[i].Subdomain == subdomain && entries[i].Token == token {
			entry = &entries[i]
			break
		}
	}
	writeJSON(w, http.StatusCreated, map[string]any{"message": "Dynamic DNS entry created", "entry": entry})
}

// apiDynamicDNSUpdate rewrites an existing dynamic DNS entry's zone line
// by line number.
func apiDynamicDNSUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body dynamicDNSAPIBody
	_ = json.NewDecoder(r.Body).Decode(&body)
	domain := strings.TrimSpace(body.Domain)
	subdomain := strings.TrimSpace(body.Subdomain)
	ip := strings.TrimSpace(body.IP)
	token := strings.TrimSpace(body.Token)
	lineNumber, lineOK := lineNumberFromBody(body.LineNumber)

	if domain == "" || !lineOK || lineNumber == 0 || subdomain == "" || ip == "" || token == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain, line_number, subdomain, ip, and token are required"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	newLine := buildZoneLine(subdomain, "A", ip, token, "")
	if !updateZoneLine(domain, lineNumber, newLine) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to update entry — line number out of range"})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "updated dynamic DNS entry "+subdomain+"."+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Dynamic DNS entry updated"})
}

// apiDynamicDNSDelete removes a dynamic DNS entry's zone line by line
// number.
func apiDynamicDNSDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body dynamicDNSAPIBody
	_ = json.NewDecoder(r.Body).Decode(&body)
	domain := strings.TrimSpace(body.Domain)
	lineNumber, lineOK := lineNumberFromBody(body.LineNumber)

	if domain == "" || !lineOK || lineNumber == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and line_number are required"})
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	deleted, ok := deleteZoneLine(domain, lineNumber)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to delete entry — line number out of range"})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted dynamic DNS entry on "+domain+": "+deleted, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Dynamic DNS entry deleted", "deleted_line": deleted})
}
