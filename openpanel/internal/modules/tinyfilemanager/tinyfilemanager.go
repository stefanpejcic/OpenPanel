// Package tinyfilemanager installs and manages TinyFileManager (github.com/prasathmani/tinyfilemanager) inside an existing domain's docroot and php-fpm container - same shape as internal/modules/tinyphotogallery, but with one extra concern: it's a single PHP file that carries its own auth as a $auth_users = array(...) literal, so install has to bake the admin username/password into that array (see install.go). No database, no CLI installer, no tagged releases upstream (always installs current master), no update mechanism, no maintenance mode, no auto-login
// install downloads tinyfilemanager.php from the master branch, hashes the admin password with password_hash() run inside the target php-fpm container, then rewrites the default sample $auth_users array to contain only the one admin account provided
// no maintenance mode, auto-login, cache-clear, version tracking, or clone here - version recorded in the sites table is a static "latest" placeholder, and only install/remove/backup-restore/manage are implemented
package tinyfilemanager

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var cms = cmsapp.New(cmsapp.App{
	Install:       handleInstallStream,
	InstallFields: []string{"domain_id", "subdirectory", "admin_username", "admin_password"},
	Slug:          "tinyfilemanager",
	Name:          "TinyFileManager",
	DBMode:        cmsapp.DBNone,
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
