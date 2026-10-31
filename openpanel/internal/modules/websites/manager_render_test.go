package websites

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestRuntimeManagerExtras(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	for _, statsOn := range []bool{true, false} {
		layout := baseLayout(mgr, "/website")
		layout.UserAllowed["docker"] = statsOn
		data := PythonNodeAppsPageData{
			pageData:     pageData{LayoutData: layout, CurrentDomain: "nodejs.example.com", Docroot: "/var/www/html/nodejs.example.com"},
			Container:    ContainerInfo{SiteName: "nodejs.example.com", Type: "NodeJS", Container: "NODEAPP", Version: "24.22.2"},
			PM2Data:      map[string]string{"prefix": "NODEAPP_NODE_", "status": "true"},
			Type:         "nodejs",
			Service:      "NODEAPP",
			StartCommand: "npm install && node /var/www/html/nodejs.example.com/index.js",
			Manifest: Manifest{File: "package.json", Path: "/var/www/html/nodejs.example.com/package.json", Found: true, Text: `{"dependencies": {"express": "^4.21.2"}}`,
				Groups: []ManifestGroup{{Title: "Dependencies", ValueLabel: "Constraint", HasInstalled: true, Rows: []ManifestRow{{"express", "^4.21.2", "4.22.3"}, {"<b>x</b>", "1", ""}}}}},
		}
		w := httptest.NewRecorder()
		if err := pythonNodeAppsPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		for _, want := range []string{"npm install &amp;&amp; node /var/www/html/nodejs.example.com/index.js", "4.22.3", "&lt;b&gt;x&lt;/b&gt;", "not installed", `class="op-btn op-btn-primary op-btn-sm install-btn"`} {
			if !strings.Contains(body, want) {
				t.Errorf("stats=%v: missing %q", statsOn, want)
			}
		}
		if got := strings.Contains(body, "op-rt-meter"); got != statsOn {
			t.Errorf("usage meters shown = %v, want %v", got, statsOn)
		}
		if statsOn && !strings.Contains(body, `var name = "nodeapp"`) {
			t.Error("usage meters must poll the lowercase podman name")
		}
		if !statsOn && !strings.Contains(body, "CPU limit") {
			t.Error("limit tiles should stay when stats aren't allowed")
		}
	}

	// missing manifest renders a note, not an empty table
	data := PythonNodeAppsPageData{
		pageData:  pageData{LayoutData: baseLayout(mgr, "/website"), CurrentDomain: "py.example.com", Docroot: "/var/www/html/py.example.com"},
		Container: ContainerInfo{Type: "Python", Container: "FLASKAPP"},
		PM2Data:   map[string]string{"prefix": "FLASKAPP_PY_", "status": "false"},
		Type:      "python",
		Manifest:  Manifest{File: "requirements.txt", Path: "/var/www/html/py.example.com/requirements.txt"},
	}
	w := httptest.NewRecorder()
	if err := pythonNodeAppsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), "was not found in the app folder") {
		t.Error("expected the missing manifest note")
	}
}

func TestN8nRouteStrip(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	pd := pageData{LayoutData: baseLayout(mgr, "/website"), CurrentDomain: "n8n.example.com", Docroot: "/var/www/html/n8n.example.com"}
	data := N8nAppPageData{pageData: pd, Service: "n8n", PM2Status: "true", CPU: "1.0", RAM: "1.0", PIDs: "500",
		Mounts: []ComposeMount{{"n8n_data", "/home/node/.n8n"}}, Networks: []string{"www", "db"}}
	w := httptest.NewRecorder()
	if err := n8nAppPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"op-dk-route is-4", "n8n_data", "/home/node/.n8n", "www, db"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}

	data.Mounts, data.Networks = nil, nil
	w = httptest.NewRecorder()
	if err := n8nAppPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render without compose info: %v", err)
	}
	if strings.Contains(w.Body.String(), "is-4") {
		t.Error("no volume means three nodes")
	}
}

func TestPHPAppComposerTable(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	pd := pageData{LayoutData: baseLayout(mgr, "/website"), CurrentDomain: "shared.example.com", Docroot: "/var/www/html/shared.example.com"}
	for _, rows := range [][]ComposerRow{{{"slim/slim", "^4.10", "4.11.0", false}, {"phpunit/phpunit", "^9.5", "", true}}, nil} {
		w := httptest.NewRecorder()
		if err := phpAppPage.Render(w, 200, PHPAppPageData{pageData: pd, ComposerRows: rows}); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if rows != nil && (!strings.Contains(body, "4.11.0") || !strings.Contains(body, "phpunit/phpunit")) {
			t.Error("missing composer rows")
		}
		if rows == nil && !strings.Contains(body, "No composer.json found") {
			t.Error("missing empty note")
		}
	}
}

func TestWebsiteBuilderSourceView(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	pd := pageData{LayoutData: baseLayout(mgr, "/website"), CurrentDomain: "builder.example.com", Docroot: "/var/www/html/builder.example.com"}
	w := httptest.NewRecorder()
	data := WebsiteBuilderPageData{pageData: pd, HasSource: true, IndexSource: `<html><script>alert(1)</script></html>`}
	if err := websiteBuilderPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "&lt;script&gt;alert(1)&lt;/script&gt;") || strings.Contains(body, "<script>alert(1)") {
		t.Error("source must be shown escaped")
	}
	w = httptest.NewRecorder()
	if err := websiteBuilderPage.Render(w, 200, WebsiteBuilderPageData{pageData: pd}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(w.Body.String(), `id="op-sb-src"`) {
		t.Error("no source means no source view")
	}
}
