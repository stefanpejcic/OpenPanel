package dashboard

import (
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// SectionItem mirrors one icon-link dict in dashboard.html's `sections` Jinja literal (e.g. {"key": "filemanager", "href": "/files", ...})
type SectionItem struct {
	Key    string
	Href   string
	Icon   string
	Label  string
	Target string

	// Disabled marks an item for a feature the current plan doesn't grant but the configured upsell plan does - rendered greyed-out with an upgrade prompt instead of being omitted entirely.
	Disabled bool
}

// Section mirrors one entry of dashboard.html's `sections` dict plus its title from `section_titles`
type Section struct {
	Key   string
	Title string
	Icon  string // only set on modern sections, shown next to the title
	Items []SectionItem
}

// buildDashboardSections builds the dashboard's section/item list, keeping only items whose key is in allowed (or in upsellAllowed, marked Disabled), for dashboard.html's {{range .Sections}}. Section order matches this slice's literal order. Titles/Labels are run through t.Get so the icon grid picks up the same catalog as the sidebar, instead of staying hardcoded English.
func buildDashboardSections(t i18n.Translator, allowed, upsellAllowed map[string]bool, menuStyle string, appKeys []string) []Section {
	all := classicSections()
	if menuStyle == "modern" {
		all = allSections()
	}
	// the classic theme draws Remix icons, Modern maps the bi- names through dash_icon
	if web.Theme() == web.ThemeClassic {
		planner := plannerClassicSections()
		if menuStyle == "modern" {
			planner = plannerSections()
		}
		all = withPlannerIcons(all, planner)
	}
	if len(appKeys) > 0 {
		for i := range all {
			if all[i].Key == "websites" {
				all[i].Items = websitesItems(all[i].Items, appKeys)
			}
		}
	}

	// websites is a main module so it's always allowed, gate the Sites tile on the apps like the sidebar does
	sitesAllowed, sitesUpsell := hasAny(allowed, web.WebsiteAppKeys), hasAny(upsellAllowed, web.WebsiteAppKeys)

	result := make([]Section, 0, len(all))
	for _, s := range all {
		var items []SectionItem
		for _, item := range s.Items {
			item.Label = t.Get(item.Label)
			ok, upsell := allowed[item.Key], upsellAllowed[item.Key]
			if item.Key == "websites" {
				ok, upsell = sitesAllowed, sitesUpsell
			}
			if ok {
				items = append(items, item)
			} else if upsell {
				item.Disabled = true
				items = append(items, item)
			}
		}
		if len(items) > 0 {
			result = append(result, Section{Key: s.Key, Title: t.Get(s.Title), Icon: s.Icon, Items: items})
		}
	}
	return result
}

// allSections is the modern (menu_style=modern) dashboard: sections and icons follow the sidebar's areas and tab order, single-page areas (Cron Jobs, Server Info) join a neighbouring section instead of standing alone, keys stay the old ones where a section survived so saved orders and custom_dashboard_section.json's before_<key> keep working
func allSections() []Section {
	return []Section{
		// mautic/flarum omitted: legacy code slated for removal entirely, not ported here (per user decision)
		{Key: "websites", Title: "Websites", Icon: "bi-window-stack", Items: []SectionItem{
			{"websites", "/sites", "bi-app-indicator", "Sites", "", false},
			{"wordpress", "/wordpress", "bi-wordpress", "WordPress", "", false},
			{"autoinstaller", "/auto-installer", "bi-download", "Install App", "", false},
			{"git", "/git", "bi-git", "Git", "", false},
		}},
		{Key: "domains", Title: "Domains", Icon: "bi-globe", Items: []SectionItem{
			{"domains", "/domains", "bi-globe", "Domains", "", false},
			{"dns", "/domains/edit-dns-zone", "bi-file-text", "DNS Zone Editor", "", false},
			{"dynamic_dns", "/domains/dynamic-dns", "bi-arrow-repeat", "Dynamic DNS", "", false},
			{"ssl", "/domains/ssl", "bi-lock", "SSL Certificates", "", false},
			{"redirects", "/domains/redirect", "bi-link-45deg", "Redirects", "", false},
		}},
		{Key: "emails", Title: "Email", Icon: "bi-envelope", Items: []SectionItem{
			{"emails", "/emails", "bi-envelope", "Email Accounts", "", false},
			{"email_aliases", "/emails/aliases", "bi-at", "Aliases", "", false},
			{"email_default", "/emails/default", "bi-envelope-at", "Catch-all Address", "", false},
			{"email_filters", "/emails/filter", "bi-funnel", "Filters", "", false},
			{"email_deliverability", "/emails/deliverability", "bi-envelope-check", "Deliverability", "", false},
			{"email_import", "/emails/import", "bi-envelope-arrow-up", "Import", "", false},
			{"email_migrate", "/emails/migrate", "bi-arrow-left-right", "Migrate", "", false},
			{"email_export", "/emails/export", "bi-envelope-arrow-down", "Export", "", false},
			{"webmail", "/webmail/", "bi-box-arrow-up-right", "Webmail", "_blank", false},
		}},
		{Key: "files", Title: "Files", Icon: "bi-folder", Items: []SectionItem{
			{"filemanager", "/files", "bi-folder-fill", "File Manager", "", false},
			{"filemanager", "/file-manager/upload?method=upload", "bi-upload", "Upload", "", false},
			{"filemanager", "/file-manager/upload?method=download", "bi-download", "Download from URL", "", false},
			{"trash", "/files.trash", "bi-trash3", "Trash", "", false},
			{"ftp", "/ftp", "bi-folder-symlink-fill", "FTP Accounts", "", false},
			{"ftp", "/ftp/connections", "bi-folder-symlink", "FTP Connections", "", false},
			{"fix_permissions", "/fix-permissions", "bi-file-binary-fill", "Fix Permissions", "", false},
		}},
		{Key: "backups", Title: "Backups", Icon: "bi-cloud-check", Items: []SectionItem{
			{"backups", "/backups", "bi-folder-check", "Backups", "", false},
			{"backups", "/backups/list", "bi-arrow-counterclockwise", "Restore", "", false},
			{"backups", "/backups/settings", "bi-gear", "Configuration", "", false},
			{"backups", "/backups/destination", "bi-cloud-upload", "Destination", "", false},
			{"backup_wizard", "/backup-wizard", "bi-cloud-arrow-down", "Backup Wizard", "", false},
		}},
		{Key: "mysql", Title: "MySQL", Icon: "bi-database", Items: []SectionItem{
			{"mysql", "/mysql", "bi-database", "Databases", "", false},
			{"mysql", "/mysql/users", "bi-people", "Users", "", false},
			{"mysql", "/mysql/wizard", "bi-database-add", "Database Wizard", "", false},
			{"mysql_import", "/mysql/import", "bi-database-fill-add", "Import", "", false},
			{"phpmyadmin", "/phpmyadmin", "bi-box-arrow-up-right", "phpMyAdmin", "_blank", false},
			{"phpmyadmin", "/phpmyadmin/link", "bi-box-arrow-in-right", "phpMyAdmin Login Form", "_blank", false},
			{"remote_mysql", "/mysql/remote-mysql", "bi-hdd-network", "Remote Access", "", false},
			{"mysql_processlist", "/mysql/processlist", "bi-database-slash", "Running Queries", "", false},
			{"mysql_conf", "/mysql/configuration", "bi-database-lock", "Configuration", "", false},
			{"mysql_root_password", "/mysql/root-password", "bi-key-fill", "Root Password", "", false},
			{"change_db", "/containers/mysql", "bi-toggle2-off", "Server Type", "", false},
		}},
		{Key: "postgresql", Title: "PostgreSQL", Icon: "bi-database", Items: []SectionItem{
			{"postgresql", "/postgresql", "bi-database", "Databases", "", false},
			{"postgresql", "/postgresql/users", "bi-people", "Users", "", false},
			{"postgresql", "/postgresql/wizard", "bi-database-add", "Database Wizard", "", false},
			{"postgresql_import", "/postgresql/import", "bi-database-fill-add", "Import", "", false},
			{"remote_postgresql", "/postgresql/remote-postgresql", "bi-diagram-3", "Remote Access", "", false},
			{"postgresql", "/postgresql/processlist", "bi-database-slash", "Running Queries", "", false},
			{"postgresql_conf", "/postgresql/configuration", "bi-database-lock", "Configuration", "", false},
		}},
		{Key: "mongodb", Title: "MongoDB", Icon: "bi-database", Items: []SectionItem{
			{"mongodb", "/mongodb", "bi-database", "Databases", "", false},
			{"mongodb", "/mongodb/users", "bi-people", "Users", "", false},
			{"mongodb", "/mongodb/wizard", "bi-database-add", "Database Wizard", "", false},
			{"mongodb_import", "/mongodb/import", "bi-database-fill-add", "Import", "", false},
			{"mongodb", "/mongodb/processlist", "bi-database-slash", "Running Queries", "", false},
		}},
		{Key: "php", Title: "PHP", Icon: "bi-filetype-php", Items: []SectionItem{
			{"php", "/php/domains", "bi-code-square", "Version per Domain", "", false},
			{"php", "/php/default", "bi-filetype-php", "Default Version", "", false},
			{"php_options", "/php/options", "bi-toggles", "Options", "", false},
			{"php_extensions", "/php/extensions", "bi-puzzle", "Extensions", "", false},
			{"php_ini", "/php/php_ini_editor", "bi-filetype-php", "php.ini Editor", "", false},
		}},
		{Key: "cache", Title: "Cache & Search", Icon: "bi-lightning-charge", Items: []SectionItem{
			{"redis", "/cache/redis", "bi-database-fill-lock", "Redis", "", false},
			{"valkey", "/cache/valkey", "bi-database-fill-lock", "Valkey", "", false},
			{"memcached", "/cache/memcached", "bi-hdd-network-fill", "Memcached", "", false},
			{"varnish", "/cache/varnish", "bi-lightning-charge-fill", "Varnish", "", false},
			{"opensearch", "/cache/opensearch", "bi-search", "OpenSearch", "", false},
			{"elasticsearch", "/cache/elasticsearch", "bi-search-heart", "Elasticsearch", "", false},
		}},
		{Key: "webserver", Title: "Web Server Config", Icon: "bi-hdd-network", Items: []SectionItem{
			{"webserver_conf", "/server/webserver_conf", "bi-hdd-network", "Server Settings", "", false},
			{"edit_vhost", "/domains/vhosts", "bi-file-text", "Domain VHosts", "", false},
			{"change_ws", "/containers/webserver", "bi-toggle2-on", "Web Server Type", "", false},
		}},
		{Key: "docker", Title: "Containers", Icon: "bi-boxes", Items: []SectionItem{
			{"docker", "/containers", "bi-boxes", "Containers", "", false},
			{"terminal", "/containers/terminal", "bi-terminal", "Terminal", "", false},
			{"docker", "/containers/logs", "bi-file-binary", "Logs", "", false},
			{"change_image", "/containers/image/change", "bi-textarea-t", "Software Versions", "", false},
			{"crons", "/cronjobs", "bi-calendar2-week", "Cron Jobs", "", false},
			{"timezone", "/server/timezone", "bi-clock", "Change TimeZone", "", false},
		}},
		{Key: "advanced", Title: "Statistics", Icon: "bi-bar-chart", Items: []SectionItem{
			{"usage", "/server/usage", "bi-speedometer2", "Resource Usage", "", false},
			{"usage", "/server/usage/history", "bi-speedometer2", "Resource Usage History", "", false},
			{"disk_usage", "/disk-usage/", "bi-folder-plus", "Disk Usage", "", false},
			{"inodes", "/inodes-explorer", "bi-folder-x", "Inode Usage", "", false},
			{"summary", "/summary", "bi-journal-text", "Daily Summary", "", false},
			{"goaccess", "/domains/stats", "bi-graph-up", "Visitor Statistics", "", false},
			{"domain_logs", "/domains/log", "bi-file-text", "Access Logs", "", false},
			{"error_logs", "/error-logs", "bi-bug", "Error Logs", "", false},
			{"info", "/server/info", "bi-info-square", "Server Info", "", false},
		}},
		{Key: "processes", Title: "Processes & Services", Icon: "bi-cpu", Items: []SectionItem{
			{"services", "/services", "bi-hdd-stack", "Services", "", false},
			{"process_manager", "/process-manager", "bi-cpu", "Processes", "", false},
		}},
		{Key: "account", Title: "Account", Icon: "bi-person-circle", Items: []SectionItem{
			{"account", "/account", "bi-person-gear", "Login Details", "", false},
			{"locale", "/account/language", "bi-translate", "Language", "", false},
			{"notifications", "/account/notifications", "bi-bell", "Notifications", "", false},
			{"favorites", "/account/favorites", "bi-star", "Favorites", "", false},
			{"api", "/account/api", "bi-braces", "API", "", false},
			{"mcp", "/account/mcp", "bi-robot", "AI Assistant (MCP)", "", false},
			{"logout", "/logout", "bi-door-open", "Log out", "", false},
		}},
		{Key: "security", Title: "Security", Icon: "bi-shield-lock", Items: []SectionItem{
			{"waf", "/server/waf", "bi-shield-lock", "Web Firewall", "", false},
			{"waf", "/server/waf/log", "bi-shield-exclamation", "Web Firewall Logs", "", false},
			{"ip_blocker", "/security/ip-blocker", "bi-ban", "IP Blocker", "", false},
			{"malware_scan", "/malware-scanner", "bi-upc-scan", "Malware Scanner", "", false},
			{"twofa", "/account/2fa", "bi-fingerprint", "Two-Factor Auth", "", false},
			{"passkeys", "/account/passkeys", "bi-key", "Passkeys", "", false},
			{"sessions", "/account/sessions", "bi-people", "Active Sessions", "", false},
			{"activity", "/account/activity", "bi-activity", "Activity Log", "", false},
			{"login_history", "/account/login-history", "bi-person-exclamation", "Login History", "", false},
		}},
	}
}

// classicSections is the dashboard for menu_style=classic, grouped like the classic sidebar
func classicSections() []Section {
	return []Section{
		{Key: "files", Title: "Files", Items: []SectionItem{
			{"filemanager", "/files", "bi-folder-fill", "File Manager", "", false},
			{"filemanager", "/file-manager/upload?method=upload", "bi-upload", "File Upload", "", false},
			{"filemanager", "/file-manager/upload?method=download", "bi-download", "Download Files", "", false},
			{"ftp", "/ftp", "bi-folder-symlink-fill", "FTP Accounts", "", false},
			{"ftp", "/ftp/connections", "bi-folder-symlink", "FTP Connections", "", false},
			{"disk_usage", "/disk-usage/", "bi-folder-plus", "Disk Usage", "", false},
			{"backups", "/backups", "bi-folder-check", "Backups", "", false},
			{"backup_wizard", "/backup-wizard", "bi-cloud-arrow-down", "Backup Wizard", "", false},
			{"inodes", "/inodes-explorer", "bi-folder-x", "Inodes Explorer", "", false},
			{"malware_scan", "/malware-scanner", "bi-upc-scan", "ClamAV Scanner", "", false},
			{"fix_permissions", "/fix-permissions", "bi-file-binary-fill", "Fix Permissions", "", false},
			{"trash", "/files.trash", "bi-trash3", "Trash", "", false},
		}},
		{Key: "domains", Title: "Domains", Items: []SectionItem{
			{"domains", "/domains", "bi-globe", "Domains", "", false},
			{"redirects", "/domains/redirect", "bi-link-45deg", "Redirects", "", false},
			{"dns", "/domains/edit-dns-zone", "bi-file-text", "DNS Zone Editor", "", false},
			{"dynamic_dns", "/domains/dynamic-dns", "bi-arrow-repeat", "Dynamic DNS", "", false},
			{"ssl", "/domains/ssl", "bi-lock", "SSL", "", false},
			{"domain_suspend", "/domains/suspend", "bi-ban", "Suspend Domain", "", false},
			{"domain_suspend", "/domains/unsuspend", "bi-ban-fill", "Unsuspend Domain", "", false},
			{"docroot", "/domains/docroot", "bi-folder2", "Change Docroot", "", false},
			{"edit_vhost", "/domains/vhosts", "bi-file-text", "Edit Virtual Hosts", "", false},
			{"goaccess", "/domains/stats", "bi-graph-up", "GoAccess", "", false},
			{"domain_logs", "/domains/log", "bi-file-text", "Raw Access Logs", "", false},
		}},
		{Key: "mysql", Title: "MySQL", Items: []SectionItem{
			{"mysql", "/mysql", "bi-database", "Databases", "", false},
			{"mysql", "/mysql/users", "bi-people", "Users", "", false},
			{"mysql", "/mysql/wizard", "bi-database-add", "Database Wizard", "", false},
			{"mysql_import", "/mysql/import", "bi-database-fill-add", "Import Database", "", false},
			{"phpmyadmin", "/phpmyadmin", "bi-box-arrow-up-right", "Open phpMyAdmin", "_blank", false},
			{"phpmyadmin", "/phpmyadmin/link", "bi-box-arrow-in-right", "phpMyAdmin Login Form", "_blank", false},
			{"mysql_processlist", "/mysql/processlist", "bi-database-slash", "Process List", "", false},
			{"remote_mysql", "/mysql/remote-mysql", "bi-hdd-network", "Remote Access", "", false},
			{"mysql_conf", "/mysql/configuration", "bi-database-lock", "MySQL Configuration", "", false},
			{"mysql_root_password", "/mysql/root-password", "bi-key-fill", "Change root password", "", false},
		}},
		{Key: "postgresql", Title: "PostgreSQL", Items: []SectionItem{
			{"postgresql", "/postgresql", "bi-database", "Databases", "", false},
			{"postgresql", "/postgresql/users", "bi-people", "Users", "", false},
			{"postgresql", "/postgresql/wizard", "bi-database-add", "Database Wizard", "", false},
			{"postgresql_import", "/postgresql/import", "bi-database-fill-add", "Import Database", "", false},
			{"postgresql", "/postgresql/processlist", "bi-database-slash", "Process List", "", false},
			{"remote_postgresql", "/postgresql/remote-postgresql", "bi-diagram-3", "Remote Access", "", false},
			{"postgresql_conf", "/postgresql/configuration", "bi-database-lock", "PostgreSQL Configuration", "", false},
		}},
		{Key: "mongodb", Title: "MongoDB", Items: []SectionItem{
			{"mongodb", "/mongodb", "bi-database", "Databases", "", false},
			{"mongodb", "/mongodb/users", "bi-people", "Users", "", false},
			{"mongodb", "/mongodb/wizard", "bi-database-add", "Database Wizard", "", false},
			{"mongodb_import", "/mongodb/import", "bi-database-fill-add", "Import Database", "", false},
			{"mongodb", "/mongodb/processlist", "bi-database-slash", "Process List", "", false},
		}},
		// mautic/flarum omitted: legacy code slated for removal entirely, not ported here (per user decision)
		{Key: "websites", Title: "Websites", Items: []SectionItem{
			{"websites", "/sites", "bi-app-indicator", "Site Manager", "", false},
			{"autoinstaller", "/auto-installer", "bi-download", "Auto Installer", "", false},
			{"wordpress", "/wordpress", "bi-wordpress", "WP Manager", "", false},
			{"git", "/git", "bi-git", "Git", "", false},
		}},
		{Key: "cache", Title: "Cache", Items: []SectionItem{
			{"redis", "/cache/redis", "bi-database-fill-lock", "Redis", "", false},
			{"valkey", "/cache/valkey", "bi-database-fill-lock", "Valkey", "", false},
			{"memcached", "/cache/memcached", "bi-hdd-network-fill", "Memcached", "", false},
			{"varnish", "/cache/varnish", "bi-lightning-charge-fill", "Varnish", "", false},
			{"opensearch", "/cache/opensearch", "bi-search", "OpenSearch", "", false},
			{"elasticsearch", "/cache/elasticsearch", "bi-search-heart", "ElasticSearch", "", false},
		}},
		{Key: "emails", Title: "Emails", Items: []SectionItem{
			{"emails", "/emails", "bi-envelope", "Email Accounts", "", false},
			{"email_aliases", "/emails/aliases", "bi-at", "Aliases", "", false},
			{"email_filters", "/emails/filter", "bi-funnel", "Filters", "", false},
			{"email_deliverability", "/emails/deliverability", "bi-envelope-check", "Email Deliverability", "", false},
			{"email_default", "/emails/default", "bi-envelope-at", "Default Address", "", false},
			{"email_import", "/emails/import", "bi-envelope-arrow-up", "Address Importer", "", false},
			{"email_migrate", "/emails/migrate", "bi-arrow-left-right", "Migrate Emails", "", false},
			{"email_export", "/emails/export", "bi-envelope-arrow-down", "Address Exporter", "", false},
			{"webmail", "/webmail/", "bi-box-arrow-up-right", "Webmail", "_blank", false},
		}},
		{Key: "php", Title: "PHP", Items: []SectionItem{
			{"php", "/php/domains", "bi-code-square", "Select PHP version", "", false},
			{"php", "/php/default", "bi-filetype-php", "Default Version", "", false},
			{"php_options", "/php/options", "bi-toggles", "PHP Options", "", false},
			{"php_extensions", "/php/extensions", "bi-puzzle", "PHP Extensions", "", false},
			{"php_ini", "/php/php_ini_editor", "bi-filetype-php", "PHP.INI Editor", "", false},
		}},
		{Key: "docker", Title: "Containers", Items: []SectionItem{
			{"docker", "/containers", "bi-boxes", "Containers", "", false},
			{"terminal", "/containers/terminal", "bi-terminal", "Terminal", "", false},
			{"docker", "/containers/logs", "bi-file-binary", "Logs", "", false},
			{"change_image", "/containers/image/change", "bi-textarea-t", "Change Image tag", "", false},
			{"change_ws", "/containers/webserver", "bi-toggle2-on", "Change webserver", "", false},
			{"change_db", "/containers/mysql", "bi-toggle2-off", "Change MySQL Type", "", false},
		}},
		{Key: "advanced", Title: "Advanced", Items: []SectionItem{
			{"services", "/services", "bi-hdd-stack", "Services", "", false},
			{"crons", "/cronjobs", "bi-calendar2-week", "Cron Jobs", "", false},
			{"ip_blocker", "/security/ip-blocker", "bi-ban", "IP Blocker", "", false},
			{"usage", "/server/usage", "bi-speedometer2", "Resource Usage", "", false},
			{"usage", "/server/usage/history", "bi-speedometer2", "Resource Usage History", "", false},
			{"error_logs", "/error-logs", "bi-bug", "Error Logs", "", false},
			{"summary", "/summary", "bi-journal-text", "Daily Summary", "", false},
			{"process_manager", "/process-manager", "bi-cpu", "Process Manager", "", false},
			{"timezone", "/server/timezone", "bi-clock", "Change TimeZone", "", false},
			{"webserver_conf", "/server/webserver_conf", "bi-hdd-network", "Webserver Configuration", "", false},
			{"waf", "/server/waf", "bi-shield-lock", "WAF Settings", "", false},
			{"waf", "/server/waf/log", "bi-shield-exclamation", "WAF Logs", "", false},
			{"info", "/server/info", "bi-info-square", "Server Information", "", false},
		}},
		{Key: "account", Title: "Account", Items: []SectionItem{
			{"account", "/account", "bi-person-gear", "Email & Password", "", false},
			{"locale", "/account/language", "bi-translate", "Change Language", "", false},
			{"notifications", "/account/notifications", "bi-bell", "Email Notifications", "", false},
			{"twofa", "/account/2fa", "bi-fingerprint", "Two-Factor Authentication", "", false},
			{"sessions", "/account/sessions", "bi-people", "Active Sessions", "", false},
			{"favorites", "/account/favorites", "bi-star", "Favorite Pages", "", false},
			{"activity", "/account/activity", "bi-activity", "Account Activity", "", false},
			{"login_history", "/account/login-history", "bi-person-exclamation", "Login History", "", false},
			{"logout", "/logout", "bi-door-open", "Log out", "", false},
		}},
	}
}

// appItems are the apps an admin can pin to the Websites section via applications_dashboard_items, keyed by module name
var appItems = map[string]SectionItem{
	"joomla":           {"joomla", "/joomla/install", "bi-grid-3x3-gap", "Joomla", "", false},
	"drupal":           {"drupal", "/drupal/install", "bi-droplet", "Drupal", "", false},
	"website_builder":  {"website_builder", "/website-builder/install", "bi-brush", "Website Builder", "", false},
	"prestashop":       {"prestashop", "/prestashop/install", "bi-cart", "PrestaShop", "", false},
	"opencart":         {"opencart", "/opencart/install", "bi-cart3", "OpenCart", "", false},
	"phpbb":            {"phpbb", "/phpbb/install", "bi-chat-square-text", "phpBB", "", false},
	"flarum":           {"flarum", "/flarum/install", "bi-chat-dots", "Flarum", "", false},
	"mediawiki":        {"mediawiki", "/mediawiki/install", "bi-book", "MediaWiki", "", false},
	"dokuwiki":         {"dokuwiki", "/dokuwiki/install", "bi-journal-text", "DokuWiki", "", false},
	"sofawiki":         {"sofawiki", "/sofawiki/install", "bi-journal", "SofaWiki", "", false},
	"matomo":           {"matomo", "/matomo/install", "bi-bar-chart-line", "Matomo", "", false},
	"nextcloud":        {"nextcloud", "/nextcloud/install", "bi-cloud", "Nextcloud", "", false},
	"tinyphotogallery": {"tinyphotogallery", "/tinyphotogallery/install", "bi-images", "TinyPhotoGallery", "", false},
	"tinyfilemanager":  {"tinyfilemanager", "/tinyfilemanager/install", "bi-folder2-open", "TinyFileManager", "", false},
	"moodle":           {"moodle", "/moodle/install", "bi-mortarboard", "Moodle", "", false},
	"ojs":              {"ojs", "/ojs/install", "bi-journal-bookmark", "OJS", "", false},
	"nodejs":           {"nodejs", "/nodejs/install", "bi-filetype-js", "Node.js", "", false},
	"python":           {"python", "/python/install", "bi-filetype-py", "Python", "", false},
	"ruby":             {"ruby", "/ruby/install", "bi-filetype-rb", "Ruby", "", false},
	"java":             {"java", "/java/install", "bi-filetype-java", "Java", "", false},
	"php":              {"php", "/php/install", "bi-filetype-php", "PHP", "", false},
	"docker":           {"docker", "/containers/new", "bi-box-seam", "Docker", "", false},
}

// websitesItems swaps the Websites section items for the admin's list, reusing the section's own sites/wordpress/autoinstaller entries so their labels match the menu style
func websitesItems(defaults []SectionItem, appKeys []string) []SectionItem {
	byKey := map[string]SectionItem{}
	for _, item := range defaults {
		byKey[item.Key] = item
	}
	var items []SectionItem
	for _, key := range appKeys {
		if item, ok := byKey[key]; ok {
			items = append(items, item)
		} else if item, ok := appItems[key]; ok {
			items = append(items, item)
		}
	}
	return items
}

func hasAny(m map[string]bool, keys []string) bool {
	for _, k := range keys {
		if m[k] {
			return true
		}
	}
	return false
}

// parseAppKeys reads applications_dashboard_items, accepting spaces or commas as separators
func parseAppKeys(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool { return r == ' ' || r == ',' })
}

