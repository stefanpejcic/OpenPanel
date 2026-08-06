package domains

import (
	"net/http"
	"os/exec"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
)

// handleDomainDocroot views or changes a domain's document root.
func handleDomainDocroot(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	var domainName string
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		domainName = r.Form.Get("domain_name")
		if domainName == "" {
			flashAndRedirect(a, w, r, "error", "Invalid request. Domain name must be provided.", "/domains")
			return
		}
	} else {
		domainName = r.URL.Query().Get("domain_name")
		if domainName == "" {
			domainsList, _ := a.AllDomainsForUser(ctx, userID)
			renderDocrootPage(a, w, r, "", "", domainsList)
			return
		}
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	currentUsername, _, _ := injected(a, ctx, userID)

	if r.Method == http.MethodPost {
		providedDocroot := r.Form.Get("new_docroot")
		docroot, ok := resolveUnderVarWWWHTML(providedDocroot)
		if providedDocroot == "" {
			flashAndRedirect(a, w, r, "error", "docroot must be provided.", "/domains/docroot")
			return
		}
		if !ok {
			flashAndRedirect(a, w, r, "error", "Docroot must be inside '/var/www/html/' directory.", "/domains/docroot")
			return
		}

		out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-docroot", domainName, "update", docroot).CombinedOutput()
		if cmdErr == nil {
			flashAndRedirect(a, w, r, "success", strings.TrimSpace(string(out)), "/domains/docroot?domain_name="+domainName)
			_ = logger.RecordUserAction(a.Config, currentUsername, "changed docroot for "+domainName+" to "+docroot, "")
		} else {
			flashAndRedirect(a, w, r, "error", strings.TrimSpace(string(out)), "/domains/docroot?domain_name="+domainName)
		}
		return
	}

	// GET
	var docroot string
	out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-docroot", domainName).CombinedOutput()
	if cmdErr == nil {
		docroot = strings.TrimSpace(string(out))
	} else {
		flashSess(a, w, r, "error", strings.TrimSpace(string(out)))
	}

	renderDocrootPage(a, w, r, domainName, docroot, nil)
}
