package joomla

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// handleJoomlaMaintenance reads (GET) or toggles (POST, action=enable|disable)
// Joomla's own offline mode - the public $offline property in
// configuration.php that Joomla's front controller already checks on every
// request to show the "Site Offline" page, so no extra code needs to ship
// into the docroot the way opencart/prestashop's flag lives in the DB.
func handleJoomlaMaintenance(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := joomlaRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	if r.Method == http.MethodGet {
		argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c",
			`sed -n "s/.*\$offline[[:space:]]*=[[:space:]]*\(true\|false\).*/\1/p" "$1/configuration.php" | head -1`, "sh", docroot)
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
	value := "false"
	if action == "enable" {
		value = "true"
	}
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "sh", "-c",
		`sed -i "s/\$offline[[:space:]]*=[[:space:]]*\(true\|false\)/\$offline = `+value+`/" "$1/configuration.php"`, "sh", docroot)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Updating configuration.php failed", "details": strings.TrimSpace(string(out))})
		return
	}

	status := "disabled"
	if action == "enable" {
		status = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, action+"d Joomla maintenance mode for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Maintenance mode " + action + "d successfully.", "status": status})
}
