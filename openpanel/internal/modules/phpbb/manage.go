package phpbb

import (
	"net/http"
	"regexp"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

// phpbbDBNameRE extracts $dbname from config.php - same shape as drupal/websites.go's settings.php scrapes, needed here so remove can drop the right database
var phpbbDBNameRE = regexp.MustCompile(`\$dbname\s*=\s*'(.*?)';`)

var phpbbDBUserRE = regexp.MustCompile(`\$dbuser\s*=\s*'(.*?)';`)

// removePhpbbDB drops the database and its user if config.php can still be read, best-effort
func removePhpbbDB(a *appctx.App, w http.ResponseWriter, r *http.Request, s *cmsapp.Site) {
	ctx := r.Context()
	if out, catErr := podmanmanager.Command(ctx, s.UserContext, podmanmanager.PodmanArgv(s.UserContext, "exec", s.PHPContainer, "cat", s.InstallPath+"/config.php")).Output(); catErr == nil {
		if m := phpbbDBNameRE.FindSubmatch(out); m != nil {
			dbName := string(m[1])
			_, _ = mysqlmanager.Exec(ctx, s.UserContext, "DROP DATABASE IF EXISTS `"+dbName+"`", "")
		}
		if m := phpbbDBUserRE.FindSubmatch(out); m != nil {
			_, _ = mysqlmanager.Exec(ctx, s.UserContext, "DROP USER IF EXISTS '"+string(m[1])+"'@'%'", "")
		}
		appkit.InvalidateMySQLCaches(ctx, a, s.UserContext, s.Username)
	}
}
