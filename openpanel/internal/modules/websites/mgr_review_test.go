package websites

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

// mgrPage is a manager page's data with every feature the review turned on
func mgrPage(t *testing.T, domain string) pageData {
	t.Helper()
	mgr := i18n.NewManager(t.TempDir(), nil)
	layout := baseLayout(mgr, "/website")
	for _, k := range []string{"docker", "services", "waf", "malware_scan", "git", "disk_usage", "inodes", "ftp", "domain_logs", "php", "backups"} {
		layout.UserAllowed[k] = true
	}
	layout.NavTrail = web.BuildNavTrail(layout.UserAllowed, nil, "/website", "/website", domain, layout.PageTabs)
	return pageData{LayoutData: layout, CurrentDomain: domain, Docroot: "/var/www/html/" + domain,
		DiskUsageHref: "/disk-usage/" + domain, InodesExplorerHref: "/inodes-explorer/" + domain}
}

func renderPage(t *testing.T, p *web.Page, data any) string {
	t.Helper()
	w := httptest.NewRecorder()
	if err := p.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return w.Body.String()
}

func mustContain(t *testing.T, body string, want ...string) {
	t.Helper()
	for _, s := range want {
		if !strings.Contains(body, s) {
			t.Errorf("missing %q", s)
		}
	}
}

func runtimeData(t *testing.T, typ, domain, svc string, running bool) PythonNodeAppsPageData {
	status := "false"
	if running {
		status = "true"
	}
	files := map[string]string{"nodejs": "index.js", "python": "app.py", "ruby": "app.rb", "java": "Main.java"}[typ]
	manifests := map[string]Manifest{
		"nodejs": {File: "package.json", Found: true, Text: "{\n  \"name\": \"nodejs-tests\",\n  \"dependencies\": {\n    \"express\": \"^4.21.2\",\n    \"morgan\": \"^1.10.0\"\n  }\n}",
			Groups: []ManifestGroup{{Title: "Scripts", ValueLabel: "Command", Rows: []ManifestRow{{"dev", "node --watch index.js", ""}, {"start", "node index.js", ""}}},
				{Title: "Dependencies", ValueLabel: "Constraint", HasInstalled: true, Rows: []ManifestRow{{"express", "^4.21.2", "4.22.3"}, {"morgan", "^1.10.0", "1.12.1"}}}}},
		"python": {File: "requirements.txt", Found: true, Text: "flask==3.1.0\npython-dotenv==1.0.1\n",
			Groups: []ManifestGroup{{Title: "Packages", ValueLabel: "Version", Rows: []ManifestRow{{"flask", "3.1.0", ""}, {"python-dotenv", "1.0.1", ""}}}}},
	}
	m := manifests[typ]
	m.Path = "/var/www/html/" + domain + "/" + m.File
	return PythonNodeAppsPageData{
		pageData:  mgrPage(t, domain),
		Container: ContainerInfo{SiteName: domain, Type: typ, Container: strings.ToUpper(svc) + "_x", Version: "24.22.2", Port: "3000", CreatedDate: "2026-10-29T11:16:43Z"},
		PM2Data:   map[string]string{"prefix": strings.ToUpper(svc) + "_NODE_", "status": status},
		Type:      typ,
		PM2Status: status,
		Service:   strings.ToUpper(svc),
		CPU:       "0.5", RAM: "0.75", PIDs: "200",
		StartupFile:          "/var/www/html/" + domain + "/" + files,
		Workdir:              "/var/www/html/" + domain,
		CurrentVersion:       "24.22.2",
		RequirementsSelected: true,
		StartCommand:         "npm install && node /var/www/html/" + domain + "/" + files,
		Manifest:             m,
		SSLState:             "valid", SSLExpires: "2027-01-27",
		WebServer: "apache",
		Image:     "node:24.22.2",
		Files: []SiteEntry{{Name: "node_modules", Dir: true, Modified: "2026-10-29"}, {Name: "index.js", Size: "480 B", Modified: "2026-10-29"},
			{Name: "package-lock.json", Size: "31 KB", Modified: "2026-10-29"}, {Name: "package.json", Size: "375 B", Modified: "2026-10-29"}},
		HiddenFiles: 2,
		EnvVars:     "NODE_ENV=production\nPORT=3000",
	}
}

