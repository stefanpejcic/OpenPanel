// Package goaccess serves the pre-rendered GoAccess HTML report per domain, generated externally (opencli/cron, once every 24h per the UI copy) and simply read from disk here.
package goaccess

import (
	"net/http"
	"os"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleDomainStats shows the domain picker when no domain_name is given, or reads and serves that domain's pre-rendered GoAccess HTML report file as-is
func handleDomainStats(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	domainName := r.PathValue("domain_name")
	if domainName == "" {
		domainsList, _ := a.AllDomainsForUser(ctx, userID)
		renderGoaccessSelectPage(a, w, r, domainsList)
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	_, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	logFilePath := "/var/log/caddy/stats/" + currentUsername + "/" + domainName + ".html"
	content, readErr := os.ReadFile(logFilePath)
	if readErr != nil {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Stats file for domain %(domain_name)s not found. Data is generated every 24h.", "domain_name", domainName), "/domains/stats")
		return
	}

	// goaccess_single.html has no {% extends %} - the report is a complete standalone HTML document, not wrapped in the panel's own layout, so this writes the file's bytes directly rather than going through a web.Page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(content)
}
