// Package sofawiki installs and manages SofaWiki (github.com/bellenuit/sofawiki) inside an existing domain's docroot and php-fpm container - same shape as internal/modules/drupal and internal/modules/flarum, but simpler since SofaWiki needs no database at all
// install is a plain download-and-extract of the master branch (no tagged releases, no composer.json): no database, no admin account, no CLI installer - SofaWiki's own 4-step browser setup wizard is left for the site owner to complete on first visit, same as if they'd uploaded the files by FTP
// PHP 8.0+ throws a fatal error in inc/async.php (fwrite() on a failed fsockopen() returning false instead of a resource), so install.go refuses to install onto a domain configured for PHP 8.0+
// no maintenance mode, admin auto-login, cache-clear, or version tracking here since SofaWiki has no such concepts and no tagged releases to compare against
package sofawiki

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var cms = cmsapp.New(cmsapp.App{
	Install:       handleInstallStream,
	InstallFields: []string{"domain_id", "subdirectory", "admin_email"},
	CloneVersion:  func(*cmsapp.Clone) string { return "master" },
	CanClone:      true,
	APICloneForm:  apiCloneForm,
	Slug:          "sofawiki",
	Name:          "SofaWiki",
	DBMode:        cmsapp.DBNone,
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
