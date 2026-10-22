package mysql

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleRootPasswordMySQL changes the MySQL root password - unlike every other password-change route in this package, no strength check is applied here, deliberately, since the root password is admin-only
func handleRootPasswordMySQL(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
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
				web.Flash(a, w, r, "error", web.Tr(a, r, "Error changing MySQL root password: %(error)s", "error", execErr.Error()))
				return false
			}
			if _, execErr := mysqlmanager.Exec(ctx, userContext, "ALTER USER 'root'@'localhost' IDENTIFIED BY '"+escapedPassword+"'", ""); execErr != nil {
				web.Flash(a, w, r, "error", web.Tr(a, r, "Error changing MySQL root password: %(error)s", "error", execErr.Error()))
				return false
			}
			if _, execErr := mysqlmanager.Exec(ctx, userContext, "FLUSH PRIVILEGES", ""); execErr != nil {
				web.Flash(a, w, r, "error", web.Tr(a, r, "Error changing MySQL root password: %(error)s", "error", execErr.Error()))
				return false
			}
			return true
		}()

		if ok {
			docker.SetEnvValue(userContext, "MYSQL_ROOT_PASSWORD", password)

			if writeErr := setMyCnfPassword(userContext, password); writeErr != nil {
				web.Flash(a, w, r, "error", "Password changed but failed to update stored configuration.")
			} else {
				mysqlmanager.InvalidatePool(userContext)
				ipAddress := reqip.ClientIP(r)
				_ = logger.RecordUserAction(a.Config, currentUsername, "changed MySQL root user password", ipAddress)
				web.Flash(a, w, r, "success", "Successfully changed root password.")
			}
		}
	}

	renderRootPasswordPage(a, w, r, mysqlVersion)
}

// setMyCnfPassword rewrites the password= line in place, a sed -i swaps the inode and the container's single-file bind mount keeps showing the old file until a restart
func setMyCnfPassword(userContext, password string) error {
	path := "/home/" + userContext + "/my.cnf"
	content, err := tenantfs.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "password=") {
			lines[i] = "password=" + password
		}
	}
	return tenantfs.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600)
}
