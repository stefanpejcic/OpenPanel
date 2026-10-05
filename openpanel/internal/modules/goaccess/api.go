package goaccess

import (
	"net/http"
	"os"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// RegisterAPI wires the GoAccess stats API routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "goaccess", "GET /api/stats", func(w http.ResponseWriter, r *http.Request) { apiStatsList(a, w, r) })
	apiregistry.Handle(mux, a, "goaccess", "GET /api/stats/{domain_name}", func(w http.ResponseWriter, r *http.Request) { apiStatsDomain(a, w, r) })
}

// apiStatsList reports, for every domain the caller owns, whether a pre-rendered GoAccess stats file currently exists for it
func apiStatsList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)

	type domainStat struct {
		DomainURL string `json:"domain_url"`
		HasStats  bool   `json:"has_stats"`
	}
	result := make([]domainStat, 0, len(domains))
	for _, d := range domains {
		logFile := "/var/log/caddy/stats/" + currentUsername + "/" + d.DomainURL + ".html"
		_, statErr := os.Stat(logFile)
		result = append(result, domainStat{DomainURL: d.DomainURL, HasStats: statErr == nil})
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"domains": result})
}

// apiStatsDomain returns the pre-rendered GoAccess HTML report for a single domain the caller owns
func apiStatsDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domainName := r.PathValue("domain_name")

	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		web.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "You do not own this domain"})
		return
	}

	_, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	logFile := "/var/log/caddy/stats/" + currentUsername + "/" + domainName + ".html"
	content, readErr := os.ReadFile(logFile)
	if readErr != nil {
		web.WriteJSON(w, http.StatusNotFound, map[string]any{
			"domain": domainName, "available": false, "message": "Stats file not found. Data is generated every 24h.",
		})
		return
	}

	web.WriteJSON(w, http.StatusOK, map[string]any{"domain": domainName, "available": true, "html": string(content)})
}
