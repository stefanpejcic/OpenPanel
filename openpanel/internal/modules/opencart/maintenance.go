package opencart

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

var maintenancePrefixRE = regexp.MustCompile(`DB_PREFIX'\s*,\s*'([^']*)'`)

// handleOpenCartMaintenance reads (GET) or toggles (POST, action=enable|
// disable) OpenCart's built-in "Maintenance Mode" store setting
// (config_maintenance in `{prefix}setting`, store_id 0) - the same flag
// OpenCart's own admin Design > Design settings page writes, which its
// front controller already checks on every storefront request. Mirrors
// prestashop/maintenance.go's DB-flag approach (OpenCart has no CLI tool
// like drush/occ to shell out to).
func handleOpenCartMaintenance(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := openCartRequestParams(ctx, a, r, userID, userContext)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}

	catArgv := podmanmanager.PodmanArgv(userContext, "exec", phpContainer, "cat", docroot+"/config.php")
	content, catErr := podmanmanager.Command(ctx, userContext, catArgv).CombinedOutput()
	if catErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Unable to read config.php"})
		return
	}
	nameMatch := removeDBNameRE.FindStringSubmatch(string(content))
	prefixMatch := maintenancePrefixRE.FindStringSubmatch(string(content))
	if nameMatch == nil || prefixMatch == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Database name or prefix not found in config.php"})
		return
	}
	dbName, prefix := nameMatch[1], prefixMatch[1]

	if r.Method == http.MethodGet {
		rows, queryErr := mysqlmanager.Exec(ctx, userContext,
			"SELECT `value` FROM `"+prefix+"setting` WHERE `key` = 'config_maintenance' AND store_id = 0 LIMIT 1", dbName)
		status := "disabled"
		if queryErr == nil && len(rows) > 0 && toStringCell(rows[0][0]) == "1" {
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
	existing, existsErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT setting_id FROM `"+prefix+"setting` WHERE `key` = 'config_maintenance' AND store_id = 0 LIMIT 1", dbName)
	var execErr error
	if existsErr == nil && len(existing) > 0 {
		_, execErr = mysqlmanager.Exec(ctx, userContext,
			"UPDATE `"+prefix+"setting` SET `value` = '"+value+"' WHERE `key` = 'config_maintenance' AND store_id = 0", dbName)
	} else {
		_, execErr = mysqlmanager.Exec(ctx, userContext,
			"INSERT INTO `"+prefix+"setting` (store_id, code, `key`, `value`, serialized) VALUES (0, 'config', 'config_maintenance', '"+value+"', 0)", dbName)
	}
	if execErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Updating config_maintenance failed", "details": execErr.Error()})
		return
	}

	status := "disabled"
	if action == "enable" {
		status = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, action+"d OpenCart maintenance mode for "+domain, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]string{"message": "Maintenance mode " + action + "d successfully.", "status": status})
}