func TestRuntimeManagerShell(t *testing.T) {
	webtest.ModernOnly(t)
	body := renderPage(t, pythonNodeAppsPage, runtimeData(t, "nodejs", "nodejs.tests.openpanel.org", "nodeapp", true))
	mustContain(t, body,
		`class="op-content op-mgr-shell"`, `id="backups-tab"`, `id="backups" role="tabpanel"`,
		"How requests reach the app", "SSL certificate valid until 2027-01-27", "Apache", "nodeapp:3000", "node:24.22.2",
		"node_modules", "31 KB", "2 hidden files not shown", "Latest log lines", `id="op-rt-tail"`,
		`name="requirements"`, "Runs on start", `name="startup_file"`, `name="workdir"`, `name="git_repo_url"`, `name="cpu"`, `name="ram"`, `name="pids"`,
		"window.loadAppLogs", ".replace(/^backup$/, 'backups')")
	for _, gone := range []string{"op-wp-side", "w-100", `id="backup"`} {
		if strings.Contains(body, gone) {
			t.Errorf("still has %q", gone)
		}
	}
	// stopped app, no files, no TLS
	d := runtimeData(t, "python", "python.tests.openpanel.org", "flaskapp", false)
	d.Files, d.HiddenFiles, d.SSLState, d.Image = nil, 0, "", ""
	body = renderPage(t, pythonNodeAppsPage, d)
	mustContain(t, body, "The folder is empty.", "Served over plain HTTP", "requirements.txt", "pip install -r requirements.txt", "Start")
}

func TestListSiteDirAndHelpers(t *testing.T) {
	if got := humanSize(480); got != "480 B" {
		t.Errorf("humanSize(480) = %q", got)
	}
	if got := humanSize(31437); got != "31 KB" {
		t.Errorf("humanSize(31437) = %q", got)
	}
	if got := humanSize(1536); got != "1.5 KB" {
		t.Errorf("humanSize(1536) = %q", got)
	}
	if got := svcImage("nobody", map[string]any{"image": "docker.io/library/node:24.22.2"}); got != "node:24.22.2" {
		t.Errorf("svcImage = %q", got)
	}
	if got := svcImage("nobody", map[string]any{"image": "n8nio/n8n:${N8N_N8N_TAG:-3.0.1}"}); got != "n8nio/n8n:3.0.1" {
		t.Errorf("svcImage = %q", got)
	}
	env := map[string]string{"TAG": "24.22.2"}
	if got := expandCompose("node:${TAG:-latest} ${NOPE:-x} ${NOPE}", func(k string) string { return env[k] }); got != "node:24.22.2 x " {
		t.Errorf("expandCompose = %q", got)
	}
	if files, hidden := listSiteDir("nobody", "/etc", 5); files != nil || hidden != 0 {
		t.Error("only /var/www/html paths are listed")
	}
	if files, _ := listSiteDir("nobody", "/var/www/html/../../etc", 5); files != nil {
		t.Error("dot-dot must not escape")
	}
}

func TestBlankBuilderPage(t *testing.T) {
	blank := "<!DOCTYPE html><html><head><link href=\"x.css\" rel=\"stylesheet\"></head><body>\n<body></body>\n<style>* { box-sizing: border-box; }</style></body></html>"
	if !blankPage(blank) {
		t.Error("an editor page with no content is blank")
	}
	for _, s := range []string{"<body><h1>Hi</h1></body>", "<body><img src=\"a.png\"></body>"} {
		if blankPage(s) {
			t.Errorf("%q is not blank", s)
		}
	}
}

func TestComposerLock(t *testing.T) {
	pk := parseComposerLock([]byte(`{"packages":[{"name":"slim/slim","version":"4.15.3"}],"packages-dev":[{"name":"phpunit/phpunit","version":"9.6.0"}]}`))
	if len(pk) != 2 || pk[0].Dev || !pk[1].Dev || pk[0].Version != "4.15.3" {
		t.Errorf("parseComposerLock = %+v", pk)
	}
	if parseComposerLock([]byte("nope")) != nil {
		t.Error("bad json is nil")
	}
}

