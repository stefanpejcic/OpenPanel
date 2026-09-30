package mysql

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const mysqlSocketMount = "./sockets/mysqld:/var/run/mysqld"

// userHomeRoot is where per-user compose files and php.ini copies live, swapped in tests
var userHomeRoot = "/home"

var (
	mysqliSocketRE = regexp.MustCompile(`(?m)^\s*mysqli\.default_socket\s*=\s*"?/var/run/mysqld/mysqld\.sock"?\s*$`)
	pdoSocketRE    = regexp.MustCompile(`(?m)^\s*pdo_mysql\.default_socket\s*=\s*"?/var/run/mysqld/mysqld\.sock"?\s*$`)
)

// AppDBHost is the DB host to write into app configs: localhost (unix socket) when every PHP container of the user can reach the socket, otherwise the mysql/mariadb service name
func AppDBHost(userContext, mysqlVersion string) string {
	if mysqlSocketReady(userContext) {
		return "localhost"
	}
	return mysqlVersion
}

// mysqlSocketReady checks all php-fpm/openlitespeed services mount the socket dir and all php.ini copies point PHP at it, users created before this was added have neither
func mysqlSocketReady(userContext string) bool {
	if userContext == "" || strings.ContainsAny(userContext, "/.") {
		return false
	}
	home := filepath.Join(userHomeRoot, userContext)

	raw, err := os.ReadFile(filepath.Join(home, "docker-compose.yml"))
	if err != nil {
		return false
	}
	var compose struct {
		Services map[string]struct {
			Volumes []any `yaml:"volumes"`
		} `yaml:"services"`
	}
	if yaml.Unmarshal(raw, &compose) != nil {
		return false
	}
	phpServices := 0
	for name, svc := range compose.Services {
		if !strings.HasPrefix(name, "php-fpm-") && name != "openlitespeed" {
			continue
		}
		phpServices++
		mounted := false
		for _, v := range svc.Volumes {
			if s, ok := v.(string); ok && strings.TrimSpace(s) == mysqlSocketMount {
				mounted = true
				break
			}
		}
		if !mounted {
			return false
		}
	}
	if phpServices == 0 {
		return false
	}

	inis, _ := filepath.Glob(filepath.Join(home, "php.ini", "*.ini"))
	if len(inis) == 0 {
		return false
	}
	for _, p := range inis {
		content, err := os.ReadFile(p)
		if err != nil || !mysqliSocketRE.Match(content) || !pdoSocketRE.Match(content) {
			return false
		}
	}
	return true
}
