package moodle

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
)

// moodleRequestParams pulls the domain/docroot query params every handler
// in this file needs, splits the main domain out of a possible
// subdirectory suffix, verifies ownership, and resolves the PHP container -
// mirrors prestashop/cli.go's prestashopRequestParams.
func moodleRequestParams(ctx context.Context, a *appctx.App, r *http.Request, userID int, userContext string) (domain, docroot, phpContainer string, ok bool) {
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
	return domain, docroot, phpContainer, true
}

// moodleApprootContainerPath returns the approot's container-visible path
// for a given site (domain), derived the same way install.go computed it -
// docroot itself is a symlink into <approot>/public, but admin/cli/*
// scripts live one level up, in approot, so cache-clear/logs/cron all need
// this path instead of docroot.
func moodleApprootContainerPath(domain string) string {
	return "/var/www/html/" + siteSlug(domain) + "_moodleapp"
}

// handleMoodleCacheClean purges all Moodle caches via its own bundled CLI
// script (confirmed present in the release tarball at
// admin/cli/purge_caches.php - the standard, documented way to do this
// without a browser).
func handleMoodleCacheClean(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, _, phpContainer, ok := moodleRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	approot := moodleApprootContainerPath(domain)
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", approot+"/admin/cli/purge_caches.php")
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Clearing cache failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared Moodle caches for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Caches purged successfully."})
}

// handleMoodleLogs tails Moodle's PHP error log if one exists under
// moodledata (Moodle itself logs most events to its DB, readable only
// through its own admin UI - there's no simple flat application log file
// the way Joomla/PrestaShop have, so this surfaces PHP-level errors
// instead, consistent in spirit with every other module's Logs tab).
func handleMoodleLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, _, phpContainer, ok := moodleRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		http.Error(w, "domain and docroot are required, or you do not own this domain", http.StatusBadRequest)
		return
	}

	datarootContainerPath := "/var/www/html/" + siteSlug(domain) + "_moodledata"
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c",
		`f=$(ls -t "$1"/error.log "$1"/*.log 2>/dev/null | head -1); [ -n "$f" ] && tail -n 300 "$f"`, "sh", datarootContainerPath)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if runErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(out)
		return
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		_, _ = w.Write([]byte("No log entries found under moodledata."))
		return
	}
	_, _ = w.Write(out)
}

// mappedApproot converts a domain into its host-side approot path - same
// mapping install.go computes as approotHostPath.
func mappedApproot(userContext, domain string) string {
	return filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", siteSlug(domain)+"_moodleapp")
}
