package drupal

import (
	"context"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// drushRequestParams pulls the domain/docroot query params every drush-backed handler needs, splits the main domain from any subdirectory suffix, verifies ownership, and resolves the PHP container to exec drush inside - shared by login/cache/logs so each handler only worries about its own drush subcommand
func drushRequestParams(ctx context.Context, a *appctx.App, r *http.Request, userID int, userContext string) (domain, docroot, phpContainer string, ok bool) {
	domain = r.URL.Query().Get("domain")
	docroot = r.URL.Query().Get("docroot")
	if domain == "" || docroot == "" {
		return "", "", "", false
	}

	mainDomain := domain
	if idx := strings.Index(domain, "/"); idx != -1 {
		mainDomain = domain[:idx]
	}
	if !a.CheckDomainBelongsToUser(ctx, userID, mainDomain) {
		return "", "", "", false
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	phpVersion := php.GetPHPVForDomain(ctx, a, userContext, mainDomain)
	phpContainer = webServer
	if !strings.Contains(strings.ToLower(webServer), "litespeed") {
		phpContainer = "php-fpm-" + phpVersion
	}

	// every drush-backed handler needs both the PHP and db containers actually running, or it fails with a confusing podman/drush low-level error instead of the real cause - starting both here (a no-op if already running) makes every action self-heal
	ensureContainerRunning(ctx, userContext, phpContainer)
	if mysqlContainer := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE"); mysqlContainer != "" {
		ensureContainerRunning(ctx, userContext, mysqlContainer)
	}

	// composer doesn't reliably leave vendor/bin/* executable here (came out 644 not 755), and drush's own wrapper chain hits three of those files, so every drush request chmods vendor/bin/ plus drush's real binary, self-healing sites installed before install.go did this too - uses `find -exec` not `sh -c` with a glob so the user-supplied docroot is a plain argv entry, not interpolated into shell syntax
	chmodArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "find",
		docroot+"/vendor/bin", docroot+"/vendor/drush/drush/drush", "-type", "f", "-exec", "chmod", "+x", "{}", "+")
	_, _ = podmanmanager.Command(ctx, userContext, chmodArgv).CombinedOutput()

	return domain, docroot, phpContainer, true
}

// handleDrupalLogin generates a one-time admin login link via Drush's built-in `user:login` (uli) - Drupal core already has this via the same one-time-login-hash system used for password resets, so no custom mu-plugin/token table is needed unlike wordpress/wpcli.go's "login" action
func handleDrupalLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := drushRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	siteURL := "https://" + domain
	drushArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, docroot+"/vendor/bin/drush"),
		"user:login", "--uri="+siteURL, "--root="+docroot)
	out, runErr := podmanmanager.Command(ctx, userContext, drushArgv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "drush user:login failed", "details": strings.TrimSpace(string(out))})
		return
	}

	loginLink := strings.TrimSpace(string(out))
	maskedLink := loginLink
	if len(loginLink) > 10 {
		maskedLink = loginLink[:len(loginLink)-10] + "*****"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "generated auto-login link for Drupal admin: "+maskedLink, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"login_link": loginLink})
}

// handleDrupalCacheRebuild runs `drush cache:rebuild` (cr), Drupal's equivalent of `wp cache flush` - Drupal has no single pluggable "cache type" like WordPress, so this rebuilds every cache bin rather than targeting one backend
func handleDrupalCacheRebuild(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := drushRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	siteURL := "https://" + domain
	drushArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, docroot+"/vendor/bin/drush"),
		"cache:rebuild", "--uri="+siteURL, "--root="+docroot)
	out, runErr := podmanmanager.Command(ctx, userContext, drushArgv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "drush cache:rebuild failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "rebuilt Drupal cache for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Cache rebuilt successfully."})
}

// handleDrupalLogs returns the last N watchdog (dblog) entries via `drush watchdog:show` as plain text, same shape python/node's /pm2/logs/ endpoint returns so the Logs tab can reuse that page's rendering JS
func handleDrupalLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := drushRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		http.Error(w, "domain and docroot are required, or you do not own this domain", http.StatusBadRequest)
		return
	}

	siteURL := "https://" + domain
	drushArgv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, docroot+"/vendor/bin/drush"),
		"watchdog:show", "--count=100", "--uri="+siteURL, "--root="+docroot)
	out, runErr := podmanmanager.Command(ctx, userContext, drushArgv).CombinedOutput()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if runErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(out)
		return
	}
	_, _ = w.Write(out)
}
