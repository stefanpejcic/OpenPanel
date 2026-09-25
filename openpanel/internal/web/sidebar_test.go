package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func tabLabels(tabs []NavLink) string {
	labels := make([]string, len(tabs))
	for i, tab := range tabs {
		labels[i] = tab.Label
	}
	return strings.Join(labels, ",")
}

func activeTab(tabs []NavLink) string {
	for _, tab := range tabs {
		if tab.Active {
			return tab.Label
		}
	}
	return ""
}

func findItem(items []NavItem, label string) *NavItem {
	for i := range items {
		if items[i].Label == label {
			return &items[i]
		}
	}
	return nil
}

func TestNavPathAndUploadDownloadActiveState(t *testing.T) {
	allowed := map[string]bool{"filemanager": true}
	cases := map[string]string{
		"/file-manager/upload":                 "Upload",
		"/file-manager/upload?method=upload":   "Upload",
		"/file-manager/upload?method=download": "Download from URL",
	}
	for url, want := range cases {
		r := httptest.NewRequest("GET", url, nil)
		if got := activeTab(BuildPageTabs(allowed, nil, NavPath(r), TabContext{})); got != want {
			t.Errorf("%s: expected %q tab active, got %q", url, want, got)
		}
	}
}

func TestBuildSidebarNavEmpty(t *testing.T) {
	if items := BuildSidebarNav(map[string]bool{}, nil, "/dashboard"); len(items) != 0 {
		t.Errorf("expected no items for an empty allowed set, got %+v", items)
	}
}

func TestSidebarLinksToFirstReachablePage(t *testing.T) {
	allowed := map[string]bool{"ftp": true, "email_aliases": true, "webmail": true}
	items := BuildSidebarNav(allowed, nil, "/ftp/new")

	files := findItem(items, "Files")
	if files == nil || files.Href != "/ftp" || !files.Active {
		t.Errorf("expected an active Files link to /ftp, got %+v", files)
	}
	// Webmail opens in a new window, so it never becomes the Email link
	email := findItem(items, "Email")
	if email == nil || email.Href != "/emails/aliases" || email.Active {
		t.Errorf("expected an inactive Email link to /emails/aliases, got %+v", email)
	}
}

func TestSidebarUpsellOnlyAreaIsDisabled(t *testing.T) {
	items := BuildSidebarNav(map[string]bool{}, map[string]bool{"domains": true, "dns": true}, "/dashboard")
	if len(items) != 1 || !items[0].Disabled || items[0].Href != "/domains" {
		t.Errorf("expected a single disabled Domains link, got %+v", items)
	}
}

func TestSidebarSections(t *testing.T) {
	allowed := map[string]bool{"domains": true, "php": true, "crons": true, "usage": true, "info": true}
	sections := map[string]string{}
	for _, item := range BuildSidebarNav(allowed, nil, "/cronjobs") {
		if item.Section != "" {
			sections[item.Section] = item.Label
		}
	}
	if sections["Configure"] != "PHP" || sections["Monitor & Secure"] != "Statistics" {
		t.Errorf("unexpected section starts: %v", sections)
	}

	// a section with no visible areas hands its heading to nobody, the next section still gets its own
	items := BuildSidebarNav(map[string]bool{"domains": true, "info": true}, nil, "/domains")
	if items[1].Label != "Server Info" || items[1].Section != "Monitor & Secure" {
		t.Errorf("expected Server Info to carry the Monitor & Secure heading, got %+v", items)
	}
}

