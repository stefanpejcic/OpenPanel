package drupal

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// handleDrupalMaintenance reads (GET) or toggles (POST, action=enable|disable)
// Drupal's built-in maintenance mode via drush's `state:get`/`state:set` on
// the system.maintenance_mode state key - Drupal core's own front controller
// already checks this on every request, so no extra code needs to ship into
// the docroot.
func handleDrupalMaintenance(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
	drush := docroot + "/vendor/bin/drush"

	if r.Method == http.MethodGet {
		argv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, drush),
			"state:get", "system.maintenance_mode", "--root="+docroot)
		out, _ := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
		status := "disabled"
		if strings.TrimSpace(string(out)) == "1" {
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
	value := "0"
	if action == "enable" {
		value = "1"
	}
	argv := append(podmanmanager.PodmanArgv(userContext, "exec", phpContainer, drush),
		"state:set", "system.maintenance_mode", value, "--input-format=integer", "--root="+docroot)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "drush state:set failed", "details": strings.TrimSpace(string(out))})
		return
	}

	status := "disabled"
	if action == "enable" {
		status = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, action+"d Drupal maintenance mode for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Maintenance mode " + action + "d successfully.", "status": status})
}
