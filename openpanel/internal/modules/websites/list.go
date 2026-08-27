package websites

import (
	"database/sql"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// SiteRow is one row from the sites table, as read by list_sites().
type SiteRow struct {
	ID          int
	SiteName    string
	DomainID    int
	AdminEmail  string
	Version     string
	CreatedDate string
	Type        string
	Container   string
	Ports       string
	// IsStatic mirrors the case-sensitive `'static' in site[5]` check in
	// sites.html's Actions column (unlike every other type check on this
	// page, that one isn't lowercased first).
	IsStatic bool
	// Docroot is the site's full container-path docroot (the owning
	// domain's docroot plus any subdirectory suffix parsed out of
	// SiteName) - same computation dispatch.go's /website handler already
	// does per-request, needed here too so the Actions column's autologin
	// button can pass it straight through to each CMS's /<type>/login
	// endpoint without a second lookup.
	Docroot string
}

// SiteGroup is one type-grouped section of the /sites table (e.g. all
// "wordpress" rows together), in first-seen order.
type SiteGroup struct {
	Type  string
	Sites []SiteRow
}

// handleListSites loads the current user's sites, grouped by type, for the
// /sites listing page (or as JSON when requested).
func handleListSites(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	domains, err := a.AllDomainsForUser(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	viewMode := r.URL.Query().Get("view")
	if viewMode == "" {
		viewMode = "table"
	}

	rows, execErr := a.DB.QueryContext(ctx, `
		SELECT sites.id, sites.site_name, sites.domain_id, sites.admin_email, sites.version, sites.created_date,
		       sites.type, sites.container, sites.ports, domains.docroot
		FROM sites
		JOIN domains ON domains.domain_id = sites.domain_id
		WHERE sites.domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?)`, userID)
	if execErr != nil {
		_, _ = w.Write([]byte("An error occurred: " + execErr.Error()))
		return
	}
	defer rows.Close()

	groupIndex := map[string]int{}
	var groups []SiteGroup

	for rows.Next() {
		var (
			s                                                                          SiteRow
			siteName, adminEmail, version, createdDate, typ, container, ports, docroot sql.NullString
			domainID                                                                   sql.NullInt64
			id                                                                         sql.NullInt64
		)
		if scanErr := rows.Scan(&id, &siteName, &domainID, &adminEmail, &version, &createdDate, &typ, &container, &ports, &docroot); scanErr != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		s.ID = int(id.Int64)
		s.SiteName, s.DomainID = siteName.String, int(domainID.Int64)
		s.AdminEmail, s.Version, s.CreatedDate = adminEmail.String, version.String, createdDate.String
		s.Type, s.Container, s.Ports = typ.String, container.String, ports.String
		s.IsStatic = strings.Contains(s.Type, "static")

		_, folder := splitDomainAndFolder(s.SiteName)
		s.Docroot = docroot.String
		if folder != "" {
			s.Docroot = strings.TrimSuffix(s.Docroot, "/") + "/" + folder
		}
		key := strings.ToLower(s.Type)
		idx, ok := groupIndex[key]
		if !ok {
			idx = len(groups)
			groupIndex[key] = idx
			groups = append(groups, SiteGroup{Type: key})
		}
		groups[idx].Sites = append(groups[idx].Sites, s)
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{"title": "Websites", "groups": groups, "view_mode": viewMode})
		return
	}

	renderSitesPage(a, w, r, domains, groups, viewMode)
}
