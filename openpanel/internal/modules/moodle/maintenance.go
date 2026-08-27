package moodle

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// handleMoodleMaintenance reads (GET) or toggles (POST, action=enable|disable)
// Moodle's built-in CLI maintenance mode via admin/cli/maintenance.php -
// the same script update.go already wraps its own update run in. Enabling
// it writes a $CFG->dataroot/climaintenance.html marker file Moodle's own
// front controller checks on every request, so status is read back by
// testing for that file rather than parsing the script's own text output.
func handleMoodleMaintenance(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
	dataroot := "/var/www/html/" + siteSlug(domain) + "_moodledata"
	markerFile := dataroot + "/climaintenance.html"

	if r.Method == http.MethodGet {
		checkArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "test", "-f", markerFile)
		status := "disabled"
		if err := podmanmanager.Command(ctx, userContext, checkArgv).Run(); err == nil {
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
	flag := "--" + action
	argv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "php", approot+"/admin/cli/maintenance.php", flag)
	out, runErr := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	if runErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "admin/cli/maintenance.php failed", "details": strings.TrimSpace(string(out))})
		return
	}

	status := "disabled"
	if action == "enable" {
		status = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, action+"d Moodle maintenance mode for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Maintenance mode " + action + "d successfully.", "status": status})
}
