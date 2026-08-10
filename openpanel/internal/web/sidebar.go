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
}

// NavGroup is one collapsible section of the sidebar navigation: a labeled
// group of NavLinks, shown only when the user has access to at least one
// feature in that group.
type NavGroup struct {
	Label  string
	Icon   template.HTML
	MenuID string
	Links  []NavLink
	Open   bool
	Active bool
}

// NavPath derives BuildSidebarNav's path argument from a request: r.URL.Path,
// with "?method=download" appended for the file-manager upload page's
// download-from-URL variant - the one nav item whose active state depends
// on a query param rather than the path alone. Every other path in
// the sidebar is compared with plain prefix/equality checks that this
// suffix never touches, since it's only ever appended to
// "/file-manager/upload".
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

// BuildSidebarNav builds the sidebar's feature-conditional menu groups from
// user_allowed and the current request path, one group per feature area
// (Websites, Files, MySQL, ...), each included only when the user has
// access to at least one feature in it.
func BuildSidebarNav(allowed map[string]bool, path string) []NavGroup {
	var groups []NavGroup

	has := func(keys ...string) bool {
		for _, k := range keys {
			if allowed[k] {
				return true
			}
		}
		return false
	}

	// Websites group.
	// mautic/flarum/drupal are excluded: legacy code slated for removal
	// from the codebase entirely, not ported here (per user decision).
	if has("wordpress", "website_builder", "nodejs", "python") {
		var links []NavLink
		if allowed["autoinstaller"] {
			links = append(links, NavLink{"/auto-installer", "Auto Installer",
				hasAnyPrefix(path, "/auto-installer", "/pm2", "/nodejs", "/python", "/website-builder/install"), ""})
		}
		links = append(links, NavLink{"/sites", "Site Manager",
			hasAnyPrefix(path, "/sites") || (strings.HasPrefix(path, "/website") && !strings.HasPrefix(path, "/website-builder")), ""})
		if allowed["wordpress"] {
			links = append(links, NavLink{"/wordpress", "WordPress Manager", strings.HasPrefix(path, "/wordpress"), ""})
		}
		open := hasAnyPrefix(path, "/auto-installer", "/sites", "/website", "/wordpress", "/pm2", "/nodejs", "/python")
		groups = append(groups, NavGroup{"Websites", websitesIcon, "websites-menu", links, open, open})
	}

	// Files group
	if has("filemanager", "trash", "ftp", "disk_usage", "backups", "backup_wizard", "inodes", "malware_scan", "fix_permissions") {
		var links []NavLink
		if allowed["filemanager"] {
			links = append(links,
				NavLink{"/files", "File Manager", (strings.HasPrefix(path, "/files") && !strings.HasPrefix(path, "/files.trash")) || strings.HasPrefix(path, "/file-manager/view-file"), ""},
				NavLink{"/file-manager/upload?method=upload", "Upload from device", strings.HasPrefix(path, "/file-manager/upload") && !strings.HasSuffix(path, "?method=download"), ""},
				NavLink{"/file-manager/upload?method=download", "Download from URL", strings.HasPrefix(path, "/file-manager/upload") && strings.HasSuffix(path, "?method=download"), ""},
			)
		}
		if allowed["ftp"] {
			links = append(links, NavLink{"/ftp", "FTP Accounts", strings.HasPrefix(path, "/ftp"), ""})
		}
		if allowed["backups"] {
			links = append(links, NavLink{"/backups", "Backups", strings.HasPrefix(path, "/backups"), ""})
		}
		if allowed["backup_wizard"] {
			links = append(links, NavLink{"/backup-wizard", "Backup Wizard", strings.HasPrefix(path, "/backup-wizard"), ""})
		}
		if allowed["malware_scan"] {
			links = append(links, NavLink{"/malware-scanner", "ClamAV Scanner", strings.HasPrefix(path, "/malware-scanner"), ""})
		}
		if allowed["disk_usage"] {
			links = append(links, NavLink{"/disk-usage/", "Disk Usage", strings.HasPrefix(path, "/disk-usage"), ""})
		}
		if allowed["inodes"] {
			links = append(links, NavLink{"/inodes-explorer/", "Inodes Explorer", strings.HasPrefix(path, "/inodes-explorer"), ""})
		}
		if allowed["fix_permissions"] {
			links = append(links, NavLink{"/fix-permissions", "Fix Permissions", strings.HasPrefix(path, "/fix-permissions"), ""})
		}
		if allowed["trash"] {
			links = append(links, NavLink{"/files.trash", "Trash", strings.HasPrefix(path, "/files.trash"), ""})
		}
		open := hasAnyPrefix(path, "/files", "/file-manager/edit-file", "/file-manager/view-file", "/backups",
			"/backup-wizard", "/disk-usage", "/inodes-explorer", "/malware-scanner", "/ftp", "/fix-permissions", "/file-manager/upload")
		groups = append(groups, NavGroup{"Files", filesIcon, "files-menu", links, open, open})
	}

	// MySQL group
	if has("mysql_conf", "remote_mysql", "mysql", "mysql_root_password", "mysql_processlist") {
		links := []NavLink{
			{"/mysql", "Databases", path == "/mysql", ""},
			{"/mysql/users", "Users", path == "/mysql/users", ""},
		}
		if allowed["phpmyadmin"] {
			links = append(links, NavLink{"/mysql/phpmyadmin", "phpMyAdmin", path == "/mysql/phpmyadmin", "_blank"})
		}
		links = append(links,
			NavLink{"/mysql/wizard", "Database Wizard", path == "/mysql/wizard", ""},
			NavLink{"/mysql/new", "Create Database", path == "/mysql/new", ""},
			NavLink{"/mysql/user", "Create User", path == "/mysql/user", ""},
			NavLink{"/mysql/assign", "Assign User to DB", path == "/mysql/assign", ""},
			NavLink{"/mysql/remove", "Remove User from DB", path == "/mysql/remove", ""},
		)
		if allowed["mysql_import"] {
			links = append(links, NavLink{"/mysql/import", "Import Database", strings.HasPrefix(path, "/mysql/import"), ""})
		}
		links = append(links, NavLink{"/mysql/remote-mysql", "Remote Access", path == "/mysql/remote-mysql", ""})
		if allowed["mysql_root_password"] {
			links = append(links, NavLink{"/mysql/root-password", "Change root password", path == "/mysql/root-password", ""})
		}
		if allowed["mysql_processlist"] {
			links = append(links, NavLink{"/mysql/processlist", "Show Processes", path == "/mysql/processlist", ""})
		}
		if allowed["mysql_conf"] {
			links = append(links, NavLink{"/mysql/configuration", "Configuration", path == "/mysql/configuration", ""})
		}
		open := hasAnyPrefix(path, "/mysql", "/database")
		groups = append(groups, NavGroup{"MySQL", dbIcon, "mysql-menu", links, open, open})
	}

	// PostgreSQL group
	if has("postgresql_conf", "remote_postgresql", "postgresql") {
		links := []NavLink{
			{"/postgresql", "Databases", path == "/postgresql", ""},
			{"/postgresql/users", "Users", path == "/postgresql/users", ""},
		}
		if allowed["pgadmin"] {
			links = append(links, NavLink{"/postgresql/pgadmin", "pgAdmin", path == "/postgresql/pgadmin", ""})
		}
		links = append(links,
			NavLink{"/postgresql/wizard", "Database Wizard", path == "/postgresql/wizard", ""},
			NavLink{"/postgresql/new", "Create Database", path == "/postgresql/new", ""},
			NavLink{"/postgresql/user", "Create User", path == "/postgresql/user", ""},
			NavLink{"/postgresql/assign", "Assign User to DB", path == "/postgresql/assign", ""},
			NavLink{"/postgresql/remove", "Remove User from DB", path == "/postgresql/remove", ""},
		)
		if allowed["import_postgresql"] {
			links = append(links, NavLink{"/postgresql/import", "Import Database", strings.HasPrefix(path, "/postgresql/import"), ""})
		}
		links = append(links,
			NavLink{"/postgresql/remote-postgresql", "Remote Access", path == "/postgresql/remote-postgresql", ""},
			NavLink{"/postgresql/processlist", "Show Processes", path == "/postgresql/processlist", ""},
		)
		if allowed["postgresql_conf"] {
			links = append(links, NavLink{"/postgresql/configuration", "Configuration", path == "/postgresql/configuration", ""})
		}
		open := hasAnyPrefix(path, "/postgresql", "/database")
		groups = append(groups, NavGroup{"PostgreSQL", postgresqlIcon, "postgresql-menu", links, open, open})
	}

	// Domains group
	if allowed["domains"] {
		links := []NavLink{
			{"/domains", "Domain Names", path == "/domains", ""},
			{"/domains/new", "Add New Domain", path == "/domains/new", ""},
		}
		if allowed["redirects"] {
			links = append(links, NavLink{"/domains/redirect", "Redirects", path == "/domains/redirect", ""})
		}
		if allowed["dns"] {
			links = append(links, NavLink{"/domains/edit-dns-zone", "DNS Zone Editor", strings.HasPrefix(path, "/domains/edit-dns-zone"), ""})
		}
		if allowed["dynamic_dns"] {
			links = append(links, NavLink{"/domains/dynamic-dns", "Dynamic DNS", strings.HasPrefix(path, "/domains/dynamic-dns"), ""})
		}
		if allowed["ssl"] {
			links = append(links, NavLink{"/domains/ssl", "SSL", path == "/domains/ssl", ""})
		}
		if allowed["edit_vhost"] {
			links = append(links, NavLink{"/domains/vhosts", "VHosts File Editor", path == "/domains/vhosts", ""})
		}
		if allowed["domain_suspend"] {
			links = append(links,
				NavLink{"/domains/suspend", "Suspend a Domain", path == "/domains/suspend", ""},
				NavLink{"/domains/unsuspend", "Unsuspend a Domain", path == "/domains/unsuspend", ""},
			)
		}
		if allowed["docroot"] {
			links = append(links, NavLink{"/domains/docroot", "Change docroot", path == "/domains/docroot", ""})
		}
		if allowed["domain_logs"] {
			links = append(links, NavLink{"/domains/log", "Raw Access Logs", strings.HasPrefix(path, "/domains/log"), ""})
		}
		if allowed["goaccess"] {
			links = append(links, NavLink{"/domains/stats", "GoAccess", path == "/domains/stats", ""})
		}
		open := strings.HasPrefix(path, "/domains")
		groups = append(groups, NavGroup{"Domains", domainsIcon, "domains-menu", links, open, open})
	}

	// Emails group
	if has("emails", "email_filters", "email_aliases", "email_default", "email_import", "email_deliverability", "webmail") {
		var links []NavLink
		if allowed["emails"] {
			links = append(links,
				NavLink{"/emails", "Email Accounts", path == "/emails", ""},
				NavLink{"/emails/new", "Create New Account", strings.HasPrefix(path, "/emails/new"), ""},
			)
		}
		if allowed["webmail"] {
			links = append(links, NavLink{"/webmail/", "Webmail", false, "_blank"})
		}
		if allowed["email_filters"] {
			links = append(links, NavLink{"/emails/filter", "Filters", strings.HasPrefix(path, "/emails/filter"), ""})
		}
		if allowed["email_aliases"] {
			links = append(links, NavLink{"/emails/aliases", "Aliases", strings.HasPrefix(path, "/emails/aliases"), ""})
		}
		if allowed["email_default"] {
			links = append(links, NavLink{"/emails/default", "Default Address", strings.HasPrefix(path, "/emails/default"), ""})
		}
		if allowed["email_import"] {
			links = append(links, NavLink{"/emails/import", "Address Importer", strings.HasPrefix(path, "/emails/import"), ""})
		}
		if allowed["email_deliverability"] {
			links = append(links, NavLink{"/emails/deliverability", "Email Deliverability", strings.HasPrefix(path, "/emails/deliverability"), ""})
		}
		if allowed["emails"] {
			links = append(links, NavLink{"/emails/delete", "Delete Accounts", strings.HasPrefix(path, "/emails/delete"), ""})
		}
		open := strings.HasPrefix(path, "/email")
		groups = append(groups, NavGroup{"Emails", emailIcon, "emails-menu", links, open, open})
	}

	// Caching group
	if has("redis", "valkey", "memcached", "varnish", "elasticsearch", "opensearch") {
		var links []NavLink
		if allowed["redis"] {
			links = append(links, NavLink{"/cache/redis", "Redis", path == "/cache/redis", ""})
		}
		if allowed["valkey"] {
			links = append(links, NavLink{"/cache/valkey", "Valkey", path == "/cache/valkey", ""})
		}
		if allowed["memcached"] {
			links = append(links, NavLink{"/cache/memcached", "Memcached", path == "/cache/memcached", ""})
		}
		if allowed["opensearch"] {
			links = append(links, NavLink{"/cache/opensearch", "Opensearch", path == "/cache/opensearch", ""})
		}
		if allowed["elasticsearch"] {
			links = append(links, NavLink{"/cache/elasticsearch", "Elasticsearch", path == "/cache/elasticsearch", ""})
		}
		if allowed["varnish"] {
			links = append(links, NavLink{"/cache/varnish", "Varnish", path == "/cache/varnish", ""})
		}
		open := strings.HasPrefix(path, "/cache")
		groups = append(groups, NavGroup{"Caching", cacheIcon, "cache-menu", links, open, open})
	}

	// PHP group
	if has("php", "php_options", "php_ini", "php_extensions") {
		links := []NavLink{
			{"/php/domains", "Select PHP version", path == "/php/domains", ""},
			{"/php/default", "Default version", path == "/php/default", ""},
		}
		if allowed["php_options"] {
			links = append(links, NavLink{"/php/options", "PHP Options", strings.HasPrefix(path, "/php") && strings.Contains(path, "/options"), ""})
		}
		if allowed["php_extensions"] {
			links = append(links, NavLink{"/php/extensions", "PHP Extensions", strings.HasPrefix(path, "/php") && strings.Contains(path, "/extensions"), ""})
		}
		if allowed["php_ini"] {
			links = append(links, NavLink{"/php/php_ini_editor", "PHP.INI Editor", strings.HasPrefix(path, "/php") && strings.Contains(path, "/php_ini_editor"), ""})
		}
		open := strings.HasPrefix(path, "/php")
		groups = append(groups, NavGroup{"PHP", phpIcon, "php-menu", links, open, open})
	}

	// Advanced group
	if has("crons", "services", "ssh", "usage", "process_manager", "webserver_conf", "timezone", "waf", "ip_blocker", "info") {
		var links []NavLink
		if allowed["services"] {
			links = append(links, NavLink{"/services", "Services", strings.HasPrefix(path, "/services"), ""})
		}
		if allowed["crons"] {
			links = append(links, NavLink{"/cronjobs", "Cron Jobs", strings.HasPrefix(path, "/cronjobs"), ""})
		}
		if allowed["ip_blocker"] {
			links = append(links, NavLink{"/security/ip-blocker", "IP Blocker", path == "/security/ip-blocker", ""})
		}
		if allowed["process_manager"] {
			links = append(links, NavLink{"/process-manager", "Process Manager", path == "/process-manager", ""})
		}
		if allowed["webserver_conf"] {
			links = append(links, NavLink{"/server/webserver_conf", "WebServer Settings", path == "/server/webserver_conf", ""})
		}
		if allowed["waf"] {
			links = append(links, NavLink{"/server/waf", "WAF", strings.HasPrefix(path, "/server/waf"), ""})
		}
		if allowed["usage"] {
			links = append(links, NavLink{"/server/usage", "Resource Usage", strings.HasPrefix(path, "/server/usage"), ""})
		}
		if allowed["info"] {
			links = append(links, NavLink{"/server/info", "Server Information", path == "/server/info", ""})
		}
		open := hasAnyPrefix(path, "/cronjobs", "/services/", "/server", "/process-manager", "/server/usage", "/security/ip-blocker")
		active := hasAnyPrefix(path, "/cronjobs", "/server", "/process-manager", "/server/usage", "/security/ip-blocker")
		groups = append(groups, NavGroup{"Advanced", advancedIcon, "advanced-menu", links, open, active})
	}

	// Docker group
	if allowed["docker"] {
		links := []NavLink{
			{"/containers", "Containers", path == "/containers" || path == "/containers/new" || strings.HasPrefix(path, "/containers/edit"), ""},
			{"/containers/terminal", "Terminal", strings.HasPrefix(path, "/containers/terminal"), ""},
			{"/containers/logs", "Logs", strings.HasPrefix(path, "/containers/logs"), ""},
			{"/containers/image/", "Image Updates", path == "/containers/image/", ""},
			{"/containers/image/change", "Change image tag", path == "/containers/image/change", ""},
			{"/containers/webserver", "Switch WebServer", path == "/containers/webserver", ""},
			{"/containers/mysql", "Switch MySQL Type", path == "/containers/mysql", ""},
		}
		open := strings.HasPrefix(path, "/containers")
		groups = append(groups, NavGroup{"Containers", dockerIcon, "docker-menu", links, open, open})
	}

	// Account group
	if has("account", "twofa", "passkeys", "favorites", "login_history", "notifications", "locale", "sessions", "activity", "mcp") {
		var links []NavLink
		if allowed["account"] {
			links = append(links, NavLink{"/account", "Email & Password", path == "/account", ""})
		}
		if allowed["locale"] {
			links = append(links, NavLink{"/account/language", "Change Language", strings.HasPrefix(path, "/account/language"), ""})
		}
		if allowed["notifications"] {
			links = append(links, NavLink{"/account/notifications", "Email Notifications", strings.HasPrefix(path, "/account/notifications"), ""})
		}
		if allowed["twofa"] {
			links = append(links, NavLink{"/account/2fa", "2FA", strings.HasPrefix(path, "/account/2fa"), ""})
		}
		if allowed["passkeys"] {
			links = append(links, NavLink{"/account/passkeys", "Passkeys", strings.HasPrefix(path, "/account/passkeys"), ""})
		}
		if allowed["sessions"] {
			links = append(links, NavLink{"/account/sessions", "Active Sessions", strings.HasPrefix(path, "/account/sessions"), ""})
		}
		if allowed["favorites"] {
			links = append(links, NavLink{"/account/favorites", "Favorite Pages", strings.HasPrefix(path, "/account/favorites"), ""})
		}
		if allowed["activity"] {
			links = append(links, NavLink{"/account/activity", "Account Activity", strings.HasPrefix(path, "/account/activity"), ""})
		}
		if allowed["login_history"] {
			links = append(links, NavLink{"/account/login-history", "Login History", strings.HasPrefix(path, "/account/login-history"), ""})
		}
		if allowed["api"] {
			links = append(links, NavLink{"/account/api", "API Reference", strings.HasPrefix(path, "/account/api"), ""})
		}
		if allowed["mcp"] {
			links = append(links, NavLink{"/account/mcp", "MCP", strings.HasPrefix(path, "/account/mcp"), ""})
		}
		open := strings.HasPrefix(path, "/account")
		groups = append(groups, NavGroup{"Account", accountIcon, "account-menu", links, open, open})
	}

	return groups
}
