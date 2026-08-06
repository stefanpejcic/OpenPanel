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

// SiteRow is one row of the delete-confirmation "websites using this
// domain" list.
type SiteRow struct {
	SiteName string
}

// handleDeleteDomain removes a domain and its associated site.
func handleDeleteDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var domainURL string
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		domainURL = r.Form.Get("domain_url")
	} else {
		domainURL = r.URL.Query().Get("domain")
	}

	if domainURL == "" {
		flashAndRedirect(a, w, r, "error", "Domain name not provided.", "/domains")
		return
	}

	var ownerUserID int
	var docroot string
	if scanErr := a.DB.QueryRowContext(ctx, "SELECT user_id, docroot FROM domains WHERE domain_url = ?", domainURL).Scan(&ownerUserID, &docroot); scanErr != nil || ownerUserID != userID {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-delete", domainURL).CombinedOutput()
		if cmdErr == nil && strings.Contains(strings.ToLower(string(out)), "deleted successfully") {
			_ = logger.RecordUserAction(a.Config, currentUsername, "deleted domain "+domainURL, reqip.ClientIP(r))
			flashAndRedirect(a, w, r, "success", "Domain "+domainURL+" deleted successfully.", "/domains")
		} else {
			flashAndRedirect(a, w, r, "error", "Failed to delete domain "+domainURL+". Output: "+string(out), "/domains")
		}
		return
	}

	// GET
	siteCount, subdomainCount := 0, 0
	var sites []SiteRow

	rows, siteErr := a.DB.QueryContext(ctx, `
		SELECT site_name FROM sites
		WHERE domain_id IN (SELECT domain_id FROM domains WHERE domain_url = ?)
	`, domainURL)
	if siteErr == nil {
		defer rows.Close()
		for rows.Next() {
			var s SiteRow
			if rows.Scan(&s.SiteName) == nil {
				sites = append(sites, s)
			}
		}
		siteCount = len(sites)

		_ = a.DB.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM domains WHERE domain_url LIKE ? AND domain_url != ?
		`, "%."+domainURL, domainURL).Scan(&subdomainCount)
	}

	renderDeleteDomainPage(a, w, r, domainURL, docroot, siteCount, subdomainCount, sites)
}
