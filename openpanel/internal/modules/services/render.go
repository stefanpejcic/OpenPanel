package services

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var servicesPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"system/services.html",
)

// ServicesPageData is system/services.html's template context - either the service picker (Services set) or one service's detail view (Service set)
type ServicesPageData struct {
	web.LayoutData

	Services []string
	Rows     []ServiceRow
	Running  int
	Stopped  int
	Stack    string

	Service       string
	UserContext   string
	IsRunning     bool
	StatusKey     string
	StatusColor   string
	StatusLabel   string
	StatusMapJSON template.JS

	ActionValue string
	ActionLabel string
	Kind        string
	StatusTone  string
	Uptime      string
}

// statusEntry pairs a Tailwind color class with the (untranslated) message-catalog key for its label
type statusEntry struct{ Color, MsgID string }

// serviceStatusMap maps a container status key to its display color and translatable label - "stopping" is libpod's real State.Status during shutdown, distinct from "exited"/"removing"; keep static/js/service-status.js's SERVICE_STATUS_TRANSITIONAL list in sync with this one or the badge gets stuck on "Unknown" mid-shutdown
var serviceStatusMap = map[string]statusEntry{
	"running":    {"emerald-500", "Running"},
	"healthy":    {"emerald-500", "Running"},
	"unhealthy":  {"red-500", "Unhealthy"},
	"starting":   {"orange-500", "Starting"},
	"not_found":  {"gray-400", "Disabled"},
	"created":    {"blue-500", "Created"},
	"restarting": {"orange-500", "Restarting"},
	"paused":     {"orange-500", "Paused"},
	"exited":     {"red-500", "Stopped"},
	"removing":   {"orange-500", "Removing"},
	"stopping":   {"orange-500", "Stopping"},
	"dead":       {"red-700", "Dead"},
}

var unknownStatus = statusEntry{"orange-500", "Unknown"}

// StatusKeyFor derives the status-map key for a container from its state and health check status
func StatusKeyFor(containerState, healthStatus string) string {
	if containerState == "running" {
		switch healthStatus {
		case "healthy", "unhealthy", "starting":
			return healthStatus
		}
		return "running"
	}
	return containerState
}

func StatusColorLabel(t i18n.Translator, key string) (string, string) {
	e, ok := serviceStatusMap[key]
	if !ok {
		e = unknownStatus
	}
	return e.Color, t.Get(e.MsgID)
}

