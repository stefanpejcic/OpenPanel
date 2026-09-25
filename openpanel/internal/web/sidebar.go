package web

import (
	"html/template"
	"net/http"
	"strings"
)

// NavLink is one link entry in the sidebar navigation.
type NavLink struct {
	Href   string
	Label  string
	Active bool
	Target string // "_blank" or ""

	// Disabled marks a link for a feature the current plan doesn't grant but the configured upsell plan does - rendered greyed-out with an upgrade prompt instead of being omitted entirely.
	Disabled bool
}

// NavItem is one sidebar entry: a single link to a feature area, whose pages are tabbed via BuildPageTabs.
type NavItem struct {
	Label    string
	Icon     template.HTML
	MenuID   string
	Href     string
	Target   string
	Active   bool
	Disabled bool // plan doesn't grant any page of the area but the upsell plan does

	// Section is the heading rendered above this item, set only on the first item of a new sidebar section
	Section string
}

// NavPath derives BuildSidebarNav's path argument from a request: r.URL.Path, with "?method=download" appended for the file-manager upload page's download-from-URL variant - the one nav item whose active state depends on a query param rather than the path alone.
func NavPath(r *http.Request) string {
	path := r.URL.Path
	if path == "/file-manager/upload" && r.URL.Query().Get("method") == "download" {
		return path + "?method=download"
	}
	return path
}

