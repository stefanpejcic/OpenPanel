package webserverconf

import (
	"testing"
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
