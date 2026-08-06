// Package modules is the Go replacement for app.py's importlib-based
// module loader: instead of dynamically importing modules.<name> based on
// openpanel.config's enabled_modules list, every module is a normal
// compiled-in Go package exposing a Register(mux, app) function, and a
// small static map resolves config names to those functions.
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
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dynamicdns"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/emails"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/filemanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/fixpermissions"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/ftp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/goaccess"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/inodes"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/ipblocker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/malwarescan"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/mysql"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/nodejs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/plugins"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/postgresql"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/processmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/python"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/search"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/serverinfo"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/services"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/trash"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/waf"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/webserverconf"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/wordpress"
)

// Registrar wires one feature module's routes onto mux.
type Registrar func(mux *http.ServeMux, a *appctx.App)

// alwaysOn mirrors app.py's unconditional top-of-file imports (`from
// modules import dashboard, websites` / `from modules.account import
// login, logout` / ...): these register regardless of enabled_modules,
// unlike everything dispatched through the for-loop below.
var alwaysOn = []Registrar{
	account.Register,              // login, login_autologin, logout
	account.RegisterPasswordReset, // modules/account/forgot_password.py: conditionally imported on password_reset config, gated internally
	account.RegisterAPILogin,      // modules/api_core.py: /api/login, no api_required gate in Python either
	api.Register,                  // modules/api_core.py: /api/endpoints, gated internally via apiregistry.Handle
	dashboard.Register,
	serverinfo.RegisterHostingJSON,    // modules/json/helpers.py: unconditional top-level import
	appinstall.RegisterShared,         // modules/json/helpers.py: docker tags + check_if_file_exists
	appinstall.RegisterPM2,            // modules/json/helpers.py: pm2 logs/action/delete
	appinstall.RegisterPM2API,         // modules/api/pm2.py: gated internally via apiregistry.Handle ("pm2" feature)
	filemanager.RegisterDirectorySize, // modules/json/helpers.py: get_folder_size
	search.Register,                   // modules/core/search.py: unconditional, gated per-what internally
	websites.Register,                 // modules/websites.py: app.py's main_modules = ["dashboard", "websites"]
	websites.RegisterSitesAPI,         // modules/api/websites.py: gated internally via apiregistry.Handle
	plugins.Register,                  // app.py: `if plugin_names: @app.route('/plugins') ...`, gated internally
}

// configured maps openpanel.config's enabled_modules entries to their Go
// package's Register function, replacing app.py's category dispatch table
// (modules.cache.<name>, modules.files.<name>, modules.account.<name>,
// ...) - Python needs that indirection to resolve a string to an import
// path; Go doesn't, since every module that exists at all is already
// compiled in.
//
// Only modules that have been ported so far appear here. A name present in
// enabled_modules but missing from this map is silently skipped, which -
// unlike Python's ImportError for the same situation - is the expected,
// normal state during an incremental port, not a broken install.
var configured = map[string]Registrar{
	// mysql and the rest land here as their phases are ported.
	"docker":          func(mux *http.ServeMux, a *appctx.App) { docker.Register(mux, a); docker.RegisterAPI(mux, a) },
	"services":        func(mux *http.ServeMux, a *appctx.App) { services.Register(mux, a); services.RegisterAPI(mux, a) },
	"filemanager":     filemanager.Register,
	"disk_usage":      func(mux *http.ServeMux, a *appctx.App) { diskusage.Register(mux, a); diskusage.RegisterAPI(mux, a) },
	"inodes":          func(mux *http.ServeMux, a *appctx.App) { inodes.Register(mux, a); inodes.RegisterAPI(mux, a) },
	"fix_permissions": fixpermissions.Register,
	"malware_scan":    func(mux *http.ServeMux, a *appctx.App) { malwarescan.Register(mux, a); malwarescan.RegisterAPI(mux, a) },
	"trash":           trash.Register,
	"ftp":             func(mux *http.ServeMux, a *appctx.App) { ftp.Register(mux, a); ftp.RegisterAPI(mux, a) },
	"backup_wizard":   backupwizard.Register,
	"backups":         backups.Register,
	"domains":         func(mux *http.ServeMux, a *appctx.App) { domains.Register(mux, a); domains.RegisterAPI(mux, a) },
	"goaccess":        func(mux *http.ServeMux, a *appctx.App) { goaccess.Register(mux, a); goaccess.RegisterAPI(mux, a) },
	"php": func(mux *http.ServeMux, a *appctx.App) {
		php.Register(mux, a)
		php.RegisterOptionsAPI(mux, a)
		php.RegisterIniAPI(mux, a)
		php.RegisterExtensionsAPI(mux, a)
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
	"email_export":         emails.RegisterExport,
	"email_filters":        emails.RegisterFilters,
	"email_import":         emails.RegisterImport,
	"webmail": func(mux *http.ServeMux, a *appctx.App) {
		emails.RegisterWebmail(mux, a)
		emails.RegisterWebmailAPI(mux, a)
	},
	"crons": func(mux *http.ServeMux, a *appctx.App) { crons.Register(mux, a); crons.RegisterAPI(mux, a) },
	"info":  serverinfo.RegisterInfo,
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
	"locale":        account.RegisterLocale,
	"twofa":         account.RegisterTwofa,
	"passkeys":      account.RegisterPasskeys,
	"notifications": account.RegisterNotifications,
	"favorites":     account.RegisterFavorites,
	"sessions":      account.RegisterSessions,
	"activity":      account.RegisterActivity,
	"login_history": account.RegisterLoginHistory,
	"mcp": func(mux *http.ServeMux, a *appctx.App) {
		account.RegisterMCP(mux, a)
		account.RegisterMCPEndpoint(mux, a)
	},
	"api":               account.RegisterAPIDocs,
	"postgresql":        func(mux *http.ServeMux, a *appctx.App) { postgresql.Register(mux, a); postgresql.RegisterAPI(mux, a) },
	"postgresql_conf":   postgresql.RegisterConf,
	"postgresql_import": postgresql.RegisterImport,
	"remote_postgresql": postgresql.RegisterRemote,
	"python":            python.Register,
	"nodejs":            nodejs.Register,
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
