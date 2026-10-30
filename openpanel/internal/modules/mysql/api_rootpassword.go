package mysql

import (
	"encoding/json"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// apiMySQLSetRootPassword changes the MySQL root user's password (both the given host, defaulting to '%', and 'localhost'), then persists it to my.cnf/.env, no restart needed - the API equivalent of POST /mysql/root-password (handleRootPasswordMySQL in rootpassword.go) - there's no GET counterpart since the root password isn't readable anywhere, so this is write-only, and deliberately has no strength check, unlike the page, which also checks length and strength
func apiMySQLSetRootPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var body struct {
		Password string `json:"password"`
		Host     string `json:"host"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	password := body.Password
	dbHost := strings.TrimSpace(body.Host)
	if dbHost == "" {
		dbHost = "%"
	}
	if password == "" {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Password is required."})
		return
	}
	if rootPwHasBadChars(password) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": rootPwBadCharsMsg})
		return
	}
	if !validators.IsValidHost(dbHost) {
		writeAPIMySQLJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid host format."})
		return
	}
	escapedPassword := escapeMySQLString(password)

	if _, execErr := mysqlmanager.Exec(ctx, userContext, "ALTER USER 'root'@'"+dbHost+"' IDENTIFIED BY '"+escapedPassword+"'", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "ALTER USER 'root'@'localhost' IDENTIFIED BY '"+escapedPassword+"'", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}
	if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
		writeAPIMySQLJSON(w, http.StatusInternalServerError, map[string]string{"error": mysqlAPIError(execErr)})
		return
	}

	docker.SetEnvValue(userContext, "MYSQL_ROOT_PASSWORD", password)

	if writeErr := setMyCnfPassword(userContext, password); writeErr != nil {
		writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"updated": true, "restarted": false, "warning": "Password changed but failed to update stored configuration."})
		return
	}
	mysqlmanager.InvalidatePool(userContext)

	_ = logger.RecordUserAction(a.Config, currentUsername, "changed MySQL root user password via API", reqip.ClientIP(r))
	writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"updated": true, "restarted": false})
}
