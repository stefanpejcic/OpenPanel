package moodle

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// moodleApprootContainerPath returns the approot's container-visible path for a given site (domain), derived the same way install.go computed it - docroot itself is a symlink into <approot>/public, but admin/cli/* scripts live one level up in approot, so cache-clear/logs/cron all need this path instead of docroot
func moodleApprootContainerPath(domain string) string {
	return "/var/www/html/" + appkit.SiteSlug(domain) + "_moodleapp"
}

// handleMoodleCacheClean purges all Moodle caches via its own bundled CLI script (admin/cli/purge_caches.php - the standard, documented way to do this without a browser)
func handleMoodleCacheClean(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, _, phpContainer, ok := cmsapp.RequestParams(ctx, a, r, userID, userContext)
	if !ok {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	approot := moodleApprootContainerPath(domain)
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", approot+"/admin/cli/purge_caches.php")
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Clearing cache failed", "details": strings.TrimSpace(string(out))})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "cleared Moodle caches for "+domain, reqip.ClientIP(r))
	web.WriteJSON(w, http.StatusOK, map[string]string{"message": "Caches purged successfully."})
}

// handleMoodleLogs tails Moodle's PHP error log if one exists under moodledata (Moodle itself logs most events to its DB, readable only through its own admin UI - there's no simple flat application log file the way Joomla/PrestaShop have, so this surfaces PHP-level errors instead, consistent in spirit with every other module's Logs tab)
func handleMoodleLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, _, phpContainer, ok := cmsapp.RequestParams(ctx, a, r, userID, userContext)
	if !ok {
		http.Error(w, "domain and docroot are required, or you do not own this domain", http.StatusBadRequest)
		return
	}

	datarootContainerPath := "/var/www/html/" + appkit.SiteSlug(domain) + "_moodledata"
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
