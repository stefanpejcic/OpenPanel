package waf

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// WAFIssue is one health-check issue surfaced on the WAF list page, e.g.
// a warning that the WAF is disabled for one or more domains.
type WAFIssue struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func wafStatusForDomain(domainName string) string {
	content, err := os.ReadFile(domainConfigPath(domainName))
	if err != nil {
		return "Not Found"
	}
	switch {
	case strings.Contains(string(content), "SecRuleEngine On"):
		return "On"
	case strings.Contains(string(content), "SecRuleEngine Off"):
		return "Off"
	default:
		return "Unknown"
	}
}

// notifySentinel fires off `opencli sentinel` without waiting for it to
// exit, so cmd.Start() is used deliberately instead of cmd.Run() - the
// caller shouldn't block on this notification.
func notifySentinel(domainName, statusText string) {
	cmd := exec.Command("opencli", "sentinel", "--action=waf_domain",
		"--title", "WAF "+statusText+" for domain",
		"--message", "CorazaWAF has been "+statusText+" for domain '"+domainName+"'.")
	_ = cmd.Start()
}

// handleWAFList handles the per-domain enable/disable toggle (POST) and
// the domain list / single domain status lookup (GET). Notably, a POST
// here does not redirect - it flashes and falls straight through to the
// GET rendering below, in the same response.
func handleWAFList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	username, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		// ParseMultipartForm (not just ParseForm) since the WAF panel
		// widget POSTs a FormData body, which the browser always encodes
		// as multipart/form-data - ParseForm alone can't see those fields
		// and silently leaves domain_name empty, which then fails the
		// ownership check below.
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
		if newStatus == "On" {
			statusText = "enabled"
		}

		sess, _ := a.Sessions.Get(r, session.CookieName)
		if content, readErr := os.ReadFile(configFilePath); readErr == nil {
			contentStr := string(content)
			switch newStatus {
			case "On":
				contentStr = strings.ReplaceAll(contentStr, "SecRuleEngine Off", "SecRuleEngine On")
			case "Off":
				contentStr = strings.ReplaceAll(contentStr, "SecRuleEngine On", "SecRuleEngine Off")
			}

			if writeErr := os.WriteFile(configFilePath, []byte(contentStr), 0o644); writeErr == nil {
				if reloadErr := reloadCaddy(r.Context()); reloadErr == nil {
					_ = logger.RecordUserAction(a.Config, username, statusText+" WAF for domain "+domainName, reqip.ClientIP(r))
					notifySentinel(domainName, statusText)
					flash.Add(sess, "success", "WAF for domain: "+domainName+" is now "+statusText)
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
			flash.Add(sess, "warning", "Config file for "+domainName+" not found")
		}
		_ = a.Sessions.Save(r, w, sess)
	}

	if requestedDomain := r.URL.Query().Get("domain"); requestedDomain != "" {
		requestedDomain = firstPathSegment(requestedDomain)
		if !a.CheckDomainBelongsToUser(r.Context(), userID, requestedDomain) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{requestedDomain: wafStatusForDomain(requestedDomain)})
		return
	}

	domains, _ := a.AllDomainsForUser(r.Context(), userID)
	modsecStatus := make(map[string]string, len(domains))
	var disabledDomains []string
	for _, d := range domains {
		status := wafStatusForDomain(d.DomainURL)
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
			Message: "WAF is disabled for " + disabledDomains[0] + ".",
		})
	default:
		issues = append(issues, WAFIssue{
			ID: "waf-disabled-summary", Severity: "warning",
			Message: fmt.Sprintf("WAF is disabled for %d domains.", len(disabledDomains)),
		})
	}

	renderWAFListPage(a, w, r, domains, modsecStatus, issues)
}
