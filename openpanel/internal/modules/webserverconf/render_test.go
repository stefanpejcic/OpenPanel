package webserverconf

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func baseLayout(mgr *i18n.Manager, path string) web.LayoutData {
	userAllowed := map[string]bool{"dashboard": true, "webserver_conf": true}
	return web.LayoutData{
		Title: "Nginx Configuration Editor", BrandName: "Test Panel", CSRFToken: "test-csrf-token", PanelDir: "ltr",
		NavGroups: web.BuildClassicSidebarNav(userAllowed, nil, path), UserAllowed: userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed), CurrentUsername: "testuser",
		RequestPath: path, AdminPort: "2087", PasswordStrength: 50, T: mgr.Translator("en"),
	}
}

func TestRenderWebserverConfPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	data := WebserverConfPageData{
		LayoutData: baseLayout(mgr, "/server/webserver_conf"), Service: "nginx", Path: "nginx.conf",
		CanRestoreDefault: true, FileContent: "server { listen 80; }",
	}
	w := httptest.NewRecorder()
	if err := webserverConfPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "server { listen 80; }") {
		t.Error("expected file content in body")
	}
	if !strings.Contains(body, "Restore Default") {
		t.Error("expected restore-default button when CanRestoreDefault is true")
	}
}

func TestRenderWebserverConfPageRejectedSave(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	layout := baseLayout(mgr, "/server/webserver_conf")
	layout.UserAllowed["edit_vhost"] = true
	data := WebserverConfPageData{
		LayoutData: layout, Service: "nginx", Path: "nginx.conf", FileContent: "http { broken",
		TestFailure: &testFailure{
			Problems:    []string{`nginx.conf, line 2: unexpected end of file`},
			Hints:       []string{"php-fpm-8.4 could not be found."},
			Line:        2,
			VhostDomain: "a.com",
		},
	}
	w := httptest.NewRecorder()
	if err := webserverConfPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	wants := []string{"http { broken", "nginx.conf, line 2: unexpected end of file", "php-fpm-8.4 could not be found.", "/domains/vhosts?domain=a.com", "Go to line 2"}
	if web.Theme() == web.ThemeClassic {
		wants = append(wants, "editor.addLineClass( 2  - 1")
	} else {
		wants = append(wants, "unsaved: true, errorLine: 2,", `@click="jump(2)"`)
	}
	for _, want := range wants {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
}

func TestRenderWebserverConfPageNoRestoreOption(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)

	data := WebserverConfPageData{
		LayoutData: baseLayout(mgr, "/server/webserver_conf"), Service: "nginx", Path: "nginx.conf",
		CanRestoreDefault: false, FileContent: "",
	}
	w := httptest.NewRecorder()
	if err := webserverConfPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(w.Body.String(), "Restore Default") {
		t.Error("did not expect restore-default button when CanRestoreDefault is false")
	}
}

func TestRenderWebserverConfEditor(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	conf, err := os.ReadFile("testdata/httpd.conf")
	if err != nil {
		t.Fatal(err)
	}
	layout := baseLayout(mgr, "/server/webserver_conf")
	layout.Title = "Apache Configuration Editor"
	data := WebserverConfPageData{
		LayoutData: layout, Service: "apache", Path: "httpd.conf", CanRestoreDefault: true,
		FileContent: string(conf),
		Summary:     ServerSummary{State: "running", Version: "Apache 2.4.62", TestCommand: "httpd -t"},
	}
	w := httptest.NewRecorder()
	if err := webserverConfPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{
		"js/conf-editor.js", `opConfEditor({ kind: &#34;apache&#34;`, `name="editor_content"`, "op-confed-find", "Previous match", "Next match",
		`x-ref="line"`, "Ctrl&#43;S saves", `value="restore_default"`, "Restore the default configuration?", "Apache configuration", "httpd -t", "LoadModule rewrite_module",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	// both toolbar inputs opt out of the global "Looks good." hint
	if n := strings.Count(body, "data-no-live-validate"); n < 2 {
		t.Errorf("want find and go-to-line inputs opted out of live validation, got %d", n)
	}
	if strings.Contains(body, "onclick=\"return confirm(") {
		t.Error("restore should use the sheet, not confirm()")
	}
}

func TestRenderWebserverConfNoFile(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := WebserverConfPageData{LayoutData: baseLayout(mgr, "/server/webserver_conf"), Service: "caddy"}
	w := httptest.NewRecorder()
	if err := webserverConfPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "No configuration file to edit") {
		t.Error("expected the empty state without a config file")
	}
	if strings.Contains(body, `name="editor_content"`) {
		t.Error("did not expect an editor without a config file")
	}
}