func TestPHPAppProject(t *testing.T) {
	webtest.ModernOnly(t)
	pd := mgrPage(t, "shared.tests.openpanel.org")
	data := PHPAppPageData{pageData: pd, Container: ContainerInfo{SiteName: "shared.tests.openpanel.org", CreatedDate: "2026-10-29T11:12:18Z"},
		PHPVersion: "8.5", InitialProject: "slim/slim-skeleton", AutorunComposerInstall: true,
		ComposerRows: []ComposerRow{{"php", "^7.4 || ^8.0", "", false}, {"ext-json", "*", "", false}, {"slim/slim", "^4.10", "4.15.3", false}, {"phpunit/phpunit", "^9.6", "9.6.0", true}},
		Packages:     []ComposerPackage{{"slim/slim", "4.15.3", false}, {"psr/log", "3.0.2", false}, {"phpunit/phpunit", "9.6.0", true}},
		Files: []SiteEntry{{Name: "app", Dir: true, Modified: "2026-10-29"}, {Name: "public", Dir: true, Modified: "2026-10-29"}, {Name: "vendor", Dir: true, Modified: "2026-10-29"},
			{Name: "composer.json", Size: "1.5 KB"}, {Name: "composer.lock", Size: "129 KB"}, {Name: "README.md", Size: "1.5 KB"}},
		HiddenFiles: 4}
	body := renderPage(t, phpAppPage, data)
	mustContain(t, body, "php-fpm-8.5", "packagist.org/packages/slim/slim-skeleton", "op-sp-tree", "composer.lock", "4 dotfiles not shown",
		"8.5 (container)", "built in", "Installed packages", "data-dev", "Composer runs", "loadPHPAppLogs", "setTimeout(loadPHPAppLogs, 0)",
		`id="optimize_autoloader"`, "Open site", "File Manager")
	if strings.Contains(body, "w-100") {
		t.Error("w-100 is a 400px utility")
	}
}

func TestN8nSideCards(t *testing.T) {
	webtest.ModernOnly(t)
	pd := mgrPage(t, "n8n.tests.openpanel.org")
	data := N8nAppPageData{pageData: pd, Container: ContainerInfo{SiteName: "n8n.tests.openpanel.org", Version: "3.0.1", Port: "5678", CreatedDate: "2026-10-29T11:22:12Z"},
		Service: "N8N", PM2Status: "true", CPU: "1.0", RAM: "1.0", PIDs: "500", CurrentVersion: "3.0.1",
		Mounts: []ComposeMount{{"n8n_data", "/home/node/.n8n"}}, Networks: []string{"www", "db"},
		Image: "n8nio/n8n:3.0.1", Restart: "always", WebhookURL: "https://n8n.tests.openpanel.org/", Timezone: "UTC", EnvVars: "N8N_HOST=n8n.tests.openpanel.org"}
	body := renderPage(t, n8nAppPage, data)
	mustContain(t, body, "op-dk-usage", "n8n:5678", "n8nio/n8n:3.0.1", "always", "www, db", "Editor and webhooks", "Open in", ">Environment<", "window.loadAppLogs")
	for _, gone := range []string{"pagespeed", "Env Vars"} {
		if strings.Contains(body, gone) {
			t.Errorf("still has %q", gone)
		}
	}
}

func TestWebsiteBuilderBlankAndFiles(t *testing.T) {
	webtest.ModernOnly(t)
	pd := mgrPage(t, "website-builder.tests.openpanel.org")
	data := WebsiteBuilderPageData{pageData: pd, Container: ContainerInfo{ID: 7, CreatedDate: "2026-10-29T11:22:37Z"}, HasSource: true, BlankPage: true,
		IndexSource: "<body></body>", Files: []SiteEntry{{Name: "index.html", Size: "345 B", Modified: "2026-10-29"}, {Name: "style.css", Size: "47 B", Modified: "2026-10-29"}}}
	body := renderPage(t, websiteBuilderPage, data)
	mustContain(t, body, "This page is still blank", "style.css", "saved 2026-10-29")
	if strings.Contains(body, "<iframe") {
		t.Error("a blank page shows the empty state, not the iframe")
	}
	data.BlankPage = false
	mustContain(t, renderPage(t, websiteBuilderPage, data), "<iframe")
}

func TestCMSLayoutAndDBVersion(t *testing.T) {
	webtest.ModernOnly(t)
	for _, slug := range []string{"joomla", "nextcloud", "matomo"} {
		pd := mgrPage(t, slug+".tests.openpanel.org")
		data := CMSAppPageData{pageData: pd, Container: ContainerInfo{ID: 3, SiteName: slug + ".tests.openpanel.org", CreatedDate: "2026-10-29T11:15:26Z"},
			Version: "6.2.0", PHPVersion: "7.4", CurrentPHPVersion: "7.4", AvailablePHPVersions: []string{"8.5", "7.4"}, MainDomain: slug + ".tests.openpanel.org",
			DBInfo: map[string]string{"name": slug + "_fec8yg", "user": "9496x8vqef", "password": "secret", "host": "localhost", "prefix": "jos_"}, MySQLType: "mariadb",
			App: web.NewAppNames(slug, cmsPages[slug].name)}
		body := renderPage(t, cmsPages[slug].page, data)
		mustContain(t, body, "op-cms-layout")
		if strings.Contains(body, "Version not found") || strings.Contains(body, "w-100") {
			t.Errorf("%s: leftover", slug)
		}
		data.MySQLVersion = "MariaDB 13.0.2"
		mustContain(t, renderPage(t, cmsPages[slug].page, data), "13.0.2")
	}
}
