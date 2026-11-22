package docker

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// socketFeature is the plan feature that lets a user give a container their podman socket
const socketFeature = "docker_socket"

// socketProxyService is the template service that legitimately mounts the podman socket dir
const socketProxyService = "docker-proxy"

// isPodmanRuntimePath matches bind sources that reach the user's rootless podman socket
func isPodmanRuntimePath(n string) bool {
	n = strings.TrimSpace(n)
	return strings.HasPrefix(n, "/run/user/") || strings.HasPrefix(n, "/run/podman") ||
		strings.HasPrefix(n, "/var/run/user/") || strings.Contains(n, "podman.sock") || strings.Contains(n, "docker.sock")
}

// socketAllowed reports whether the user's plan grants the docker_socket feature
func socketAllowed(a *appctx.App, r *http.Request) bool {
	_, username, userContext, err := auth.Injected(a, r)
	if err != nil || username == "" {
		return false
	}
	features, err := a.LoadUserFeatures(r.Context(), username, userContext)
	if err != nil {
		return false
	}
	return slices.Contains(features, socketFeature)
}

// checkSocketAccess blocks the podman socket (checkbox or a raw /run/user/... volume) unless the plan allows it
func checkSocketAccess(allowed bool, names []string, addSocket bool, service string) error {
	if allowed || service == socketProxyService {
		return nil
	}
	msg := fmt.Errorf("Giving a container the Docker socket isn't enabled for your plan. Turn off the socket option or ask your hosting provider.")
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
