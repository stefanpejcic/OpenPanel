package ojs

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func removeOJSDB(a *appctx.App, w http.ResponseWriter, r *http.Request, s *cmsapp.Site) {
	ctx := r.Context()
	dbInfo := extractOJSDatabaseInfoForLogin(s.UserContext, s.Name)
	if dbInfo["error"] != "" || dbInfo["database_name"] == "" {
		web.Flash(a, w, r, "warning", "Database name or user not found in config.inc.php")
		return
	}
	dbName := dbInfo["database_name"]
	_, _ = mysqlmanager.Exec(ctx, s.UserContext, "DROP DATABASE IF EXISTS `"+dbName+"`", "")
	if dbUser := dbInfo["database_username"]; dbUser != "" {
		_, _ = mysqlmanager.Exec(ctx, s.UserContext, "DROP USER IF EXISTS '"+dbUser+"'@'%'", "")
	}
	appkit.InvalidateMySQLCaches(ctx, a, s.UserContext, s.Username)
}
