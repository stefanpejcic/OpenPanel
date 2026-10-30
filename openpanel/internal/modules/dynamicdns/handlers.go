package dynamicdns

import (
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dns"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var dynDNSTokenRE = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// hostProblem returns why a host can't be saved as an untranslated message with a %(host)s slot, or "" when it's fine
func hostProblem(zone, domain, sub, ip string, skipLine int) string {
	switch {
	case sub == "":
		return "Enter a subdomain, like home."
	case !validateSubdomain(sub):
		return "Use letters, numbers and hyphens (up to 63), not starting or ending with a hyphen."
	case !validateIPv4(ip):
		return "Enter an IPv4 address like 198.51.100.23. Dynamic DNS hosts are A records."
	}
	switch hostConflict(zone, domain, sub, skipLine) {
	case conflictDDNS:
		return "%(host)s is already a Dynamic DNS host."
	case conflictRecord:
		return "%(host)s already has a DNS record. Pick another subdomain or remove that record first."
	}
	return ""
}

func readZone(domain string) string {
	content, _ := tenantfs.ReadFile(zoneFilePath(domain))
	return string(content)
}

// handleDynamicDNS serves the hosts page and its create/edit/delete form posts
func handleDynamicDNS(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	const back = "/domains/dynamic-dns"

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		action := r.Form.Get("action")
		domain := strings.TrimSpace(r.Form.Get("domain"))

		// the ownership check stops at the first slash, the zone path doesn't
		if strings.ContainsAny(domain, "/\\") || !a.CheckDomainBelongsToUser(ctx, userID, domain) {
			web.FlashRedirect(a, w, r, "error", "You do not own this domain.", back)
			return
		}

		ipAddress := reqip.ClientIP(r)
		subdomain := strings.TrimSpace(r.Form.Get("subdomain"))
		ip := strings.TrimSpace(r.Form.Get("ip"))
		host := subdomain + "." + domain

		switch action {
		case "create":
			if ip == "" {
				ip = "0.0.0.0"
			}
			if msg := hostProblem(readZone(domain), domain, subdomain, ip, 0); msg != "" {
				web.FlashRedirect(a, w, r, "error", web.Tr(a, r, msg, "host", host), back)
				return
			}
			if _, ok := addDynamicDNSEntry(domain, subdomain, "A", ip); !ok {
				web.FlashRedirect(a, w, r, "error", "Failed to create Dynamic DNS entry.", back)
				return
			}
			_ = logger.RecordUserAction(a.Config, currentUsername, "created dynamic DNS entry "+host, ipAddress)
			web.FlashRedirect(a, w, r, "success", "Dynamic DNS entry created.", back+"?host="+url.QueryEscape(host))
			return

		case "edit":
			lineNumber, convErr := strconv.Atoi(r.Form.Get("line_number"))
			entry, found := entryAt(domain, lineNumber)
			if convErr != nil || !found || (r.Form.Get("token") != "" && r.Form.Get("token") != entry.Token) {
				web.FlashRedirect(a, w, r, "error", "That Dynamic DNS entry changed or no longer exists. Reload the page and try again.", back)
				return
			}
			if msg := hostProblem(readZone(domain), domain, subdomain, ip, lineNumber); msg != "" {
				web.FlashRedirect(a, w, r, "error", web.Tr(a, r, msg, "host", host), back)
				return
			}
			if !updateZoneLine(domain, lineNumber, buildZoneLine(subdomain, "A", ip, entry.Token, "")) {
				web.FlashRedirect(a, w, r, "error", "Failed to update Dynamic DNS entry.", back)
				return
			}
			_ = logger.RecordUserAction(a.Config, currentUsername, "updated dynamic DNS entry "+host, ipAddress)
			web.FlashRedirect(a, w, r, "success", "Dynamic DNS entry updated.", back+"?host="+url.QueryEscape(host))
			return

		case "regenerate":
			lineNumber, convErr := strconv.Atoi(r.Form.Get("line_number"))
			entry, found := entryAt(domain, lineNumber)
			if convErr != nil || !found || (r.Form.Get("token") != "" && r.Form.Get("token") != entry.Token) {
				web.FlashRedirect(a, w, r, "error", "That Dynamic DNS entry changed or no longer exists. Reload the page and try again.", back)
				return
			}
			if !regenerateToken(domain, entry) {
				web.FlashRedirect(a, w, r, "error", "Failed to create a new token.", back)
				return
			}
			host = entry.Subdomain + "." + domain
			_ = logger.RecordUserAction(a.Config, currentUsername, "created a new update token for dynamic DNS entry "+host, ipAddress)
			web.FlashRedirect(a, w, r, "success", "New token created. The old update URL stopped working, put the new one on your device.", back+"?host="+url.QueryEscape(host))
			return

		case "delete":
			lineNumber, convErr := strconv.Atoi(r.Form.Get("line_number"))
			entry, found := entryAt(domain, lineNumber)
			if convErr != nil || !found || (r.Form.Get("token") != "" && r.Form.Get("token") != entry.Token) {
				web.FlashRedirect(a, w, r, "error", "That Dynamic DNS entry changed or no longer exists. Reload the page and try again.", back)
				return
			}
			if deleted, ok := deleteZoneLine(domain, lineNumber); ok {
				_ = logger.RecordUserAction(a.Config, currentUsername, "deleted dynamic DNS entry on "+domain+": "+deleted, ipAddress)
				web.FlashRedirect(a, w, r, "success", "Dynamic DNS entry deleted.", back)
			} else {
				web.FlashRedirect(a, w, r, "error", "Failed to delete Dynamic DNS entry.", back)
			}
			return
		}

		http.Redirect(w, r, back, http.StatusFound)
		return
	}

	userDomains, _ := a.AllDomainsForUser(ctx, userID)
	var hosts []DynHost
	for _, d := range userDomains {
		for _, e := range parseDynamicDNSFromZone(d.DomainURL) {
			hosts = append(hosts, newDynHost(d.DomainURL, e))
		}
	}

	renderDynamicDNSPage(a, w, r, hosts, userDomains)
}

