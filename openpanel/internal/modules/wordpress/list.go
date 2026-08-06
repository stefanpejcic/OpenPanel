package wordpress

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// SiteRow is one row of manager/wp/list.html's table/cards.
type SiteRow struct {
	SiteName    string
	DomainID    int
	AdminEmail  string
	Version     string
	CreatedDate string
	Type        string
	ID          int
}

// handleListWordPress mirrors list_wordpress().
func handleListWordPress(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	domains, err := a.AllDomainsForUser(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	viewMode := r.URL.Query().Get("view")
	if viewMode == "" {
		viewMode = "cards"
	}

	rows, execErr := a.DB.QueryContext(ctx, `
		SELECT site_name, domain_id, admin_email, version, created_date, type, id
		FROM sites
		WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?)
		AND type = 'WordPress'`, userID)
	if execErr != nil {
		_, _ = w.Write([]byte("An error occurred: " + execErr.Error()))
		return
	}
	defer rows.Close()

	var sites []SiteRow
	for rows.Next() {
		var s SiteRow
		if scanErr := rows.Scan(&s.SiteName, &s.DomainID, &s.AdminEmail, &s.Version, &s.CreatedDate, &s.Type, &s.ID); scanErr != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		sites = append(sites, s)
	}

	renderListPage(a, w, r, domains, sites, viewMode)
}
