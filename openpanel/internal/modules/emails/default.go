package emails

import (
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// defaultAliasesFile keeps the domains' default addresses as "/*@domain/ destination" lines; docker-mailserver never reads it, `opencli email-default_sync` turns it into the postfix-regexp.cf rules it does read
const defaultAliasesFile = "/usr/local/mail/openmail/docker-data/dms/config/postfix-regex.cf"

// checkCurrentDefaultAliasForDomain looks up the current default (catch-all) alias for domain.
func checkCurrentDefaultAliasForDomain(domain string) string {
	content, err := tenantfs.ReadFile(defaultAliasesFile)
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
	// this file is shared by every account, one line per domain
	if domain == "" || strings.ContainsAny(domain+destination, "/ \t\r\n\x00") {
		return fmt.Errorf("invalid domain or destination")
	}
	pattern := regexp.MustCompile(`(?m)^/\*@` + regexp.QuoteMeta(domain) + `/[ \t]+\S+[ \t]*\n?`)

	content := ""
	if data, err := tenantfs.ReadFile(defaultAliasesFile); err == nil {
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

	if err := tenantfs.WriteFile(defaultAliasesFile, []byte(content), 0o644); err != nil {
		return err
	}
	if out, err := exec.Command("opencli", "email-default_sync").CombinedOutput(); err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
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
			web.FlashRedirect(a, w, r, "error", "Invalid destination email address.", "/emails/default/"+domain)
			return
		}

		if err := setDefaultAliasForDomain(domain, destination); err != nil {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Failed to update default email: %(error)s", "error", err.Error()), "/emails/default/"+domain)
			return
		}

		_, currentUsername, _, _ := auth.Injected(a, r)
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
