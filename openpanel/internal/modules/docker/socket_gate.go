package docker

import (
	"fmt"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

// openpanel.config key from OpenAdmin, off unless "yes" since the socket is the account's whole podman API
const socketSetting = "docker_socket"

// only the template's cron and backup join this network, anything on it can call docker-proxy
const socketProxyNetwork = "socket-proxy"

// built-ins that hold the socket or sit on socket-proxy, their Edit is hidden and a direct request is refused too
var socketServices = map[string]bool{"docker-proxy": true, "cron": true, "backup": true}

// panelConfigPath is a var so tests can point it at a temp file
var panelConfigPath = appctx.ConfigPath

// isPodmanRuntimePath matches bind sources that reach the user's rootless podman socket
func isPodmanRuntimePath(n string) bool {
	n = strings.TrimSpace(n)
	return strings.HasPrefix(n, "/run/user/") || strings.HasPrefix(n, "/run/podman") ||
		strings.HasPrefix(n, "/var/run/user/") || strings.Contains(n, "podman.sock") || strings.Contains(n, "docker.sock")
}

// read from disk each time so turning it off in OpenAdmin works without a panel restart
func socketAllowed() bool {
	cfg, err := config.Load(panelConfigPath)
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Get(socketSetting, "no")), "yes")
}

// checkSocketAccess blocks the podman socket (checkbox or a raw /run/user/... volume) unless the server allows it
func checkSocketAccess(allowed bool, names []string, addSocket bool) error {
	if allowed {
		return nil
	}
	msg := fmt.Errorf("Giving a container the Docker socket is turned off on this server. Turn off the socket option or ask your hosting provider.")
	if addSocket {
		return msg
	}
	for _, n := range names {
		if isPodmanRuntimePath(n) {
			return msg
		}
	}
	return nil
}

// every compose network except socket-proxy
func userNetworks(composeData map[string]any) []string {
	all := GetAvailableNetworks(composeData)
	out := make([]string, 0, len(all))
	for _, n := range all {
		if n != socketProxyNetwork {
			out = append(out, n)
		}
	}
	return out
}
