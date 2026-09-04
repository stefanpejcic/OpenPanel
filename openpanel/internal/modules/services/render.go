package services

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

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

// ServicesPageData is system/services.html's template context - either the
// service picker (Services set) or one service's detail view (Service
// set).
type ServicesPageData struct {
	web.LayoutData

	Services []string

	Service       string
	UserContext   string
	IsRunning     bool
	StatusKey     string
	StatusColor   string
	StatusLabel   string
	StatusMapJSON template.JS

	ActionValue string
	ActionLabel string
}

// statusEntry pairs a Tailwind color class with the (untranslated)
// message-catalog key for its label.
type statusEntry struct{ Color, MsgID string }

// serviceStatusMap maps a container status key to its display color and
// translatable label. "stopping" is libpod's real State.Status during
// shutdown (distinct from "exited"/"removing", which apply once it's
// actually stopped/removed) - a render landing in that window used to show
// "Unknown", and since "stopping" was also missing from
// static/js/service-status.js's SERVICE_STATUS_TRANSITIONAL list, the
// auto-refresh poller never recognized it as in-progress either, so the
// badge stayed stuck on "Unknown" until a manual reload happened to land
// outside that window. Keep that file's list in sync with this one.
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

// StatusKeyFor derives the status-map key for a container from its state
// and health check status.
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

// StatusMapJSON marshals every status key's [color, translated label] pair
// to JSON, for service-status.js's client-side polling to redraw the badge
// without another server round-trip.
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

func renderServicesListPage(a *appctx.App, w http.ResponseWriter, r *http.Request, allowedServices []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Choose Service")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ServicesPageData{LayoutData: layout, Services: allowedServices}
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
	}
	if err := servicesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("SERVICES - detail template render error: %v", err)
	}
}
