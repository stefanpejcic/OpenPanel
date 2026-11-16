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
			{"bot_control", "/server/bot-control", "ri-robot-2-line", "AI & Bot Control", "", false},
			{"ip_blocker", "/security/ip-blocker", "ri-forbid-line", "IP blocker", "", false},
			{"malware_scan", "/malware-scanner", "ri-barcode-line", "Malware scanner", "", false},
			{"imunifyav", "/imunifyav", "ri-shield-check-line", "ImunifyAV", "", false},
			{"twofa", "/account/2fa", "ri-fingerprint-line", "Two-factor auth", "", false},
			{"passkeys", "/account/passkeys", "ri-key-2-line", "Passkeys", "", false},
			{"sessions", "/account/sessions", "ri-group-line", "Active sessions", "", false},
			{"activity", "/account/activity", "ri-pulse-line", "Activity log", "", false},
			{"login_history", "/account/login-history", "ri-user-unfollow-line", "Login history", "", false},
		}},
	}
}

// classicSections is the dashboard for menu_style=classic, grouped like the classic sidebar
func classicSections() []Section {
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
			{"bot_control", "/server/bot-control", "ri-robot-2-line", "AI & Bot Control", "", false},
			{"imunifyav", "/imunifyav", "ri-shield-check-line", "ImunifyAV", "", false},
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

// appItems are the apps an admin can pin to the Websites section via applications_dashboard_items, keyed by module name
var appItems = map[string]SectionItem{
	"joomla":           {"joomla", "/joomla/install", "ri-grid-line", "Joomla", "", false},
	"drupal":           {"drupal", "/drupal/install", "ri-drop-line", "Drupal", "", false},
	"website_builder":  {"website_builder", "/website-builder/install", "ri-brush-line", "Website builder", "", false},
	"prestashop":       {"prestashop", "/prestashop/install", "ri-shopping-cart-line", "PrestaShop", "", false},
	"opencart":         {"opencart", "/opencart/install", "ri-shopping-cart-line", "OpenCart", "", false},
	"phpbb":            {"phpbb", "/phpbb/install", "ri-chat-1-line", "phpBB", "", false},
	"flarum":           {"flarum", "/flarum/install", "ri-chat-3-line", "Flarum", "", false},
	"mediawiki":        {"mediawiki", "/mediawiki/install", "ri-book-open-line", "MediaWiki", "", false},
	"dokuwiki":         {"dokuwiki", "/dokuwiki/install", "ri-file-list-3-line", "DokuWiki", "", false},
	"sofawiki":         {"sofawiki", "/sofawiki/install", "ri-book-2-line", "SofaWiki", "", false},
	"matomo":           {"matomo", "/matomo/install", "ri-bar-chart-line", "Matomo", "", false},
	"nextcloud":        {"nextcloud", "/nextcloud/install", "ri-cloud-line", "Nextcloud", "", false},
	"tinyphotogallery": {"tinyphotogallery", "/tinyphotogallery/install", "ri-image-2-line", "TinyPhotoGallery", "", false},
	"tinyfilemanager":  {"tinyfilemanager", "/tinyfilemanager/install", "ri-folder-open-line", "TinyFileManager", "", false},
	"moodle":           {"moodle", "/moodle/install", "ri-graduation-cap-line", "Moodle", "", false},
	"ojs":              {"ojs", "/ojs/install", "ri-book-marked-line", "OJS", "", false},
	"nodejs":           {"nodejs", "/nodejs/install", "ri-javascript-line", "Node.js", "", false},
	"python":           {"python", "/python/install", "ri-code-s-slash-line", "Python", "", false},
	"ruby":             {"ruby", "/ruby/install", "ri-code-s-slash-line", "Ruby", "", false},
	"java":             {"java", "/java/install", "ri-code-s-slash-line", "Java", "", false},
	"php":              {"php", "/php/install", "ri-code-s-slash-line", "PHP", "", false},
	"docker":           {"docker", "/containers/new", "ri-archive-2-line", "Docker", "", false},
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
