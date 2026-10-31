package websites

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
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