func TestEachPageBelongsToOneArea(t *testing.T) {
	allowed := map[string]bool{"domains": true, "goaccess": true, "domain_logs": true, "edit_vhost": true, "webserver_conf": true,
		"docker": true, "terminal": true, "mysql": true, "change_db": true, "change_ws": true, "usage": true}
	cases := map[string]string{
		"/domains/stats":        "Statistics",
		"/domains/log":          "Statistics",
		"/domains/vhosts":       "Web Server Config",
		"/domains/ssl":          "Domains",
		"/containers/mysql":     "MySQL",
		"/containers/webserver": "Web Server Config",
		"/containers/terminal":  "Containers",
		"/server/usage/history": "Statistics",
	}
	for path, want := range cases {
		var active []string
		for _, item := range BuildSidebarNav(allowed, nil, path) {
			if item.Active {
				active = append(active, item.Label)
			}
		}
		if len(active) != 1 || active[0] != want {
			t.Errorf("%s: expected only %s active, got %v", path, want, active)
		}
	}
}

func TestBuildPageTabsMySQL(t *testing.T) {
	allowed := map[string]bool{"mysql": true, "mysql_conf": true, "phpmyadmin": true, "remote_mysql": true, "change_db": true}
	tabs := BuildPageTabs(allowed, nil, "/mysql/assign", TabContext{})
	if got := tabLabels(tabs); got != "Databases,Users,phpMyAdmin,Remote Access,Configuration,Server Type" {
		t.Errorf("unexpected MySQL tabs: %s", got)
	}
	if activeTab(tabs) != "Users" {
		t.Errorf("expected Users active on /mysql/assign, got %q", activeTab(tabs))
	}
	if tabs[2].Target != "_blank" {
		t.Error("expected phpMyAdmin to open in a new window")
	}
	if activeTab(BuildPageTabs(allowed, nil, "/containers/mysql", TabContext{})) != "Server Type" {
		t.Error("expected Server Type active on /containers/mysql")
	}
}

func TestDatabasesSectionListsEnabledEngines(t *testing.T) {
	allowed := map[string]bool{"backups": true, "postgresql": true, "mongodb": true, "php": true}
	items := BuildSidebarNav(allowed, nil, "/postgresql/users")

	var labels []string
	for _, item := range items {
		labels = append(labels, item.Label)
	}
	if strings.Join(labels, ",") != "Backups,PostgreSQL,MongoDB,PHP" {
		t.Errorf("unexpected sidebar: %v", labels)
	}
	// MySQL isn't enabled, so the heading moves to the first engine that is
	if pg := findItem(items, "PostgreSQL"); pg == nil || pg.Section != "Databases" || !pg.Active {
		t.Errorf("expected an active PostgreSQL item carrying the Databases heading, got %+v", pg)
	}
	if php := findItem(items, "PHP"); php == nil || php.Section != "Configure" {
		t.Errorf("expected PHP to keep the Configure heading, got %+v", php)
	}
}
func TestBuildPageTabsBackups(t *testing.T) {
	allowed := map[string]bool{"backups": true, "backup_wizard": true}
	tabs := BuildPageTabs(allowed, nil, "/backups/settings", TabContext{})
	if tabLabels(tabs) != "Overview,Restore,Configuration,Destination,Backup Wizard" || activeTab(tabs) != "Configuration" {
		t.Errorf("unexpected backup tabs: %+v", tabs)
	}
	tabs = BuildPageTabs(allowed, nil, "/backups", TabContext{BackupsAdminManaged: true})
	if tabLabels(tabs) != "Overview,Restore,Backup Wizard" {
		t.Errorf("expected Configuration/Destination hidden when admin-managed, got %s", tabLabels(tabs))
	}
}

func TestBuildPageTabsNeedsTwoTabs(t *testing.T) {
	if tabs := BuildPageTabs(map[string]bool{"domains": true}, nil, "/domains", TabContext{}); tabs != nil {
		t.Errorf("expected no tab bar with a single tab, got %+v", tabs)
	}
}

func TestBuildPageTabsServerInfo(t *testing.T) {
	tabs := BuildPageTabs(map[string]bool{"info": true}, nil, "/server/info", TabContext{})
	if tabLabels(tabs) != "Server,Hosting Plan,Panel" || activeTab(tabs) != "Server" || tabs[1].Href != "/server/info#plan" {
		t.Errorf("unexpected server info tabs: %+v", tabs)
	}
}

