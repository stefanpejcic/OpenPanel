package autoinstaller

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// RegisterAPI wires the autoinstaller API route onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "autoinstaller", "GET /api/autoinstaller", func(w http.ResponseWriter, r *http.Request) {
		apiHandleAutoinstaller(a, w, r)
	})
}

// apiHandleAutoinstaller returns the user's site counts per technology, plus their site and domain lists, as JSON
func apiHandleAutoinstaller(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	domains, err := a.AllDomainsForUser(ctx, userID)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	rows, err := a.DB.QueryContext(ctx,
		"SELECT site_name, type FROM sites WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?)", userID)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	counts := make(map[string]int, len(technologies))
	for _, t := range technologies {
		counts[t] = 0
	}
	type siteEntry struct {
		SiteName string `json:"site_name"`
		Type     string `json:"type"`
	}
	sites := []siteEntry{}
	for rows.Next() {
		var siteName, siteType string
		if scanErr := rows.Scan(&siteName, &siteType); scanErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": scanErr.Error()})
			return
		}
		sites = append(sites, siteEntry{SiteName: siteName, Type: siteType})
		lowerType := strings.ToLower(siteType)
		for _, t := range technologies {
			if strings.Contains(lowerType, t) {
				counts[t]++
			}
		}
	}
	if err := rows.Err(); err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	web.WriteJSON(w, http.StatusOK, map[string]any{
		"sites": sites, "counts": counts, "technologies": technologies, "domain_count": len(domains),
	})
}