func hasAnyPrefix(path string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// navGate holds the plan/upsell feature maps and hands out links gated on them
type navGate struct {
	allowed, upsellAllowed map[string]bool
}

func (g navGate) has(keys ...string) bool {
	for _, k := range keys {
		if g.allowed[k] || g.upsellAllowed[k] {
			return true
		}
	}
	return false
}

// add appends a link gated by a single feature key: shown active when the plan grants it, greyed-out (Disabled) when only the upsell plan would, omitted otherwise.
func (g navGate) add(links []NavLink, key, href, label string, active bool, target string) []NavLink {
	if g.allowed[key] {
		return append(links, NavLink{Href: href, Label: label, Active: active, Target: target})
	}
	if g.upsellAllowed[key] {
		return append(links, NavLink{Href: href, Label: label, Target: target, Disabled: true})
	}
	return links
}

// Page-area path matchers, each area's matcher excludes paths owned by a more specific area so a page belongs to exactly one
func isDomainsPath(path string) bool {
	return strings.HasPrefix(path, "/domain") && !isWebServerConfPath(path) && !isStatsPath(path)
}

func isEmailPath(path string) bool {
	return strings.HasPrefix(path, "/email")
}

func isFileManagerPath(path string) bool {
	return hasAnyPrefix(path, "/files", "/file-manager")
}

func isFilesPath(path string) bool {
	return isFileManagerPath(path) || hasAnyPrefix(path, "/ftp", "/fix-permissions")
}

func isSitesPath(path string) bool {
	return hasAnyPrefix(path, "/sites", "/wordpress") || (strings.HasPrefix(path, "/website") && !strings.HasPrefix(path, "/website-builder/install"))
}

func isAppInstallPath(path string) bool {
	return hasAnyPrefix(path, "/auto-installer", "/pm2", "/nodejs", "/python", "/ruby", "/java", "/n8n", "/website-builder/install",
		"/drupal/install", "/joomla/install", "/opencart/install", "/nextcloud/install",
		"/prestashop/install", "/matomo/install", "/moodle/install", "/mediawiki/install")
}

func isWebsitesPath(path string) bool {
	return isSitesPath(path) || isAppInstallPath(path) || hasAnyPrefix(path, "/drupal", "/joomla", "/opencart", "/nextcloud", "/prestashop", "/matomo", "/moodle", "/mediawiki")
}

func isMySQLPath(path string) bool {
	return strings.HasPrefix(path, "/mysql") || path == "/containers/mysql"
}

func isBackupsPath(path string) bool {
	return strings.HasPrefix(path, "/backup")
}

func isWebServerConfPath(path string) bool {
	return path == "/server/webserver_conf" || path == "/containers/webserver" || strings.HasPrefix(path, "/domains/vhosts")
}

func isContainersPath(path string) bool {
	return strings.HasPrefix(path, "/containers") && !isMySQLPath(path) && !isWebServerConfPath(path)
}

func isStatsPath(path string) bool {
	return hasAnyPrefix(path, "/server/usage", "/disk-usage", "/inodes-explorer", "/domains/stats", "/domains/log")
}

func isProcessesPath(path string) bool {
	return hasAnyPrefix(path, "/services", "/process-manager")
}

func isSecurityPath(path string) bool {
	return hasAnyPrefix(path, "/server/waf", "/security/ip-blocker", "/malware-scanner") || isAccountSecurityPath(path) || isAccountActivityPath(path)
}

func isAccountPath(path string) bool {
	return path == "/account" || hasAnyPrefix(path, "/account/language", "/account/notifications", "/account/favorites", "/account/api", "/account/mcp")
}

func isAccountSecurityPath(path string) bool {
	return hasAnyPrefix(path, "/account/2fa", "/account/passkeys", "/account/sessions")
}

func isAccountActivityPath(path string) bool {
	return hasAnyPrefix(path, "/account/activity", "/account/login-history")
}

// TabContext is the per-account state BuildPageTabs needs beyond feature flags.
type TabContext struct {
	Service             string // compose service the page manages (see ServiceForPath), adds Terminal/Logs tabs for it
	BackupsAdminManaged bool   // backup destination/settings are locked by the admin, so those tabs are hidden
	ActiveTool          string // "terminal" or "logs" when that page was opened from a service's own tabs
}

// BackupsAdminManaged is set by the backups module, which owns the admin-managed marker but can't be imported here.
var BackupsAdminManaged = func(userContext string) bool { return false }

// navArea is one feature area: a sidebar link plus the tabs across its pages. The link goes to the area's first reachable tab.
type navArea struct {
	label   string
	icon    template.HTML
	menuID  string
	section string   // sidebar heading that starts before this area
	parent  string   // extra crumb before the area in the header trail, linking to the first reachable area with the same parent
	keys    []string // area shows only when one of these is granted/upsellable, nil means "when it has any tab"
	match   func(path string) bool
	tabs    func(g navGate, path string, ctx TabContext) []NavLink
}

func mysqlTabs(g navGate, path string) []NavLink {
	var tabs []NavLink
	tabs = g.add(tabs, "mysql", "/mysql", "Databases", path == "/mysql" || path == "/mysql/new" || path == "/mysql/wizard", "")
	tabs = g.add(tabs, "mysql", "/mysql/users", "Users", path == "/mysql/users" || hasAnyPrefix(path, "/mysql/user", "/mysql/assign", "/mysql/remove", "/mysql/password"), "")
	tabs = g.add(tabs, "mysql_import", "/mysql/import", "Import", strings.HasPrefix(path, "/mysql/import"), "")
	tabs = g.add(tabs, "phpmyadmin", "/mysql/phpmyadmin", "phpMyAdmin", path == "/mysql/phpmyadmin", "_blank")
	tabs = g.add(tabs, "remote_mysql", "/mysql/remote-mysql", "Remote Access", path == "/mysql/remote-mysql", "")
	tabs = g.add(tabs, "mysql_processlist", "/mysql/processlist", "Running Queries", path == "/mysql/processlist", "")
	tabs = g.add(tabs, "mysql_conf", "/mysql/configuration", "Configuration", path == "/mysql/configuration", "")
	tabs = g.add(tabs, "mysql_root_password", "/mysql/root-password", "Root Password", path == "/mysql/root-password", "")
	tabs = g.add(tabs, "change_db", "/containers/mysql", "Server Type", path == "/containers/mysql", "")
	return tabs
}

func postgresqlTabs(g navGate, path string) []NavLink {
	var tabs []NavLink
	tabs = g.add(tabs, "postgresql", "/postgresql", "Databases", path == "/postgresql" || path == "/postgresql/new" || path == "/postgresql/wizard", "")
	tabs = g.add(tabs, "postgresql", "/postgresql/users", "Users", path == "/postgresql/users" || hasAnyPrefix(path, "/postgresql/user", "/postgresql/assign", "/postgresql/remove", "/postgresql/password"), "")
	tabs = g.add(tabs, "postgresql_import", "/postgresql/import", "Import", strings.HasPrefix(path, "/postgresql/import"), "")
	tabs = g.add(tabs, "remote_postgresql", "/postgresql/remote-postgresql", "Remote Access", path == "/postgresql/remote-postgresql", "")
	tabs = g.add(tabs, "postgresql", "/postgresql/processlist", "Running Queries", path == "/postgresql/processlist", "")
	tabs = g.add(tabs, "postgresql_conf", "/postgresql/configuration", "Configuration", path == "/postgresql/configuration", "")
	return tabs
}

func mongodbTabs(g navGate, path string) []NavLink {
	var tabs []NavLink
	tabs = g.add(tabs, "mongodb", "/mongodb", "Databases", path == "/mongodb" || path == "/mongodb/new" || path == "/mongodb/wizard", "")
	tabs = g.add(tabs, "mongodb", "/mongodb/users", "Users", path == "/mongodb/users" || hasAnyPrefix(path, "/mongodb/user", "/mongodb/assign", "/mongodb/remove", "/mongodb/password"), "")
	tabs = g.add(tabs, "mongodb_import", "/mongodb/import", "Import", strings.HasPrefix(path, "/mongodb/import"), "")
	tabs = g.add(tabs, "mongodb", "/mongodb/processlist", "Running Queries", path == "/mongodb/processlist", "")
	return tabs
}

// sidebarAreas is the sidebar in display order
var sidebarAreas = []navArea{
	{label: "Domains", icon: domainsIcon, menuID: "domains-menu", keys: []string{"domains"}, match: isDomainsPath,
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			if !g.allowed["domains"] {
				// sub-features are moot without domains itself
				return g.add(nil, "domains", "/domains", "Domains", false, "")
			}
			tabs := []NavLink{{Href: "/domains", Label: "Domains", Active: isDomainsPath(path) && !hasAnyPrefix(path, "/domains/edit-dns-zone", "/domains/dynamic-dns", "/domains/ssl", "/domains/redirect")}}
			tabs = g.add(tabs, "dns", "/domains/edit-dns-zone", "DNS Zone Editor", strings.HasPrefix(path, "/domains/edit-dns-zone"), "")
			tabs = g.add(tabs, "dynamic_dns", "/domains/dynamic-dns", "Dynamic DNS", strings.HasPrefix(path, "/domains/dynamic-dns"), "")
			tabs = g.add(tabs, "ssl", "/domains/ssl", "SSL Certificates", strings.HasPrefix(path, "/domains/ssl"), "")
			tabs = g.add(tabs, "redirects", "/domains/redirect", "Redirects", strings.HasPrefix(path, "/domains/redirect"), "")
			return tabs
		}},

	// mautic/flarum are excluded, legacy code slated for removal entirely, not ported here per user decision
	{label: "Websites", icon: websitesIcon, menuID: "websites-menu", match: isWebsitesPath,
		keys: []string{"wordpress", "drupal", "joomla", "opencart", "nextcloud", "prestashop", "matomo", "moodle", "mediawiki", "website_builder", "nodejs", "python"},
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			tabs := []NavLink{{Href: "/sites", Label: "Sites", Active: isWebsitesPath(path) && !strings.HasPrefix(path, "/wordpress") && !isAppInstallPath(path)}}
			tabs = g.add(tabs, "wordpress", "/wordpress", "WordPress", strings.HasPrefix(path, "/wordpress"), "")
			tabs = g.add(tabs, "autoinstaller", "/auto-installer", "Install App", isAppInstallPath(path), "")
			return tabs
		}},

	{label: "Email", icon: emailIcon, menuID: "emails-menu", match: isEmailPath,
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			var tabs []NavLink
			tabs = g.add(tabs, "emails", "/emails", "Accounts", isEmailPath(path) && !hasAnyPrefix(path, "/emails/aliases", "/emails/default", "/emails/filter", "/emails/deliverability", "/emails/import", "/emails/delete"), "")
			tabs = g.add(tabs, "email_aliases", "/emails/aliases", "Aliases", strings.HasPrefix(path, "/emails/aliases"), "")
			tabs = g.add(tabs, "email_default", "/emails/default", "Catch-all Address", strings.HasPrefix(path, "/emails/default"), "")
			tabs = g.add(tabs, "email_filters", "/emails/filter", "Filters", strings.HasPrefix(path, "/emails/filter"), "")
			tabs = g.add(tabs, "email_deliverability", "/emails/deliverability", "Deliverability", strings.HasPrefix(path, "/emails/deliverability"), "")
			tabs = g.add(tabs, "email_import", "/emails/import", "Import", strings.HasPrefix(path, "/emails/import"), "")
			tabs = g.add(tabs, "emails", "/emails/delete", "Delete Accounts", strings.HasPrefix(path, "/emails/delete"), "")
			tabs = g.add(tabs, "webmail", "/webmail/", "Webmail", false, "_blank")
			return tabs
		}},

	{label: "Files", icon: filesIcon, menuID: "files-menu", match: isFilesPath,
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			var tabs []NavLink
			tabs = g.add(tabs, "filemanager", "/files", "File Manager", strings.HasPrefix(path, "/files") && !strings.HasPrefix(path, "/files.trash") || hasAnyPrefix(path, "/file-manager/view-file", "/file-manager/edit-file"), "")
			tabs = g.add(tabs, "filemanager", "/file-manager/upload?method=upload", "Upload", strings.HasPrefix(path, "/file-manager/upload") && !strings.HasSuffix(path, "?method=download"), "")
			tabs = g.add(tabs, "filemanager", "/file-manager/upload?method=download", "Download from URL", strings.HasPrefix(path, "/file-manager/upload") && strings.HasSuffix(path, "?method=download"), "")
			tabs = g.add(tabs, "trash", "/files.trash", "Trash", strings.HasPrefix(path, "/files.trash"), "")
			tabs = g.add(tabs, "ftp", "/ftp", "FTP Accounts", strings.HasPrefix(path, "/ftp"), "")
			tabs = g.add(tabs, "fix_permissions", "/fix-permissions", "Fix Permissions", strings.HasPrefix(path, "/fix-permissions"), "")
			return tabs
		}},

	{label: "Backups", icon: backupsIcon, menuID: "backups-menu", match: isBackupsPath,
		tabs: func(g navGate, path string, ctx TabContext) []NavLink {
			var tabs []NavLink
			tabs = g.add(tabs, "backups", "/backups", "Overview", path == "/backups", "")
			tabs = g.add(tabs, "backups", "/backups/list", "Restore", strings.HasPrefix(path, "/backups/list"), "")
			if !ctx.BackupsAdminManaged {
				tabs = g.add(tabs, "backups", "/backups/settings", "Configuration", strings.HasPrefix(path, "/backups/settings"), "")
				tabs = g.add(tabs, "backups", "/backups/destination", "Destination", strings.HasPrefix(path, "/backups/destination"), "")
			}
			tabs = g.add(tabs, "backup_wizard", "/backup-wizard", "Backup Wizard", strings.HasPrefix(path, "/backup-wizard"), "")
			return tabs
		}},

	{label: "MySQL", icon: dbIcon, menuID: "mysql-menu", section: "Databases", parent: "Databases", match: isMySQLPath,
		tabs: func(g navGate, path string, _ TabContext) []NavLink { return mysqlTabs(g, path) }},

	{label: "PostgreSQL", icon: postgresqlIcon, menuID: "postgresql-menu", parent: "Databases", match: func(path string) bool { return strings.HasPrefix(path, "/postgresql") },
		tabs: func(g navGate, path string, _ TabContext) []NavLink { return postgresqlTabs(g, path) }},

	{label: "MongoDB", icon: mongodbIcon, menuID: "mongodb-menu", parent: "Databases", match: func(path string) bool { return strings.HasPrefix(path, "/mongodb") },
		tabs: func(g navGate, path string, _ TabContext) []NavLink { return mongodbTabs(g, path) }},

	{label: "PHP", icon: phpIcon, menuID: "php-menu", section: "Configure", match: func(path string) bool { return strings.HasPrefix(path, "/php") },
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			var tabs []NavLink
			tabs = g.add(tabs, "php", "/php/domains", "Version per Domain", path == "/php/domains", "")
			tabs = g.add(tabs, "php", "/php/default", "Default Version", path == "/php/default", "")
			tabs = g.add(tabs, "php_options", "/php/options", "Options", strings.Contains(path, "/options"), "")
			tabs = g.add(tabs, "php_extensions", "/php/extensions", "Extensions", strings.Contains(path, "/extensions"), "")
			tabs = g.add(tabs, "php_ini", "/php/php_ini_editor", "php.ini Editor", strings.Contains(path, "/php_ini_editor"), "")
			return tabs
		}},

	{label: "Cron Jobs", icon: cronIcon, menuID: "crons-menu", match: func(path string) bool { return strings.HasPrefix(path, "/cronjobs") },
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			return g.add(nil, "crons", "/cronjobs", "Cron Jobs", strings.HasPrefix(path, "/cronjobs"), "")
		}},

	{label: "Cache & Search", icon: cacheIcon, menuID: "cache-menu", match: func(path string) bool { return strings.HasPrefix(path, "/cache") },
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			var tabs []NavLink
			for _, c := range [][2]string{{"redis", "Redis"}, {"valkey", "Valkey"}, {"memcached", "Memcached"}, {"varnish", "Varnish"}, {"opensearch", "OpenSearch"}, {"elasticsearch", "Elasticsearch"}} {
				tabs = g.add(tabs, c[0], "/cache/"+c[0], c[1], path == "/cache/"+c[0], "")
			}
			return tabs
		}},

	{label: "Web Server Config", icon: webserverIcon, menuID: "webserver-menu", match: isWebServerConfPath,
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			var tabs []NavLink
			tabs = g.add(tabs, "webserver_conf", "/server/webserver_conf", "Server Settings", path == "/server/webserver_conf", "")
			tabs = g.add(tabs, "edit_vhost", "/domains/vhosts", "Domain VHosts", strings.HasPrefix(path, "/domains/vhosts"), "")
			tabs = g.add(tabs, "change_ws", "/containers/webserver", "Web Server Type", path == "/containers/webserver", "")
			return tabs
		}},

	{label: "Containers", icon: dockerIcon, menuID: "docker-menu", match: isContainersPath,
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			var tabs []NavLink
			tabs = g.add(tabs, "docker", "/containers", "Containers", path == "/containers" || path == "/containers/new" || hasAnyPrefix(path, "/containers/edit", "/containers/delete"), "")
			tabs = g.add(tabs, "terminal", "/containers/terminal", "Terminal", strings.HasPrefix(path, "/containers/terminal"), "")
			tabs = g.add(tabs, "docker", "/containers/logs", "Logs", strings.HasPrefix(path, "/containers/logs"), "")
			tabs = g.add(tabs, "change_image", "/containers/image/change", "Software Versions", strings.HasPrefix(path, "/containers/image"), "")
			return tabs
		}},

	{label: "Account", icon: accountIcon, menuID: "account-menu", match: isAccountPath,
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			var tabs []NavLink
			tabs = g.add(tabs, "account", "/account", "Login Details", path == "/account", "")
			tabs = g.add(tabs, "locale", "/account/language", "Language", strings.HasPrefix(path, "/account/language"), "")
			tabs = g.add(tabs, "notifications", "/account/notifications", "Notifications", strings.HasPrefix(path, "/account/notifications"), "")
			tabs = g.add(tabs, "favorites", "/account/favorites", "Favorites", strings.HasPrefix(path, "/account/favorites"), "")
			tabs = g.add(tabs, "api", "/account/api", "API", strings.HasPrefix(path, "/account/api"), "")
			tabs = g.add(tabs, "mcp", "/account/mcp", "AI Assistant (MCP)", strings.HasPrefix(path, "/account/mcp"), "")
			return tabs
		}},

	{label: "Statistics", icon: statsIcon, menuID: "stats-menu", section: "Monitor & Secure", match: isStatsPath,
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			var tabs []NavLink
			tabs = g.add(tabs, "usage", "/server/usage", "Resource Usage", strings.HasPrefix(path, "/server/usage"), "")
			tabs = g.add(tabs, "disk_usage", "/disk-usage/", "Disk Usage", strings.HasPrefix(path, "/disk-usage"), "")
			tabs = g.add(tabs, "inodes", "/inodes-explorer/", "Inode Usage", strings.HasPrefix(path, "/inodes-explorer"), "")
			tabs = g.add(tabs, "goaccess", "/domains/stats", "Visitor Statistics", strings.HasPrefix(path, "/domains/stats"), "")
			tabs = g.add(tabs, "domain_logs", "/domains/log", "Access Logs", strings.HasPrefix(path, "/domains/log"), "")
			return tabs
		}},

	{label: "Processes & Services", icon: processIcon, menuID: "processes-menu", match: isProcessesPath,
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			var tabs []NavLink
			tabs = g.add(tabs, "services", "/services", "Services", strings.HasPrefix(path, "/services"), "")
			tabs = g.add(tabs, "process_manager", "/process-manager", "Processes", strings.HasPrefix(path, "/process-manager"), "")
			return tabs
		}},

	{label: "Security", icon: securityIcon, menuID: "security-menu", match: isSecurityPath,
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			var tabs []NavLink
			tabs = g.add(tabs, "waf", "/server/waf", "Web Firewall", strings.HasPrefix(path, "/server/waf"), "")
			tabs = g.add(tabs, "ip_blocker", "/security/ip-blocker", "IP Blocker", strings.HasPrefix(path, "/security/ip-blocker"), "")
			tabs = g.add(tabs, "malware_scan", "/malware-scanner", "Malware Scanner", strings.HasPrefix(path, "/malware-scanner"), "")
			tabs = g.add(tabs, "twofa", "/account/2fa", "Two-Factor Auth", strings.HasPrefix(path, "/account/2fa"), "")
			tabs = g.add(tabs, "passkeys", "/account/passkeys", "Passkeys", strings.HasPrefix(path, "/account/passkeys"), "")
			tabs = g.add(tabs, "sessions", "/account/sessions", "Active Sessions", strings.HasPrefix(path, "/account/sessions"), "")
			tabs = g.add(tabs, "activity", "/account/activity", "Activity Log", strings.HasPrefix(path, "/account/activity"), "")
			tabs = g.add(tabs, "login_history", "/account/login-history", "Login History", strings.HasPrefix(path, "/account/login-history"), "")
			return tabs
		}},

	{label: "Server Info", icon: infoIcon, menuID: "info-menu", match: func(path string) bool { return path == "/server/info" },
		tabs: func(g navGate, path string, _ TabContext) []NavLink {
			return g.add(nil, "info", "/server/info", "Server Info", path == "/server/info", "")
		}},
}

