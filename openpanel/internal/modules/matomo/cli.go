package matomo

import (
	"net/http"
	"net/url"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleMatomoCacheClean runs Matomo's own `console core:clear-caches`
func handleMatomoCacheClean(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := cmsapp.RequestParams(ctx, a, r, userID, userContext)
	if !ok {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", docroot+"/console", "core:clear-caches", "--no-interaction")
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Clearing cache failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared Matomo cache for "+domain, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Cache cleared successfully."})
}

// handleMatomoLogin returns a one-time login link pointing at the openpanel-login.php helper deployed at install time (see login_php.go) - unlike the other CMS modules, no per-request DB token is minted here, since the helper's own baked-in secret token (generated once at install, see login_support.go's saveMatomoCredentials) already gates it
func handleMatomoLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "domain is required"})
		return
	}
	mainDomain := domain
	if idx := strings.Index(domain, "/"); idx != -1 {
		mainDomain = domain[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, mainDomain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	creds, credErr := loadMatomoCredentials(domain)
	if credErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "No stored admin credentials found for this install"})
		return
	}

	loginLink := "https://" + domain + "/" + openpanelLoginFileName + "?op_login=" + url.QueryEscape(creds.Token)
	maskedLink := loginLink
	if len(loginLink) > 10 {
		maskedLink = loginLink[:len(loginLink)-10] + "*****"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "generated auto-login link for Matomo admin: "+maskedLink, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"login_link": loginLink})
}
