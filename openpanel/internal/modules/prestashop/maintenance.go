package prestashop

import (
	"net/http"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

var maintenancePrefixRE = regexp.MustCompile(`'database_prefix'\s*=>\s*'([^']*)'`)

// handlePrestashopMaintenance reads (GET) or toggles (POST, action=enable|
// disable) PrestaShop's maintenance mode via the PS_SHOP_ENABLE row in
// `{prefix}configuration` - mirroring joomla/drupal/nextcloud's
// maintenance.go handlers for the same feature.
func handlePrestashopMaintenance(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := prestashopRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	catArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "cat", docroot+"/app/config/parameters.php")
	content, catErr := podmanmanager.Command(ctx, userContext, catArgv).CombinedOutput()
	if catErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to read app/config/parameters.php"})
		return
	}
	nameMatch := removeDBNameRE.FindStringSubmatch(string(content))
	prefixMatch := maintenancePrefixRE.FindStringSubmatch(string(content))
	if nameMatch == nil || prefixMatch == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Database name or prefix not found in app/config/parameters.php"})
		return
	}
	dbName, prefix := nameMatch[1], prefixMatch[1]

	if r.Method == http.MethodGet {
		rows, queryErr := mysqlmanager.Exec(ctx, userContext,
			"SELECT value FROM `"+prefix+"configuration` WHERE name = 'PS_SHOP_ENABLE' LIMIT 1", dbName)
		status := "disabled"
		if queryErr == nil && len(rows) > 0 && toStringCell(rows[0][0]) == "0" {
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
	value := "1"
	if action == "enable" {
		value = "0"
	}
	_, execErr := mysqlmanager.Exec(ctx, userContext,
		"UPDATE `"+prefix+"configuration` SET value = '"+value+"' WHERE name = 'PS_SHOP_ENABLE'", dbName)
	if execErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Updating PS_SHOP_ENABLE failed", "details": execErr.Error()})
		return
	}

	status := "disabled"
	if action == "enable" {
		status = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, action+"d PrestaShop maintenance mode for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Maintenance mode " + action + "d successfully.", "status": status})
}