// areaEntry picks the link an area's sidebar entry points at: its first reachable same-window tab, else a greyed-out upsell entry, ok=false when the area has nothing to show
func areaEntry(g navGate, a navArea, path string) (NavLink, bool) {
	if a.keys != nil && !g.has(a.keys...) {
		return NavLink{}, false
	}
	candidates := a.tabs(g, path, TabContext{})
	for _, t := range candidates {
		if !t.Disabled && t.Target == "" {
			return NavLink{Href: t.Href, Label: a.label, Active: a.match(path)}, true
		}
	}
	for _, t := range candidates {
		if t.Target == "" {
			return NavLink{Href: t.Href, Label: a.label, Disabled: true}, true
		}
	}
	return NavLink{}, false
}

// BuildSidebarNav builds the sidebar from user_allowed/user_upsell_allowed and the current request path, one link per feature area shown when the user has access to (or could upsell into) any page in it.
func BuildSidebarNav(allowed, upsellAllowed map[string]bool, path string) []NavItem {
	g := navGate{allowed, upsellAllowed}
	var items []NavItem
	pendingSection := ""
	for _, a := range sidebarAreas {
		if a.section != "" {
			pendingSection = a.section
		}
		entry, ok := areaEntry(g, a, path)
		if !ok {
			continue
		}
		items = append(items, NavItem{Label: a.label, Icon: a.icon, MenuID: a.menuID, Href: entry.Href,
			Active: entry.Active, Disabled: entry.Disabled, Section: pendingSection})
		pendingSection = ""
	}
	return items
}

