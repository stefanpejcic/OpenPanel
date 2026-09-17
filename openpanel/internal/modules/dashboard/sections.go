package dashboard

import "gist.github.com/stefanpejcic/openpanel/internal/core/i18n"

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
	Items []SectionItem
}

// buildDashboardSections builds the dashboard's section/item list, keeping only items whose key is in allowed (or in upsellAllowed, marked Disabled), for dashboard.html's {{range .Sections}}. Section order matches this slice's literal order. Titles/Labels are run through t.Get so the icon grid picks up the same catalog as the sidebar, instead of staying hardcoded English.
func buildDashboardSections(t i18n.Translator, allowed, upsellAllowed map[string]bool) []Section {
	all := []Section{
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
		}},
		// mautic/flarum omitted: legacy code slated for removal entirely, not ported here (per user decision)
		{Key: "websites", Title: "Websites", Items: []SectionItem{
			{"websites", "/sites", "bi-app-indicator", "Site Manager", "", false},
			{"autoinstaller", "/auto-installer", "bi-download", "Auto Installer", "", false},
			{"wordpress", "/wordpress", "bi-wordpress", "WP Manager", "", false},
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

	result := make([]Section, 0, len(all))
	for _, s := range all {
		var items []SectionItem
		for _, item := range s.Items {
			item.Label = t.Get(item.Label)
			if allowed[item.Key] {
				items = append(items, item)
			} else if upsellAllowed[item.Key] {
				item.Disabled = true
				items = append(items, item)
			}
		}
		if len(items) > 0 {
			result = append(result, Section{Key: s.Key, Title: t.Get(s.Title), Items: items})
		}
	}
	return result
}
