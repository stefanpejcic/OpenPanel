package domains

import (
	"net/http"
	"os/exec"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// handleSuspendDomain suspends a domain (serves a static page instead of proxying it).
func handleSuspendDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domainsList, _ := a.AllDomainsForUser(ctx, userID)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		domainName := r.Form.Get("domain_name")
		comment := strings.TrimSpace(r.Form.Get("comment"))

		if domainName == "" {
			flashAndRedirect(a, w, r, "error", "Invalid request. Domain name must be provided.", "/domains/suspend")
			return
		}
		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}

		args := []string{"domains-suspend", domainName}
		if comment != "" {
			args = append(args, "--comment="+comment)
		}
		out, cmdErr := exec.CommandContext(ctx, "opencli", args...).CombinedOutput()
		if cmdErr == nil {
			invalidateRewriteCondCache(ctx, a, domainName)
			flashAndRedirect(a, w, r, "success", string(out), "/domains")
			currentUsername, _, _ := injected(a, ctx, userID)
			_ = logger.RecordUserAction(a.Config, currentUsername, "suspended domain "+domainName, reqip.ClientIP(r))
		} else {
			flashAndRedirect(a, w, r, "error", string(out), "/domains/suspend?domain="+domainName)
		}
		return
	}

	domainName := r.URL.Query().Get("domain")
	if domainName == "" {
		renderSuspendPage(a, w, r, "", domainsList)
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	renderSuspendPage(a, w, r, domainName, domainsList)
}

// handleUnsuspendDomain reactivates a suspended domain.
func handleUnsuspendDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domainsList, _ := a.AllDomainsForUser(ctx, userID)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		domainName := r.Form.Get("domain_name")

		if domainName == "" {
			flashAndRedirect(a, w, r, "error", "Invalid request. Domain name must be provided.", "/domains/unsuspend")
			return
		}
		if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
			http.Error(w, "You do not own this domain.", http.StatusForbidden)
			return
		}

		out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-unsuspend", domainName).CombinedOutput()
		if cmdErr == nil {
			invalidateRewriteCondCache(ctx, a, domainName)
			flashAndRedirect(a, w, r, "success", string(out), "/domains")
			currentUsername, _, _ := injected(a, ctx, userID)
			_ = logger.RecordUserAction(a.Config, currentUsername, "unsuspended domain "+domainName, reqip.ClientIP(r))
		} else {
			flashAndRedirect(a, w, r, "error", string(out), "/domains/unsuspend?domain="+domainName)
		}
		return
	}

	domainName := r.URL.Query().Get("domain")
	if domainName == "" {
		renderUnsuspendPage(a, w, r, "", domainsList)
		return
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	renderUnsuspendPage(a, w, r, domainName, domainsList)
}
