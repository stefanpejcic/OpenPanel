// Package prestashop installs and manages a PrestaShop site via install/index_cli.php inside an existing domain's docroot and php-fpm container - same shape as internal/modules/opencart and internal/modules/nextcloud, MySQL/MariaDB only
package prestashop

import (
	"net/http"
	"os"
	"path/filepath"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

// findAdminDir locates the single admin*/ directory under a PrestaShop install - install.go renames it to a random name right after install, but every later operation (login link generation, uninstall) still needs to resolve whatever that current name is, in case it was renamed again by hand
func findAdminDir(hostOSPath string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(hostOSPath, "admin*"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", os.ErrNotExist
	}
	return filepath.Base(matches[0]), nil
}

var cms = cmsapp.New(cmsapp.App{
	Install:          handleInstallStream,
	InstallFields:    []string{"domain_id", "subdirectory", "prestashop_version", "admin_firstname", "admin_lastname", "admin_password", "admin_email", "db_name", "db_user", "db_password"},
	Login:            handlePrestashopLogin,
	Cache:            handlePrestashopCacheClean,
	Logs:             cmsapp.ShellLogs(`f=$(ls -t "$1"/var/logs/*.log 2>/dev/null | head -1); [ -n "$f" ] && tail -n 300 "$f"`, "No log entries yet."),
	Versions:         handlePrestashopVersions,
	Maintenance:      handlePrestashopMaintenance,
	CloneDBPrefix:    "presta_clone_",
	CloneAfterCopy:   cloneClearCache,
	CloneConfig:      cloneConfig,
	CloneAfterConfig: cloneShopURL,
	CanClone:         true,
	APICloneForm:     apiCloneForm,
	APICache:         cmsapp.SiteQueryPost("prestashop", handlePrestashopCacheClean),
	Slug:             "prestashop",
	Name:             "PrestaShop",
	RemoveConfig:     "app/config/parameters.php",
	RemoveDBNameRE:   removeDBNameRE,
	RemoveDBUserRE:   removeDBUserRE,
	DBMode:           cmsapp.DBPrefixed,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return extractPrestashopDatabaseInfoForLogin(userContext, docroot)
	},
	ConfigFile: "app/config/parameters.php",
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
