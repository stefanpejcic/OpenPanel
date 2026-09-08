// Package websites (this file) exposes the /api/sites API equivalents of three UI-only routes: the temporary preview link generator (temporarylink.go), the live-visitors counter (visitors.go), and the WordPress info panel (websites.go's handleWebsiteWPInfo). Each is wired as a literal-suffix case in api_sites.go's apiSitesGetDispatch, alongside safebrowsing/pagespeed/wp-vulnerability - see RegisterSitesAPI's doc comment for why a single {rest...} wildcard is used instead of separate mux patterns.
package websites

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// apiTemporaryLink is the /api/sites/{domain}/temporary-link equivalent of handleTemporaryLink, taking the domain from the path instead of a query param - GET, not POST, since the result is memoized/idempotent within its 800s cache window, same as GET /api/sites/{domain}/pagespeed.
func apiTemporaryLink(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")
	domainRoot, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	result, status := temporaryLinkForDomain(ctx, a, currentUsername, domain)
	writeJSON(w, status, result)
}

// apiVisitors is the /api/sites/{domain}/visitors equivalent of handleVisitors - unlike the UI route, this enforces domain ownership, since handleVisitors deliberately skips that check but every other /api/sites/{domain}/... route requires it and this API follows that convention.
func apiVisitors(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	domainRoot, _ := splitDomainAndFolder(domain)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainRoot) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	writeJSON(w, http.StatusOK, visitorsForDomain(domain, visitorsSeconds(r)))
}

// apiWPInfo is the /api/sites/{domain}/wp-info equivalent of handleWebsiteWPInfo (database credentials, WP/PHP/MySQL versions for the site-manager info panel) - the path segment may include a subfolder, matching the original /website/wp_info/{site_name...} wildcard behavior.
func apiWPInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	siteName := r.PathValue("domain")
	domainNameUsed, _ := splitDomainAndFolder(siteName)

	if !a.CheckDomainBelongsToUser(ctx, userID, domainNameUsed) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	info, ok := wpInfoForSite(a, r, userContext, siteName)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Unable to detect docroot for the domain"})
		return
	}

	writeJSON(w, http.StatusOK, info)
}
