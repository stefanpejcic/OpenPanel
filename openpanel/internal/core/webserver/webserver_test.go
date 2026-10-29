package webserver

import (
	"strings"
	"testing"
)

func TestConfigTestPassed(t *testing.T) {
	cases := []struct {
		service string
		output  string
		want    bool
	}{
		{"nginx", "nginx: the configuration file /etc/nginx/nginx.conf syntax is ok\nnginx: configuration file /etc/nginx/nginx.conf test is successful", true},
		{"nginx", "nginx: [emerg] unexpected \"}\" in /etc/nginx/nginx.conf:12\nnginx: configuration file /etc/nginx/nginx.conf test failed", false},
		{"openresty", "nginx: configuration file /usr/local/openresty/nginx/conf/nginx.conf test is successful", true},
		{"apache", "Syntax OK", true},
		{"apache", "AH00526: Syntax error on line 3 of /usr/local/apache2/conf/httpd.conf:", false},
		// warnings only, openlitespeed still exits 1 for these
		{"openlitespeed", "2026-10-27 12:55:40.310143 [WARN] [91] [config:server:tuning] invalid value of <maxCachedFileSize>:65536, use default=4096", true},
		{"openlitespeed", "", true},
		{"openlitespeed", "2026-10-27 12:55:40.310143 [ERROR] [config:server:basics] Invalid User Name((null)) or Group Name((null))!", false},
	}
	for _, c := range cases {
		test, ok := webserverConfigTests[c.service]
		if !ok {
			t.Fatalf("no config test for %s", c.service)
		}
		if got := test.passed(c.output); got != c.want {
			t.Errorf("%s passed(%q) = %v, want %v", c.service, c.output, got, c.want)
		}
	}
}

func TestHasConfigTest(t *testing.T) {
	for _, ws := range []string{"nginx", "openresty", "apache", "openlitespeed"} {
		if !HasConfigTest(ws) {
			t.Errorf("HasConfigTest(%q) = false, want true", ws)
		}
	}
	if HasConfigTest("litespeed") {
		t.Error("HasConfigTest(\"litespeed\") = true, LiteSpeed Enterprise has no test")
	}
}

func TestComposeHostnames(t *testing.T) {
	compose := map[string]any{"services": map[string]any{
		"openresty":   map[string]any{"container_name": "openresty", "image": "openresty/openresty:alpine"},
		"php-fpm-8.4": map[string]any{"container_name": "php-fpm-8.4"},
		"mysql":       map[string]any{"container_name": "mysqldb"},
		"redis":       map[string]any{},
	}}
	got := strings.Join(composeHostnames(compose), " ")
	if want := "mysql mysqldb openresty php-fpm-8.4 redis"; got != want {
		t.Errorf("composeHostnames = %q, want %q", got, want)
	}
	if got := composeHostnames(map[string]any{}); len(got) != 0 {
		t.Errorf("composeHostnames without services = %v, want none", got)
	}
}

func TestLookupFailedOnlyMatchesDNS(t *testing.T) {
	if !lookupFailed.MatchString(`nginx: [emerg] host not found in upstream "php-fpm-8.4:9000" in /etc/nginx/conf.d/a.com.conf:30`) {
		t.Error("lookup failure not retried")
	}
	if lookupFailed.MatchString(`nginx: [emerg] unknown directive "servr_name" in /etc/nginx/nginx.conf:20`) {
		t.Error("syntax error would be retried")
	}
}
