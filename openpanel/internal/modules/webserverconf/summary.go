package webserverconf

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/dbtuning"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// ServerSummary feeds the tiles above the editor
type ServerSummary struct {
	Version     string // like "Apache 2.4.69", "" when the container didn't say
	Image       string // like "httpd:alpine"
	State       string // podman state, not_found when there's no container
	Uptime      string
	CPU         string
	RAM         string
	PIDs        string
	TestCommand string // "" when the server has no syntax check
}

// Running reports whether the web server container is up
func (s ServerSummary) Running() bool { return s.State == "running" }

// versionCommands print the server's version inside its container
var versionCommands = map[string][]string{
	"apache":        {"httpd", "-v"},
	"nginx":         {"nginx", "-v"},
	"openresty":     {"openresty", "-v"},
	"openlitespeed": {"/usr/local/lsws/bin/lshttpd", "-v"},
	"litespeed":     {"/usr/local/lsws/bin/lshttpd", "-v"},
}

var versionRE = regexp.MustCompile(`(?i)\b(apache|nginx|openresty|litespeed)/(\d+(?:\.\d+)*)`)

// parseServerVersion turns "Server version: Apache/2.4.69 (Unix)" into "Apache 2.4.69"
func parseServerVersion(out string) string {
	m := versionRE.FindStringSubmatch(out)
	if m == nil {
		return ""
	}
	name := apiWebserverConfLabels[strings.ToLower(m[1])]
	if name == "" {
		name = m[1]
	}
	return name + " " + m[2]
}

// serverVersion asks the running container for its version, cached until it restarts
func serverVersion(ctx context.Context, a *appctx.App, userContext, service string, started time.Time) string {
	args, ok := versionCommands[service]
	if !ok {
		return ""
	}
	key := "webserver_version:" + userContext + ":" + service + ":" + strconv.FormatInt(started.Unix(), 10)
	v, _ := cache.Memoize(ctx, a.Cache, key, 24*time.Hour, func() (string, error) {
		execCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		out, _ := podmanmanager.Command(execCtx, userContext, podmanmanager.PodmanArgv(userContext, append([]string{"exec", service}, args...)...)).CombinedOutput()
		if v := parseServerVersion(string(out)); v != "" {
			return v, nil
		}
		return "", errors.New("no version in output")
	})
	return v
}

var imageVarRE = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::?-([^}]*))?\}`)

// resolveImage fills ${VAR:-default} in a compose image from the .env values
func resolveImage(image string, env map[string]string) string {
	return imageVarRE.ReplaceAllStringFunc(image, func(m string) string {
		sub := imageVarRE.FindStringSubmatch(m)
		if v, ok := env[sub[1]]; ok && v != "" {
			return v
		}
		return sub[2]
	})
}

// composeImage is the image a compose service runs, with its tag resolved
func composeImage(userContext, service string, env map[string]string) string {
	data, err := docker.LoadCompose(userContext)
	if err != nil {
		return ""
	}
	services, _ := data["services"].(map[string]any)
	svc, _ := services[service].(map[string]any)
	image, _ := svc["image"].(string)
	return strings.TrimPrefix(resolveImage(strings.TrimSpace(image), env), "docker.io/library/")
}

// buildServerSummary gathers version, container state and limits, each piece optional
func buildServerSummary(ctx context.Context, a *appctx.App, userContext, service string) ServerSummary {
	if service == "" {
		return ServerSummary{}
	}
	status := docker.GetContainerStatus(ctx, userContext, service)
	env := docker.LoadEnvFile(userContext)
	prefix := docker.ServiceKeyPrefix(service)
	s := ServerSummary{
		State: status.State, Image: composeImage(userContext, service, env),
		CPU: env[prefix+"_CPU"], RAM: env[prefix+"_RAM"], PIDs: env[prefix+"_PIDS"],
		TestCommand: webserver.ConfigTestCommand(service),
	}
	if s.Running() {
		if !status.Started.IsZero() {
			s.Uptime = dbtuning.FormatUptime(int64(time.Since(status.Started).Seconds()))
		}
		s.Version = serverVersion(ctx, a, userContext, service, status.Started)
	}
	return s
}

// lastLines keeps the tail of a long command output
func lastLines(out string, n int) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// handleConfigTest runs the web server's syntax check, it changes nothing
func handleConfigTest(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	entry := lookupWebserverConf(webserver.GetEnvFileValue(userContext, "WEB_SERVER"))
	command := webserver.ConfigTestCommand(entry.ServiceName)
	switch {
	case command == "":
		web.WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "supported": false, "output": "This web server has no config test."})
		return
	case !docker.IsServiceRunning(r.Context(), userContext, entry.ServiceName):
		web.WriteJSON(w, http.StatusOK, map[string]any{"ok": false, "supported": true, "output": "The " + entry.ServiceName + " container is not running."})
		return
	}
	ok, output := webserver.TestWebserverConfig(r.Context(), userContext, entry.ServiceName)
	web.WriteJSON(w, http.StatusOK, map[string]any{"ok": ok, "supported": true, "command": command, "output": lastLines(output, 20)})
}
