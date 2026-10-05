// Package joomla installs and manages a Joomla site (downloaded from GitHub releases + Joomla's own CLI installer, `installation/joomla.php install`) inside an existing domain's docroot, same shape as internal/modules/drupal - deliberately minimal matching drupal's scope: just install, a small read-only manage/overview page, a Logs tab, cache clearing, a one-time admin login link, and uninstall, MySQL/MariaDB only
package joomla

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var cms = cmsapp.New(cmsapp.App{
	Install:        handleInstallStream,
	InstallFields:  []string{"domain_id", "subdirectory", "site_name", "joomla_version", "admin_name", "admin_username", "admin_password", "admin_email", "db_name", "db_user", "db_password"},
	Login:          handleJoomlaLogin,
	Cache:          handleJoomlaCacheClean,
	Logs:           cmsapp.ShellLogs(`cd "$1/administrator/logs" 2>/dev/null && for f in *.php; do [ -e "$f" ] || continue; echo "=== $f ==="; tail -n 200 "$f" | grep -v "^<?php die"; done`, "No log entries yet - Joomla only writes here when a PHP warning/error occurs."),
	Maintenance:    handleJoomlaMaintenance,
	CloneDBPrefix:  "joomla_clone_",
	CloneConfig:    cloneConfig,
	CanClone:       true,
	APICloneForm:   apiCloneForm,
	APICache:       cmsapp.SiteQueryPost("joomla", handleJoomlaCacheClean),
	APIClonePath:   "POST /api/joomla/clone",
	Slug:           "joomla",
	Name:           "Joomla",
	RemoveConfig:   "configuration.php",
	RemoveDBNameRE: removeDBNameRE,
	RemoveDBUserRE: removeDBUserRE,
	DBMode:         cmsapp.DBPrefixed,
	DBInfo: func(userContext, docroot, selectedDomain string) map[string]string {
		return extractJoomlaDatabaseInfoForLogin(userContext, docroot)
	},
	ConfigFile: "configuration.php",
})

func Register(mux *http.ServeMux, a *appctx.App)    { cms.Register(mux, a) }
func RegisterAPI(mux *http.ServeMux, a *appctx.App) { cms.RegisterAPI(mux, a) }
