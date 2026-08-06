package emails

import (
	"net/http"
	"os"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// defaultAliasesFile is the postfix-regex.cf file used for domain default addresses.
//
// https://docker-mailserver.github.io/docker-mailserver/v10.4/config/user-management/aliases/#configuring-regexp-aliases
// TODO: /usr/local/mail/openmail/docker-data/dms/config/postfix-regex.cf is readonly in docker compose!
const defaultAliasesFile = "/usr/local/mail/openmail/docker-data/dms/config/postfix-regex.cf"

// checkCurrentDefaultAliasForDomain looks up the current default (catch-all) alias for domain.
func checkCurrentDefaultAliasForDomain(domain string) string {
	content, err := os.ReadFile(defaultAliasesFile)
	if err != nil {
		return ""
	}
	pattern := regexp.MustCompile(`(?m)^/\*@` + regexp.QuoteMeta(domain) + `/\s+(\S+)`)
	m := pattern.FindStringSubmatch(string(content))
	if m == nil {
		return ""
	}
	return m[1]
}

// setDefaultAliasForDomain sets, replaces, or removes the default (catch-all) alias for domain.
func setDefaultAliasForDomain(domain, destination string) error {
	pattern := regexp.MustCompile(`(?m)^/\*@` + regexp.QuoteMeta(domain) + `/[ \t]+\S+[ \t]*\n?`)

	content := ""
	if data, err := os.ReadFile(defaultAliasesFile); err == nil {
		content = string(data)
	}

	if destination != "" {
		newLine := "/*@" + domain + "/ " + destination + "\n"
		if pattern.MatchString(content) {
			content = pattern.ReplaceAllLiteralString(content, newLine)
		} else {
			if content != "" && !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += newLine
		}
	} else {
		content = pattern.ReplaceAllString(content, "")
	}

	return os.WriteFile(defaultAliasesFile, []byte(content), 0o644)
}

// handleDefaultAlias views and updates a domain's default (catch-all) email address.
func handleDefaultAlias(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	domain := r.PathValue("domain")

	if domain == "" {
		renderDefaultAddressPage(a, w, r, "", domains, "")
		return
	}

	if !userDomains[domain] {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		destination := strings.TrimSpace(r.Form.Get("destination"))

		if destination != "" && !isValidEmail(destination) {
			flashAndRedirect(a, w, r, "error", "Invalid destination email address.", "/emails/default/"+domain)
			return
		}

		if err := setDefaultAliasForDomain(domain, destination); err != nil {
			flashAndRedirect(a, w, r, "error", "Failed to update default email: "+err.Error(), "/emails/default/"+domain)
			return
		}

		currentUsername, _, _ := injected(a, r)
		ipAddress := reqip.ClientIP(r)
		if destination != "" {
			_ = logger.RecordUserAction(a.Config, currentUsername, "set default email for "+domain+" to "+destination, ipAddress)
		} else {
			_ = logger.RecordUserAction(a.Config, currentUsername, "removed default email for "+domain, ipAddress)
		}
		http.Redirect(w, r, "/emails/default/"+domain, http.StatusFound)
		return
	}

	currentAlias := checkCurrentDefaultAliasForDomain(domain)
	renderDefaultAddressPage(a, w, r, domain, domains, currentAlias)
}
