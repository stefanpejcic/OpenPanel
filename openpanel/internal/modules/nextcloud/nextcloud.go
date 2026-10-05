// Package nextcloud installs and manages a Nextcloud site via occ maintenance:install inside an existing domain's docroot and php-fpm container - same shape as internal/modules/opencart, MySQL/MariaDB only
package nextcloud

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var cms = cmsapp.New(cmsapp.App{
	Install:        handleInstallStream,
	InstallFields:  []string{"domain_id", "subdirectory", "nextcloud_version", "admin_username", "admin_password", "admin_email", "db_name", "db_user", "db_password"},
	Login:          handleNextcloudLogin,
	Cache:          handleNextcloudCacheClean,
	Logs:           cmsapp.ShellLogs(`tail -n 300 "$1/data/nextcloud.log" 2>/dev/null`, "No log entries yet."),
	Versions:       handleNextcloudVersions,
	Maintenance:    handleNextcloudMaintenance,
	CloneDBPrefix:  "nc_clone_",
	CloneConfig:    cloneConfig,
	Update:         handleNextcloudUpdate,
	CanClone:       true,
	APICloneForm:   apiCloneForm,
	APIUpdate:      cmsapp.SiteQuery("nextcloud", handleNextcloudUpdate),
	APICache:       cmsapp.SiteQuery("nextcloud", handleNextcloudCacheClean),
	Slug:           "nextcloud",
	Name:           "Nextcloud",
	RemoveConfig:   "config/config.php",
	RemoveDBNameRE: removeDBNameRE,
	RemoveDBUserRE: removeDBUserRE,
	DBMode:         cmsapp.DBPrefixed,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return extractNextcloudDatabaseInfoForLogin(userContext, docroot)
	},
	ConfigFile: "config/config.php",
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