// ServiceForPath returns the compose service a page manages, for its Terminal/Logs tabs, or "" when the page isn't about one service. env reads the user's .env, called only for areas whose service name is configurable.
func ServiceForPath(path string, env func(key string) string) string {
	switch {
	case isMySQLPath(path):
		if t := env("MYSQL_TYPE"); t != "" {
			return t
		}
		return "mysql"
	case strings.HasPrefix(path, "/postgresql"):
		return "postgres"
	case strings.HasPrefix(path, "/mongodb"):
		return "mongodb"
	case strings.HasPrefix(path, "/cache/"):
		return strings.Trim(strings.TrimPrefix(path, "/cache/"), "/")
	case isWebServerConfPath(path):
		return env("WEB_SERVER")
	case strings.HasPrefix(path, "/php"):
		if v := env("DEFAULT_PHP_VERSION"); v != "" {
			return "php-fpm-" + v
		}
	}
	return ""
}

// BuildPageTabs returns the tab bar across the current page's area, plus Terminal/Logs tabs when the page manages a single service, or nil when there are fewer than two tabs.
func BuildPageTabs(allowed, upsellAllowed map[string]bool, path string, ctx TabContext) []NavLink {
	g := navGate{allowed, upsellAllowed}
	var tabs []NavLink
	for _, a := range sidebarAreas {
		if a.match(path) {
			tabs = a.tabs(g, path, ctx)
			break
		}
	}

	if ctx.Service != "" && g.has("terminal", "docker") {
		// the service's own terminal/logs page is showing, so none of the area's page tabs are current
		if ctx.ActiveTool != "" {
			for i := range tabs {
				tabs[i].Active = false
			}
		}
		// ctx=service keeps those pages inside this area instead of jumping to Containers
		tabs = g.add(tabs, "terminal", "/containers/terminal/"+ctx.Service+"?ctx=service", "Terminal", ctx.ActiveTool == "terminal", "")
		tabs = g.add(tabs, "docker", "/containers/logs?container="+ctx.Service+"&lines=20&ctx=service", "Logs", ctx.ActiveTool == "logs", "")
	}

	if len(tabs) < 2 {
		return nil
	}
	return tabs
}

