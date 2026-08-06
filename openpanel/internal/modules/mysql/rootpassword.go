package mysql

import (
	"net/http"
	"os/exec"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// handleRootPasswordMySQL changes the MySQL root password. Notably, unlike
// every other password-change route in this package, no strength check is
// applied here - deliberately, since the root password is admin-only.
func handleRootPasswordMySQL(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := GetMySQLVersion(ctx, a, userContext)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		password := r.Form.Get("password")
		dbHost := r.Form.Get("db_host")
		if dbHost == "" {
			dbHost = "%"
		}
		escapedPassword := escapeMySQLString(password)

		ok := func() bool {
			if _, execErr := mysqlmanager.Exec(ctx, userContext, "ALTER USER 'root'@'"+dbHost+"' IDENTIFIED BY '"+escapedPassword+"'", ""); execErr != nil {
				flashSess(a, w, r, "error", "Error changing MySQL root password: "+execErr.Error())
				return false
			}
			if _, execErr := mysqlmanager.Exec(ctx, userContext, "ALTER USER 'root'@'localhost' IDENTIFIED BY '"+escapedPassword+"'", ""); execErr != nil {
				flashSess(a, w, r, "error", "Error changing MySQL root password: "+execErr.Error())
				return false
			}
			if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
				flashSess(a, w, r, "error", "Error changing MySQL root password: "+execErr.Error())
				return false
			}
			return true
		}()

		if ok {
			docker.SetEnvValue(userContext, "MYSQL_ROOT_PASSWORD", password)

			sedCmd := "s/^password=.*/password=" + escapedPassword + "/"
			myCnfPath := "/home/" + userContext + "/my.cnf"
			if sedErr := exec.CommandContext(ctx, "sed", "-i", "-e", sedCmd, myCnfPath).Run(); sedErr != nil {
				flashSess(a, w, r, "error", "Password changed but failed to restart MySQL.")
			} else {
				mysqlmanager.InvalidatePool(userContext)

				argv := podmanmanager.PodmanArgv(userContext, "restart", mysqlVersion)
				if restartErr := podmanmanager.Command(ctx, userContext, argv).Run(); restartErr != nil {
					flashSess(a, w, r, "error", "Password changed but failed to restart MySQL.")
				} else {
					ipAddress := reqip.ClientIP(r)
					_ = logger.RecordUserAction(a.Config, currentUsername, "changed MySQL root user password", ipAddress)
					flashSess(a, w, r, "success", "Successfully changed root password.")
				}
			}
		}
	}

	renderRootPasswordPage(a, w, r, mysqlVersion)
}
