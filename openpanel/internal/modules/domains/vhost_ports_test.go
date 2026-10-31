package domains

import (
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestParseVhostFile(t *testing.T) {
	apache := "<VirtualHost *:80>\n    ServerName php.example.com\n</VirtualHost>\n\n# <VirtualHost *:8443>\n<VirtualHost *:443>\n    SSLEngine on\n</VirtualHost>\n"
	nginx := "server {\n    listen 80;\n    listen [::]:80;\n}\nserver {\n    listen 443 ssl http2;\n    listen [::]:443 ssl;\n}\n"
	varnish := "<VirtualHost *:8080>\n</VirtualHost>\n<VirtualHost *:443>\n</VirtualHost>"
	ols := "docRoot /var/www/html\nvhssl  {\n  keyFile /etc/ssl/key.pem\n}\n"
	cases := []struct {
		name  string
		in    string
		lines int
		ports []string
		ssl   bool
	}{
		{"apache", apache, 8, []string{"80", "443 SSL"}, true},
		{"nginx", nginx, 8, []string{"80", "443 SSL"}, true},
		{"varnish apache", varnish, 4, []string{"8080", "443 SSL"}, true},
		{"openlitespeed", ols, 4, nil, true},
		{"plain", "listen 127.0.0.1:8081;\n", 1, []string{"8081"}, false},
	}
	for _, c := range cases {
		got := parseVhostFile(c.in)
		if got.Lines != c.lines || got.SSL != c.ssl || !reflect.DeepEqual(got.Ports, c.ports) {
			t.Errorf("%s: got %+v, want lines=%d ports=%v ssl=%v", c.name, got, c.lines, c.ports, c.ssl)
		}
	}
}

func TestRenderVhostListPorts(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := VhostPageData{
		LayoutData: baseLayout(mgr, "/domains/vhosts"), WebServerPreference: "apache",
		Domains:    []appctx.Domain{{DomainURL: "example.com"}, {DomainURL: "nofile.com"}},
		VhostFiles: map[string]VhostFileInfo{"example.com": {Lines: 61, Ports: []string{"80", "443 SSL"}, SSL: true}},
	}
	w := httptest.NewRecorder()
	if err := vhostPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"61 lines", ">443 SSL<", ">Ports<"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}
}
