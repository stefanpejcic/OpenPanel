package mysql

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSocketFixture(t *testing.T, compose string, inis map[string]string) {
	t.Helper()
	root := t.TempDir()
	orig := userHomeRoot
	userHomeRoot = root
	t.Cleanup(func() { userHomeRoot = orig })

	home := filepath.Join(root, "alice")
	if err := os.MkdirAll(filepath.Join(home, "php.ini"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "docker-compose.yml"), []byte(compose), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, body := range inis {
		if err := os.WriteFile(filepath.Join(home, "php.ini", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

const socketIni = "[PHP]\npdo_mysql.default_socket = /var/run/mysqld/mysqld.sock\nmysqli.default_socket = /var/run/mysqld/mysqld.sock\n"
const emptySocketIni = "[PHP]\npdo_mysql.default_socket=\nmysqli.default_socket =\n"

const composeWithMounts = `services:
  openlitespeed:
    volumes:
      - ./sockets/mysqld:/var/run/mysqld
  php-fpm-8.3:
    volumes:
      - html_data:/var/www/html/
      - ./sockets/mysqld:/var/run/mysqld
  php-fpm-8.5:
    volumes:
      - html_data:/var/www/html/
      - ./sockets/mysqld:/var/run/mysqld
  mariadb:
    volumes:
      - ./sockets/mysqld:/var/run/mysqld
`

func TestAppDBHostLocalhostWhenEverythingMounted(t *testing.T) {
	writeSocketFixture(t, composeWithMounts, map[string]string{"8.3.ini": socketIni, "8.5.ini": socketIni})
	if got := AppDBHost("alice", "mariadb"); got != "localhost" {
		t.Fatalf("expected localhost, got %q", got)
	}
}

func TestAppDBHostFallsBackWhenOnePHPContainerLacksMount(t *testing.T) {
	compose := composeWithMounts + `  php-fpm-7.4:
    volumes:
      - html_data:/var/www/html/
`
	writeSocketFixture(t, compose, map[string]string{"8.3.ini": socketIni})
	if got := AppDBHost("alice", "mysql"); got != "mysql" {
		t.Fatalf("expected service name fallback, got %q", got)
	}
}

func TestAppDBHostFallsBackWhenIniNotSet(t *testing.T) {
	writeSocketFixture(t, composeWithMounts, map[string]string{"8.3.ini": socketIni, "7.4.ini": emptySocketIni})
	if got := AppDBHost("alice", "mariadb"); got != "mariadb" {
		t.Fatalf("expected service name fallback, got %q", got)
	}
}

func TestAppDBHostFallsBackWithoutFiles(t *testing.T) {
	orig := userHomeRoot
	userHomeRoot = t.TempDir()
	t.Cleanup(func() { userHomeRoot = orig })
	if got := AppDBHost("alice", "mariadb"); got != "mariadb" {
		t.Fatalf("expected service name fallback, got %q", got)
	}
	if got := AppDBHost("../etc", "mariadb"); got != "mariadb" {
		t.Fatalf("expected fallback for odd context, got %q", got)
	}
}
