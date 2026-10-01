package waf

import (
	"log"
	"net/http"
	"os"
	"os/exec"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// WAFIssue is one health-check issue surfaced on the WAF list page, e.g. a warning that the WAF is disabled for one or more domains
type WAFIssue struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Link     string `json:"link,omitempty"`
	// button text for Link, without it health-toast shows the old red Cancel link
	LinkLabel string `json:"link_label,omitempty"`
}

// StatusForDomain is the domain's SecRuleEngine value, "Not Found" when it has no conf
func StatusForDomain(domainName string) string {
	content, err := os.ReadFile(domainConfigPath(domainName))
	if err != nil {
		return "Not Found"
	}
	return engineStatus(string(content))
}

// notifySentinel fires off `opencli sentinel` without blocking the caller, Wait runs in a goroutine so the finished child doesn't stay a zombie
func notifySentinel(domainName, statusText string) {
	cmd := exec.Command("opencli", "sentinel", "--action=waf_domain",
		"--title", "WAF "+statusText+" for domain",
		"--message", "CorazaWAF has been "+statusText+" for domain '"+domainName+"'.")
	if err := cmd.Start(); err == nil {
		go func() { _ = cmd.Wait() }()
	}
}

// handleWAFList handles the per-domain enable/disable toggle (POST) and the domain list/single domain status lookup (GET) - a POST here does not redirect, it flashes and falls straight through to the GET rendering below in the same response
func handleWAFList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	username, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		// ParseMultipartForm since the WAF panel widget POSTs a FormData body, always encoded multipart/form-data - ParseForm alone can't see those fields and silently leaves domain_name empty
		_ = r.ParseMultipartForm(32 << 20)
		domainName := firstPathSegment(r.Form.Get("domain_name"))
		newStatus := r.Form.Get("modsec_action")

		if !a.CheckDomainBelongsToUser(r.Context(), userID, domainName) {
			log.Printf("WAF - Domain %s is not owned by the user.", domainName)
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}

		configFilePath := domainConfigPath(domainName)
		statusText := "disabled"
		switch newStatus {
		case "On":
			statusText = "enabled"
		case engineMonitor:
			statusText = "set to monitor only"
		}

		sess, _ := a.Sessions.Get(r, session.CookieName)
		if content, readErr := os.ReadFile(configFilePath); readErr == nil {
			contentStr := string(content)
			if validEngine(newStatus) {
				contentStr = setEngine(contentStr, newStatus)
			}

			if writeErr := os.WriteFile(configFilePath, []byte(contentStr), 0o644); writeErr == nil {
				if reloadErr := reloadCaddy(r.Context()); reloadErr == nil {
					_ = logger.RecordUserAction(a.Config, username, statusText+" WAF for domain "+domainName, reqip.ClientIP(r))
					notifySentinel(domainName, statusText)
					flash.Add(sess, "success", web.Tr(a, r, "WAF for domain: %(domain_name)s is now %(status_text)s", "domain_name", domainName, "status_text", web.Tr(a, r, statusText)))
				} else {
					log.Printf("WAF - Error changing WAF status for domain: %v", reloadErr)
					flash.Add(sess, "error", "Error changing WAF status.")
				}
			} else {
				log.Printf("WAF - Error changing WAF status for domain: %v", writeErr)
				flash.Add(sess, "error", "Error changing WAF status.")
			}
		} else {
			log.Printf("WAF - Error: config file for domain %s does not exist.", domainName)
			flash.Add(sess, "warning", web.Tr(a, r, "Config file for %(domain_name)s not found", "domain_name", domainName))
		}
		_ = a.Sessions.Save(r, w, sess)
		// toggle on the domain's own WAF page should land back there, not on the list
		switch r.Form.Get("return_to") {
		case "domain":
			http.Redirect(w, r, "/server/waf/"+domainName, http.StatusFound)
			return
		case "domains":
			http.Redirect(w, r, "/domains", http.StatusFound)
			return
		}
	}

	if requestedDomain := r.URL.Query().Get("domain"); requestedDomain != "" {
		requestedDomain = firstPathSegment(requestedDomain)
		if !a.CheckDomainBelongsToUser(r.Context(), userID, requestedDomain) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{requestedDomain: StatusForDomain(requestedDomain)})
		return
	}

	domains, _ := a.AllDomainsForUser(r.Context(), userID)
	modsecStatus := make(map[string]string, len(domains))
	rows := make(map[string]DomainWAFRow, len(domains))
	detected := detectProfilesForUser(r.Context(), a, userID)
	var disabledDomains []string
	for _, d := range domains {
		row := domainWAFRow(d.DomainURL, detected[d.DomainURL])
		rows[d.DomainURL] = row
		status := row.Status
		modsecStatus[d.DomainURL] = status
		if status == "Off" {
			disabledDomains = append(disabledDomains, d.DomainURL)
		}
	}

	var issues []WAFIssue
	switch len(disabledDomains) {
	case 0:
	case 1:
		issues = append(issues, WAFIssue{
			ID: "waf-disabled:" + disabledDomains[0], Severity: "warning",
			Message: web.Tr(a, r, "WAF is disabled for %(domain)s.", "domain", disabledDomains[0]),
		})
	default:
		issues = append(issues, WAFIssue{
			ID: "waf-disabled-summary", Severity: "warning",
			Message: web.Tr(a, r, "WAF is disabled for %(count)s domains.", "count", len(disabledDomains)),
		})
	}

	accountEnabled := true
	if _, userContext, ctxErr := injectedContext(a, r); ctxErr == nil && userContext != "" {
		accountEnabled = AccountWAFEnabled(userContext)
	}
	var recs []WAFIssue
	for _, d := range domains {
		for _, name := range rows[d.DomainURL].Recommended {
			recs = append(recs, WAFIssue{
				ID: "waf-profile:" + d.DomainURL + ":" + name, Severity: "info", Link: "/server/waf/" + d.DomainURL, LinkLabel: web.Tr(a, r, "Open"),
				Message: web.Tr(a, r, "%(app)s found on %(domain)s. Turn on the %(app)s firewall profile to avoid false blocks.", "app", name, "domain", d.DomainURL),
			})
		}
	}
	renderWAFListPage(a, w, r, WAFListPageData{Domains: domains, ModsecStatus: modsecStatus, Issues: issues, Rows: rows, AccountEnabled: accountEnabled, Recommendations: recs})
}
