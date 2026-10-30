package webserverconf

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestParseServerVersion(t *testing.T) {
	cases := map[string]string{
		"Server version: Apache/2.4.69 (Unix)\nServer built:   Oct 21 2026 22:47:29": "Apache 2.4.69",
		"nginx version: nginx/1.28.0":                    "Nginx 1.28.0",
		"nginx version: openresty/1.27.1.2":              "OpenResty 1.27.1.2",
		"LiteSpeed/1.8.3 Open \n\tmodule versions:":      "LiteSpeed 1.8.3",
		"Error: no container with name or ID \"apache\"": "",
	}
	for in, want := range cases {
		if got := parseServerVersion(in); got != want {
			t.Errorf("parseServerVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveImage(t *testing.T) {
	env := map[string]string{"APACHE_VERSION": "alpine", "EMPTY": ""}
	cases := map[string]string{
		"httpd:${APACHE_VERSION:-latest}":  "httpd:alpine",
		"nginx:${NGINX_VERSION:-latest}":   "nginx:latest",
		"img:${EMPTY:-stable}":             "img:stable",
		"shinsenter/php:8.5-fpm":           "shinsenter/php:8.5-fpm",
		"openresty/openresty:${OR-alpine}": "openresty/openresty:alpine",
	}
	for in, want := range cases {
		if got := resolveImage(in, env); got != want {
			t.Errorf("resolveImage(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLastLines(t *testing.T) {
	if got := lastLines("a\nb\nc\n", 2); got != "b\nc" {
		t.Errorf("lastLines = %q", got)
	}
}

func TestRenderWebserverConfSummary(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := WebserverConfPageData{
		LayoutData: baseLayout(mgr, "/server/webserver_conf"), Service: "apache", Path: "httpd.conf",
		Summary: ServerSummary{
			Version: "Apache 2.4.69", Image: "httpd:alpine", State: "running", Uptime: "1d 19h",
			CPU: "0.5", RAM: "1G", PIDs: "2000", TestCommand: "apachectl configtest",
		},
	}
	w := httptest.NewRecorder()
	if err := webserverConfPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"Apache 2.4.69", "httpd:alpine", "up 1d 19h", "0.5 <small>CPU</small> · 1G", "2000 processes max", "apachectl configtest", "/server/webserver_conf/test", "Test config"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in body", want)
		}
	}

	data.Summary = ServerSummary{}
	w = httptest.NewRecorder()
	if err := webserverConfPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(w.Body.String(), "Test config") {
		t.Error("tiles should be left out without a summary")
	}
}
