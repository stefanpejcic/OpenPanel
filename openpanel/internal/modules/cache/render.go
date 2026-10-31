package cache

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/services"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var cacheServicePage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"cache/service.html",
)

// DomainVarnishRow is one row of the varnish page's per-domain toggle table
type DomainVarnishRow struct {
	DomainURL string
	Status    string // "On" | "Off" | "Unknown"
}

// CachePageData is cache/service.html's template context, shared by all six cache routes. Port == 0 selects the varnish branch (no fixed TCP port, per-domain toggle table instead of a port info card).
type CachePageData struct {
	web.LayoutData

	Service        string
	Description    string
	Port           int
	ContainerState string
	HealthStatus   string
	StatusKey      string
	StatusColor    string
	StatusLabel    string
	StatusMapJSON  template.JS
	IsRunning      bool

	// Varnish-only.
	VarnishDomains []DomainVarnishRow
}

// VarnishDomainNames feeds the domain search's empty state
func (d CachePageData) VarnishDomainNames() []string {
	names := make([]string, 0, len(d.VarnishDomains))
	for _, v := range d.VarnishDomains {
		names = append(names, v.DomainURL)
	}
	return names
}

// StatusTone picks the kit colour class for the status pill
func (d CachePageData) StatusTone() string {
	switch d.StatusKey {
	case "running", "healthy":
		return "is-success"
	case "unhealthy", "exited", "dead":
		return "is-danger"
	case "not_found":
		return "is-muted"
	}
	return "is-warning"
}

// ConnectAddress is how apps in the account reach the service
func (d CachePageData) ConnectAddress() string {
	host := fmt.Sprintf("%s:%d", d.Service, d.Port)
	switch d.Service {
	case "redis", "valkey":
		return "tcp://" + host
	case "elasticsearch", "opensearch":
		return "http://" + host
	}
	return host
}

// PingCommand is a quick connection check to run from the account's terminal
func (d CachePageData) PingCommand() string {
	switch d.Service {
	case "redis", "valkey":
		return fmt.Sprintf("%s-cli -h %s -p %d ping", d.Service, d.Service, d.Port)
	case "memcached":
		return fmt.Sprintf("echo stats | nc %s %d", d.Service, d.Port)
	case "elasticsearch", "opensearch":
		return fmt.Sprintf("curl -s http://%s:%d", d.Service, d.Port)
	}
	return ""
}

func renderCacheServicePage(a *appctx.App, w http.ResponseWriter, r *http.Request, def serviceDef, status docker.ContainerStatus) {
	layout, _, err := web.BuildLayoutData(a, w, r, def.Title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	statusKey := services.StatusKeyFor(status.State, status.Health)
	color, label := services.StatusColorLabel(layout.T, statusKey)

	data := CachePageData{
		LayoutData: layout, Service: def.Name, Description: layout.T.Get(def.Description), Port: def.Port,
		ContainerState: status.State, HealthStatus: status.Health,
		StatusKey: statusKey, StatusColor: color, StatusLabel: label, StatusMapJSON: services.StatusMapJSON(layout.T),
		IsRunning: status.State == "running",
	}
	if err := cacheServicePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("CACHE - %s template render error: %v", def.Name, err)
	}
}

func renderVarnishPage(a *appctx.App, w http.ResponseWriter, r *http.Request, status docker.ContainerStatus, varnishStatus map[string]string, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Varnish")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	statusKey := services.StatusKeyFor(status.State, status.Health)
	color, label := services.StatusColorLabel(layout.T, statusKey)

	rows := make([]DomainVarnishRow, 0, len(domains))
	for _, d := range domains {
		rows = append(rows, DomainVarnishRow{DomainURL: d.DomainURL, Status: varnishStatus[d.DomainURL]})
	}

	data := CachePageData{
		LayoutData: layout, Service: "varnish",
		Description:    layout.T.Get("Varnish is a reverse caching proxy used as HTTP accelerator for content-heavy dynamic web sites as well as APIs."),
		ContainerState: status.State, HealthStatus: status.Health,
		StatusKey: statusKey, StatusColor: color, StatusLabel: label, StatusMapJSON: services.StatusMapJSON(layout.T),
		IsRunning: status.State == "running", VarnishDomains: rows,
	}
	if err := cacheServicePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("CACHE - varnish template render error: %v", err)
	}
}
