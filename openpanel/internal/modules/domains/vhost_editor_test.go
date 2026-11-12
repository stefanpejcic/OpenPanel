package domains

import (
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestRenderVhostEditor(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	content := "<VirtualHost *:80>\n    ServerName example.com\n</VirtualHost>\n<VirtualHost *:443>\n    SSLEngine on\n</VirtualHost>\n"
	data := VhostPageData{
		LayoutData: baseLayout(mgr, "/domains/vhosts"), DomainName: "example.com", WebServerPreference: "apache", VhostContent: content,
		Domain: appctx.Domain{DomainURL: "example.com", Docroot: "/var/www/html/example.com", PHPVersion: "8.3"}, File: parseVhostFile(content),
	}
	w := httptest.NewRecorder()
	if err := vhostPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{
		"js/conf-editor.js", `opConfEditor({ kind: &#34;apache&#34;`, `name="vhost_content"`, `name="domain_name" value="example.com"`,
		"example.com.conf", "/var/www/html/example.com", "php-fpm-8.3", "80 &#43; 443 SSL", "SSL is set up", "op-confed-find", "Show first issue",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Count(body, "data-no-live-validate") < 2 {
		t.Error("want find and go-to-line inputs opted out of live validation")
	}
}

func TestRenderVhostEditorNginx(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	content := "server {\n    listen 80;\n}\n"
	data := VhostPageData{LayoutData: baseLayout(mgr, "/domains/vhosts"), DomainName: "example.com", WebServerPreference: "nginx", VhostContent: content, Domain: appctx.Domain{DomainURL: "example.com"}, File: parseVhostFile(content)}
	w := httptest.NewRecorder()
	if err := vhostPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	if !strings.Contains(body, `kind: &#34;brace&#34;`) || !strings.Contains(body, "No PHP version set") || !strings.Contains(body, "No SSL block") {
		t.Error("expected brace check and empty handler/SSL tiles for nginx without PHP")
	}
}
