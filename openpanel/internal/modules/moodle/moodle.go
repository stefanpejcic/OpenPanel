// Package moodle installs and manages a Moodle LMS site (downloaded from download.moodle.org's packaged tarballs + Moodle's own genuine non-interactive `admin/cli/install.php` CLI installer) inside an existing domain's docroot, same shape as internal/modules/prestashop and matomo - Moodle 5.x restructured its tree so the actual web-served files live under a `public/` subdirectory, with config.php, admin/, lib/, etc. one level above it outside the web root by design (public/config.php is a tiny shim doing `require __DIR__.'/../config.php'`), so this module extracts the full release into a sibling "app root" directory next to the domain's docroot, then replaces the otherwise-empty docroot with a symlink to <approot>/public - this is the standard supported deployment shape for Moodle 5.x, not an OpenPanel-specific workaround; and since Moodle also requires a periodic `admin/cli/cron.php` run (no emails, enrolments, or scheduled tasks without it), install.go registers a per-minute job via internal/modules/crons.AddJob, and manage.go's uninstall handler removes it via crons.RemoveJobByComment
package moodle

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
)

// moodleCronComment is the crons.ini comment used both by install.go's crons.AddJob call and manage.go's crons.RemoveJobByComment call - it must be derived identically in both places to actually find/remove the same job later
func moodleCronComment(selectedDomain string) string {
	return "moodle-" + appkit.SiteSlug(selectedDomain)
}

var cms = cmsapp.New(cmsapp.App{
	Install:        handleInstallStream,
	InstallFields:  []string{"domain_id", "subdirectory", "moodle_version", "site_name", "site_shortname", "admin_username", "admin_password", "admin_email", "db_name", "db_user", "db_password"},
	Cache:          handleMoodleCacheClean,
	Logs:           handleMoodleLogs,
	Maintenance:    handleMoodleMaintenance,
	Clone:          handleMoodleClone,
	Update:         handleMoodleUpdate,
	APIClone:       apiMoodleClone,
	APIUpdate:      cmsapp.SiteQuery("moodle", handleMoodleUpdate),
	APICache:       cmsapp.SiteQuery("moodle", handleMoodleCacheClean),
	Slug:           "moodle",
	Name:           "Moodle",
	RemoveConfig:   "config.php",
	RemoveDBNameRE: removeDBNameRE,
	RemoveDBUserRE: removeDBUserRE,
	RemoveConfigPath: func(s *cmsapp.Site) string {
		return filepath.Join(siteHTMLVolume(s.UserContext), appkit.SiteSlug(s.Name)+"_moodleapp", "config.php")
	},
	RemoveFiles: func(ctx context.Context, s *cmsapp.Site) {
		_ = crons.RemoveJobByComment(ctx, s.UserContext, moodleCronComment(s.Name))
		slug := appkit.SiteSlug(s.Name)
		_ = os.Remove(s.HostPath)
		_ = os.RemoveAll(filepath.Join(siteHTMLVolume(s.UserContext), slug+"_moodleapp"))
		_ = os.RemoveAll(filepath.Join(siteHTMLVolume(s.UserContext), slug+"_moodledata"))
	},
	DBMode: cmsapp.DBPrefixed,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return extractMoodleDatabaseInfoForBackup(userContext, selectedDomain)
	},
	ConfigFile: "config.php",
	FilesDir: func(selectedDomain string) string {
		return "/var/www/html/" + appkit.SiteSlug(selectedDomain) + "_moodledata"
	},
	FilesLabel: "files (moodledata)",
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }

func siteHTMLVolume(userContext string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
}
