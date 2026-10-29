// Package webserver reads a per-user .env value, runs each web server's own config syntax test, and reloads it.
package webserver

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
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
	cmd            string
	successPattern *regexp.Regexp
	// used instead of successPattern by a test that has no success line, it passes unless this matches
	failPattern *regexp.Regexp
}

func (t configTest) passed(output string) bool {
	if t.failPattern != nil {
		return !t.failPattern.MatchString(output)
	}
	return t.successPattern.MatchString(output)
}

// webserverConfigTests holds the syntax-check shell command for each web server. No entry (e.g. LiteSpeed) means "can't validate" - callers should treat that as ok.
var webserverConfigTests = map[string]configTest{
	"nginx":     {cmd: "nginx -t", successPattern: regexp.MustCompile(`successful`)},
	"openresty": {cmd: "nginx -t", successPattern: regexp.MustCompile(`successful`)},
	"apache":    {cmd: "apachectl configtest", successPattern: regexp.MustCompile(`Syntax OK`)},
	// -t exits 1 on plain warnings, writes nothing without /tmp/lshttpd (only a running server makes it) and leaves its log behind for the next run when a config won't load at all
	"openlitespeed": {cmd: "mkdir -p /tmp/lshttpd && rm -f /tmp/lshttpd/testconf && exec /usr/local/lsws/bin/openlitespeed -t", failPattern: regexp.MustCompile(`\[ERROR\]`)},
}

// HasConfigTest reports whether serviceName has a built-in syntax-check command, LiteSpeed Enterprise and anything else don't
func HasConfigTest(serviceName string) bool {
	_, ok := webserverConfigTests[serviceName]
	return ok
}

// lookupFailed is nginx -t failing on a container name that aardvark-dns can't resolve yet
var lookupFailed = regexp.MustCompile(`host not found`)

// oneOffTestTimeout caps a test in a throwaway container, podman-compose pulls the image first if it's missing
const oneOffTestTimeout = 2 * time.Minute

// TestWebserverConfig runs the web server's config syntax test in its container, or in a throwaway one from the account's compose file while it's stopped, returning (ok, output)
func TestWebserverConfig(ctx context.Context, userContext, serviceName string) (bool, string) {
	test, ok := webserverConfigTests[serviceName]
	if !ok {
		return true, ""
	}
	if !isRunning(ctx, userContext, serviceName) {
		return testInThrowaway(ctx, userContext, serviceName, test)
	}
	argv := podmanmanager.PodmanArgv(userContext, "exec", serviceName, "sh", "-c", test.cmd)
	// retried since aardvark-dns briefly fails lookups right after a container is created or removed
	ok, output := runConfigTest(ctx, userContext, argv, test)
	for attempt := 1; !ok && attempt < 5 && lookupFailed.MatchString(output); attempt++ {
		select {
		case <-ctx.Done():
			return false, output
		case <-time.After(time.Second):
		}
		ok, output = runConfigTest(ctx, userContext, argv, test)
	}
	return ok, output
}

// mapComposeHosts points the account's container names at the throwaway container, nginx -t has to resolve upstreams whose containers may be stopped too
const mapComposeHosts = `for h in "$@"; do echo "127.0.0.1 $h" >> /etc/hosts; done; `

func testInThrowaway(ctx context.Context, userContext, serviceName string, test configTest) (bool, string) {
	ctx, cancel := context.WithTimeout(ctx, oneOffTestTimeout)
	defer cancel()
	// podman-compose run leaves out the ports, so this can't clash with whatever holds them now
	args := []string{"-f", "/home/" + userContext + "/docker-compose.yml",
		"run", "--rm", "--no-deps", "-T", "--entrypoint", "sh", serviceName, "-c", mapComposeHosts + test.cmd, "sh"}
	var hosts []string
	if compose, err := podmanmanager.LoadComposeConfig(ctx, userContext); err == nil {
		hosts = composeHostnames(compose)
	}
	ok, output := runConfigTest(ctx, userContext, podmanmanager.PodmanComposeArgv(append(args, hosts...)...), test)
	if !ok && ctx.Err() == context.DeadlineExceeded {
		return false, TestTimedOut
	}
	return ok, output
}

// composeHostnames is every service and container name in a compose config
func composeHostnames(compose map[string]any) []string {
	services, _ := compose["services"].(map[string]any)
	names := make([]string, 0, len(services))
	for name, svc := range services {
		names = append(names, name)
		if s, ok := svc.(map[string]any); ok {
			if cn, _ := s["container_name"].(string); cn != "" && cn != name {
				names = append(names, cn)
			}
		}
	}
	sort.Strings(names)
	return names
}

func runConfigTest(ctx context.Context, userContext string, argv []string, test configTest) (bool, string) {
	out, err := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil && output == "" {
		return false, TestNoOutput + ": " + err.Error()
	}
	return test.passed(output), output
}

func isRunning(ctx context.Context, userContext, container string) bool {
	out, err := podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "inspect", "-f", "{{.State.Running}}", container)).Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
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

	if !isRunning(ctx, userContext, ws) {
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
	if ok, testOut := TestWebserverConfig(ctx, userContext, ws); !ok {
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
