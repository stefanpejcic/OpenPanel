// Package ojs installs and manages Open Journal Systems (OJS, pkp/ojs) sites inside an existing domain's docroot and php-fpm container, same overall shape as internal/modules/moodle; OJS ships no ready-made GitHub release tarball since the repo uses git submodules a plain archive/zip silently omits, so releases are pulled from https://pkp.sfu.ca/ojs/download/ instead (see version.go); unlike Moodle's approot+public/ split, OJS's tarball root is itself the full web root, so this module still uses a sibling app-root dir + docroot symlink anyway, just to give update.go an atomic swap-and-rollback target, plus a separate "_ojsfiles" sibling dir for files_dir kept outside the web-accessible tree
package ojs

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

// ojsCronComment is the crons.ini comment shared by install.go's AddJob and manage.go's RemoveJobByComment calls - must be derived identically in both to find/remove the same job later
func ojsCronComment(selectedDomain string) string {
	return "ojs-" + appkit.SiteSlug(selectedDomain)
}

var cms = cmsapp.New(cmsapp.App{
	Install:       handleInstallStream,
	InstallFields: []string{"domain_id", "subdirectory", "ojs_version", "admin_username", "admin_password", "admin_email", "db_name", "db_user", "db_password"},
	Login:         handleOJSLogin,
	Cache:         handleOJSCacheClean,
	Logs:          handleOJSLogs,
	Clone:         handleOJSClone,
	Update:        handleOJSUpdate,
	APIClone:      apiOJSClone,
	APIUpdate:     cmsapp.SiteQuery("ojs", handleOJSUpdate),
	APICache:      cmsapp.SiteQuery("ojs", handleOJSCacheClean),
	Slug:          "ojs",
	Name:          "OJS",
	RemoveDB:      removeOJSDB,
	RemoveFiles: func(ctx context.Context, s *cmsapp.Site) {
		_ = crons.RemoveJobByComment(ctx, s.UserContext, ojsCronComment(s.Name))
		slug := appkit.SiteSlug(s.Name)
		_ = os.Remove(s.HostPath)
		_ = os.RemoveAll(filepath.Join(siteHTMLVolume(s.UserContext), slug+"_ojsapp"))
		_ = os.RemoveAll(filepath.Join(siteHTMLVolume(s.UserContext), slug+"_ojsfiles"))
	},
	DBMode: cmsapp.DBWhole,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return extractOJSDatabaseInfoForLogin(userContext, selectedDomain)
	},
	ConfigFile: "config.inc.php",
	FilesDir: func(selectedDomain string) string {
		return "/var/www/html/" + appkit.SiteSlug(selectedDomain) + "_ojsfiles"
	},
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }

func siteHTMLVolume(userContext string) string {
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"
}
