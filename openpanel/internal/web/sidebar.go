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

// NavGroup is one collapsible section of the sidebar navigation: a labeled group of NavLinks, shown only when the user has access to at least one feature in that group.
type NavGroup struct {
	Label  string
	Icon   template.HTML
	MenuID string
	Links  []NavLink
	Open   bool
	Active bool
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

// BuildSidebarNav builds the sidebar's feature-conditional menu groups from user_allowed/user_upsell_allowed and the current request path, one group per feature area, each included only when the user has access to (or could upsell into) at least one feature in it.
func BuildSidebarNav(allowed, upsellAllowed map[string]bool, path string) []NavGroup {
	var groups []NavGroup

	has := func(keys ...string) bool {
		for _, k := range keys {
			if allowed[k] || upsellAllowed[k] {
				return true
			}
		}
		return false
	}

	// add appends a link gated by a single feature key: shown active when the plan grants it, greyed-out (Disabled) when only the upsell plan would, omitted otherwise.
	add := func(links []NavLink, key, href, label string, active bool, target string) []NavLink {
		if allowed[key] {
			return append(links, NavLink{Href: href, Label: label, Active: active, Target: target})
		}
		if upsellAllowed[key] {
			return append(links, NavLink{Href: href, Label: label, Target: target, Disabled: true})
		}
		return links
	}

	// Websites group - mautic/flarum are excluded, legacy code slated for removal entirely, not ported here per user decision
	if has("wordpress", "drupal", "joomla", "opencart", "nextcloud", "prestashop", "matomo", "moodle", "mediawiki", "website_builder", "nodejs", "python") {
		var links []NavLink
		links = add(links, "autoinstaller", "/auto-installer", "Auto Installer",
			hasAnyPrefix(path, "/auto-installer", "/pm2", "/nodejs", "/python", "/website-builder/install",
				"/drupal/install", "/joomla/install", "/opencart/install", "/nextcloud/install",
				"/prestashop/install", "/matomo/install", "/moodle/install", "/mediawiki/install"), "")
		links = append(links, NavLink{Href: "/sites", Label: "Site Manager",
			Active: hasAnyPrefix(path, "/sites") || (strings.HasPrefix(path, "/website") && !strings.HasPrefix(path, "/website-builder"))})
		links = add(links, "wordpress", "/wordpress", "WordPress Manager", strings.HasPrefix(path, "/wordpress"), "")
		open := hasAnyPrefix(path, "/auto-installer", "/sites", "/website", "/wordpress", "/drupal", "/joomla", "/opencart", "/nextcloud", "/prestashop", "/matomo", "/moodle", "/mediawiki", "/pm2", "/nodejs", "/python")
		groups = append(groups, NavGroup{"Websites", websitesIcon, "websites-menu", links, open, open})
	}

	// Files group
	if has("filemanager", "trash", "ftp", "disk_usage", "backups", "backup_wizard", "inodes", "malware_scan", "fix_permissions") {
		var links []NavLink
		if allowed["filemanager"] {
			links = append(links,
				NavLink{Href: "/files", Label: "File Manager", Active: (strings.HasPrefix(path, "/files") && !strings.HasPrefix(path, "/files.trash")) || strings.HasPrefix(path, "/file-manager/view-file")},
				NavLink{Href: "/file-manager/upload?method=upload", Label: "Upload from device", Active: strings.HasPrefix(path, "/file-manager/upload") && !strings.HasSuffix(path, "?method=download")},
				NavLink{Href: "/file-manager/upload?method=download", Label: "Download from URL", Active: strings.HasPrefix(path, "/file-manager/upload") && strings.HasSuffix(path, "?method=download")},
			)
		} else if upsellAllowed["filemanager"] {
			links = append(links,
				NavLink{Href: "/files", Label: "File Manager", Disabled: true},
				NavLink{Href: "/file-manager/upload?method=upload", Label: "Upload from device", Disabled: true},
				NavLink{Href: "/file-manager/upload?method=download", Label: "Download from URL", Disabled: true},
			)
		}
		links = add(links, "ftp", "/ftp", "FTP Accounts", strings.HasPrefix(path, "/ftp"), "")
		links = add(links, "backups", "/backups", "Backups", strings.HasPrefix(path, "/backups"), "")
		links = add(links, "backup_wizard", "/backup-wizard", "Backup Wizard", strings.HasPrefix(path, "/backup-wizard"), "")
		links = add(links, "malware_scan", "/malware-scanner", "ClamAV Scanner", strings.HasPrefix(path, "/malware-scanner"), "")
		links = add(links, "disk_usage", "/disk-usage/", "Disk Usage", strings.HasPrefix(path, "/disk-usage"), "")
		links = add(links, "inodes", "/inodes-explorer/", "Inodes Explorer", strings.HasPrefix(path, "/inodes-explorer"), "")
		links = add(links, "fix_permissions", "/fix-permissions", "Fix Permissions", strings.HasPrefix(path, "/fix-permissions"), "")
		links = add(links, "trash", "/files.trash", "Trash", strings.HasPrefix(path, "/files.trash"), "")
		open := hasAnyPrefix(path, "/files", "/file-manager/edit-file", "/file-manager/view-file", "/backups",
			"/backup-wizard", "/disk-usage", "/inodes-explorer", "/malware-scanner", "/ftp", "/fix-permissions", "/file-manager/upload")
		groups = append(groups, NavGroup{"Files", filesIcon, "files-menu", links, open, open})
	}

	// MySQL group
	if has("mysql_conf", "remote_mysql", "mysql", "mysql_root_password", "mysql_processlist") {
		links := []NavLink{
			{Href: "/mysql", Label: "Databases", Active: path == "/mysql"},
			{Href: "/mysql/users", Label: "Users", Active: path == "/mysql/users"},
		}
		links = add(links, "phpmyadmin", "/mysql/phpmyadmin", "phpMyAdmin", path == "/mysql/phpmyadmin", "_blank")
		links = append(links,
			NavLink{Href: "/mysql/wizard", Label: "Database Wizard", Active: path == "/mysql/wizard"},
			NavLink{Href: "/mysql/new", Label: "Create Database", Active: path == "/mysql/new"},
			NavLink{Href: "/mysql/user", Label: "Create User", Active: path == "/mysql/user"},
			NavLink{Href: "/mysql/assign", Label: "Assign User to DB", Active: path == "/mysql/assign"},
			NavLink{Href: "/mysql/remove", Label: "Remove User from DB", Active: path == "/mysql/remove"},
		)
		links = add(links, "mysql_import", "/mysql/import", "Import Database", strings.HasPrefix(path, "/mysql/import"), "")
		links = append(links, NavLink{Href: "/mysql/remote-mysql", Label: "Remote Access", Active: path == "/mysql/remote-mysql"})
		links = add(links, "mysql_root_password", "/mysql/root-password", "Change root password", path == "/mysql/root-password", "")
		links = add(links, "mysql_processlist", "/mysql/processlist", "Show Processes", path == "/mysql/processlist", "")
		links = add(links, "mysql_conf", "/mysql/configuration", "Configuration", path == "/mysql/configuration", "")
		open := hasAnyPrefix(path, "/mysql", "/database")
		groups = append(groups, NavGroup{"MySQL", dbIcon, "mysql-menu", links, open, open})
	}

	// PostgreSQL group
	if has("postgresql_conf", "remote_postgresql", "postgresql") {
		links := []NavLink{
			{Href: "/postgresql", Label: "Databases", Active: path == "/postgresql"},
			{Href: "/postgresql/users", Label: "Users", Active: path == "/postgresql/users"},
		}
		links = append(links,
			NavLink{Href: "/postgresql/wizard", Label: "Database Wizard", Active: path == "/postgresql/wizard"},
			NavLink{Href: "/postgresql/new", Label: "Create Database", Active: path == "/postgresql/new"},
			NavLink{Href: "/postgresql/user", Label: "Create User", Active: path == "/postgresql/user"},
			NavLink{Href: "/postgresql/assign", Label: "Assign User to DB", Active: path == "/postgresql/assign"},
			NavLink{Href: "/postgresql/remove", Label: "Remove User from DB", Active: path == "/postgresql/remove"},
		)
		links = add(links, "import_postgresql", "/postgresql/import", "Import Database", strings.HasPrefix(path, "/postgresql/import"), "")
		links = append(links,
			NavLink{Href: "/postgresql/remote-postgresql", Label: "Remote Access", Active: path == "/postgresql/remote-postgresql"},
			NavLink{Href: "/postgresql/processlist", Label: "Show Processes", Active: path == "/postgresql/processlist"},
		)
		links = add(links, "postgresql_conf", "/postgresql/configuration", "Configuration", path == "/postgresql/configuration", "")
		open := hasAnyPrefix(path, "/postgresql", "/database")
		groups = append(groups, NavGroup{"PostgreSQL", postgresqlIcon, "postgresql-menu", links, open, open})
	}

	// MongoDB group
	if has("mongodb") {
		links := []NavLink{
			{Href: "/mongodb", Label: "Databases", Active: path == "/mongodb"},
			{Href: "/mongodb/users", Label: "Users", Active: path == "/mongodb/users"},
		}
		links = append(links,
			NavLink{Href: "/mongodb/wizard", Label: "Database Wizard", Active: path == "/mongodb/wizard"},
			NavLink{Href: "/mongodb/new", Label: "Create Database", Active: path == "/mongodb/new"},
			NavLink{Href: "/mongodb/user", Label: "Create User", Active: path == "/mongodb/user"},
			NavLink{Href: "/mongodb/assign", Label: "Assign User to DB", Active: path == "/mongodb/assign"},
			NavLink{Href: "/mongodb/remove", Label: "Remove User from DB", Active: path == "/mongodb/remove"},
		)
		links = add(links, "mongodb_import", "/mongodb/import", "Import Database", strings.HasPrefix(path, "/mongodb/import"), "")
		open := hasAnyPrefix(path, "/mongodb", "/database")
		groups = append(groups, NavGroup{"MongoDB", mongodbIcon, "mongodb-menu", links, open, open})
	}

	// Domains group
	if allowed["domains"] || upsellAllowed["domains"] {
		var links []NavLink
		if allowed["domains"] {
			links = append(links,
				NavLink{Href: "/domains", Label: "Domain Names", Active: path == "/domains"},
				NavLink{Href: "/domains/new", Label: "Add New Domain", Active: path == "/domains/new"},
			)
			links = add(links, "redirects", "/domains/redirect", "Redirects", path == "/domains/redirect", "")
			links = add(links, "dns", "/domains/edit-dns-zone", "DNS Zone Editor", strings.HasPrefix(path, "/domains/edit-dns-zone"), "")
			links = add(links, "dynamic_dns", "/domains/dynamic-dns", "Dynamic DNS", strings.HasPrefix(path, "/domains/dynamic-dns"), "")
			links = add(links, "ssl", "/domains/ssl", "SSL", path == "/domains/ssl", "")
			links = add(links, "edit_vhost", "/domains/vhosts", "VHosts File Editor", path == "/domains/vhosts", "")
			if allowed["domain_suspend"] {
				links = append(links,
					NavLink{Href: "/domains/suspend", Label: "Suspend a Domain", Active: path == "/domains/suspend"},
					NavLink{Href: "/domains/unsuspend", Label: "Unsuspend a Domain", Active: path == "/domains/unsuspend"},
				)
			} else if upsellAllowed["domain_suspend"] {
				links = append(links,
					NavLink{Href: "/domains/suspend", Label: "Suspend a Domain", Disabled: true},
					NavLink{Href: "/domains/unsuspend", Label: "Unsuspend a Domain", Disabled: true},
				)
			}
			links = add(links, "docroot", "/domains/docroot", "Change docroot", path == "/domains/docroot", "")
			links = add(links, "domain_logs", "/domains/log", "Raw Access Logs", strings.HasPrefix(path, "/domains/log"), "")
			links = add(links, "goaccess", "/domains/stats", "GoAccess", path == "/domains/stats", "")
		} else {
			// domains itself is only upsell-eligible - sub-features are moot without it, so just offer the two base links greyed out
			links = append(links,
				NavLink{Href: "/domains", Label: "Domain Names", Disabled: true},
				NavLink{Href: "/domains/new", Label: "Add New Domain", Disabled: true},
			)
		}
		open := strings.HasPrefix(path, "/domains")
		groups = append(groups, NavGroup{"Domains", domainsIcon, "domains-menu", links, open, open})
	}

	// Emails group
	if has("emails", "email_filters", "email_aliases", "email_default", "email_import", "email_deliverability", "webmail") {
		var links []NavLink
		if allowed["emails"] {
			links = append(links,
				NavLink{Href: "/emails", Label: "Email Accounts", Active: path == "/emails"},
				NavLink{Href: "/emails/new", Label: "Create New Account", Active: strings.HasPrefix(path, "/emails/new")},
			)
		} else if upsellAllowed["emails"] {
			links = append(links,
				NavLink{Href: "/emails", Label: "Email Accounts", Disabled: true},
				NavLink{Href: "/emails/new", Label: "Create New Account", Disabled: true},
			)
		}
		links = add(links, "webmail", "/webmail/", "Webmail", false, "_blank")
		links = add(links, "email_filters", "/emails/filter", "Filters", strings.HasPrefix(path, "/emails/filter"), "")
		links = add(links, "email_aliases", "/emails/aliases", "Aliases", strings.HasPrefix(path, "/emails/aliases"), "")
		links = add(links, "email_default", "/emails/default", "Default Address", strings.HasPrefix(path, "/emails/default"), "")
		links = add(links, "email_import", "/emails/import", "Address Importer", strings.HasPrefix(path, "/emails/import"), "")
		links = add(links, "email_deliverability", "/emails/deliverability", "Email Deliverability", strings.HasPrefix(path, "/emails/deliverability"), "")
		links = add(links, "emails", "/emails/delete", "Delete Accounts", strings.HasPrefix(path, "/emails/delete"), "")
		open := strings.HasPrefix(path, "/email")
		groups = append(groups, NavGroup{"Emails", emailIcon, "emails-menu", links, open, open})
	}

	// Caching group
	if has("redis", "valkey", "memcached", "varnish", "elasticsearch", "opensearch") {
		var links []NavLink
		links = add(links, "redis", "/cache/redis", "Redis", path == "/cache/redis", "")
		links = add(links, "valkey", "/cache/valkey", "Valkey", path == "/cache/valkey", "")
		links = add(links, "memcached", "/cache/memcached", "Memcached", path == "/cache/memcached", "")
		links = add(links, "opensearch", "/cache/opensearch", "Opensearch", path == "/cache/opensearch", "")
		links = add(links, "elasticsearch", "/cache/elasticsearch", "Elasticsearch", path == "/cache/elasticsearch", "")
		links = add(links, "varnish", "/cache/varnish", "Varnish", path == "/cache/varnish", "")
		open := strings.HasPrefix(path, "/cache")
		groups = append(groups, NavGroup{"Caching", cacheIcon, "cache-menu", links, open, open})
	}

	// PHP group
	if has("php", "php_options", "php_ini", "php_extensions") {
		links := []NavLink{
			{Href: "/php/domains", Label: "Select PHP version", Active: path == "/php/domains"},
			{Href: "/php/default", Label: "Default version", Active: path == "/php/default"},
		}
		links = add(links, "php_options", "/php/options", "PHP Options", strings.HasPrefix(path, "/php") && strings.Contains(path, "/options"), "")
		links = add(links, "php_extensions", "/php/extensions", "PHP Extensions", strings.HasPrefix(path, "/php") && strings.Contains(path, "/extensions"), "")
		links = add(links, "php_ini", "/php/php_ini_editor", "PHP.INI Editor", strings.HasPrefix(path, "/php") && strings.Contains(path, "/php_ini_editor"), "")
		open := strings.HasPrefix(path, "/php")
		groups = append(groups, NavGroup{"PHP", phpIcon, "php-menu", links, open, open})
	}

	// Advanced group
	if has("crons", "services", "ssh", "usage", "process_manager", "webserver_conf", "timezone", "waf", "ip_blocker", "info") {
		var links []NavLink
		links = add(links, "services", "/services", "Services", strings.HasPrefix(path, "/services"), "")
		links = add(links, "crons", "/cronjobs", "Cron Jobs", strings.HasPrefix(path, "/cronjobs"), "")
		links = add(links, "ip_blocker", "/security/ip-blocker", "IP Blocker", path == "/security/ip-blocker", "")
		links = add(links, "process_manager", "/process-manager", "Process Manager", path == "/process-manager", "")
		links = add(links, "webserver_conf", "/server/webserver_conf", "WebServer Settings", path == "/server/webserver_conf", "")
		links = add(links, "waf", "/server/waf", "WAF", strings.HasPrefix(path, "/server/waf"), "")
		links = add(links, "usage", "/server/usage", "Resource Usage", strings.HasPrefix(path, "/server/usage"), "")
		links = add(links, "info", "/server/info", "Server Information", path == "/server/info", "")
		open := hasAnyPrefix(path, "/cronjobs", "/services/", "/server", "/process-manager", "/server/usage", "/security/ip-blocker")
		active := hasAnyPrefix(path, "/cronjobs", "/server", "/process-manager", "/server/usage", "/security/ip-blocker")
		groups = append(groups, NavGroup{"Advanced", advancedIcon, "advanced-menu", links, open, active})
	}

	// Docker group
	if has("docker", "terminal", "change_image", "change_ws", "change_db") {
		var links []NavLink
		links = add(links, "docker", "/containers", "Containers", path == "/containers" || path == "/containers/new" || strings.HasPrefix(path, "/containers/edit"), "")
		links = add(links, "terminal", "/containers/terminal", "Terminal", strings.HasPrefix(path, "/containers/terminal"), "")
		links = add(links, "docker", "/containers/logs", "Logs", strings.HasPrefix(path, "/containers/logs"), "")
		links = add(links, "change_image", "/containers/image/change", "Change image tag", path == "/containers/image/change", "")
		links = add(links, "change_ws", "/containers/webserver", "Switch WebServer", path == "/containers/webserver", "")
		links = add(links, "change_db", "/containers/mysql", "Switch MySQL Type", path == "/containers/mysql", "")
		open := strings.HasPrefix(path, "/containers")
		groups = append(groups, NavGroup{"Containers", dockerIcon, "docker-menu", links, open, open})
	}

	// Account group
	if has("account", "twofa", "passkeys", "favorites", "login_history", "notifications", "locale", "sessions", "activity", "mcp") {
		var links []NavLink
		links = add(links, "account", "/account", "Email & Password", path == "/account", "")
		links = add(links, "locale", "/account/language", "Change Language", strings.HasPrefix(path, "/account/language"), "")
		links = add(links, "notifications", "/account/notifications", "Email Notifications", strings.HasPrefix(path, "/account/notifications"), "")
		links = add(links, "twofa", "/account/2fa", "2FA", strings.HasPrefix(path, "/account/2fa"), "")
		links = add(links, "passkeys", "/account/passkeys", "Passkeys", strings.HasPrefix(path, "/account/passkeys"), "")
		links = add(links, "sessions", "/account/sessions", "Active Sessions", strings.HasPrefix(path, "/account/sessions"), "")
		links = add(links, "favorites", "/account/favorites", "Favorite Pages", strings.HasPrefix(path, "/account/favorites"), "")
		links = add(links, "activity", "/account/activity", "Account Activity", strings.HasPrefix(path, "/account/activity"), "")
		links = add(links, "login_history", "/account/login-history", "Login History", strings.HasPrefix(path, "/account/login-history"), "")
		links = add(links, "api", "/account/api", "API Reference", strings.HasPrefix(path, "/account/api"), "")
		links = add(links, "mcp", "/account/mcp", "MCP", strings.HasPrefix(path, "/account/mcp"), "")
		open := strings.HasPrefix(path, "/account")
		groups = append(groups, NavGroup{"Account", accountIcon, "account-menu", links, open, open})
	}

	return groups
}
