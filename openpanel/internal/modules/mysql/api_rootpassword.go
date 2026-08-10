package mysql

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// apiMySQLSetRootPassword changes the MySQL root user's password (both the
// given host, defaulting to '%', and 'localhost'), then persists it to
// my.cnf/.env and restarts the service. This is the API equivalent of
// POST /mysql/root-password (handleRootPasswordMySQL in rootpassword.go).
// There's no GET counterpart: the root password isn't readable anywhere
// (not even by the web page, which is a set-only form), so this is a
// write-only resource - no strength check is applied, deliberately,
// matching handleRootPasswordMySQL's own comment that the root password is
// admin-only.
func apiMySQLSetRootPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := GetMySQLVersion(ctx, a, userContext)

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

	sedCmd := "s/^password=.*/password=" + escapedPassword + "/"
	myCnfPath := "/home/" + userContext + "/my.cnf"
	if sedErr := exec.CommandContext(ctx, "sed", "-i", "-e", sedCmd, myCnfPath).Run(); sedErr != nil {
		writeAPIMySQLJSON(w, http.StatusOK, map[string]any{"updated": true, "restarted": false, "warning": "Password changed but failed to update stored configuration."})
		return
	}
	mysqlmanager.InvalidatePool(userContext)

	restarted := false
	argv := podmanmanager.PodmanArgv(userContext, "restart", mysqlVersion)
	if restartErr := podmanmanager.Command(ctx, userContext, argv).Run(); restartErr == nil {
		restarted = true
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "changed MySQL root user password via API", reqip.ClientIP(r))
	result := map[string]any{"updated": true, "restarted": restarted}
	if !restarted {
		result["warning"] = "Password changed but failed to restart MySQL."
	}
	writeAPIMySQLJSON(w, http.StatusOK, result)
}
