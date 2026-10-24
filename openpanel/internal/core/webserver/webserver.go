// Package webserver reads a per-user .env value, runs each web server's own config syntax test inside its container, and reloads it.
package webserver

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// GetEnvFileValue reads one KEY=value line from /home/<context>/.env, stripping a wrapping pair of quotes
func GetEnvFileValue(userContext, key string) string {
	content, err := tenantfs.ReadFile("/home/" + userContext + "/.env")
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

// reloadSignals is the graceful reload signal each image's pid 1 (the server's master process) takes
var reloadSignals = map[string]string{"nginx": "HUP", "openresty": "HUP", "apache": "USR1"}

// failLog always reaches the container log, the default logger is silenced outside dev_mode and a failed reload leaves a site on its old config
var failLog = log.New(os.Stderr, "webserver: ", log.LstdFlags)

// ReloadWebserver applies config/vhost changes without dropping connections, a stopped webserver gets started and a failed signal falls back to a full restart
func ReloadWebserver(ctx context.Context, userContext, ws string) error {
	err := reloadWebserver(ctx, userContext, ws)
	if err != nil {
		failLog.Printf("reloading %s for %s: %v", ws, userContext, err)
	}
	return err
}

func reloadWebserver(ctx context.Context, userContext, ws string) error {
	run := func(args ...string) ([]byte, error) {
		return podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, args...)).CombinedOutput()
	}
	restart := func() error {
		if out, err := run("restart", ws); err != nil {
			return fmt.Errorf("restarting %s: %s", ws, strings.TrimSpace(string(out)))
		}
		return nil
	}

	if out, err := run("inspect", "-f", "{{.State.Running}}", ws); err != nil || strings.TrimSpace(string(out)) != "true" {
		return restart()
	}
	// pid 1 in the litespeed container is our start.sh, lswsctrl restart is litespeed's own graceful restart
	if strings.Contains(ws, "litespeed") {
		if _, err := run("exec", ws, "/usr/local/lsws/bin/lswsctrl", "restart"); err != nil {
			return restart()
		}
		return nil
	}
	signal, ok := reloadSignals[ws]
	if !ok {
		return restart()
	}
	// a reload with a broken config would be ignored silently, a restart would take every site down
	// retried since aardvark-dns briefly fails lookups right after a container is created or removed
	ok, testOut := TestWebserverConfig(ctx, userContext, ws)
	for attempt := 1; !ok && attempt < 5; attempt++ {
		time.Sleep(time.Second)
		ok, testOut = TestWebserverConfig(ctx, userContext, ws)
	}
	if !ok {
		return fmt.Errorf("%s config test failed, still running the previous config: %s", ws, testOut)
	}
	if _, err := run("kill", "--signal", signal, ws); err != nil {
		return restart()
	}
	return nil
}

// WaitListening waits up to a minute for container to have a listening tcp socket, a php-fpm container runs a few seconds before fpm accepts connections
func WaitListening(ctx context.Context, userContext, container string) bool {
	argv := podmanmanager.PodmanArgv(userContext, "exec", container, "sh", "-c", `awk '$4 == "0A"' /proc/net/tcp /proc/net/tcp6 2>/dev/null | grep -q .`)
	deadline := time.Now().Add(time.Minute)
	for time.Now().Before(deadline) {
		if podmanmanager.Command(ctx, userContext, argv).Run() == nil {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(2 * time.Second):
		}
	}
	return false
}