// plannerSections is allSections for the classic theme
func plannerSections() []Section {
	return []Section{
		// mautic/flarum omitted: legacy code slated for removal entirely, not ported here (per user decision)
		{Key: "websites", Title: "Websites", Icon: "ri-window-2-line", Items: []SectionItem{
			{"websites", "/sites", "ri-notification-badge-line", "Sites", "", false},
			{"wordpress", "/wordpress", "ri-wordpress-line", "WordPress", "", false},
			{"autoinstaller", "/auto-installer", "ri-download-line", "Install app", "", false},
			{"git", "/git", "ri-git-branch-line", "Git", "", false},
		}},
		{Key: "domains", Title: "Domains", Icon: "ri-global-line", Items: []SectionItem{
			{"domains", "/domains", "ri-global-line", "Domains", "", false},
			{"dns", "/domains/edit-dns-zone", "ri-file-text-line", "DNS zone editor", "", false},
			{"dynamic_dns", "/domains/dynamic-dns", "ri-loop-right-line", "Dynamic DNS", "", false},
			{"ssl", "/domains/ssl", "ri-lock-line", "SSL certificates", "", false},
			{"redirects", "/domains/redirect", "ri-link", "Redirects", "", false},
		}},
		{Key: "emails", Title: "Email", Icon: "ri-mail-line", Items: []SectionItem{
			{"emails", "/emails", "ri-mail-line", "Email accounts", "", false},
			{"email_aliases", "/emails/aliases", "ri-at-line", "Aliases", "", false},
			{"email_default", "/emails/default", "ri-at-line", "Catch-all address", "", false},
			{"email_filters", "/emails/filter", "ri-filter-3-line", "Filters", "", false},
			{"email_deliverability", "/emails/deliverability", "ri-mail-check-line", "Deliverability", "", false},
			{"email_import", "/emails/import", "ri-mail-send-line", "Import", "", false},
			{"email_migrate", "/emails/migrate", "ri-arrow-left-right-line", "Migrate", "", false},
			{"email_export", "/emails/export", "ri-mail-download-line", "Export", "", false},
			{"webmail", "/webmail/", "ri-external-link-line", "Webmail", "_blank", false},
		}},
		{Key: "files", Title: "Files", Icon: "ri-folder-line", Items: []SectionItem{
			{"filemanager", "/files", "ri-folder-line", "File manager", "", false},
			{"filemanager", "/file-manager/upload?method=upload", "ri-upload-line", "Upload", "", false},
			{"filemanager", "/file-manager/upload?method=download", "ri-download-line", "Download from URL", "", false},
			{"trash", "/files.trash", "ri-delete-bin-line", "Trash", "", false},
			{"ftp", "/ftp", "ri-folder-shared-line", "FTP accounts", "", false},
			{"ftp", "/ftp/connections", "ri-folder-shared-line", "FTP connections", "", false},
			{"fix_permissions", "/fix-permissions", "ri-file-code-line", "Fix permissions", "", false},
		}},
		{Key: "backups", Title: "Backups", Icon: "ri-cloud-line", Items: []SectionItem{
			{"backups", "/backups", "ri-folder-check-line", "Backups", "", false},
			{"backups", "/backups/list", "ri-arrow-go-back-line", "Restore", "", false},
			{"backups", "/backups/settings", "ri-settings-3-line", "Configuration", "", false},
			{"backups", "/backups/destination", "ri-upload-cloud-2-line", "Destination", "", false},
			{"backup_wizard", "/backup-wizard", "ri-download-cloud-2-line", "Backup wizard", "", false},
		}},
		{Key: "mysql", Title: "MySQL", Icon: "ri-database-2-line", Items: []SectionItem{
			{"mysql", "/mysql", "ri-database-2-line", "Databases", "", false},
			{"mysql", "/mysql/users", "ri-group-line", "Users", "", false},
			{"mysql", "/mysql/wizard", "ri-database-2-line", "Database wizard", "", false},
			{"mysql_import", "/mysql/import", "ri-database-2-line", "Import", "", false},
			{"phpmyadmin", "/phpmyadmin", "ri-external-link-line", "phpMyAdmin", "_blank", false},
			{"phpmyadmin", "/phpmyadmin/link", "ri-login-box-line", "phpMyAdmin Login Form", "_blank", false},
			{"remote_mysql", "/mysql/remote-mysql", "ri-server-line", "Remote access", "", false},
			{"mysql_processlist", "/mysql/processlist", "ri-database-2-line", "Running queries", "", false},
			{"mysql_conf", "/mysql/configuration", "ri-database-2-line", "Configuration", "", false},
			{"mysql_root_password", "/mysql/root-password", "ri-key-2-line", "Root password", "", false},
			{"change_db", "/containers/mysql", "ri-toggle-line", "Server type", "", false},
		}},
		{Key: "postgresql", Title: "PostgreSQL", Icon: "ri-database-2-line", Items: []SectionItem{
			{"postgresql", "/postgresql", "ri-database-2-line", "Databases", "", false},
			{"postgresql", "/postgresql/users", "ri-group-line", "Users", "", false},
			{"postgresql", "/postgresql/wizard", "ri-database-2-line", "Database wizard", "", false},
			{"postgresql_import", "/postgresql/import", "ri-database-2-line", "Import", "", false},
			{"remote_postgresql", "/postgresql/remote-postgresql", "ri-organization-chart", "Remote access", "", false},
			{"postgresql", "/postgresql/processlist", "ri-database-2-line", "Running queries", "", false},
			{"postgresql_conf", "/postgresql/configuration", "ri-database-2-line", "Configuration", "", false},
		}},
		{Key: "mongodb", Title: "MongoDB", Icon: "ri-database-2-line", Items: []SectionItem{
			{"mongodb", "/mongodb", "ri-database-2-line", "Databases", "", false},
			{"mongodb", "/mongodb/users", "ri-group-line", "Users", "", false},
			{"mongodb", "/mongodb/wizard", "ri-database-2-line", "Database wizard", "", false},
			{"mongodb_import", "/mongodb/import", "ri-database-2-line", "Import", "", false},
			{"mongodb", "/mongodb/processlist", "ri-database-2-line", "Running queries", "", false},
		}},
		{Key: "php", Title: "PHP", Icon: "ri-php-line", Items: []SectionItem{
			{"php", "/php/domains", "ri-code-box-line", "Version per domain", "", false},
			{"php", "/php/default", "ri-code-s-slash-line", "Default version", "", false},
			{"php_options", "/php/options", "ri-toggle-line", "Options", "", false},
			{"php_extensions", "/php/extensions", "ri-puzzle-line", "Extensions", "", false},
			{"php_ini", "/php/php_ini_editor", "ri-code-s-slash-line", "php.ini Editor", "", false},
		}},
		{Key: "cache", Title: "Cache & search", Icon: "ri-flashlight-line", Items: []SectionItem{
			{"redis", "/cache/redis", "ri-database-2-line", "Redis", "", false},
			{"valkey", "/cache/valkey", "ri-database-2-line", "Valkey", "", false},
			{"memcached", "/cache/memcached", "ri-server-line", "Memcached", "", false},
			{"varnish", "/cache/varnish", "ri-flashlight-line", "Varnish", "", false},
			{"opensearch", "/cache/opensearch", "ri-search-line", "OpenSearch", "", false},
			{"elasticsearch", "/cache/elasticsearch", "ri-search-eye-line", "Elasticsearch", "", false},
		}},
		{Key: "webserver", Title: "Web server config", Icon: "ri-server-line", Items: []SectionItem{
			{"webserver_conf", "/server/webserver_conf", "ri-server-line", "Server settings", "", false},
			{"edit_vhost", "/domains/vhosts", "ri-file-text-line", "Domain VHosts", "", false},
			{"change_ws", "/containers/webserver", "ri-toggle-fill", "Web server type", "", false},
		}},
		{Key: "docker", Title: "Containers", Icon: "ri-stack-line", Items: []SectionItem{
			{"docker", "/containers", "ri-stack-line", "Containers", "", false},
			{"terminal", "/containers/terminal", "ri-terminal-box-line", "Terminal", "", false},
			{"docker", "/containers/logs", "ri-file-code-line", "Logs", "", false},
			{"change_image", "/containers/image/change", "ri-text-block", "Software versions", "", false},
			{"crons", "/cronjobs", "ri-calendar-line", "Cron jobs", "", false},
			{"timezone", "/server/timezone", "ri-time-line", "Change timezone", "", false},
		}},
		{Key: "account", Title: "Account", Icon: "ri-account-circle-line", Items: []SectionItem{
			{"account", "/account", "ri-user-settings-line", "Login details", "", false},
			{"locale", "/account/language", "ri-translate-2", "Language", "", false},
			{"notifications", "/account/notifications", "ri-notification-3-line", "Notifications", "", false},
			{"favorites", "/account/favorites", "ri-star-line", "Favorites", "", false},
			{"api", "/account/api", "ri-braces-line", "API", "", false},
			{"mcp", "/account/mcp", "ri-robot-2-line", "AI Assistant (MCP)", "", false},
			{"logout", "/logout", "ri-door-open-line", "Log out", "", false},
		}},
		{Key: "advanced", Title: "Statistics", Icon: "ri-bar-chart-line", Items: []SectionItem{
			{"usage", "/server/usage", "ri-dashboard-3-line", "Resource usage", "", false},
			{"usage", "/server/usage/history", "ri-dashboard-3-line", "Resource usage history", "", false},
			{"disk_usage", "/disk-usage/", "ri-folder-add-line", "Disk usage", "", false},
			{"inodes", "/inodes-explorer", "ri-folder-forbid-line", "Inode usage", "", false},
			{"summary", "/summary", "ri-file-list-3-line", "Daily summary", "", false},
			{"goaccess", "/domains/stats", "ri-line-chart-line", "Visitor statistics", "", false},
			{"domain_logs", "/domains/log", "ri-file-text-line", "Access logs", "", false},
			{"error_logs", "/error-logs", "ri-bug-line", "Error logs", "", false},
			{"info", "/server/info", "ri-information-line", "Server info", "", false},
		}},
		{Key: "processes", Title: "Processes & services", Icon: "ri-cpu-line", Items: []SectionItem{
			{"services", "/services", "ri-stack-line", "Services", "", false},
			{"process_manager", "/process-manager", "ri-cpu-line", "Processes", "", false},
		}},
		{Key: "security", Title: "Security", Icon: "ri-shield-keyhole-line", Items: []SectionItem{
			{"waf", "/server/waf", "ri-shield-keyhole-line", "Web firewall", "", false},
			{"waf", "/server/waf/log", "ri-error-warning-line", "Web firewall logs", "", false},
			{"ip_blocker", "/security/ip-blocker", "ri-forbid-line", "IP blocker", "", false},
			{"malware_scan", "/malware-scanner", "ri-barcode-line", "Malware scanner", "", false},
			{"twofa", "/account/2fa", "ri-fingerprint-line", "Two-factor auth", "", false},
			{"passkeys", "/account/passkeys", "ri-key-2-line", "Passkeys", "", false},
			{"sessions", "/account/sessions", "ri-group-line", "Active sessions", "", false},
			{"activity", "/account/activity", "ri-pulse-line", "Activity log", "", false},
			{"login_history", "/account/login-history", "ri-user-unfollow-line", "Login history", "", false},
		}},
	}
}

