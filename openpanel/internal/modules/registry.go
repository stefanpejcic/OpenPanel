// Package modules resolves openpanel.config's enabled_modules list to Go
// packages: every module is a normal compiled-in Go package exposing a
// Register(mux, app) function, and a small static map resolves config
// names to those functions.
package modules

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/account"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/api"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/appinstall"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/autoinstaller"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/backups"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/backupwizard"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dashboard"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/diskusage"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dns"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/domains"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/drupal"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dynamicdns"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/emails"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/filemanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/fixpermissions"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/flarum"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/ftp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/goaccess"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/inodes"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/ipblocker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/joomla"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/malwarescan"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/matomo"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/mediawiki"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/moodle"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/mysql"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/nextcloud"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/nodejs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/opencart"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/phpapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/plugins"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/postgresql"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/prestashop"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/processmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/python"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/ruby"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/search"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/serverinfo"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/services"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/sofawiki"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/trash"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/waf"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/webserverconf"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/wordpress"
)

// Registrar wires one feature module's routes onto mux.
type Registrar func(mux *http.ServeMux, a *appctx.App)

// alwaysOn is the set of routes that register regardless of
// enabled_modules (core login/dashboard/search/etc.), unlike everything
// dispatched through the for-loop below.
var alwaysOn = []Registrar{
	account.Register,              // login, login_autologin, logout
	account.RegisterPasswordReset, // gated internally on the password_reset config
	account.RegisterAPILogin,      // /api/login, no additional API-key gate
	api.Register,                  // /api/endpoints, gated internally via apiregistry.Handle
	dashboard.Register,
	dashboard.RegisterAPI,             // gated internally via apiregistry.Handle
	serverinfo.RegisterHostingJSON,    // core helper endpoint, always available
	serverinfo.RegisterHostingAPI,     // API twin of RegisterHostingJSON, always available
	appinstall.RegisterShared,         // docker tags + check_if_file_exists helpers
	appinstall.RegisterSharedAPI,      // API twin of RegisterShared, gated internally via apiregistry.Handle
	appinstall.RegisterPM2,            // pm2 logs/action/delete
	appinstall.RegisterPM2API,         // gated internally via apiregistry.Handle ("pm2" feature)
	appinstall.RegisterPM2Clone,       // pm2 clone (copy files + settings onto a new domain)
	filemanager.RegisterDirectorySize, // get_folder_size helper
	search.Register,                   // unconditional, gated per-what internally
	search.RegisterAPI,                // API twin of search.Register, gated internally via apiregistry.Handle
	websites.Register,                 // part of mainModules = ["dashboard", "websites"] (see internal/app)
	websites.RegisterSitesAPI,         // gated internally via apiregistry.Handle
	plugins.Register,                  // only registers a route if plugin_names is non-empty, gated internally
	plugins.RegisterAPI,               // same guard, API twin
}

