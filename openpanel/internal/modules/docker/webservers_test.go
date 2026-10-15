package docker

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const webserversTestCompose = `services:
  caddy:
    image: caddy
    ports: ["${HTTP_PORT}"]
    volumes:
      - type: volume
        source: webserver_data
        target: /etc/caddy
  nginx:
    image: nginx
    ports: ["${HTTP_PORT}", "${HTTPS_PORT}"]
    volumes: ["webserver_data:/etc/nginx/sites-available/"]
  apache:
    image: httpd
    ports: ["${PROXY_HTTP_PORT}"]
    volumes: ["webserver_data:/usr/local/apache2/conf/sites-available/"]
  backup:
    image: backup
    volumes: ["webserver_data:/backup/webserver"]
  varnish:
    image: varnish
    ports: ["${HTTP_PORT}"]
`

func TestAvailableWebservers(t *testing.T) {
	dir := t.TempDir()
	old := homeDirOverride
	homeDirOverride = dir
	defer func() { homeDirOverride = old }()

	if got := AvailableWebservers("u"); !reflect.DeepEqual(got, fallbackWebservers) {
		t.Errorf("no compose file should fall back, got %v", got)
	}

	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(webserversTestCompose), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, want := AvailableWebservers("u"), []string{"apache", "nginx", "caddy"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	choices := webserverChoicesFor("u", "openlitespeed")
	var values []string
	for _, c := range choices {
		values = append(values, c.Value)
	}
	if want := []string{"openlitespeed", "apache", "nginx", "caddy"}; !reflect.DeepEqual(values, want) {
		t.Errorf("choices = %v, want %v", values, want)
	}
	if !choices[0].Current || choices[3].Label != "caddy" || choices[3].Icon != "" {
		t.Errorf("unexpected choices %+v", choices)
	}
}
