package domains

import (
	"net/http"
	"os"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// DomainRow is the per-row shape the domains list page renders.
type DomainRow struct {
	DomainID       int
	Docroot        string
	DomainURL      string
	PHPVersion     string
	SiteCount      int
	RedirectURL    string
	SuspendComment string
	DNS            bool
	IsSubdomain    bool
	HTTPS          string
	Status         string
}

// handleDomainsPage lists all domains for this account, paginated.
func handleDomainsPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	domainData, err := domainsWithSites(ctx, a, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	subdomainURLs := subdomainURLSet(domainData)

	rows := make([]DomainRow, 0, len(domainData))
	for _, d := range domainData {
		row := DomainRow{
			DomainID: d.DomainID, Docroot: d.Docroot, DomainURL: d.DomainURL,
			PHPVersion: d.PHPVersion, SiteCount: d.SiteCount,
		}
		row.RedirectURL = getRedirectURL(d.DomainURL)
		if _, statErr := os.Stat("/etc/bind/zones/" + d.DomainURL + ".zone"); statErr == nil {
			row.DNS = true
		}
		row.IsSubdomain = subdomainURLs[d.DomainURL]
		status := isRewriteCondEnabled(ctx, a, d.DomainURL)
		row.HTTPS, row.Status, row.SuspendComment = status.HTTPS, status.Suspended, status.SuspendComment
		rows = append(rows, row)
	}

	page := atoiDefault(r.URL.Query().Get("page"), 1)
	if page < 1 {
		page = 1
	}
	perPage := atoiDefault(a.Config.Get("domains_per_page", "100"), 100)
	if perPage < 1 {
		perPage = 100
	}

	totalDomains := len(rows)
	startIndex := (page - 1) * perPage
	endIndex := startIndex + perPage
	if startIndex > totalDomains {
		startIndex = totalDomains
	}
	if endIndex > totalDomains {
		endIndex = totalDomains
	}
	paginated := rows[startIndex:endIndex]

	totalPages := (totalDomains + perPage - 1) / perPage

	renderDomainsPage(a, w, r, paginated, totalPages, page, startIndex+1, endIndex, totalDomains)
}

// subdomainURLSet reduces domain categorization to just the set of URLs
// classified as subdomains - the domains list only needs
// domain.is_subdomain, not the full main/sub split.
func subdomainURLSet(domainData []DomainWithSite) map[string]bool {
	urls := make([]appctx.Domain, len(domainData))
	for i, d := range domainData {
		urls[i] = appctx.Domain{DomainURL: d.DomainURL}
	}
	_, subs := appctx.Categorize(urls)
	set := make(map[string]bool, len(subs))
	for _, s := range subs {
		set[s.DomainURL] = true
	}
	return set
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