// configured maps openpanel.config's enabled_modules entries to their Go
// package's Register function - every module that exists at all is
// already compiled in, so this is just a name-to-function lookup.
//
// Only modules that have been ported so far appear here. A name present in
// enabled_modules but missing from this map is silently skipped - the
// expected, normal state during an incremental port, not a broken install.
var configured = map[string]Registrar{
	// mysql and the rest land here as their phases are ported.
	"docker": func(mux *http.ServeMux, a *appctx.App) {
		docker.Register(mux, a)
		docker.RegisterAPI(mux, a)
		docker.RegisterContainerManageAPI(mux, a)
	},
	"terminal": docker.RegisterTerminal,
	"change_image": func(mux *http.ServeMux, a *appctx.App) {
		docker.RegisterChangeImage(mux, a)
		docker.RegisterChangeImageAPI(mux, a)
	},
	"change_ws": func(mux *http.ServeMux, a *appctx.App) {
		docker.RegisterChangeWS(mux, a)
		docker.RegisterChangeWSAPI(mux, a)
	},
	"change_db": func(mux *http.ServeMux, a *appctx.App) {
		docker.RegisterChangeDB(mux, a)
		docker.RegisterChangeDBAPI(mux, a)
	},
	"services":    func(mux *http.ServeMux, a *appctx.App) { services.Register(mux, a); services.RegisterAPI(mux, a) },
	"filemanager": func(mux *http.ServeMux, a *appctx.App) { filemanager.Register(mux, a); filemanager.RegisterAPI(mux, a) },
	"disk_usage":  func(mux *http.ServeMux, a *appctx.App) { diskusage.Register(mux, a); diskusage.RegisterAPI(mux, a) },
	"inodes":      func(mux *http.ServeMux, a *appctx.App) { inodes.Register(mux, a); inodes.RegisterAPI(mux, a) },
	"fix_permissions": func(mux *http.ServeMux, a *appctx.App) {
		fixpermissions.Register(mux, a)
		fixpermissions.RegisterAPI(mux, a)
	},
	"malware_scan": func(mux *http.ServeMux, a *appctx.App) { malwarescan.Register(mux, a); malwarescan.RegisterAPI(mux, a) },
	"trash":        func(mux *http.ServeMux, a *appctx.App) { trash.Register(mux, a); trash.RegisterAPI(mux, a) },
	"ftp":          func(mux *http.ServeMux, a *appctx.App) { ftp.Register(mux, a); ftp.RegisterAPI(mux, a) },
	"backup_wizard": func(mux *http.ServeMux, a *appctx.App) {
		backupwizard.Register(mux, a)
		backupwizard.RegisterAPI(mux, a)
	},
	"backups":    func(mux *http.ServeMux, a *appctx.App) { backups.Register(mux, a); backups.RegisterAPI(mux, a) },
	"domains":    func(mux *http.ServeMux, a *appctx.App) { domains.Register(mux, a); domains.RegisterAPI(mux, a) },
	"drupal":     func(mux *http.ServeMux, a *appctx.App) { drupal.Register(mux, a); drupal.RegisterAPI(mux, a) },
	"flarum":     func(mux *http.ServeMux, a *appctx.App) { flarum.Register(mux, a); flarum.RegisterAPI(mux, a) },
	"sofawiki":   func(mux *http.ServeMux, a *appctx.App) { sofawiki.Register(mux, a); sofawiki.RegisterAPI(mux, a) },
	"joomla":     func(mux *http.ServeMux, a *appctx.App) { joomla.Register(mux, a); joomla.RegisterAPI(mux, a) },
	"opencart":   func(mux *http.ServeMux, a *appctx.App) { opencart.Register(mux, a); opencart.RegisterAPI(mux, a) },
	"nextcloud":  func(mux *http.ServeMux, a *appctx.App) { nextcloud.Register(mux, a); nextcloud.RegisterAPI(mux, a) },
	"prestashop": func(mux *http.ServeMux, a *appctx.App) { prestashop.Register(mux, a); prestashop.RegisterAPI(mux, a) },
	"matomo":     func(mux *http.ServeMux, a *appctx.App) { matomo.Register(mux, a); matomo.RegisterAPI(mux, a) },
	"moodle":     func(mux *http.ServeMux, a *appctx.App) { moodle.Register(mux, a); moodle.RegisterAPI(mux, a) },
	"mediawiki":  func(mux *http.ServeMux, a *appctx.App) { mediawiki.Register(mux, a); mediawiki.RegisterAPI(mux, a) },
	"goaccess":   func(mux *http.ServeMux, a *appctx.App) { goaccess.Register(mux, a); goaccess.RegisterAPI(mux, a) },
	"php": func(mux *http.ServeMux, a *appctx.App) {
		php.Register(mux, a)
		php.RegisterOptionsAPI(mux, a)
		php.RegisterIniAPI(mux, a)
		php.RegisterExtensionsAPI(mux, a)
		php.RegisterDefaultAPI(mux, a)
		php.RegisterDomainsAPI(mux, a)
		phpapp.Register(mux, a)
		phpapp.RegisterAPI(mux, a)
	},
	"dns":         func(mux *http.ServeMux, a *appctx.App) { dns.Register(mux, a); dns.RegisterAPI(mux, a) },
	"dynamic_dns": func(mux *http.ServeMux, a *appctx.App) { dynamicdns.Register(mux, a); dynamicdns.RegisterAPI(mux, a) },
	"redis":       func(mux *http.ServeMux, a *appctx.App) { cache.RegisterRedis(mux, a); cache.RegisterRedisAPI(mux, a) },
	"memcached": func(mux *http.ServeMux, a *appctx.App) {
		cache.RegisterMemcached(mux, a)
		cache.RegisterMemcachedAPI(mux, a)
	},
	"elasticsearch": func(mux *http.ServeMux, a *appctx.App) {
		cache.RegisterElasticsearch(mux, a)
		cache.RegisterElasticsearchAPI(mux, a)
	},
	"opensearch": func(mux *http.ServeMux, a *appctx.App) {
		cache.RegisterOpensearch(mux, a)
		cache.RegisterOpensearchAPI(mux, a)
	},
	"valkey": func(mux *http.ServeMux, a *appctx.App) { cache.RegisterValkey(mux, a); cache.RegisterValkeyAPI(mux, a) },
	"varnish": func(mux *http.ServeMux, a *appctx.App) {
		cache.RegisterVarnish(mux, a)
		cache.RegisterVarnishAPI(mux, a)
	},
	"mysql":               func(mux *http.ServeMux, a *appctx.App) { mysql.Register(mux, a); mysql.RegisterAPI(mux, a) },
	"mysql_conf":          mysql.RegisterConf,
	"mysql_import":        mysql.RegisterImport,
	"mysql_processlist":   mysql.RegisterProcesslist,
	"mysql_root_password": mysql.RegisterRootPassword,
	"remote_mysql":        mysql.RegisterRemote,
	"emails": func(mux *http.ServeMux, a *appctx.App) {
		emails.RegisterAccounts(mux, a)
		emails.RegisterEmailsAPI(mux, a)
	},
	"email_aliases":        emails.RegisterAliases,
	"email_default":        emails.RegisterDefault,
	"email_deliverability": emails.RegisterDeliverability,
	"email_export": func(mux *http.ServeMux, a *appctx.App) {
		emails.RegisterExport(mux, a)
		emails.RegisterEmailExportAPI(mux, a)
	},
	"email_filters": emails.RegisterFilters,
	"email_import": func(mux *http.ServeMux, a *appctx.App) {
		emails.RegisterImport(mux, a)
		emails.RegisterEmailImportAPI(mux, a)
	},
	"webmail": func(mux *http.ServeMux, a *appctx.App) {
		emails.RegisterWebmail(mux, a)
		emails.RegisterWebmailAPI(mux, a)
	},
	"crons": func(mux *http.ServeMux, a *appctx.App) { crons.Register(mux, a); crons.RegisterAPI(mux, a) },
	"info": func(mux *http.ServeMux, a *appctx.App) {
		serverinfo.RegisterInfo(mux, a)
		serverinfo.RegisterInfoAPI(mux, a)
	},
	"usage": func(mux *http.ServeMux, a *appctx.App) {
		serverinfo.RegisterUsage(mux, a)
		serverinfo.RegisterUsageAPI(mux, a)
	},
	"process_manager": func(mux *http.ServeMux, a *appctx.App) {
		processmanager.Register(mux, a)
		processmanager.RegisterAPI(mux, a)
	},
	"ip_blocker": func(mux *http.ServeMux, a *appctx.App) { ipblocker.Register(mux, a); ipblocker.RegisterAPI(mux, a) },
	"webserver_conf": func(mux *http.ServeMux, a *appctx.App) {
		webserverconf.Register(mux, a)
		webserverconf.RegisterAPI(mux, a)
	},
	"waf": func(mux *http.ServeMux, a *appctx.App) { waf.Register(mux, a); waf.RegisterAPI(mux, a) },
	"account": func(mux *http.ServeMux, a *appctx.App) {
		account.RegisterSettings(mux, a)
		account.RegisterAccountAPI(mux, a)
	},
	"locale": func(mux *http.ServeMux, a *appctx.App) {
		account.RegisterLocale(mux, a)
		account.RegisterLocaleAPI(mux, a)
	},
	"twofa": func(mux *http.ServeMux, a *appctx.App) {
		account.RegisterTwofa(mux, a)
		account.RegisterTwofaAPI(mux, a)
	},
	"passkeys": func(mux *http.ServeMux, a *appctx.App) {
		account.RegisterPasskeys(mux, a)
		account.RegisterPasskeysAPI(mux, a)
	},
	"notifications": func(mux *http.ServeMux, a *appctx.App) {
		account.RegisterNotifications(mux, a)
		account.RegisterNotificationsAPI(mux, a)
	},
	"favorites": func(mux *http.ServeMux, a *appctx.App) {
		account.RegisterFavorites(mux, a)
		account.RegisterFavoritesAPI(mux, a)
	},
	"sessions": account.RegisterSessions,
	"activity": func(mux *http.ServeMux, a *appctx.App) {
		account.RegisterActivity(mux, a)
		account.RegisterActivityAPI(mux, a)
	},
	"login_history": func(mux *http.ServeMux, a *appctx.App) {
		account.RegisterLoginHistory(mux, a)
		account.RegisterLoginHistoryAPI(mux, a)
	},
	"mcp": func(mux *http.ServeMux, a *appctx.App) {
		account.RegisterMCP(mux, a)
		account.RegisterMCPEndpoint(mux, a)
		account.RegisterMCPAPI(mux, a)
	},
	"api":               account.RegisterAPIDocs,
	"postgresql":        func(mux *http.ServeMux, a *appctx.App) { postgresql.Register(mux, a); postgresql.RegisterAPI(mux, a) },
	"postgresql_conf":   postgresql.RegisterConf,
	"postgresql_import": postgresql.RegisterImport,
	"remote_postgresql": postgresql.RegisterRemote,
	"python":            func(mux *http.ServeMux, a *appctx.App) { python.Register(mux, a); python.RegisterAPI(mux, a) },
	"nodejs":            func(mux *http.ServeMux, a *appctx.App) { nodejs.Register(mux, a); nodejs.RegisterAPI(mux, a) },
	"ruby":              func(mux *http.ServeMux, a *appctx.App) { ruby.Register(mux, a); ruby.RegisterAPI(mux, a) },
	"autoinstaller": func(mux *http.ServeMux, a *appctx.App) {
		autoinstaller.Register(mux, a)
		autoinstaller.RegisterAPI(mux, a)
	},
	"wordpress": func(mux *http.ServeMux, a *appctx.App) { wordpress.Register(mux, a); wordpress.RegisterAPI(mux, a) },
	"website_builder": func(mux *http.ServeMux, a *appctx.App) {
		websites.RegisterWebsiteBuilder(mux, a)
		websites.RegisterWebsiteBuilderAPI(mux, a)
	},
}

// RegisterAll wires up the always-on modules plus every configured module
// that has a Go implementation yet.
func RegisterAll(mux *http.ServeMux, a *appctx.App) {
	for _, reg := range alwaysOn {
		reg(mux, a)
	}

	for _, name := range a.EnabledModules {
		reg, ok := configured[name]
		if !ok {
			continue
		}
		reg(mux, a)
		log.Printf("APP - Registered module: %s", name)
	}
}
