// Package tinyphotogallery installs and manages TinyPhotoGallery (github.com/stefanpejcic/tinyphotogallery) inside an existing domain's docroot and php-fpm container - same shape as internal/modules/sofawiki, but even simpler: a single PHP file (index.php) plus an empty "photos/" folder next to it. No database, no admin account, no CLI installer, no versioning, no update mechanism
// install is a plain curl download of index.php from the main branch followed by `mkdir photos` - PHP 8.1+ with GD is what the README asks for, but nothing here enforces it
// no maintenance mode, auto-login, or cache-clear here since TinyPhotoGallery has no such concepts, and version recorded in the sites table is a static "main" placeholder since there are no tagged releases
package tinyphotogallery

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var cms = cmsapp.New(cmsapp.App{
	Install:       handleInstallStream,
	InstallFields: []string{"domain_id", "subdirectory"},
	Slug:          "tinyphotogallery",
	Name:          "TinyPhotoGallery",
	DBMode:        cmsapp.DBNone,
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
