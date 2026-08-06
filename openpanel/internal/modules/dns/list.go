package dns

import (
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/publicsuffix"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// DomainZoneRow represents a user's domain plus whether it has a zone
// file on disk. DomainName is populated from the domain's docroot - odd,
// but that's what the JSON output actually contains; nothing in dns.html
// ever reads it.
type DomainZoneRow struct {
	DomainID       int
	DomainName     string
	DomainURL      string
	ZoneFileExists bool
}

func (d DomainZoneRow) toJSONMap() map[string]any {
	return map[string]any{
		"domain_id": d.DomainID, "domain_name": d.DomainName, "domain_url": d.DomainURL,
		"redirect_url": nil, "zone_file_exists": d.ZoneFileExists,
	}
}

// hasSubdomainLabel reports whether domain has a label beyond its
// registrable (eTLD+1) form. Used only to pick the right "zone file not
// found" wording.
func hasSubdomainLabel(domain string) bool {
	etldPlusOne, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return false
	}
	return !strings.EqualFold(domain, etldPlusOne)
}

// handleEditDNSZone serves the domain-list landing page, and (with a
// domain) either the table view or the raw code view of its zone file.
func handleEditDNSZone(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domainParam := r.PathValue("domain")

	domainData, _ := a.AllDomainsForUser(ctx, userID)
	viewMode := r.URL.Query().Get("view")
	if viewMode == "" {
		viewMode = "table"
	}
	outputJSON := r.URL.Query().Get("output") == "json"

	if domainParam == "" {
		rows := make([]DomainZoneRow, 0, len(domainData))
		for _, d := range domainData {
			rows = append(rows, DomainZoneRow{
				DomainID: d.DomainID, DomainName: d.Docroot, DomainURL: d.DomainURL,
				ZoneFileExists: fileExists(zoneFilePath(d.DomainURL)),
			})
		}
		if outputJSON {
			dicts := make([]map[string]any, len(rows))
			for i, row := range rows {
				dicts[i] = row.toJSONMap()
			}
			writeJSON(w, http.StatusOK, dicts)
			return
		}
		renderDNSListPage(a, w, r, rows)
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domainParam) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	path := zoneFilePath(domainParam)
	if !fileExists(path) {
		var errorMessage string
		if hasSubdomainLabel(domainParam) {
			errorMessage = "Zone file not found for the subdomain - try editing the zone for apex domain."
		} else {
			errorMessage = "Zone file not found."
		}
		if outputJSON {
			writeJSON(w, http.StatusOK, errorMessage)
			return
		}
		flashAndRedirect(a, w, r, "error", errorMessage, "/domains/edit-dns-zone")
		return
	}

	content, err := os.ReadFile(path)
	if err != nil {
		flashAndRedirect(a, w, r, "error", "Zone file not found.", "/domains/edit-dns-zone")
		return
	}
	zoneContent := string(content)

	if outputJSON {
		writeJSON(w, http.StatusOK, zoneContent)
		return
	}

	var dnsZoneIssues []HealthIssue
	if zoneError := validateZoneFile(domainParam, path); zoneError != "" {
		dnsZoneIssues = append(dnsZoneIssues, HealthIssue{
			ID: "dns-zone-syntax:" + domainParam, Severity: "error",
			Message: "Syntax error in the DNS zone file for " + domainParam + ": " + zoneError,
		})
	}

	if viewMode == "code" {
		renderDNSCodePage(a, w, r, domainParam, zoneContent, dnsZoneIssues)
		return
	}

	serialNumber := readSerialNumber(strings.Split(zoneContent, "\n"))
	entries := parseZoneWithLineNumbers(zoneContent)
	rows := buildZoneRows(entries)
	renderDNSTablePage(a, w, r, domainParam, rows, serialNumber, dnsZoneIssues)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