// handleDynamicDNSUpdate is the public, unauthenticated webcall a router/IoT device hits with just its token to push its current IP - not wrapped in auth.RequireLogin, and needs no CSRF exemption since it's GET-only and gorilla/csrf only enforces the token on unsafe methods
func handleDynamicDNSUpdate(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	entries, err := tenantfs.ReadDir(dns.ZoneFileDir)
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

		content, readErr := tenantfs.ReadFile(filepath.Join(dns.ZoneFileDir, name))
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
			// an A record can't take the IPv6 address a dual-stack device calls from by default
			if !ipFitsRecordType(ip, recordType) {
				http.Error(w, wrongFamilyMessage(recordType), http.StatusBadRequest)
				return
			}
			// one stamp for the zone and the reply, the DNS reload in between takes a few seconds
			updated := nowUTCStr()
			lines[i] = buildZoneLine(subdomain, recordType, ip, token, updated)

			if writeErr := writeZoneLines(domain, lines); writeErr != nil {
				http.Error(w, "Invalid token", http.StatusNotFound)
				return
			}
			dns.RestartDNSService(domain)

			web.WriteJSON(w, http.StatusOK, map[string]string{
				"status": "updated", "host": subdomain + "." + domain, "ip": ip, "updated": updated,
			})
			return
		}
	}

	http.Error(w, "Invalid token", http.StatusNotFound)
}

// wrongFamilyMessage is the update URL's reply when the caller's IP family doesn't fit the record
func wrongFamilyMessage(recordType string) string {
	if strings.EqualFold(recordType, "AAAA") {
		return "IPv6 address required"
	}
	return "IPv4 address required"
}
