package appinstall

import (
	"context"
	"errors"
	"strings"

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
	case webServerType == "apache":
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

// refreshProxyUpstreams reloads nginx/openresty after an app container is (re)created, they resolved its old IP at load time and would keep proxying there, apache looks it up per request
func refreshProxyUpstreams(ctx context.Context, userContext string) {
	if ws, _ := docker.GetEnvValue(userContext, "WEB_SERVER"); ws == "nginx" || ws == "openresty" {
		reloadWebserver(ctx, userContext, ws)
	}
}

func reloadWebserver(ctx context.Context, userContext, webServerType string) {
	_ = webserver.ReloadWebserver(ctx, userContext, webServerType)
}
