package botcontrol

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// RegisterAPI wires the bot-control JSON API routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "bot_control", "GET /api/bot-control", func(w http.ResponseWriter, r *http.Request) { apiListPolicies(a, w, r) })
	apiregistry.Handle(mux, a, "bot_control", "POST /api/bot-control/{domain}", func(w http.ResponseWriter, r *http.Request) { apiSetPolicy(a, w, r) })
	apiregistry.Handle(mux, a, "bot_control", "GET /api/bot-control/{domain}/stats", func(w http.ResponseWriter, r *http.Request) { apiDomainStats(a, w, r) })
}

func apiListPolicies(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domainsList, err := a.AllDomainsForUser(ctx, userID)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list domains"})
		return
	}

	items := make([]map[string]any, 0, len(domainsList))
	for _, d := range domainsList {
		items = append(items, map[string]any{
			"domain": d.DomainURL,
			"policy": getDomainPolicy(d.DomainURL),
		})
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"domains": items})
}

func apiSetPolicy(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, username, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domainName := strings.TrimSpace(r.PathValue("domain"))
	if domainName == "" || !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		web.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "Domain not owned by user"})
		return
	}

	var body struct {
		Policy string `json:"policy"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	policy := Policy(strings.TrimSpace(body.Policy))
	if !validPolicy(policy) {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "policy must be allow, ratelimit, or block"})
		return
	}

	if err := setDomainPolicy(ctx, domainName, policy); err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update policy: " + err.Error()})
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "set AI bot policy to "+string(policy)+" for "+domainName+" via API", reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"domain": domainName,
		"policy": policy,
	})
}

func apiDomainStats(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domainName := strings.TrimSpace(r.PathValue("domain"))
	if domainName == "" || !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		web.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "Domain not owned by user"})
		return
	}

	window := r.URL.Query().Get("window")
	report, err := analyzeDomainTraffic(domainName, window, time.Now())
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to analyze traffic: " + err.Error()})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"policy": getDomainPolicy(domainName),
		"report": report,
	})
}
