package nextcloud

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// handleNextcloudMaintenance reads (GET) or toggles (POST, action=enable|
// disable) Nextcloud's own maintenance mode via `occ maintenance:mode`,
// mirroring joomla/drupal's maintenance.go handlers for the same feature.
func handleNextcloudMaintenance(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := nextcloudRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}
	occ := docroot + "/occ"

	if r.Method == http.MethodGet {
		argv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
			occ, "config:system:get", "maintenance")
		out, _ := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
		status := "disabled"
		if strings.TrimSpace(string(out)) == "true" {
			status = "enabled"
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": status})
		return
	}

	action := strings.ToLower(r.FormValue("action"))
	if action != "enable" && action != "disable" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be 'enable' or 'disable'"})
		return
	}
	flag := "--off"
	if action == "enable" {
		flag = "--on"
	}
	argv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php"),
		occ, "maintenance:mode", flag)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "occ maintenance:mode failed", "details": strings.TrimSpace(string(out))})
		return
	}

	status := "disabled"
	if action == "enable" {
		status = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, action+"d Nextcloud maintenance mode for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Maintenance mode " + action + "d successfully.", "status": status})
}
