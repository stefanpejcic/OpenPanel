package dynamicdns

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dns"
)

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

var dynDNSTokenRE = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// handleDynamicDNS serves the create/edit/delete panel and its POST
// action dispatch.
func handleDynamicDNS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		action := r.Form.Get("action")
		domain := r.Form.Get("domain")

		if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
			flashAndRedirect(a, w, r, "error", "You do not own this domain.", "/domains/dynamic-dns")
			return
		}

		ipAddress := reqip.ClientIP(r)

		switch action {
		case "create":
			subdomain := r.Form.Get("subdomain")
			ip := r.Form.Get("ip")
			if ip == "" {
				ip = "0.0.0.0"
			}
			if !validateSubdomain(subdomain) {
				flashAndRedirect(a, w, r, "error", "Invalid subdomain.", "/domains/dynamic-dns")
				return
			}
			if !validateIP(ip) {
				flashAndRedirect(a, w, r, "error", "Invalid IP address.", "/domains/dynamic-dns")
				return
			}
			if _, ok := addDynamicDNSEntry(domain, subdomain, "A", ip); ok {
				_ = logger.RecordUserAction(a.Config, currentUsername, "created dynamic DNS entry "+subdomain+"."+domain, ipAddress)
			}
			flashAndRedirect(a, w, r, "success", "Dynamic DNS entry created.", "/domains/dynamic-dns")
			return

		case "edit":
			lineNumber, convErr := strconv.Atoi(r.Form.Get("line_number"))
			if convErr != nil {
				flashAndRedirect(a, w, r, "error", "Invalid line number.", "/domains/dynamic-dns")
				return
			}
			subdomain := r.Form.Get("subdomain")
			ip := r.Form.Get("ip")
			token := r.Form.Get("token")

			if !validateSubdomain(subdomain) {
				flashAndRedirect(a, w, r, "error", "Invalid subdomain.", "/domains/dynamic-dns")
				return
			}
			if !validateIP(ip) {
				flashAndRedirect(a, w, r, "error", "Invalid IP address.", "/domains/dynamic-dns")
				return
			}
			if !dynDNSTokenRE.MatchString(token) {
				flashAndRedirect(a, w, r, "error", "Invalid token.", "/domains/dynamic-dns")
				return
			}

			newLine := buildZoneLine(subdomain, "A", ip, token, "")
			if updateZoneLine(domain, lineNumber, newLine) {
				_ = logger.RecordUserAction(a.Config, currentUsername, "updated dynamic DNS entry "+subdomain+"."+domain, ipAddress)
				flashAndRedirect(a, w, r, "success", "Dynamic DNS entry updated.", "/domains/dynamic-dns")
			} else {
				flashAndRedirect(a, w, r, "error", "Failed to update Dynamic DNS entry.", "/domains/dynamic-dns")
			}
			return

		case "delete":
			lineNumber, convErr := strconv.Atoi(r.Form.Get("line_number"))
			if convErr != nil {
				flashAndRedirect(a, w, r, "error", "Invalid line number.", "/domains/dynamic-dns")
				return
			}
			if deleted, ok := deleteZoneLine(domain, lineNumber); ok {
				_ = logger.RecordUserAction(a.Config, currentUsername, "deleted dynamic DNS entry on "+domain+": "+deleted, ipAddress)
				flashAndRedirect(a, w, r, "success", "Dynamic DNS entry deleted.", "/domains/dynamic-dns")
			} else {
				flashAndRedirect(a, w, r, "error", "Failed to delete Dynamic DNS entry.", "/domains/dynamic-dns")
			}
			return
		}

		http.Redirect(w, r, "/domains/dynamic-dns", http.StatusFound)
		return
	}

	userDomains, _ := a.AllDomainsForUser(ctx, userID)
	var domainEntries []DomainEntries
	globalIndex := 0
	for _, d := range userDomains {
		entries := parseDynamicDNSFromZone(d.DomainURL)
		if len(entries) > 0 {
			for i := range entries {
				entries[i].Index = globalIndex
				globalIndex++
			}
			domainEntries = append(domainEntries, DomainEntries{DomainName: d.DomainURL, Entries: entries})
		}
	}

	renderDynamicDNSPage(a, w, r, domainEntries, userDomains)
}

// handleDynamicDNSUpdate is the public, unauthenticated webcall a
// router/IoT device hits with just its token to push its current IP. Not
// wrapped in auth.RequireLogin, and needs no CSRF exemption - it's
// GET-only, and gorilla/csrf only enforces the token on unsafe methods by
// default.
func handleDynamicDNSUpdate(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	entries, err := os.ReadDir(dns.ZoneFileDir)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusNotFound)
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".zone") {
			continue
		}
		domain := strings.TrimSuffix(name, ".zone")

		content, readErr := os.ReadFile(filepath.Join(dns.ZoneFileDir, name))
		if readErr != nil {
			continue
		}
		lines := strings.Split(string(content), "\n")

		for i, rawLine := range lines {
			line := strings.TrimSpace(rawLine)
			m := webcallTokenRE.FindStringSubmatch(line)
			if m == nil || m[1] != token {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			subdomain := fields[0]
			recordType := fields[3]
			ip := reqip.ClientIP(r)
			lines[i] = buildZoneLine(subdomain, recordType, ip, token, "")

			if writeErr := writeZoneLines(domain, lines); writeErr != nil {
				http.Error(w, "Invalid token", http.StatusNotFound)
				return
			}
			dns.RestartDNSService(domain)

			writeJSON(w, http.StatusOK, map[string]string{
				"status": "updated", "host": subdomain + "." + domain, "ip": ip, "updated": nowUTCStr(),
			})
			return
		}
	}

	http.Error(w, "Invalid token", http.StatusNotFound)
}
