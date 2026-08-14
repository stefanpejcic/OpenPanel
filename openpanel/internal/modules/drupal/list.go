package drupal

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// SiteRow is one row of manager/drupal_list.html's table.
type SiteRow struct {
	SiteName    string
	DomainID    int
	AdminEmail  string
	Version     string
	CreatedDate string
	Type        string
	ID          int
}

// handleListDrupal lists the user's Drupal sites.
func handleListDrupal(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	rows, execErr := a.DB.QueryContext(ctx, `
		SELECT site_name, domain_id, admin_email, version, created_date, type, id
		FROM sites
		WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?)
		AND type = 'drupal'`, userID)
	if execErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
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

	renderListPage(a, w, r, sites)
}
