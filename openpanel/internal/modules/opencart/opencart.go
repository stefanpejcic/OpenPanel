// Package opencart installs and manages an OpenCart site via install/cli_install.php inside an existing domain's docroot and php-fpm container - same shape as internal/modules/joomla, MySQL/MariaDB only
package opencart

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var cms = cmsapp.New(cmsapp.App{
	Install:        handleInstallStream,
	InstallFields:  []string{"domain_id", "subdirectory", "opencart_version", "admin_username", "admin_password", "admin_email", "db_name", "db_user", "db_password"},
	Login:          handleOpenCartLogin,
	Cache:          handleOpenCartCacheClean,
	Logs:           cmsapp.ShellLogs(`tail -n 300 "$1/system/storage/logs/error.log" 2>/dev/null`, "No log entries yet - OpenCart only writes here when a PHP warning/error occurs."),
	Maintenance:    handleOpenCartMaintenance,
	CloneDBPrefix:  "oc_clone_",
	CloneConfig:    cloneConfig,
	CanClone:       true,
	APICloneForm:   apiCloneForm,
	APICache:       cmsapp.SiteQueryPost("opencart", handleOpenCartCacheClean),
	Slug:           "opencart",
	Name:           "OpenCart",
	RemoveConfig:   "config.php",
	RemoveDBNameRE: removeDBNameRE,
	RemoveDBUserRE: removeDBUserRE,
	DBMode:         cmsapp.DBPrefixed,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return opencartDBInfoForBackup(userContext, docroot)
	},
	ConfigFile: "config.php",
})

var backupDBPrefixRE = regexp.MustCompile(`DB_PREFIX'\s*,\s*'([^']*)'`)

// opencartDBInfoForBackup extends extractOpenCartDatabaseInfoForLogin (database_name only) with the table prefix backup/restore also need
func opencartDBInfoForBackup(userContext, docroot string) map[string]string {
	info := extractOpenCartDatabaseInfoForLogin(userContext, docroot)
	if info["error"] != "" {
		return info
	}
	const wwwPrefix = "/var/www/html/"
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "config.php"))
	if err != nil {
		return info
	}
	if m := backupDBPrefixRE.FindStringSubmatch(string(content)); m != nil {
		info["database_prefix"] = m[1]
	}
	return info
}

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
