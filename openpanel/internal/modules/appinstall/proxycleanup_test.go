package appinstall

import (
	"strings"
	"testing"
)

const nginxHeaders = "proxy_http_version 1.1;\nproxy_set_header Host $host;\nproxy_set_header X-Real-IP $remote_addr;\nproxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\nproxy_set_header X-Forwarded-Proto $scheme;" +
	"\n            proxy_set_header Upgrade $http_upgrade;\n            proxy_set_header Connection \"upgrade\";"

const nginxVhost = "server {\n    listen 80;\n    location / {\n        try_files $uri $uri/ /index.php?$args;\n        autoindex on;\n    }\n}\n"

func TestStripNginxRootProxy(t *testing.T) {
	// same insert editNginxConfig does for a root install
	end := strings.Index(nginxVhost, "    }\n}")
	insert := "\n        proxy_pass http://kuma:3001/;\n            " + nginxHeaders + "\n        "
	withProxy := nginxVhost[:end] + "\n" + insert + nginxVhost[end:]

	got, changed := stripServiceProxies(withProxy, "kuma")
	if !changed || strings.Contains(got, "proxy_") || !strings.Contains(got, "location / {") || !strings.Contains(got, "autoindex on;") {
		t.Fatalf("bad strip:\n%s", got)
	}
	if strings.Count(got, "{") != strings.Count(got, "}") {
		t.Fatalf("unbalanced braces:\n%s", got)
	}
}

func TestStripNginxSubdirProxy(t *testing.T) {
	marker := strings.Index(nginxVhost, "    location / {")
	block := "\n        location /app/ {\n            proxy_pass http://kuma/;\n            " + nginxHeaders + "\n        }\n        "
	withProxy := nginxVhost[:marker] + block + "\n" + nginxVhost[marker:]

	got, _ := stripServiceProxies(withProxy, "kuma")
	if strings.Contains(got, "/app/") || strings.Contains(got, "proxy_") || !strings.Contains(got, "location / {") {
		t.Fatalf("bad strip:\n%s", got)
	}
}

func TestStripApacheAndOtherServicesKept(t *testing.T) {
	in := "<VirtualHost *:80>\n\tDirectoryIndex index.php\n\tProxyPass / http://kuma:3001/ upgrade=websocket\n\tProxyPassReverse / http://kuma:3001/\n\tProxyPass /x/ http://kuma2/\n</VirtualHost>\n"
	got, changed := stripServiceProxies(in, "kuma")
	if !changed || strings.Contains(got, "http://kuma:") || !strings.Contains(got, "http://kuma2/") {
		t.Fatalf("bad strip:\n%s", got)
	}
}

func TestStripLitespeedBlock(t *testing.T) {
	in := "docRoot /var/www\n\n# PM2-PROXY-START:kuma\nextprocessor kuma {\n  address http://kuma:3001\n}\n# PM2-PROXY-END:kuma\n"
	got, _ := stripServiceProxies(in, "kuma")
	if strings.Contains(got, "kuma") || !strings.Contains(got, "docRoot") {
		t.Fatalf("bad strip:\n%q", got)
	}
}

func TestStripNoMatch(t *testing.T) {
	if _, changed := stripServiceProxies(nginxVhost, "kuma"); changed {
		t.Fatal("changed a vhost without a proxy")
	}
}

func TestStripApacheKeepsMarkerIndent(t *testing.T) {
	orig := "<VirtualHost *:80>\n    DirectoryIndex index.php index.html default_page.html\n</VirtualHost>\n"
	// what editApacheConfig produces for a root proxy
	marker := "DirectoryIndex index.php index.html default_page.html"
	i := strings.Index(orig, marker)
	with := orig[:i] + "\tProxyPass / http://n8n:5678/ upgrade=websocket\n\tProxyPassReverse / http://n8n:5678/\n" + orig[i:]
	if got, _ := stripServiceProxies(with, "n8n"); got != orig {
		t.Fatalf("got %q", got)
	}
}

// shape of the shipped nginx domain template, with the locations that used to win over the app
const nginxDomainVhost = `server {
    listen 80;
    root /var/www/html/example.com;

    location / {
        try_files $uri $uri/ /index.php?$args;
        index index.php index.html /default_page.html;
        autoindex on;
    }

    location ~ \.php$ {
        fastcgi_pass php-fpm-8.5:9000;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico)$ {
        expires max;
    }
}
`

func rootProxyLines(svc string) string {
	return "\n        proxy_pass http://" + svc + ":5678/;\n            " + nginxHeaders + "\n        "
}

func TestRootAppTakesLocationAndDeleteRestoresIt(t *testing.T) {
	added, errMsg := insertNginxProxy(nginxDomainVhost, "", "n8n", rootProxyLines("n8n"))
	if errMsg != "" {
		t.Fatal(errMsg)
	}
	if !strings.Contains(added, "location ^~ / { # OPENPANEL-APP:n8n\n") || !strings.Contains(added, "# OPENPANEL-APP:n8n try_files $uri") {
		t.Fatalf("root app didn't take location / over:\n%s", added)
	}
	if strings.Count(added, "{") != strings.Count(added, "}") {
		t.Fatalf("unbalanced braces:\n%s", added)
	}

	got, changed := stripServiceProxies(added, "n8n")
	if !changed || strings.Contains(got, "OPENPANEL-APP") || strings.Contains(got, "proxy_") {
		t.Fatalf("bad strip:\n%s", got)
	}
	if !strings.Contains(got, "    location / {\n        try_files $uri $uri/ /index.php?$args;") {
		t.Fatalf("location / or try_files not restored:\n%s", got)
	}
	if strings.Count(got, "{") != strings.Count(got, "}") {
		t.Fatalf("unbalanced braces:\n%s", got)
	}
}

func TestRootAppStripLeavesOtherServiceAlone(t *testing.T) {
	added, _ := insertNginxProxy(nginxDomainVhost, "", "kuma2", rootProxyLines("kuma2"))
	got, changed := stripServiceProxies(added, "kuma")
	if changed || got != added {
		t.Fatalf("stripping kuma touched kuma2:\n%s", got)
	}
}

func TestSubdirAppUsesPrefixPriorityAndDeleteRemovesIt(t *testing.T) {
	lines := "\n        location ^~ /app/ {\n            proxy_pass http://kuma/;\n            " + nginxHeaders + "\n        }\n        "
	added, errMsg := insertNginxProxy(nginxDomainVhost, "app", "kuma", lines)
	if errMsg != "" || !strings.Contains(added, "location ^~ /app/ {") || !strings.Contains(added, "location / {\n        try_files") {
		t.Fatalf("bad subdirectory insert (%s):\n%s", errMsg, added)
	}
	got, _ := stripServiceProxies(added, "kuma")
	if strings.Contains(got, "/app/") || strings.Contains(got, "proxy_") || strings.Count(got, "{") != strings.Count(got, "}") {
		t.Fatalf("bad strip:\n%s", got)
	}
}
