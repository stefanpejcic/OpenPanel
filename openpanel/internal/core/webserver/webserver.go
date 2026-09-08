// Package webserver reads a per-user .env value and runs each web server's own config syntax test inside its container.
package webserver

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

// GetEnvFileValue reads one KEY=value line from /home/<context>/.env, stripping a wrapping pair of quotes
func GetEnvFileValue(userContext, key string) string {
	content, err := os.ReadFile("/home/" + userContext + "/.env")
	if err != nil {
		return ""
	}
	prefix := key + "="
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, prefix) {
			return strings.Trim(strings.TrimSpace(line[len(prefix):]), `'"`)
		}
	}
	return ""
}

type configTest struct {
	args           []string
	successPattern *regexp.Regexp
}

// webserverConfigTests holds the syntax-check command for each web server, run inside its container. No entry (e.g. LiteSpeed) means "can't validate" - callers should treat that as ok.
var webserverConfigTests = map[string]configTest{
	"nginx":     {[]string{"nginx", "-t"}, regexp.MustCompile(`successful`)},
	"openresty": {[]string{"nginx", "-t"}, regexp.MustCompile(`successful`)},
	"apache":    {[]string{"apachectl", "configtest"}, regexp.MustCompile(`Syntax OK`)},
}

// HasConfigTest reports whether serviceName has a built-in syntax-check command (nginx/openresty/apache) - LiteSpeed and anything else don't
func HasConfigTest(serviceName string) bool {
	_, ok := webserverConfigTests[serviceName]
	return ok
}

// TestWebserverConfig runs the web server's config syntax test inside its container, returning (ok, output). Only meaningful while the container is running.
func TestWebserverConfig(ctx context.Context, userContext, serviceName string) (bool, string) {
	test, ok := webserverConfigTests[serviceName]
	if !ok {
		return true, ""
	}

	argv := podmanmanager.PodmanArgv(userContext, append([]string{"exec", serviceName}, test.args...)...)
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = podmanmanager.PodmanEnv(userContext)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil && output == "" {
		return false, err.Error()
	}
	return test.successPattern.MatchString(output), output
}
