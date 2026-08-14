package dashboard

// SectionItem mirrors one icon-link dict in dashboard.html's `sections`
// Jinja literal (e.g. {"key": "filemanager", "href": "/files", ...}).
type SectionItem struct {
	Key    string
	Href   string
	Icon   string
	Label  string
	Target string
}

// Section mirrors one entry of dashboard.html's `sections` dict plus its
// title from `section_titles`.
type Section struct {
	Key   string
	Title string
	Items []SectionItem
}

// buildDashboardSections builds the dashboard's section/item list,
// keeping only items whose key is in allowed, for dashboard.html's
// {{range .Sections}}. Section order matches this slice's literal order.
func buildDashboardSections(allowed map[string]bool) []Section {
	all := []Section{
		{Key: "files", Title: "Files", Items: []SectionItem{
			{"filemanager", "/files", "bi-folder-fill", "File Manager", ""},
			{"filemanager", "/file-manager/upload?method=upload", "bi-upload", "File Upload", ""},
			{"filemanager", "/file-manager/upload?method=download", "bi-download", "Download Files", ""},
			{"ftp", "/ftp", "bi-folder-symlink-fill", "FTP Accounts", ""},
			{"ftp", "/ftp/connections", "bi-folder-symlink", "FTP Connections", ""},
			{"disk_usage", "/disk-usage/", "bi-folder-plus", "Disk Usage", ""},
			{"backups", "/backups", "bi-folder-check", "Backups", ""},
			{"backup_wizard", "/backup-wizard", "bi-cloud-arrow-down", "Backup Wizard", ""},
			{"inodes", "/inodes-explorer", "bi-folder-x", "Inodes Explorer", ""},
			{"malware_scan", "/malware-scanner", "bi-upc-scan", "ClamAV Scanner", ""},
			{"fix_permissions", "/fix-permissions", "bi-file-binary-fill", "Fix Permissions", ""},
			{"trash", "/files.trash", "bi-trash3", "Trash", ""},
		}},
		{Key: "domains", Title: "Domains", Items: []SectionItem{
			{"domains", "/domains", "bi-globe", "Domains", ""},
			{"redirects", "/domains/redirect", "bi-link-45deg", "Redirects", ""},
			{"dns", "/domains/edit-dns-zone", "bi-file-text", "DNS Zone Editor", ""},
			{"dynamic_dns", "/domains/dynamic-dns", "bi-arrow-repeat", "Dynamic DNS", ""},
			{"ssl", "/domains/ssl", "bi-lock", "SSL", ""},
			{"domain_suspend", "/domains/suspend", "bi-ban", "Suspend Domain", ""},
			{"domain_suspend", "/domains/unsuspend", "bi-ban-fill", "Unsuspend Domain", ""},
			{"docroot", "/domains/docroot", "bi-folder2", "Change Docroot", ""},
			{"edit_vhost", "/domains/vhosts", "bi-file-text", "Edit Virtual Hosts", ""},
			{"goaccess", "/domains/stats", "bi-graph-up", "GoAccess", ""},
			{"domain_logs", "/domains/log", "bi-file-text", "Raw Access Logs", ""},
		}},
		{Key: "mysql", Title: "MySQL", Items: []SectionItem{
			{"mysql", "/mysql", "bi-database", "Databases", ""},
			{"mysql", "/mysql/users", "bi-people", "Users", ""},
			{"mysql", "/mysql/wizard", "bi-database-add", "Database Wizard", ""},
			{"mysql_import", "/mysql/import", "bi-database-fill-add", "Import Database", ""},
			{"phpmyadmin", "/phpmyadmin", "bi-box-arrow-up-right", "Open phpMyAdmin", "_blank"},
			{"phpmyadmin", "/phpmyadmin/link", "bi-box-arrow-in-right", "phpMyAdmin Login Form", "_blank"},
			{"mysql_processlist", "/mysql/processlist", "bi-database-slash", "Process List", ""},
			{"remote_mysql", "/mysql/remote-mysql", "bi-hdd-network", "Remote Access", ""},
			{"mysql_conf", "/mysql/configuration", "bi-database-lock", "MySQL Configuration", ""},
			{"mysql_root_password", "/mysql/root-password", "bi-key-fill", "Change root password", ""},
		}},
		{Key: "postgresql", Title: "PostgreSQL", Items: []SectionItem{
			{"postgresql", "/postgresql", "bi-database", "Databases", ""},
			{"postgresql", "/postgresql/users", "bi-people", "Users", ""},
			{"postgresql", "/postgresql/wizard", "bi-database-add", "Database Wizard", ""},
			{"postgresql_import", "/postgresql/import", "bi-database-fill-add", "Import Database", ""},
			{"postgresql", "/postgresql/processlist", "bi-database-slash", "Process List", ""},
			{"remote_postgresql", "/postgresql/remote-postgresql", "bi-diagram-3", "Remote Access", ""},
			{"postgresql_conf", "/postgresql/configuration", "bi-database-lock", "PostgreSQL Configuration", ""},
		}},
		// mautic/flarum omitted: legacy code slated for removal entirely,
		// not ported here (per user decision).
		{Key: "websites", Title: "Websites", Items: []SectionItem{
			{"websites", "/sites", "bi-app-indicator", "Site Manager", ""},
			{"autoinstaller", "/auto-installer", "bi-download", "Auto Installer", ""},
			{"wordpress", "/wordpress", "bi-wordpress", "WP Manager", ""},
			{"drupal", "/drupal", "bi-droplet", "Drupal Manager", ""},
		}},
		{Key: "cache", Title: "Cache", Items: []SectionItem{
			{"redis", "/cache/redis", "bi-database-fill-lock", "Redis", ""},
			{"valkey", "/cache/valkey", "bi-database-fill-lock", "Valkey", ""},
			{"memcached", "/cache/memcached", "bi-hdd-network-fill", "Memcached", ""},
			{"varnish", "/cache/varnish", "bi-lightning-charge-fill", "Varnish", ""},
			{"opensearch", "/cache/opensearch", "bi-search", "OpenSearch", ""},
			{"elasticsearch", "/cache/elasticsearch", "bi-search-heart", "ElasticSearch", ""},
		}},
		{Key: "emails", Title: "Emails", Items: []SectionItem{
			{"emails", "/emails", "bi-envelope", "Email Accounts", ""},
			{"email_aliases", "/emails/aliases", "bi-at", "Aliases", ""},
			{"email_filters", "/emails/filter", "bi-funnel", "Filters", ""},
			{"email_deliverability", "/emails/deliverability", "bi-envelope-check", "Email Deliverability", ""},
			{"email_default", "/emails/default", "bi-envelope-at", "Default Address", ""},
			{"email_import", "/emails/import", "bi-envelope-arrow-up", "Address Importer", ""},
			{"email_export", "/emails/export", "bi-envelope-arrow-down", "Address Exporter", ""},
			{"webmail", "/webmail/", "bi-box-arrow-up-right", "Webmail", "_blank"},
		}},
		{Key: "php", Title: "PHP", Items: []SectionItem{
			{"php", "/php/domains", "bi-code-square", "Select PHP version", ""},
			{"php", "/php/default", "bi-filetype-php", "Default Version", ""},
			{"php_options", "/php/options", "bi-toggles", "PHP Options", ""},
			{"php_extensions", "/php/extensions", "bi-puzzle", "PHP Extensions", ""},
			{"php_ini", "/php/php_ini_editor", "bi-filetype-php", "PHP.INI Editor", ""},
		}},
		{Key: "docker", Title: "Containers", Items: []SectionItem{
			{"docker", "/containers", "bi-boxes", "Containers", ""},
			{"docker", "/containers/terminal", "bi-terminal", "Terminal", ""},
			{"docker", "/containers/logs", "bi-file-binary", "Logs", ""},
			{"docker", "/containers/image/", "bi-arrow-clockwise", "Image Updates", ""},
			{"docker", "/containers/image/change", "bi-textarea-t", "Change Image tag", ""},
			{"docker", "/containers/webserver", "bi-toggle2-on", "Change webserver", ""},
			{"docker", "/containers/mysql", "bi-toggle2-off", "Change MySQL Type", ""},
		}},
		{Key: "advanced", Title: "Advanced", Items: []SectionItem{
			{"services", "/services", "bi-hdd-stack", "Services", ""},
			{"crons", "/cronjobs", "bi-calendar2-week", "Cron Jobs", ""},
			{"ip_blocker", "/security/ip-blocker", "bi-ban", "IP Blocker", ""},
			{"usage", "/server/usage", "bi-speedometer2", "Resource Usage", ""},
			{"usage", "/server/usage/history", "bi-speedometer2", "Resource Usage History", ""},
			{"process_manager", "/process-manager", "bi-cpu", "Process Manager", ""},
			{"timezone", "/server/timezone", "bi-clock", "Change TimeZone", ""},
			{"webserver_conf", "/server/webserver_conf", "bi-hdd-network", "Webserver Configuration", ""},
			{"waf", "/server/waf", "bi-shield-lock", "WAF Settings", ""},
			{"waf", "/server/waf/log", "bi-shield-exclamation", "WAF Logs", ""},
			{"info", "/server/info", "bi-info-square", "Server Information", ""},
		}},
		{Key: "account", Title: "Account", Items: []SectionItem{
			{"account", "/account", "bi-person-gear", "Email & Password", ""},
			{"locale", "/account/language", "bi-translate", "Change Language", ""},
			{"notifications", "/account/notifications", "bi-bell", "Email Notifications", ""},
			{"twofa", "/account/2fa", "bi-fingerprint", "Two-Factor Authentication", ""},
			{"sessions", "/account/sessions", "bi-people", "Active Sessions", ""},
			{"favorites", "/account/favorites", "bi-star", "Favorite Pages", ""},
			{"activity", "/account/activity", "bi-activity", "Account Activity", ""},
			{"login_history", "/account/login-history", "bi-person-exclamation", "Login History", ""},
			{"logout", "/logout", "bi-door-open", "Log out", ""},
		}},
	}

	result := make([]Section, 0, len(all))
	for _, s := range all {
		var items []SectionItem
		for _, item := range s.Items {
			if allowed[item.Key] {
				items = append(items, item)
			}
		}
		if len(items) > 0 {
			result = append(result, Section{Key: s.Key, Title: s.Title, Items: items})
		}
	}
	return result
}
