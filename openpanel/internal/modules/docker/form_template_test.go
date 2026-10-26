package docker

import (
	"net/url"
	"strings"
	"testing"
)

func TestTemplateFormValues(t *testing.T) {
	if templateFormValues(url.Values{"foo": {"1"}}, []string{"www"}) != nil {
		t.Fatal("expected nil without known keys")
	}
	q, _ := url.ParseQuery("image=louislam/uptime-kuma:1&name=Kuma&ram=2&network=www,bogus&env=TZ=UTC&env=A=b&volume=kuma_data:/app/data&volume=bad&socket=1&domain=Status.Example.com&port=3001")
	v := templateFormValues(q, []string{"default", "www"})
	checks := map[string]string{
		"image": "louislam/uptime-kuma:1", "service_name": "kuma", "ram": "2G", "cpu": "0.5", "pids": "500",
		"environment": "TZ=UTC\nA=b", "add_socket": "on", "domain": "status.example.com", "port": "3001",
		"volume_name": "kuma_data", "volume_mount": "/app/data", "volume_readonly": "off",
	}
	for k, want := range checks {
		if got := v.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if len(v["network"]) != 1 || v["network"][0] != "www" {
		t.Errorf("network = %v", v["network"])
	}
	if len(v["volume_name"]) != 1 {
		t.Errorf("volume_name = %v", v["volume_name"])
	}
}

func TestCheckNewVolumes(t *testing.T) {
	available := []string{"html_data"}
	if err := checkNewVolumes(available, []string{"html_data", "kuma_data", "/run/user/${USER_ID}/podman/podman.sock", ""}); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"Kuma", "ab", "1data", "a b"} {
		if checkNewVolumes(available, []string{bad}) == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestDeclareNewVolumes(t *testing.T) {
	compose := map[string]any{"volumes": map[string]any{"html_data": map[string]any{"driver": "local"}}}
	declareNewVolumes(compose, []string{"html_data", "kuma_data", "/etc"}, "kuma")
	vols := compose["volumes"].(map[string]any)
	if len(vols) != 2 || vols["kuma_data"] == nil {
		t.Errorf("got %v", vols)
	}
}

func TestTemplateFormValuesN8N(t *testing.T) {
	q, _ := url.ParseQuery("image=n8nio%2Fn8n%3Alatest&name=n8n&cpu=1&ram=1&pids=500&network=www%2Cdb&volume=n8n_data%3A%2Fhome%2Fnode%2F.n8n&env=N8N_HOST%3D%7Bdomain%7D&env=N8N_PORT%3D5678&env=N8N_PROTOCOL%3Dhttps&env=N8N_EDITOR_BASE_URL%3Dhttps%3A%2F%2F%7Bdomain%7D%2F&env=WEBHOOK_URL%3Dhttps%3A%2F%2F%7Bdomain%7D%2F&env=GENERIC_TIMEZONE%3DUTC&env=N8N_RUNNERS_ENABLED%3Dtrue&port=5678")
	v := templateFormValues(q, []string{"www", "db", "default"})
	if v.Get("image") != "n8nio/n8n:latest" || v.Get("port") != "5678" || v.Get("volume_mount") != "/home/node/.n8n" {
		t.Fatalf("got %v", v)
	}
	if len(v["network"]) != 2 || !strings.Contains(v.Get("environment"), "WEBHOOK_URL=https://{domain}/") {
		t.Fatalf("got %v", v)
	}
}

func TestCheckHostPaths(t *testing.T) {
	compose := map[string]any{"services": map[string]any{
		"php": map[string]any{"volumes": []any{"/etc/openpanel/wordpress/wp-cli.phar:/usr/local/bin/wp:ro"}},
	}}
	ok := []string{"html_data", "", "./data", "./sockets/mysqld", "/etc/openpanel/wordpress/wp-cli.phar", "/run/user/${USER_ID}/podman/podman.sock"}
	if err := checkHostPaths(compose, ok, "alice"); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"/", "/etc", "/home/ctx", "./", ".", "./..", "./../x", "./a/../../x", "./features.txt", "./docker-compose.yml", "./.env", "./.config/containers", "${HOME}/x", "./a:/b", "~/x"} {
		if checkHostPaths(compose, []string{bad}, "alice") == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}