var cacheServices = map[string]bool{"redis": true, "valkey": true, "memcached": true, "varnish": true, "opensearch": true, "elasticsearch": true}

// AreaPathForService maps a compose service to the page of the area that manages it, "" for services no area owns
func AreaPathForService(service string, env func(key string) string) string {
	switch {
	case service == "mysql" || service == "mariadb":
		return "/mysql"
	case service == "postgres":
		return "/postgresql"
	case service == "mongodb":
		return "/mongodb"
	case cacheServices[service]:
		return "/cache/" + service
	case strings.HasPrefix(service, "php-fpm-"):
		return "/php/domains"
	case service != "" && service == env("WEB_SERVER"):
		return "/server/webserver_conf"
	}
	return ""
}

// ResolveNav returns the path the navigation should treat this request as, plus its tab context. A terminal/logs page opened from a service's tabs (?ctx=service) is treated as a page of that service's area.
func ResolveNav(r *http.Request, env func(key string) string) (string, TabContext) {
	if r.URL.Query().Get("ctx") == "service" {
		var service, tool string
		switch {
		case strings.HasPrefix(r.URL.Path, "/containers/terminal/"):
			service, tool = strings.Trim(strings.TrimPrefix(r.URL.Path, "/containers/terminal/"), "/"), "terminal"
		case r.URL.Path == "/containers/logs":
			service, tool = r.URL.Query().Get("container"), "logs"
		}
		if area := AreaPathForService(service, env); area != "" {
			return area, TabContext{Service: service, ActiveTool: tool}
		}
	}
	return NavPath(r), TabContext{Service: ServiceForPath(r.URL.Path, env)}
}

