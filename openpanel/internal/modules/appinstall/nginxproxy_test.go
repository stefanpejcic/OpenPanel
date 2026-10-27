package appinstall

import (
	"strings"
	"testing"
)

const testNginxVhost = `server {
    listen 80;
    server_name example.com www.example.com;
    root /var/www/html/example.com;

    location / {
        real_ip_header    X-Forwarded-For;
        try_files $uri $uri/ /index.php?$args;
        index index.php index.html /default_page.html;
        autoindex on;
    }

    location = /default_page.html {
        alias /etc/nginx/default_page.html;
    }

    location ~ \.php$ {
        fastcgi_pass php-fpm-8.5:9000;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico)$ {
        expires max;
    }

    location = /favicon.ico {
        log_not_found off;
    }
}

server {
    listen 443 ssl;
    server_name example.com www.example.com;

    location / {
        try_files $uri $uri/ /index.php?$args;
    }
}
`

var testNginxHeaders = []string{"proxy_http_version 1.1;", "proxy_set_header Host $host;", "proxy_set_header X-Real-IP $remote_addr;", "proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;", "proxy_set_header X-Forwarded-Proto $scheme;"}

// activeLines drops comment lines so only what nginx actually reads is checked
func activeLines(conf string) string {
	var kept []string
	for _, l := range strings.Split(conf, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(l), "#") {
			kept = append(kept, l)
		}
	}
	return strings.Join(kept, "\n")
}

func TestAddNginxRootProxy(t *testing.T) {
	got, err := addNginxProxy(testNginxVhost, nginxProxy{Service: "nodeapp", Port: 3000, Resolver: "10.89.0.1", Headers: testNginxHeaders})
	if err != nil {
		t.Fatal(err)
	}
	active := activeLines(got)
	if strings.Count(active, "location ^~ / {") != 2 || strings.Contains(active, "location / {") || strings.Contains(active, "try_files") {
		t.Fatalf("location / not replaced in both server blocks:\n%s", got)
	}
	for _, want := range []string{"resolver 10.89.0.1 valid=10s ipv6=off;", "set $openpanel_upstream nodeapp:3000;", "proxy_pass http://$openpanel_upstream;", "proxy_set_header X-Forwarded-Proto $scheme;"} {
		if strings.Count(active, want) != 2 {
			t.Fatalf("want %q in both server blocks:\n%s", want, got)
		}
	}
	if strings.Count(active, "{") != strings.Count(active, "}") {
		t.Fatalf("unbalanced braces:\n%s", got)
	}
	if strings.Contains(active, "rewrite") || strings.Contains(active, "http://nodeapp") {
		t.Fatalf("root proxy should pass the URI as is:\n%s", got)
	}
	if strings.Contains(active, "location =") || !strings.Contains(active, `location ~ \.php$ {`) {
		t.Fatalf("exact-match locations should go to the app, regex ones stay:\n%s", got)
	}
}

func TestAddNginxSubdirProxy(t *testing.T) {
	got, err := addNginxProxy(testNginxVhost, nginxProxy{Service: "api", Port: 8000, Subdirectory: "my.api", Resolver: "10.89.0.1", Headers: testNginxHeaders})
	if err != nil {
		t.Fatal(err)
	}
	active := activeLines(got)
	if strings.Count(active, "location ^~ /my.api/ {") != 2 || strings.Count(active, `rewrite ^/my\.api/(.*)$ /$1 break;`) != 2 {
		t.Fatalf("missing subdirectory location:\n%s", got)
	}
	if strings.Count(active, "location / {") != 2 || strings.Count(active, "try_files") != 2 || strings.Count(active, "location =") != 2 {
		t.Fatalf("site's own locations should stay:\n%s", got)
	}
}