// plannerClassicSections is classicSections for the classic theme
func plannerClassicSections() []Section {
	return []Section{
		{Key: "files", Title: "Files", Items: []SectionItem{
			{"filemanager", "/files", "ri-folder-line", "File manager", "", false},
			{"filemanager", "/file-manager/upload?method=upload", "ri-upload-line", "File upload", "", false},
			{"filemanager", "/file-manager/upload?method=download", "ri-download-line", "Download files", "", false},
			{"ftp", "/ftp", "ri-folder-shared-line", "FTP accounts", "", false},
			{"ftp", "/ftp/connections", "ri-folder-shared-line", "FTP connections", "", false},
			{"disk_usage", "/disk-usage/", "ri-folder-add-line", "Disk usage", "", false},
			{"backups", "/backups", "ri-folder-check-line", "Backups", "", false},
			{"backup_wizard", "/backup-wizard", "ri-download-cloud-2-line", "Backup wizard", "", false},
			{"inodes", "/inodes-explorer", "ri-folder-forbid-line", "Inodes explorer", "", false},
			{"malware_scan", "/malware-scanner", "ri-barcode-line", "ClamAV scanner", "", false},
			{"fix_permissions", "/fix-permissions", "ri-file-code-line", "Fix permissions", "", false},
			{"trash", "/files.trash", "ri-delete-bin-line", "Trash", "", false},
		}},
		{Key: "domains", Title: "Domains", Items: []SectionItem{
			{"domains", "/domains", "ri-global-line", "Domains", "", false},
			{"redirects", "/domains/redirect", "ri-link", "Redirects", "", false},
			{"dns", "/domains/edit-dns-zone", "ri-file-text-line", "DNS zone editor", "", false},
			{"dynamic_dns", "/domains/dynamic-dns", "ri-loop-right-line", "Dynamic DNS", "", false},
			{"ssl", "/domains/ssl", "ri-lock-line", "SSL", "", false},
			{"domain_suspend", "/domains/suspend", "ri-forbid-line", "Suspend domain", "", false},
			{"domain_suspend", "/domains/unsuspend", "ri-forbid-line", "Unsuspend domain", "", false},
			{"docroot", "/domains/docroot", "ri-folder-line", "Change docroot", "", false},
			{"edit_vhost", "/domains/vhosts", "ri-file-text-line", "Edit virtual hosts", "", false},
			{"goaccess", "/domains/stats", "ri-line-chart-line", "GoAccess", "", false},
			{"domain_logs", "/domains/log", "ri-file-text-line", "Raw access logs", "", false},
		}},
		{Key: "mysql", Title: "MySQL", Items: []SectionItem{
			{"mysql", "/mysql", "ri-database-2-line", "Databases", "", false},
			{"mysql", "/mysql/users", "ri-group-line", "Users", "", false},
			{"mysql", "/mysql/wizard", "ri-database-2-line", "Database wizard", "", false},
			{"mysql_import", "/mysql/import", "ri-database-2-line", "Import database", "", false},
			{"phpmyadmin", "/phpmyadmin", "ri-external-link-line", "Open phpMyAdmin", "_blank", false},
			{"phpmyadmin", "/phpmyadmin/link", "ri-login-box-line", "phpMyAdmin Login Form", "_blank", false},
			{"mysql_processlist", "/mysql/processlist", "ri-database-2-line", "Process list", "", false},
			{"remote_mysql", "/mysql/remote-mysql", "ri-server-line", "Remote access", "", false},
			{"mysql_conf", "/mysql/configuration", "ri-database-2-line", "MySQL configuration", "", false},
			{"mysql_root_password", "/mysql/root-password", "ri-key-2-line", "Change root password", "", false},
		}},
		{Key: "postgresql", Title: "PostgreSQL", Items: []SectionItem{
			{"postgresql", "/postgresql", "ri-database-2-line", "Databases", "", false},
			{"postgresql", "/postgresql/users", "ri-group-line", "Users", "", false},
			{"postgresql", "/postgresql/wizard", "ri-database-2-line", "Database wizard", "", false},
			{"postgresql_import", "/postgresql/import", "ri-database-2-line", "Import database", "", false},
			{"postgresql", "/postgresql/processlist", "ri-database-2-line", "Process list", "", false},
			{"remote_postgresql", "/postgresql/remote-postgresql", "ri-organization-chart", "Remote access", "", false},
			{"postgresql_conf", "/postgresql/configuration", "ri-database-2-line", "PostgreSQL configuration", "", false},
		}},
		{Key: "mongodb", Title: "MongoDB", Items: []SectionItem{
			{"mongodb", "/mongodb", "ri-database-2-line", "Databases", "", false},
			{"mongodb", "/mongodb/users", "ri-group-line", "Users", "", false},
			{"mongodb", "/mongodb/wizard", "ri-database-2-line", "Database wizard", "", false},
			{"mongodb_import", "/mongodb/import", "ri-database-2-line", "Import database", "", false},
			{"mongodb", "/mongodb/processlist", "ri-database-2-line", "Process list", "", false},
		}},
		// mautic/flarum omitted: legacy code slated for removal entirely, not ported here (per user decision)
		{Key: "websites", Title: "Websites", Items: []SectionItem{
			{"websites", "/sites", "ri-notification-badge-line", "Site manager", "", false},
			{"autoinstaller", "/auto-installer", "ri-download-line", "Auto installer", "", false},
			{"wordpress", "/wordpress", "ri-wordpress-line", "WP manager", "", false},
			{"git", "/git", "ri-git-branch-line", "Git", "", false},
		}},
		{Key: "cache", Title: "Cache", Items: []SectionItem{
			{"redis", "/cache/redis", "ri-database-2-line", "Redis", "", false},
			{"valkey", "/cache/valkey", "ri-database-2-line", "Valkey", "", false},
			{"memcached", "/cache/memcached", "ri-server-line", "Memcached", "", false},
			{"varnish", "/cache/varnish", "ri-flashlight-line", "Varnish", "", false},
			{"opensearch", "/cache/opensearch", "ri-search-line", "OpenSearch", "", false},
			{"elasticsearch", "/cache/elasticsearch", "ri-search-eye-line", "ElasticSearch", "", false},
		}},
		{Key: "emails", Title: "Emails", Items: []SectionItem{
			{"emails", "/emails", "ri-mail-line", "Email accounts", "", false},
			{"email_aliases", "/emails/aliases", "ri-at-line", "Aliases", "", false},
			{"email_filters", "/emails/filter", "ri-filter-3-line", "Filters", "", false},
			{"email_deliverability", "/emails/deliverability", "ri-mail-check-line", "Email deliverability", "", false},
			{"email_default", "/emails/default", "ri-at-line", "Default address", "", false},
			{"email_import", "/emails/import", "ri-mail-send-line", "Address importer", "", false},
			{"email_migrate", "/emails/migrate", "ri-arrow-left-right-line", "Migrate emails", "", false},
			{"email_export", "/emails/export", "ri-mail-download-line", "Address exporter", "", false},
			{"webmail", "/webmail/", "ri-external-link-line", "Webmail", "_blank", false},
		}},
		{Key: "php", Title: "PHP", Items: []SectionItem{
			{"php", "/php/domains", "ri-code-box-line", "Select PHP version", "", false},
			{"php", "/php/default", "ri-code-s-slash-line", "Default version", "", false},
			{"php_options", "/php/options", "ri-toggle-line", "PHP options", "", false},
			{"php_extensions", "/php/extensions", "ri-puzzle-line", "PHP extensions", "", false},
			{"php_ini", "/php/php_ini_editor", "ri-code-s-slash-line", "PHP INI editor", "", false},
		}},
		{Key: "docker", Title: "Containers", Items: []SectionItem{
			{"docker", "/containers", "ri-stack-line", "Containers", "", false},
			{"terminal", "/containers/terminal", "ri-terminal-box-line", "Terminal", "", false},
			{"docker", "/containers/logs", "ri-file-code-line", "Logs", "", false},
			{"change_image", "/containers/image/change", "ri-text-block", "Change image tag", "", false},
			{"change_ws", "/containers/webserver", "ri-toggle-fill", "Change webserver", "", false},
			{"change_db", "/containers/mysql", "ri-toggle-line", "Change MySQL type", "", false},
		}},
		{Key: "advanced", Title: "Advanced", Items: []SectionItem{
			{"services", "/services", "ri-stack-line", "Services", "", false},
			{"crons", "/cronjobs", "ri-calendar-line", "Cron jobs", "", false},
			{"ip_blocker", "/security/ip-blocker", "ri-forbid-line", "IP blocker", "", false},
			{"usage", "/server/usage", "ri-dashboard-3-line", "Resource usage", "", false},
			{"usage", "/server/usage/history", "ri-dashboard-3-line", "Resource usage history", "", false},
			{"error_logs", "/error-logs", "ri-bug-line", "Error logs", "", false},
			{"summary", "/summary", "ri-file-list-3-line", "Daily summary", "", false},
			{"process_manager", "/process-manager", "ri-cpu-line", "Process manager", "", false},
			{"timezone", "/server/timezone", "ri-time-line", "Change timezone", "", false},
			{"webserver_conf", "/server/webserver_conf", "ri-server-line", "Webserver configuration", "", false},
			{"waf", "/server/waf", "ri-shield-keyhole-line", "WAF settings", "", false},
			{"waf", "/server/waf/log", "ri-error-warning-line", "WAF logs", "", false},
			{"info", "/server/info", "ri-information-line", "Server information", "", false},
		}},
		{Key: "account", Title: "Account", Items: []SectionItem{
			{"account", "/account", "ri-user-settings-line", "Email & password", "", false},
			{"locale", "/account/language", "ri-translate-2", "Change language", "", false},
			{"notifications", "/account/notifications", "ri-notification-3-line", "Email notifications", "", false},
			{"twofa", "/account/2fa", "ri-fingerprint-line", "Two-factor authentication", "", false},
			{"sessions", "/account/sessions", "ri-group-line", "Active sessions", "", false},
			{"favorites", "/account/favorites", "ri-star-line", "Favorite pages", "", false},
			{"activity", "/account/activity", "ri-pulse-line", "Account activity", "", false},
			{"login_history", "/account/login-history", "ri-user-unfollow-line", "Login history", "", false},
			{"logout", "/logout", "ri-door-open-line", "Log out", "", false},
		}},
	}
}

// withPlannerIcons swaps the section and item icons for the classic theme's Remix ones, keeping the order and labels that match the sidebar
func withPlannerIcons(sections, planner []Section) []Section {
	sectionIcons, itemIcons := map[string]string{}, map[string]string{}
	for _, s := range planner {
		sectionIcons[s.Key] = s.Icon
		for _, it := range s.Items {
			itemIcons[it.Href] = it.Icon
		}
	}
	out := make([]Section, len(sections))
	for i, s := range sections {
		if icon, ok := sectionIcons[s.Key]; ok {
			s.Icon = icon
		}
		s.Items = append([]SectionItem(nil), s.Items...)
		for j := range s.Items {
			if icon, ok := itemIcons[s.Items[j].Href]; ok {
				s.Items[j].Icon = icon
			}
		}
		out[i] = s
	}
	return out
}
