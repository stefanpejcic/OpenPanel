// Package matomo installs and manages a Matomo (self-hosted web analytics) site downloaded from GitHub releases inside an existing domain's docroot, same shape as internal/modules/opencart, nextcloud and prestashop - unlike those, Matomo ships no non-interactive CLI installer, so install.go instead drives Matomo's own browser installation wizard (plugins/Installation/Controller.php) as a sequence of plain HTTP requests matching what a real browser would submit (DB setup -> table creation -> superuser -> first site -> finish -> login) - no maintenance-mode toggle exists here since Matomo has no offline-mode primitive, but backups.go is included (pure DB dump + file tar, no CMS-specific dependency beyond DB name/prefix)
package matomo

import (
	"context"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var cms = cmsapp.New(cmsapp.App{
	Install:        handleInstallStream,
	InstallFields:  []string{"domain_id", "subdirectory", "matomo_version", "admin_login", "admin_password", "admin_email", "db_name", "db_user", "db_password"},
	Login:          handleMatomoLogin,
	Cache:          handleMatomoCacheClean,
	Logs:           cmsapp.ShellLogs(`f=$(ls -t "$1"/tmp/logs/*.log 2>/dev/null | head -1); [ -n "$f" ] && tail -n 300 "$f"`, "No log entries yet."),
	Versions:       handleMatomoVersions,
	CloneDBPrefix:  "matomo_clone_",
	CloneConfig:    cloneConfig,
	Update:         handleMatomoUpdate,
	CanClone:       true,
	APICloneForm:   apiCloneForm,
	APIUpdate:      cmsapp.SiteQuery("matomo", handleMatomoUpdate),
	APICache:       cmsapp.SiteQuery("matomo", handleMatomoCacheClean),
	Slug:           "matomo",
	Name:           "Matomo",
	RemoveConfig:   "config/config.ini.php",
	RemoveDBNameRE: removeDBNameRE,
	RemoveDBUserRE: removeDBUserRE,
	OnRemove: func(ctx context.Context, s *cmsapp.Site) {
		removeMatomoCredentials(s.Name)
	},
	DBMode: cmsapp.DBPrefixed,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return extractMatomoDatabaseInfoForBackup(userContext, docroot)
	},
	ConfigFile: "config/config.ini.php",
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