func TestNginxProxyRoundTrip(t *testing.T) {
	cases := map[string]nginxProxy{
		"root":                {Service: "nodeapp", Port: 3000, Resolver: "10.89.0.1", Headers: testNginxHeaders},
		"subdir":              {Service: "api", Port: 8000, Subdirectory: "api", Resolver: "10.89.0.1", Headers: testNginxHeaders},
		"root no resolver":    {Service: "kuma", Port: 3001, Headers: append(testNginxHeaders, `proxy_set_header Connection "upgrade";`)},
		"subdir no resolver":  {Service: "kuma", Subdirectory: "a/b", Headers: testNginxHeaders},
		"no port, no headers": {Service: "web", Resolver: "10.89.0.1"},
	}
	for name, p := range cases {
		with, err := addNginxProxy(testNginxVhost, p)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		got, changed := stripServiceProxies(with, p.Service)
		if !changed || got != testNginxVhost {
			t.Fatalf("%s: not restored byte for byte:\n%s", name, got)
		}
	}
}

func TestNginxProxyWithoutResolverUsesPlainProxyPass(t *testing.T) {
	got, _ := addNginxProxy(testNginxVhost, nginxProxy{Service: "api", Port: 8000, Subdirectory: "api"})
	if !strings.Contains(got, "proxy_pass http://api:8000/;") || strings.Contains(got, "resolver") {
		t.Fatalf("bad fallback:\n%s", got)
	}
}

func TestAddNginxProxyTwiceIsNoop(t *testing.T) {
	p := nginxProxy{Service: "nodeapp", Port: 3000, Resolver: "10.89.0.1"}
	once, _ := addNginxProxy(testNginxVhost, p)
	twice, err := addNginxProxy(once, p)
	if err != nil || twice != once {
		t.Fatalf("second add changed the file: %v\n%s", err, twice)
	}
}

func TestStripNginxProxyKeepsOtherApps(t *testing.T) {
	with, _ := addNginxProxy(testNginxVhost, nginxProxy{Service: "kuma2", Subdirectory: "two", Resolver: "10.89.0.1"})
	with, _ = addNginxProxy(with, nginxProxy{Service: "kuma", Subdirectory: "one", Resolver: "10.89.0.1"})
	got, _ := stripServiceProxies(with, "kuma")
	if strings.Contains(got, "/one/") || strings.Count(got, "location ^~ /two/ {") != 2 {
		t.Fatalf("bad strip:\n%s", got)
	}
	if back, _ := stripServiceProxies(got, "kuma2"); back != testNginxVhost {
		t.Fatalf("not restored:\n%s", back)
	}
}

func TestAddNginxProxyRejectsOddSubdirectory(t *testing.T) {
	for _, sub := range []string{"a b", "a;b", "a{b", "a\nb", `a"b`, "a//b", "a/../b"} {
		if _, err := addNginxProxy(testNginxVhost, nginxProxy{Service: "x", Subdirectory: sub}); err == nil {
			t.Fatalf("accepted subdirectory %q", sub)
		}
	}
}

func TestAddNginxRootProxyNestedBraces(t *testing.T) {
	conf := "server {\n  location / {\n    # a } in a comment\n    add_header X-Test \"}\";\n    if ($bad) { return 403; }\n  }\n  location /x { return 404; }\n}\n"
	with, err := addNginxProxy(conf, nginxProxy{Service: "app", Resolver: "10.89.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(with, "  location /x { return 404; }\n") || strings.Contains(activeLines(with), "return 403") {
		t.Fatalf("wrong block replaced:\n%s", with)
	}
	if got, _ := stripServiceProxies(with, "app"); got != conf {
		t.Fatalf("not restored:\n%s", got)
	}
}

func TestAddNginxProxyNeedsLocationRoot(t *testing.T) {
	if _, err := addNginxProxy("server {\n  listen 80;\n}\n", nginxProxy{Service: "x"}); err == nil {
		t.Fatal("expected an error without a location / block")
	}
	if _, err := addNginxProxy("server {\n  location / {\n    root /x;\n", nginxProxy{Service: "x"}); err == nil {
		t.Fatal("expected an error for an unclosed location / block")
	}
}