func TestBuildPageTabsCronjobs(t *testing.T) {
	allowed := map[string]bool{"crons": true}
	for path, active := range map[string]string{"/cronjobs": "Cron Jobs", "/cronjobs/new": "Cron Jobs", "/cronjobs/editor": "File Editor", "/cronjobs/logs": "Logs"} {
		tabs := BuildPageTabs(allowed, nil, path, TabContext{})
		if tabLabels(tabs) != "Cron Jobs,File Editor,Logs" || activeTab(tabs) != active {
			t.Errorf("%s: tabs %s, active %s", path, tabLabels(tabs), activeTab(tabs))
		}
	}
}

func TestBuildPageTabsWebsitesAndEmail(t *testing.T) {
	allowed := map[string]bool{"emails": true, "email_aliases": true, "webmail": true, "wordpress": true, "autoinstaller": true}

	tabs := BuildPageTabs(allowed, nil, "/emails/aliases", TabContext{})
	if tabLabels(tabs) != "Accounts,Aliases,Delete Accounts,Webmail" || activeTab(tabs) != "Aliases" {
		t.Errorf("unexpected email tabs: %+v", tabs)
	}

	tabs = BuildPageTabs(allowed, nil, "/drupal/install", TabContext{})
	if tabLabels(tabs) != "Sites,WordPress,Install App" || activeTab(tabs) != "Install App" {
		t.Errorf("unexpected website tabs: %+v", tabs)
	}
}

