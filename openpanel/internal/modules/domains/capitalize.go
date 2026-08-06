package domains

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// capitalizedDomainsFile returns the per-account file storing display-case
// overrides for domain names (e.g. "MyBrand.com" instead of "mybrand.com"),
// using the same /home/<context>/ convention as every other per-account
// file in this codebase.
func capitalizedDomainsFile(userContext string) string {
	return "/home/" + userContext + "/capitalized_domains.json"
}

func capitalizedDomainsCacheKey(userContext string) string {
	return "capitalized_domains:" + userContext
}

// loadCapitalizedDomains reads (or returns the cached copy of) the full
// domain -> display-case map for one account.
func loadCapitalizedDomains(ctx context.Context, a *appctx.App, userContext string) map[string]string {
	result, _ := cache.Memoize(ctx, a.Cache, capitalizedDomainsCacheKey(userContext), cache.DefaultTTL, func() (map[string]string, error) {
		content, err := os.ReadFile(capitalizedDomainsFile(userContext))
		if err != nil {
			return map[string]string{}, nil
		}
		var m map[string]string
		if json.Unmarshal(content, &m) != nil {
			return map[string]string{}, nil
		}
		return m, nil
	})
	return result
}

func saveCapitalizedDomain(ctx context.Context, a *appctx.App, userContext, domainURL, capitalizedDomain string) error {
	path := capitalizedDomainsFile(userContext)
	userDomains := map[string]string{}
	if content, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(content, &userDomains)
	}
	userDomains[domainURL] = capitalizedDomain

	data, err := json.MarshalIndent(userDomains, "", "    ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}

	_ = a.Cache.Delete(ctx, capitalizedDomainsCacheKey(userContext))
	return nil
}

// handleCapitalizeDomains sets a display-case override for one domain.
func handleCapitalizeDomains(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	currentUsername, userContext, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		capitalizedDomain := r.Form.Get("capitalized_domain")
		if saveErr := saveCapitalizedDomain(ctx, a, userContext, domain, capitalizedDomain); saveErr == nil {
			flashSess(a, w, r, "success", "Domain has been capitalized to "+capitalizedDomain)
			_ = logger.RecordUserAction(a.Config, currentUsername, "capitalized domain "+domain+" to "+capitalizedDomain, reqip.ClientIP(r))
		}
	}

	userDomains := loadCapitalizedDomains(ctx, a, userContext)
	capitalizedDomain, ok := userDomains[domain]
	if !ok {
		capitalizedDomain = domain
	}

	renderCapitalizePage(a, w, r, domain, capitalizedDomain)
}

// handleDisplayCapitalizedDomains returns the display-case override map as JSON.
func handleDisplayCapitalizedDomains(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	_, userContext, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userDomains := loadCapitalizedDomains(ctx, a, userContext)
	writeJSON(w, http.StatusOK, userDomains)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
