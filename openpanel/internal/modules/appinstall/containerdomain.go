package appinstall

import (
	"context"
	"errors"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// custom containers from /containers/new reuse the same reverse proxy the app installers write
func init() {
	docker.AttachDomainProxy = attachContainerDomain
	docker.RemoveServiceProxies = RemoveServiceProxies
}

// attachContainerDomain proxies the domain root to service:port, websockets on since apps like Uptime Kuma need them
func attachContainerDomain(ctx context.Context, userContext, domain, service string, port int) error {
	webServerType, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
	switch {
	case webServerType == "apache", webServerType == "litespeed":
		editApacheConfig(userContext, domain, "", service, port, true)
	case webServerType == "nginx" || webServerType == "openresty":
		if msg := editNginxConfig(ctx, userContext, domain, "", service, webServerType, port, true); msg != "" {
			return errors.New(strings.TrimPrefix(msg, "Error: "))
		}
	case strings.Contains(strings.ToLower(webServerType), "litespeed"):
		if msg := editLswsConfig(userContext, domain, "", service, port); msg != "" {
			return errors.New(strings.TrimPrefix(msg, "Error: "))
		}
	default:
		return errors.New("unknown webserver, add the proxy manually")
	}
	reloadWebserver(ctx, userContext, webServerType)
	return nil
}

// refreshProxyUpstreams reloads the webserver after an app container is (re)created, nginx, openresty and litespeed resolved its old IP and would keep proxying there, apache looks it up per request
func refreshProxyUpstreams(ctx context.Context, userContext string) {
	switch ws, _ := docker.GetEnvValue(userContext, "WEB_SERVER"); ws {
	case "nginx", "openresty", "litespeed", "openlitespeed":
		reloadWebserver(ctx, userContext, ws)
	}
}

func reloadWebserver(ctx context.Context, userContext, webServerType string) {
	if webServerType == "litespeed" || webServerType == "openlitespeed" {
		// litespeed keeps resolved names in /dev/shm, which even a container restart keeps, so the cache goes before it starts again
		cmd := "/usr/local/lsws/bin/lswsctrl stop >/dev/null 2>&1; rm -f /dev/shm/lsws/adns_cache.*; /usr/local/lsws/bin/lswsctrl start >/dev/null 2>&1"
		_ = podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "exec", webServerType, "sh", "-c", cmd)).Run()
		return
	}
	_ = webserver.ReloadWebserver(ctx, userContext, webServerType)
}