func stripQuery(href string) string {
	if i := strings.IndexByte(href, '?'); i >= 0 {
		return href[:i]
	}
	return href
}

// BuildNavTrail returns the header trail [parent] > area > tab > sub-page, the last crumb has no Href. navPath is ResolveNav's path, requestPath the real one, title the page's own title.
func BuildNavTrail(allowed, upsellAllowed map[string]bool, navPath, requestPath, title string, tabs []NavLink) []NavLink {
	g := navGate{allowed, upsellAllowed}
	var crumbs []NavLink
	// a label already in the trail isn't repeated, e.g. the Databases tab under Databases > MySQL
	add := func(href, label string) {
		for _, c := range crumbs {
			if strings.EqualFold(c.Label, label) {
				return
			}
		}
		crumbs = append(crumbs, NavLink{Href: href, Label: label})
	}

	current := ""
	if strings.HasPrefix(navPath, "/dashboard/") {
		add("/dashboard", "Dashboard")
	}
	for _, a := range sidebarAreas {
		if !a.match(navPath) {
			continue
		}
		if a.parent != "" {
			for _, sibling := range sidebarAreas {
				if sibling.parent != a.parent {
					continue
				}
				if entry, ok := areaEntry(g, sibling, navPath); ok {
					add(entry.Href, a.parent)
					break
				}
			}
		}
		if entry, ok := areaEntry(g, a, navPath); ok {
			add(entry.Href, a.label)
			current = entry.Href
		}
		for _, t := range tabs {
			if t.Active {
				add(t.Href, t.Label)
				current = t.Href
			}
		}
		break
	}

	// a sub-page (e.g. /mysql/assign under Users) gets its own title as the last crumb
	if title != "" && (len(crumbs) == 0 || stripQuery(current) != requestPath) {
		add("", title)
	}
	if len(crumbs) > 0 {
		crumbs[len(crumbs)-1].Href = ""
	}
	return crumbs
}

