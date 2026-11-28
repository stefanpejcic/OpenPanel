package docker

import (
	"fmt"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

// openpanel.config key from OpenAdmin, off unless "yes" since the socket is the account's whole podman API
const socketSetting = "docker_socket"

// socketProxyService is the template service that legitimately mounts the podman socket dir
const socketProxyService = "docker-proxy"

// socketProxyNetwork reaches docker-proxy; only the template's cron and backup services may join it
const socketProxyNetwork = "socket-proxy"

var socketProxyMembers = map[string]bool{"cron": true, "backup": true}

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
func checkSocketAccess(allowed bool, names []string, addSocket bool, service string) error {
	if allowed || service == socketProxyService {
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

// every compose network except socket-proxy, which only cron and backup may join since it reaches docker-proxy
func userNetworks(composeData map[string]any, service string) []string {
	all := GetAvailableNetworks(composeData)
	if socketProxyMembers[service] {
		return all
	}
	out := make([]string, 0, len(all))
	for _, n := range all {
		if n != socketProxyNetwork {
			out = append(out, n)
		}
	}
	return out
}
