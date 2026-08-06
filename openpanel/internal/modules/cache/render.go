package cache

import (
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

// DomainVarnishRow is one row of the varnish page's per-domain toggle
// table.
type DomainVarnishRow struct {
	DomainURL string
	Status    string // "On" | "Off" | "Unknown"
}

// CachePageData is cache/service.html's template context, shared by all
// six cache routes. Port == 0 selects the varnish branch (no fixed TCP
// port, per-domain toggle table instead of a port info card).
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