// DashboardTabs returns the Dashboard/Upgrade tabs, nil without an upsell plan so the dashboard keeps no tab bar at all
func DashboardTabs(path string, upsellAvailable bool) []NavLink {
	if !upsellAvailable {
		return nil
	}
	return []NavLink{
		{Href: "/dashboard", Label: "Dashboard", Active: path == "/dashboard"},
		{Href: "/dashboard/upgrade", Label: "Upgrade now", Active: path == "/dashboard/upgrade"},
	}
}

// FeatureGroup is one sidebar area's pages that an upgrade would unlock
type FeatureGroup struct {
	Area  string
	Pages []string
}

// UpgradeFeatures lists, per sidebar area, the pages the upsell plan unlocks that the current plan can't reach
func UpgradeFeatures(allowed, upsellAllowed map[string]bool) []FeatureGroup {
	combined := make(map[string]bool, len(allowed)+len(upsellAllowed))
	for k, v := range allowed {
		combined[k] = v
	}
	for k, v := range upsellAllowed {
		combined[k] = combined[k] || v
	}
	current, upgraded := navGate{allowed: allowed}, navGate{allowed: combined}

	var groups []FeatureGroup
	for _, a := range sidebarAreas {
		if a.keys != nil && !upgraded.has(a.keys...) {
			continue
		}
		have := map[string]bool{}
		if a.keys == nil || current.has(a.keys...) {
			for _, t := range a.tabs(current, "", TabContext{}) {
				have[t.Href] = true
			}
		}
		var pages []string
		for _, t := range a.tabs(upgraded, "", TabContext{}) {
			if !have[t.Href] {
				pages = append(pages, t.Label)
			}
		}
		if len(pages) > 0 {
			groups = append(groups, FeatureGroup{Area: a.label, Pages: pages})
		}
	}
	return groups
}
