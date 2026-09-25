package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImageVarRE(t *testing.T) {
	cases := map[string][3]string{
		"redis:${REDIS_VERSION:-8.6.2-alpine}":             {"redis", "REDIS_VERSION", "8.6.2-alpine"},
		"openresty/openresty:${OPENRESTY_VERSION:-alpine}": {"openresty/openresty", "OPENRESTY_VERSION", "alpine"},
		"mariadb:${MYSQL_VERSION:-latest}":                 {"mariadb", "MYSQL_VERSION", "latest"},
		"mysql:${MYSQL_VERSION}":                           {"mysql", "MYSQL_VERSION", ""},
		"ghcr.io/org/app:${APP_VERSION:-1}":                {"ghcr.io/org/app", "APP_VERSION", "1"},
		"registry.local:5000/team/app:${APP_VERSION-2}":    {"registry.local:5000/team/app", "APP_VERSION", "2"},
	}
	for in, want := range cases {
		m := imageVarRE.FindStringSubmatch(in)
		if m == nil || m[1] != want[0] || m[2] != want[1] || m[3] != want[2] {
			t.Errorf("%s: got %v", in, m)
		}
	}
	for _, in := range []string{"shinsenter/php:8.4-fpm", "mcuadros/ofelia:latest", "tecnativa/docker-socket-proxy"} {
		if imageVarRE.MatchString(in) {
			t.Errorf("%s has no tag variable", in)
		}
	}
}

func TestDockerHubRepo(t *testing.T) {
	for in, want := range map[string]string{"redis": "library/redis", "openresty/openresty": "openresty/openresty", "docker.io/library/nginx": "library/nginx", "docker.io/valkey/valkey": "valkey/valkey"} {
		if got, ok := dockerHubRepo(in); !ok || got != want {
			t.Errorf("%s: got %q %v", in, got, ok)
		}
	}
	for _, in := range []string{"ghcr.io/org/app", "quay.io/x/y", "localhost/app", "registry.local:5000/a/b"} {
		if _, ok := dockerHubRepo(in); ok {
			t.Errorf("%s isn't on Docker Hub", in)
		}
	}
}

func TestTagChecks(t *testing.T) {
	for _, tag := range []string{"8.6.2-alpine", "latest", "1.6.41-alpine", "v2", "20_04"} {
		if !imageTagRE.MatchString(tag) || !keepTag(tag) {
			t.Errorf("%q should be accepted", tag)
		}
	}
	for _, tag := range []string{"", "-x", "a b", "8\"\nEVIL=1", "sha256-abc.sig", "x.att"} {
		if imageTagRE.MatchString(tag) && keepTag(tag) {
			t.Errorf("%q should be rejected", tag)
		}
	}
}

func TestImageTagVar(t *testing.T) {
	dir := t.TempDir()
	old := homeDirOverride
	homeDirOverride = dir
	defer func() { homeDirOverride = old }()
	compose := "services:\n  mariadb:\n    image: mariadb:${MYSQL_VERSION:-latest}\n  redis:\n    image: redis:${REDIS_VERSION:-8.6.2-alpine}\n  backup:\n    image: offen/docker-volume-backup:v2\n  openresty:\n    image: openresty/openresty:${OPENRESTY_VERSION:-alpine}\n  app:\n    image: ghcr.io/org/app:${APP_VERSION:-1}\n"
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(compose), 0o644); err != nil {
		t.Fatal(err)
	}
	for service, want := range map[string][2]string{
		"mariadb":    {"MYSQL_VERSION", "latest"},
		"redis":      {"REDIS_VERSION", "8.6.2-alpine"},
		"backup":     {"BACKUP_VERSION", ""},
		"phpmyadmin": {"PMA_VERSION", ""},
		"mysql":      {"MYSQL_VERSION", ""},
		"u":          {"OS", ""},
	} {
		v, d := imageTagVar("u", service)
		if v != want[0] || d != want[1] {
			t.Errorf("%s: got %s %q", service, v, d)
		}
	}

	for service, want := range map[string]string{
		"redis":     "https://hub.docker.com/_/redis/tags",
		"openresty": "https://hub.docker.com/r/openresty/openresty/tags",
		"app":       "",
		"backup":    "",
	} {
		if got := dockerHubPage("u", service); got != want {
			t.Errorf("hub page for %s = %q, want %q", service, got, want)
		}
	}
}

func TestImageChangeable(t *testing.T) {
	for _, s := range []string{"redis", "mariadb", "phpmyadmin", "nginx"} {
		if !imageChangeable(s) {
			t.Errorf("%s should be changeable", s)
		}
	}
	for _, s := range []string{"", "cron", "backup", "docker-proxy", "php-fpm-8.4"} {
		if imageChangeable(s) {
			t.Errorf("%s should not be changeable", s)
		}
	}
}