// StatusMapJSON marshals every status key's [color, translated label] pair to JSON, for service-status.js's client-side polling to redraw the badge without another server round-trip
func StatusMapJSON(t i18n.Translator) template.JS {
	m := make(map[string][2]string, len(serviceStatusMap))
	for k, e := range serviceStatusMap {
		m[k] = [2]string{e.Color, t.Get(e.MsgID)}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return template.JS(b) //nolint:gosec // server-computed status labels/colors, not user input
}

// ServiceRow is one line of the services list
type ServiceRow struct {
	Name        string
	Kind        string
	Icon        string
	Hue         string
	StatusLabel string
	StatusTone  string
	Running     bool
	Uptime      string
}

// serviceKind describes a compose service by its name, for the list's icon and subtitle
func serviceKind(name string) (kind, icon, hue string) {
	switch {
	case webserverServices[name]:
		return "Web server", "globe", "cyan"
	case mysqlServices[name], name == "postgres", name == "postgresql", name == "mongodb", name == "mongo":
		return "Database", "db", "indigo"
	case name == "phpmyadmin", name == "pgadmin", name == "mongo-express":
		return "Database admin", "browser", "blue"
	case strings.HasPrefix(name, "php-fpm"):
		return "PHP " + strings.TrimPrefix(strings.TrimPrefix(name, "php-fpm"), "-"), "code", "purple"
	case name == "redis", name == "memcached":
		return "Object cache", "refresh", "red"
	case name == "varnish":
		return "HTTP cache", "cloud", "orange"
	case name == "elasticsearch", name == "opensearch", name == "sphinx", name == "solr", name == "meilisearch":
		return "Search engine", "search", "yellow"
	case strings.Contains(name, "mail") || name == "roundcube":
		return "Mail", "mail", "green"
	}
	return "Service", "apps", "blue"
}

// uptimeLabel says how long a container has been up, in its largest whole unit
func uptimeLabel(t i18n.Translator, d time.Duration) string {
	switch {
	case d < time.Minute:
		return t.Get("Just started")
	case d < time.Hour:
		return t.Get("%(n)s minutes", "n", strconv.Itoa(int(d/time.Minute)))
	case d < 48*time.Hour:
		return t.Get("%(n)s hours", "n", strconv.Itoa(int(d/time.Hour)))
	}
	return t.Get("%(n)s days", "n", strconv.Itoa(int(d/(24*time.Hour))))
}

// statusTone picks the kit colour for a status key: green when up, amber while changing, red when broken, plain when off
func statusTone(key string) string {
	switch key {
	case "running", "healthy":
		return "is-success"
	case "unhealthy", "dead":
		return "is-danger"
	case "exited", "not_found":
		return "is-muted"
	}
	return "is-warning"
}

// servicesListData builds the list page's rows and summary from each service's container status
func servicesListData(layout web.LayoutData, allowedServices []string, statuses []docker.ContainerStatus, webserver, mysqlType string) ServicesPageData {
	data := ServicesPageData{LayoutData: layout, Services: allowedServices}
	for i, name := range allowedServices {
		key := StatusKeyFor(statuses[i].State, statuses[i].Health)
		_, label := StatusColorLabel(layout.T, key)
		kind, icon, hue := serviceKind(name)
		row := ServiceRow{Name: name, Kind: kind, Icon: icon, Hue: hue, StatusLabel: label, StatusTone: statusTone(key), Running: statuses[i].State == "running"}
		if row.Running && !statuses[i].Started.IsZero() {
			row.Uptime = uptimeLabel(layout.T, time.Since(statuses[i].Started))
		}
		if row.Running {
			data.Running++
		}
		data.Rows = append(data.Rows, row)
	}
	// compose order isn't stable, so running services come first, then by name
	sort.SliceStable(data.Rows, func(i, j int) bool {
		if data.Rows[i].Running != data.Rows[j].Running {
			return data.Rows[i].Running
		}
		return data.Rows[i].Name < data.Rows[j].Name
	})
	data.Stopped = len(data.Rows) - data.Running
	data.Stack = strings.Trim(webserver+" · "+mysqlType, " ·")
	return data
}

func renderServicesListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, allowedServices []string, statuses []docker.ContainerStatus, webserver, mysqlType string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Services")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.BulkActions = servicesBulkActions(layout.T)
	data := servicesListData(layout, allowedServices, statuses, webserver, mysqlType)
	if err := servicesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("SERVICES - list template render error: %v", err)
	}
}

func renderServiceDetailPage(a *appctx.App, w http.ResponseWriter, r *http.Request, service, userContext string, status docker.ContainerStatus) {
	layout, _, err := web.BuildLayoutData(a, w, r, service)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	statusKey := StatusKeyFor(status.State, status.Health)
	color, label := StatusColorLabel(layout.T, statusKey)

	actionValue, actionLabel := "disable", layout.T.Get("Click to Disable")
	if status.State == "not_found" {
		actionValue, actionLabel = "enable", layout.T.Get("Click to Enable")
	}

	data := ServicesPageData{
		LayoutData: layout, Service: service, UserContext: userContext,
		IsRunning: status.State == "running",
		StatusKey: statusKey, StatusColor: color, StatusLabel: label,
		StatusMapJSON: StatusMapJSON(layout.T),
		ActionValue:   actionValue, ActionLabel: actionLabel,
		StatusTone: statusTone(statusKey),
	}
	data.Kind, _, _ = serviceKind(service)
	if data.IsRunning && !status.Started.IsZero() {
		data.Uptime = uptimeLabel(layout.T, time.Since(status.Started))
	}
	if err := servicesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("SERVICES - detail template render error: %v", err)
	}
}
