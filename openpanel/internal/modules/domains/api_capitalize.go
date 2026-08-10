package domains

import (
	"encoding/json"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// apiDomainsGetCapitalize returns a domain's display-case override, if any.
func apiDomainsGetCapitalize(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	_, userContext, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userDomains := loadCapitalizedDomains(ctx, a, userContext)
	capitalizedDomain, ok := userDomains[domain]
	if !ok {
		capitalizedDomain = domain
	}
	writeAPIDomainsJSON(w, http.StatusOK, map[string]string{"domain": domain, "capitalized_domain": capitalizedDomain})
}

// apiDomainsSetCapitalize sets a domain's display-case override.
func apiDomainsSetCapitalize(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")
	if !apiOwnDomainOr403Domains(a, w, r, userID, domain) {
		return
	}
	currentUsername, userContext, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		CapitalizedDomain string `json:"capitalized_domain"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.CapitalizedDomain == "" {
		writeAPIDomainsJSON(w, http.StatusBadRequest, map[string]string{"error": "capitalized_domain is required."})
		return
	}

	if saveErr := saveCapitalizedDomain(ctx, a, userContext, domain, body.CapitalizedDomain); saveErr != nil {
		writeAPIDomainsJSON(w, http.StatusInternalServerError, map[string]string{"error": saveErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "capitalized domain "+domain+" to "+body.CapitalizedDomain+" via API", reqip.ClientIP(r))
	writeAPIDomainsJSON(w, http.StatusOK, map[string]string{"domain": domain, "capitalized_domain": body.CapitalizedDomain})
}
