package emails

import (
	"context"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

const (
	dkimKeysBasePath = "/usr/local/mail/openmail/docker-data/dms/config/opendkim/keys"
	zoneTemplatePath = "/etc/openpanel/bind9/zone_template.txt"
)

var quotedChunkRE = regexp.MustCompile(`"([^"]*)"`)

// readDKIMExpectedRecord reads the DKIM TXT record we generated for domain
// and reconstructs the full value from its (possibly split) quoted chunks.
func readDKIMExpectedRecord(domain string) (string, bool) {
	keyFile := dkimKeysBasePath + "/" + domain + "/mail.txt"
	content, err := os.ReadFile(keyFile)
	if err != nil {
		return "", false
	}
	chunks := quotedChunkRE.FindAllStringSubmatch(string(content), -1)
	if len(chunks) == 0 {
		return "", false
	}
	var sb strings.Builder
	for _, m := range chunks {
		sb.WriteString(m[1])
	}
	return sb.String(), true
}

// zoneTemplateText reads the BIND zone template file, memoized 5 minutes.
func zoneTemplateText(ctx context.Context, a *appctx.App) string {
	text, _ := cache.Memoize(ctx, a.Cache, "email_zone_template_text", 5*time.Minute, func() (string, error) {
		data, err := os.ReadFile(zoneTemplatePath)
		if err != nil {
			return "", nil //nolint:nilerr // a missing template file just yields ""
		}
		return string(data), nil
	})
	return text
}

// defaultSPFRecord returns the default SPF record for serverIP, parsed
// out of the zone template if present, falling back to a generic record.
func defaultSPFRecord(ctx context.Context, a *appctx.App, serverIP string) string {
	for _, line := range strings.Split(zoneTemplateText(ctx, a), "\n") {
		if strings.Contains(line, "spf1") {
			if m := quotedChunkRE.FindStringSubmatch(line); m != nil {
				return strings.ReplaceAll(m[1], "{server_ip}", serverIP)
			}
		}
	}
	return "v=spf1 ip4:" + serverIP + " +a +mx ~all"
}

// defaultDMARCRecord returns the default DMARC record, parsed out of the
// zone template if present, falling back to a generic "p=none" policy.
func defaultDMARCRecord(ctx context.Context, a *appctx.App) string {
	for _, line := range strings.Split(zoneTemplateText(ctx, a), "\n") {
		if strings.Contains(line, "_dmarc") {
			if m := quotedChunkRE.FindStringSubmatch(line); m != nil {
				return m[1]
			}
		}
	}
	return "v=DMARC1; p=none;"
}

// queryTXTRecords looks up TXT records for name; each returned string is
// one record's chunks joined together.
func queryTXTRecords(name string) []string {
	records, err := net.LookupTXT(name)
	if err != nil {
		return nil
	}
	return records
}

// DeliverabilityCheck is the result of checking a domain's DKIM/SPF/DMARC
// deliverability status.
type DeliverabilityCheck struct {
	Domain   string            `json:"domain"`
	ServerIP string            `json:"server_ip"`
	OK       bool              `json:"ok"`
	DKIM     DeliverabilityRec `json:"dkim"`
	SPF      DeliverabilityRec `json:"spf"`
	DMARC    DeliverabilityRec `json:"dmarc"`
}

// DeliverabilityRec is one of DKIM/SPF/DMARC's status+current+expected.
type DeliverabilityRec struct {
	Status   string  `json:"status"`
	Current  *string `json:"current"`
	Expected *string `json:"expected"`
}

func strPtr(s string, ok bool) *string {
	if !ok {
		return nil
	}
	return &s
}

// checkDomainDeliverability compares the domain's live DKIM/SPF/DMARC DNS
// records against the expected values and reports a status for each.
func checkDomainDeliverability(ctx context.Context, a *appctx.App, domain, serverIP string) DeliverabilityCheck {
	dkimExpected, dkimExpectedOK := readDKIMExpectedRecord(domain)
	dkimTXT := queryTXTRecords("mail._domainkey." + domain)
	var dkimCurrent string
	dkimCurrentOK := len(dkimTXT) > 0
	if dkimCurrentOK {
		dkimCurrent = dkimTXT[0]
	}

	dkimStatus := "missing"
	if dkimCurrentOK && dkimExpectedOK && strings.TrimSpace(dkimCurrent) == strings.TrimSpace(dkimExpected) {
		dkimStatus = "ok"
	} else if dkimCurrentOK {
		dkimStatus = "mismatch"
	}

	spfRecords := queryTXTRecords(domain)
	var spfCurrent string
	spfCurrentOK := false
	for _, r := range spfRecords {
		if strings.HasPrefix(strings.ToLower(r), "v=spf1") {
			spfCurrent = r
			spfCurrentOK = true
			break
		}
	}
	spfExpected := defaultSPFRecord(ctx, a, serverIP)

	spfStatus := "missing"
	if spfCurrentOK && serverIP != "" && strings.Contains(spfCurrent, "ip4:"+serverIP) {
		spfStatus = "ok"
	} else if spfCurrentOK {
		spfStatus = "wrong_ip"
	}

	dmarcRecords := queryTXTRecords("_dmarc." + domain)
	var dmarcCurrent string
	dmarcCurrentOK := false
	for _, r := range dmarcRecords {
		if strings.HasPrefix(strings.ToLower(r), "v=dmarc1") {
			dmarcCurrent = r
			dmarcCurrentOK = true
			break
		}
	}
	dmarcExpected := defaultDMARCRecord(ctx, a)
	dmarcStatus := "missing"
	if dmarcCurrentOK {
		dmarcStatus = "ok"
	}

	return DeliverabilityCheck{
		Domain: domain, ServerIP: serverIP,
		OK:    dkimStatus == "ok" && spfStatus == "ok" && dmarcStatus == "ok",
		DKIM:  DeliverabilityRec{Status: dkimStatus, Current: strPtr(dkimCurrent, dkimCurrentOK), Expected: strPtr(dkimExpected, dkimExpectedOK)},
		SPF:   DeliverabilityRec{Status: spfStatus, Current: strPtr(spfCurrent, spfCurrentOK), Expected: strPtr(spfExpected, true)},
		DMARC: DeliverabilityRec{Status: dmarcStatus, Current: strPtr(dmarcCurrent, dmarcCurrentOK), Expected: strPtr(dmarcExpected, true)},
	}
}

// handleEmailsDeliverability renders (or, with ?output=json, returns) the
// deliverability status of every domain owned by the current user.
func handleEmailsDeliverability(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)

	if r.URL.Query().Get("output") == "json" {
		serverIP := getDedicatedOrSharedIP(ctx, currentUsername)

		results := make([]DeliverabilityCheck, len(domains))
		var wg sync.WaitGroup
		sem := make(chan struct{}, 8)
		for i, d := range domains {
			wg.Add(1)
			sem <- struct{}{}
			go func(i int, domainURL string) {
				defer wg.Done()
				defer func() { <-sem }()
				results[i] = checkDomainDeliverability(ctx, a, domainURL, serverIP)
			}(i, d.DomainURL)
		}
		wg.Wait()

		writeJSON(w, http.StatusOK, map[string]any{"domains": results})
		return
	}

	renderDeliverabilityPage(a, w, r, domains)
}

// handleEmailDeliverabilityDomain renders (or, with ?output=json, returns)
// the deliverability status for a single domain.
func handleEmailDeliverabilityDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := r.PathValue("domain")
	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if r.URL.Query().Get("output") == "json" {
		serverIP := getDedicatedOrSharedIP(ctx, currentUsername)
		writeJSON(w, http.StatusOK, checkDomainDeliverability(ctx, a, domain, serverIP))
		return
	}

	renderDeliverabilityDomainPage(a, w, r, domain)
}