func TestServiceForPath(t *testing.T) {
	env := map[string]string{"MYSQL_TYPE": "mariadb", "WEB_SERVER": "nginx", "DEFAULT_PHP_VERSION": "8.3"}
	get := func(k string) string { return env[k] }
	cases := map[string]string{
		"/mysql/users":           "mariadb",
		"/mysql/phpmyadmin":      "mariadb",
		"/containers/mysql":      "mariadb",
		"/postgresql":            "postgres",
		"/mongodb/import":        "mongodb",
		"/cache/redis":           "redis",
		"/server/webserver_conf": "nginx",
		"/containers/webserver":  "nginx",
		"/php/options":           "php-fpm-8.3",
		"/files":                 "",
	}
	for path, want := range cases {
		if got := ServiceForPath(path, get); got != want {
			t.Errorf("ServiceForPath(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestBuildPageTabsServiceTabs(t *testing.T) {
	allowed := map[string]bool{"redis": true, "valkey": true, "terminal": true, "docker": true}
	tabs := BuildPageTabs(allowed, nil, "/cache/redis", TabContext{Service: "redis"})
	if tabLabels(tabs) != "Redis,Valkey,Terminal,Logs" || activeTab(tabs) != "Redis" {
		t.Fatalf("unexpected cache tabs: %+v", tabs)
	}
	if tabs[2].Href != "/containers/terminal/redis?ctx=service" || tabs[3].Href != "/containers/logs?container=redis&lines=20&ctx=service" {
		t.Errorf("unexpected Terminal/Logs hrefs: %+v", tabs[2:])
	}
	if tabs := BuildPageTabs(map[string]bool{"redis": true}, nil, "/cache/redis", TabContext{Service: "redis"}); tabs != nil {
		t.Errorf("expected no tabs without terminal/docker features, got %+v", tabs)
	}
}

func TestAccountIsSidebarItemBelowContainers(t *testing.T) {
	allowed := map[string]bool{"docker": true, "locale": true, "favorites": true, "api": true, "mcp": true, "usage": true}
	items := BuildSidebarNav(allowed, nil, "/account/favorites")
	for i, item := range items {
		if item.Label != "Account" {
			continue
		}
		if i == 0 || items[i-1].Label != "Containers" {
			t.Errorf("expected Account right below Containers, got %+v", items)
		}
		if item.Href != "/account/language" || !item.Active {
			t.Errorf("expected an active Account link to /account/language, got %+v", item)
		}
		if tabs := BuildPageTabs(allowed, nil, "/account/favorites", TabContext{}); tabLabels(tabs) != "Language,Favorites,API,AI Assistant (MCP)" || activeTab(tabs) != "Favorites" {
			t.Errorf("unexpected account tabs: %+v", tabs)
		}
		if activeTab(BuildPageTabs(allowed, nil, "/account/mcp", TabContext{})) != "AI Assistant (MCP)" {
			t.Error("expected AI Assistant (MCP) active on /account/mcp")
		}
		return
	}
	t.Error("expected an Account sidebar item")
}

func TestHasAnyPrefix(t *testing.T) {
	if !hasAnyPrefix("/mysql/users", "/mysql", "/postgresql") {
		t.Error("expected /mysql/users to match /mysql prefix")
	}
	if hasAnyPrefix("/domains", "/mysql", "/postgresql") {
		t.Error("did not expect /domains to match")
	}
}

func TestAccountSecurityPagesAreSecurityTabs(t *testing.T) {
	allowed := map[string]bool{"waf": true, "twofa": true, "passkeys": true, "sessions": true, "activity": true, "login_history": true}

	tabs := BuildPageTabs(allowed, nil, "/account/passkeys", TabContext{})
	if tabLabels(tabs) != "Web Firewall,Two-Factor Auth,Passkeys,Active Sessions,Activity Log,Login History" || activeTab(tabs) != "Passkeys" {
		t.Errorf("unexpected security tabs: %+v", tabs)
	}
	if activeTab(BuildPageTabs(allowed, nil, "/account/login-history", TabContext{})) != "Login History" {
		t.Error("expected Login History active on /account/login-history")
	}

	security := findItem(BuildSidebarNav(allowed, nil, "/account/2fa"), "Security")
	if security == nil || !security.Active {
		t.Errorf("expected the Security sidebar link active on /account/2fa, got %+v", security)
	}

	// with only account security features, the sidebar link lands on 2FA
	security = findItem(BuildSidebarNav(map[string]bool{"twofa": true}, nil, "/dashboard"), "Security")
	if security == nil || security.Href != "/account/2fa" {
		t.Errorf("expected Security to link to /account/2fa, got %+v", security)
	}
}

func TestResolveNavInlinesServiceTools(t *testing.T) {
	env := func(k string) string { return map[string]string{"MYSQL_TYPE": "mariadb", "WEB_SERVER": "apache"}[k] }
	allowed := map[string]bool{"mysql": true, "terminal": true, "docker": true}

	r := httptest.NewRequest("GET", "/containers/terminal/mariadb?ctx=service", nil)
	path, ctx := ResolveNav(r, env)
	if path != "/mysql" || ctx.Service != "mariadb" || ctx.ActiveTool != "terminal" {
		t.Fatalf("expected the MySQL area with an active terminal, got %q %+v", path, ctx)
	}
	tabs := BuildPageTabs(allowed, nil, path, ctx)
	if activeTab(tabs) != "Terminal" {
		t.Errorf("expected only Terminal active, got %+v", tabs)
	}
	if db := findItem(BuildSidebarNav(allowed, nil, path), "MySQL"); db == nil || !db.Active {
		t.Error("expected MySQL to stay active in the sidebar")
	}

	r = httptest.NewRequest("GET", "/containers/logs?container=apache&lines=20&ctx=service", nil)
	if path, ctx := ResolveNav(r, env); path != "/server/webserver_conf" || ctx.ActiveTool != "logs" {
		t.Errorf("expected the web server area with active logs, got %q %+v", path, ctx)
	}

	// without ctx=service, or for a service no area owns, the page stays in Containers
	for _, url := range []string{"/containers/terminal/mariadb", "/containers/terminal/cron?ctx=service"} {
		if path, _ := ResolveNav(httptest.NewRequest("GET", url, nil), env); !isContainersPath(path) {
			t.Errorf("%s: expected a Containers path, got %q", url, path)
		}
	}
}

func TestBuildNavTrail(t *testing.T) {
	allowed := map[string]bool{"mysql": true, "postgresql": true, "domains": true, "ssl": true, "crons": true, "terminal": true}
	trail := func(navPath, requestPath, title string, ctx TabContext) string {
		tabs := BuildPageTabs(allowed, nil, navPath, ctx)
		var parts []string
		for _, c := range BuildNavTrail(allowed, nil, navPath, requestPath, title, tabs) {
			parts = append(parts, c.Label)
		}
		return strings.Join(parts, " > ")
	}
	cases := []struct{ nav, req, title, want string }{
		{"/mysql/users", "/mysql/users", "Users", "Databases > MySQL > Users"},
		{"/mysql", "/mysql", "Databases", "Databases > MySQL"},
		{"/postgresql/import", "/postgresql/import", "Import", "Databases > PostgreSQL > Import"},
		{"/mysql/assign", "/mysql/assign", "Assign User", "Databases > MySQL > Users > Assign User"},
		{"/domains", "/domains", "Domains", "Domains"},
		{"/domains/ssl", "/domains/ssl", "SSL", "Domains > SSL Certificates"},
		{"/cronjobs/new", "/cronjobs/new", "New Cron Job", "Cron Jobs > New Cron Job"},
		{"/somewhere", "/somewhere", "Some Page", "Some Page"},
	}
	for _, c := range cases {
		if got := trail(c.nav, c.req, c.title, TabContext{}); got != c.want {
			t.Errorf("%s: got %q, want %q", c.req, got, c.want)
		}
	}
	if got := trail("/mysql", "/containers/terminal/mariadb", "Terminal", TabContext{Service: "mariadb", ActiveTool: "terminal"}); got != "Databases > MySQL > Terminal" {
		t.Errorf("inline terminal: got %q", got)
	}

	crumbs := BuildNavTrail(allowed, nil, "/mysql/users", "/mysql/users", "Users", BuildPageTabs(allowed, nil, "/mysql/users", TabContext{}))
	if crumbs[0].Href != "/mysql" || crumbs[len(crumbs)-1].Href != "" {
		t.Errorf("expected linked parents and an unlinked current crumb, got %+v", crumbs)
	}
}

func TestDashboardTabs(t *testing.T) {
	if DashboardTabs("/dashboard", false) != nil {
		t.Error("expected no dashboard tabs without an upsell plan")
	}
	tabs := DashboardTabs("/dashboard/upgrade", true)
	if tabLabels(tabs) != "Dashboard,Upgrade now" || activeTab(tabs) != "Upgrade now" {
		t.Errorf("unexpected dashboard tabs: %+v", tabs)
	}
	var parts []string
	for _, c := range BuildNavTrail(nil, nil, "/dashboard/upgrade", "/dashboard/upgrade", "Upgrade now", tabs) {
		parts = append(parts, c.Label)
	}
	if strings.Join(parts, " > ") != "Dashboard > Upgrade now" {
		t.Errorf("unexpected upgrade trail: %v", parts)
	}
}

func TestUpgradeFeatures(t *testing.T) {
	allowed := map[string]bool{"mysql": true, "filemanager": true}
	upsell := map[string]bool{"phpmyadmin": true, "postgresql": true, "ftp": true}

	got := map[string]string{}
	for _, g := range UpgradeFeatures(allowed, upsell) {
		got[g.Area] = strings.Join(g.Pages, ",")
	}
	want := map[string]string{
		"Files":      "FTP Accounts",
		"MySQL":      "phpMyAdmin",
		"PostgreSQL": "Databases,Users,Running Queries",
	}
	if len(got) != len(want) {
		t.Errorf("unexpected feature groups: %v", got)
	}
	for area, pages := range want {
		if got[area] != pages {
			t.Errorf("%s: got %q, want %q", area, got[area], pages)
		}
	}
}
