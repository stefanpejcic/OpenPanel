package docker

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// compose labels that remember which domain proxies to a custom container, so delete can undo it
const (
	domainLabel = "openpanel.domain"
	portLabel   = "openpanel.port"
)

// set by appinstall, it owns the vhost proxy editing and already imports this package
var (
	AttachDomainProxy    func(ctx context.Context, userContext, domain, service string, port int) error
	RemoveServiceProxies func(ctx context.Context, userContext string, names ...string) []string
)

// proxyNetwork is the network the webservers sit on, a proxied container has to join it
const proxyNetwork = "www"

var newVolumeNameRE = regexp.MustCompile(`^[a-z][a-z0-9_]{2,63}$`)

// serviceDomain reads the domain and port a service is proxied from, if any
func serviceDomain(svc map[string]any) (string, int) {
	labels, _ := svc["labels"].(map[string]any)
	domain := toStr(labels[domainLabel])
	port, _ := strconv.Atoi(toStr(labels[portLabel]))
	return domain, port
}

func setServiceDomain(svc map[string]any, domain string, port int) {
	labels, _ := svc["labels"].(map[string]any)
	if labels == nil {
		labels = map[string]any{}
	}
	labels[domainLabel] = domain
	labels[portLabel] = strconv.Itoa(port)
	svc["labels"] = labels
}

// validateContainerDomain makes sure the domain is the user's and nothing else already answers on its root
func validateContainerDomain(ctx context.Context, a *appctx.App, userID int, composeData map[string]any, domain, portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("Enter the port the app listens on inside the container, between 1 and 65535.")
	}
	if strings.ContainsAny(domain, "/ \t\r\n") || !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		return 0, fmt.Errorf("Choose one of your domains.")
	}
	if AttachDomainProxy == nil {
		return 0, fmt.Errorf("Domain proxying is not available.")
	}
	for name, raw := range servicesRaw(composeData) {
		svc, _ := raw.(map[string]any)
		if d, _ := serviceDomain(svc); d == domain {
			return 0, fmt.Errorf("Domain %s is already used by container %s.", domain, name)
		}
	}
	var appType string
	if a.DB.QueryRowContext(ctx, "SELECT type FROM sites WHERE site_name = ? AND type IN ('NodeJS','Python','Ruby','Java','n8n','Docker')", domain).Scan(&appType) == nil {
		return 0, fmt.Errorf("Domain %s already has a %s application at its root, remove it first or choose another domain.", domain, appType)
	}
	var subdirSites int
	if a.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM sites WHERE site_name LIKE ?", domain+"/%").Scan(&subdirSites) == nil && subdirSites > 0 {
		return 0, fmt.Errorf("Domain %s has %d site(s) in subdirectories that would stop working, choose another domain.", domain, subdirSites)
	}
	return port, nil
}

// checkNewVolumes rejects new volume names compose or podman would choke on, paths are left alone like before
func checkNewVolumes(available, names []string) error {
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n == "" || containsString(available, n) || strings.Contains(n, "/") {
			continue
		}
		if !newVolumeNameRE.MatchString(n) {
			return fmt.Errorf("Volume name %q is not valid. Use 3 to 64 lowercase letters, digits or _, starting with a letter.", n)
		}
	}
	return nil
}

// declareNewVolumes adds the named volumes the compose file doesn't have yet
func declareNewVolumes(composeData map[string]any, names []string, service string) {
	volumes, _ := composeData["volumes"].(map[string]any)
	if volumes == nil {
		volumes = map[string]any{}
	}
	for _, n := range names {
		n = strings.TrimSpace(n)
		if _, ok := volumes[n]; ok || !newVolumeNameRE.MatchString(n) {
			continue
		}
		volumes[n] = map[string]any{"driver": "local", "labels": map[string]any{
			"description": "Created for container " + service + ".", "purpose": "storage",
		}}
	}
	composeData["volumes"] = volumes
}

// containerSiteType is the sites.type for a domain proxied to a custom container, /sites links it to the container edit page
const containerSiteType = "Docker"

// registerContainerSite lists the domain on /sites, best effort since the proxy already works without it
func registerContainerSite(ctx context.Context, a *appctx.App, userID int, domain, service, image string, port int) {
	version := "latest"
	if i := strings.LastIndex(image, ":"); i > strings.LastIndex(image, "/") {
		version = image[i+1:]
	}
	_, err := a.DB.ExecContext(ctx, `INSERT INTO sites (site_name, domain_id, version, type, container, ports)
		SELECT ?, domain_id, ?, ?, ?, ? FROM domains WHERE domain_url = ? AND user_id = ?`,
		domain, version, containerSiteType, service, port, domain, userID)
	if err != nil {
		log.Printf("DOCKER - could not add %s to sites: %v", domain, err)
	}
}

func unregisterContainerSite(ctx context.Context, a *appctx.App, userID int, service string) {
	_, _ = a.DB.ExecContext(ctx, `DELETE FROM sites WHERE type = ? AND container = ?
		AND domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?)`, containerSiteType, service, userID)
}
