package docker

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var containersPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"docker/containers.html",
)

// coreServices mirrors containers.html's core_services set: services that
// never show Edit/Delete (built-in, not user-added).
var coreServices = map[string]bool{
	"elasticsearch": true, "redis": true, "valkey": true, "postgres": true,
	"mysql": true, "mariadb": true, "phpmyadmin": true,
	"opensearch": true, "memcached": true, "openresty": true, "nginx": true,
	"apache": true, "openlitespeed": true, "litespeed": true, "varnish": true,
	"cron": true, "backup": true, "tor": true, "docker-proxy": true,
}

// imageTrustKeywords mirrors containers.html's keywords list: image
// references containing one of these get the "verified" badge.
var imageTrustKeywords = []string{
	"httpd", "openlitespeed", "litespeed", "offen/docker-volume-backup", "mcuadros/ofelia", "openresty", "postgres",
	"elasticsearch", "mariadb", "memcached", "mysql", "redis", "valkey", "opensearchproject/opensearch",
	"nginx", "-fpm", "openpanel", "phpmyadmin", "varnish", "docker-socket-proxy", "mongo",
}

// ContainerRow is one containers.html table row, pre-resolved from the
// (already env-substituted, via `podman-compose config`) service details.
type ContainerRow struct {
	Service      string
	DisplayName  string
	Image        string
	ImageTrusted bool
	CPUUnlimited  bool
	CPUValue      string
	RAMUnlimited  bool
	RAMGB         string
	PIDsUnlimited bool
	PIDsValue     string
	ShowManage    bool
}

// ContainersPageData is containers.html's full template context.
type ContainersPageData struct {
	web.LayoutData
	TotalCPU int
	TotalRAM int
	Rows     []ContainerRow
}

func serviceLimit(details map[string]any, key string) string {
	deploy, _ := details["deploy"].(map[string]any)
	if deploy == nil {
		return ""
	}
	resources, _ := deploy["resources"].(map[string]any)
	if resources == nil {
		return ""
	}
	limits, _ := resources["limits"].(map[string]any)
	if limits == nil {
		return ""
	}
	return toStr(limits[key])
}

func formatGB(bytes float64) string {
	gb := bytes / (1024 * 1024 * 1024)
	s := strconv.FormatFloat(gb, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" {
		s = "0"
	}
	return s
}

var memoryValueRE = regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*([kmgt]?)i?b?$`)

// parseMemoryBytes parses a compose-spec memory limit value into bytes.
// `podman-compose config` resolves deploy.resources.limits.memory to a
// bare compose-style string like "0.5G" or "512M" (confirmed against a
// live server's resolved config), not a raw byte count, so the unit
// suffix has to be parsed explicitly for the real value to display.
func parseMemoryBytes(s string) (float64, bool) {
	m := memoryValueRE.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, false
	}
	val, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	mult := 1.0
	switch strings.ToLower(m[2]) {
	case "k":
		mult = 1024
	case "m":
		mult = 1024 * 1024
	case "g":
		mult = 1024 * 1024 * 1024
	case "t":
		mult = 1024 * 1024 * 1024 * 1024
	}
	return val * mult, true
}

// buildContainerRows converts the podman-compose "services" map into
// sorted, display-ready rows.
func buildContainerRows(services map[string]any) []ContainerRow {
	rows := make([]ContainerRow, 0, len(services))
	for name, raw := range services {
		details, _ := raw.(map[string]any)
		row := ContainerRow{Service: name, DisplayName: name}
		if details != nil {
			if cn, ok := details["container_name"].(string); ok && cn != "" {
				row.DisplayName = cn
			}
			row.Image = toStr(details["image"])
		}
		lowerImage := strings.ToLower(row.Image)
		for _, kw := range imageTrustKeywords {
			if strings.Contains(lowerImage, strings.ToLower(kw)) {
				row.ImageTrusted = true
				break
			}
		}

		if details != nil {
			cpuRaw := serviceLimit(details, "cpus")
			if cpuRaw == "" || cpuRaw == "0" {
				row.CPUUnlimited = true
			} else {
				row.CPUValue = cpuRaw
			}

			memRaw := serviceLimit(details, "memory")
			if memRaw == "" || memRaw == "0" {
				row.RAMUnlimited = true
			} else if memBytes, ok := parseMemoryBytes(memRaw); ok && memBytes > 0 {
				row.RAMGB = formatGB(memBytes)
			} else {
				row.RAMUnlimited = true
			}

			pidsRaw := serviceLimit(details, "pids")
			if pidsRaw == "" || pidsRaw == "0" {
				row.PIDsUnlimited = true
			} else {
				row.PIDsValue = pidsRaw
			}
		} else {
			row.CPUUnlimited = true
			row.RAMUnlimited = true
			row.PIDsUnlimited = true
		}

		row.ShowManage = !coreServices[name] && !strings.HasPrefix(name, "php-fpm-")
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Service < rows[j].Service })
	return rows
}

// renderContainersPage renders the containers list page. mysqlType/webserver
// aren't referenced by containers.html itself - they're only used earlier
// to filter dockerData's services - but are kept as parameters so the
// caller's intent is visible at the call site.
func renderContainersPage(a *appctx.App, w http.ResponseWriter, r *http.Request, totalCPU, totalRAM int, mysqlType, webserver string, dockerData map[string]any) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Containers")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := ContainersPageData{LayoutData: layout, TotalCPU: totalCPU, TotalRAM: totalRAM}
	if services, ok := dockerData["services"].(map[string]any); ok {
		data.Rows = buildContainerRows(services)
	}

	if err := containersPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOCKER - containers template render error: %v", err)
	}
}

var containerFormPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"docker/container_form.html",
)

// prefilledContainerForm holds the edit-form prefill values, read from
// the existing compose-file service definition.
type prefilledContainerForm struct {
	ServiceName string
	Image       string
	Environment string
	CPU         string
	RAM         string
	PIDs        string
	Volumes     []VolumeEntry
	AddSocket   bool
	Network     string
	Healthcheck string
}

// containerFormView is what the add/edit container handlers pass to
// renderContainerFormPage: either a fresh form (GET add, neither FormData
// nor PrefilledForm set), a failed-validation redisplay (FormData set to
// the submitted form values), or a GET-edit prefill (PrefilledForm set).
type containerFormView struct {
	Volumes          []string
	Networks         []string
	ExistingServices []string
	Title            string
	Editing          bool
	Error            string
	FormData         url.Values
	PrefilledForm    *prefilledContainerForm
}

// ContainerFormPageData is container_form.html's template context.
type ContainerFormPageData struct {
	web.LayoutData
	Title                string
	Editing              bool
	Error                string
	AvailableVolumes     []string
	AvailableNetworks    []string
	ExistingServicesJSON template.JS
	ServiceName          string
	ServiceNameReadonly  bool
	Image                string
	Environment          string
	CPU                  string
	RAM                  string
	PIDs                 string
	Network              string
	Healthcheck          string
	AddSocket            bool
	VolumeEntries        []VolumeEntry
}

func toJSONStrings(v []string) template.JS {
	if v == nil {
		v = []string{}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return template.JS(b) //nolint:gosec // server-computed list of existing service names, not user input
}

// renderContainerFormPage renders the add/edit container form, choosing
// field values by priority: PrefilledForm wins (GET-edit), then FormData
// as-submitted with no defaulting (POST validation failure - an empty
// field stays empty rather than falling back to a default), then
// hardcoded defaults (GET-add).
func renderContainerFormPage(a *appctx.App, w http.ResponseWriter, r *http.Request, v containerFormView) {
	layout, _, err := web.BuildLayoutData(a, w, r, v.Title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := ContainerFormPageData{
		LayoutData:           layout,
		Title:                v.Title,
		Editing:              v.Editing,
		Error:                v.Error,
		AvailableVolumes:     v.Volumes,
		AvailableNetworks:    v.Networks,
		ExistingServicesJSON: toJSONStrings(v.ExistingServices),
		ServiceNameReadonly:  v.Editing,
	}

	switch {
	case v.PrefilledForm != nil:
		p := v.PrefilledForm
		data.ServiceName = p.ServiceName
		data.Image = p.Image
		data.Environment = p.Environment
		data.CPU = p.CPU
		data.RAM = p.RAM
		data.PIDs = p.PIDs
		data.Network = p.Network
		data.Healthcheck = p.Healthcheck
		data.AddSocket = p.AddSocket
		data.VolumeEntries = p.Volumes
	case v.FormData != nil:
		data.ServiceName = v.FormData.Get("service_name")
		data.Image = v.FormData.Get("image")
		data.Environment = v.FormData.Get("environment")
		data.CPU = v.FormData.Get("cpu")
		data.RAM = v.FormData.Get("ram")
		data.PIDs = v.FormData.Get("pids")
		data.Network = v.FormData.Get("network")
		data.Healthcheck = v.FormData.Get("healthcheck")
		data.AddSocket = v.FormData.Get("add_socket") != ""
	default:
		data.CPU = "0.5"
		data.RAM = "1G"
		data.PIDs = "100"
	}

	if err := containerFormPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOCKER - container_form template render error: %v", err)
	}
}

var deleteConfirmPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"docker/delete_confirm.html",
)

type DeleteConfirmPageData struct {
	web.LayoutData
	Service string
}

// renderDeleteConfirmPage renders the delete-container confirmation page.
func renderDeleteConfirmPage(a *appctx.App, w http.ResponseWriter, r *http.Request, service string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Delete container "+service)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DeleteConfirmPageData{LayoutData: layout, Service: service}
	if err := deleteConfirmPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOCKER - delete_confirm template render error: %v", err)
	}
}

var changeMySQLPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"docker/change_mysql.html",
)

type ChangeMySQLPageData struct {
	web.LayoutData
	Title     string
	MySQLType string
	Available string
}

// renderChangeMySQLPage renders the MySQL-switch page. Unlike
// renderChangeWebserverPage below, it never passes a domains value, so the
// switch form and green highlighting always show with no domain check at
// all for MySQL switching - matching handleContainersMySQL, which
// likewise never computes user domains.
func renderChangeMySQLPage(a *appctx.App, w http.ResponseWriter, r *http.Request, mysqlType, available string) {
	title := "Switch from " + mysqlType + " to " + available
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ChangeMySQLPageData{LayoutData: layout, Title: title, MySQLType: mysqlType, Available: available}
	if err := changeMySQLPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOCKER - change_mysql template render error: %v", err)
	}
}

var changeWebserverPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"docker/change_webserver.html",
)

type ChangeWebserverPageData struct {
	web.LayoutData
	Title                       string
	Webserver                   string
	AvailableOptions            []string
	AvailableOptionsCapitalized []string
	HasDomains                  bool
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// renderChangeWebserverPage renders the webserver-switch page.
func renderChangeWebserverPage(a *appctx.App, w http.ResponseWriter, r *http.Request, webserver string, available []string, userDomains []appctx.Domain) {
	title := "Switch from " + webserver
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	capitalized := make([]string, len(available))
	for i, ws := range available {
		capitalized[i] = capitalizeFirst(ws)
	}
	data := ChangeWebserverPageData{
		LayoutData: layout, Title: title, Webserver: webserver,
		AvailableOptions: available, AvailableOptionsCapitalized: capitalized,
		HasDomains: len(userDomains) > 0,
	}
	if err := changeWebserverPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOCKER - change_webserver template render error: %v", err)
	}
}

var imagesPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"docker/images.html",
)

// cupData mirrors the JSON shape written by the `cup`-based
// "opencli docker-images" image-update checker into cup.json.
type cupData struct {
	Metrics cupMetrics `json:"metrics"`
	Images  []cupImage `json:"images"`
}

type cupMetrics struct {
	MonitoredImages  int `json:"monitored_images"`
	UpdatesAvailable int `json:"updates_available"`
	MinorUpdates     int `json:"minor_updates"`
	MajorUpdates     int `json:"major_updates"`
	UpToDate         int `json:"up_to_date"`
	Unknown          int `json:"unknown"`
}

type cupImage struct {
	Parts struct {
		Repository string `json:"repository"`
		Tag        string `json:"tag"`
	} `json:"parts"`
	URL    string `json:"url"`
	Result struct {
		HasUpdate *bool `json:"has_update"`
		Info      *struct {
			Type              string `json:"type"`
			NewTag            string `json:"new_tag"`
			CurrentVersion    string `json:"current_version"`
			NewVersion        string `json:"new_version"`
			VersionUpdateType string `json:"version_update_type"`
		} `json:"info"`
		Error string `json:"error"`
	} `json:"result"`
}

// ImageRow is one images.html table row.
type ImageRow struct {
	Repository                                            string
	Tag                                                   string
	ImageRef                                              string
	URL                                                   string
	UpdateStatus                                          string // "available" | "uptodate" | "unknown"
	InfoType                                              string // "version" | "digest" | ""
	NewTag, CurrentVersion, NewVersion, VersionUpdateType string
	Error                                                 string
}

// ImagesPageData is images.html's template context.
type ImagesPageData struct {
	web.LayoutData
	LastModified string
	Metrics      cupMetrics
	Rows         []ImageRow
}

// renderImagesPage renders the image-updates page. `data` may come from
// either a cached file read or a fresh `opencli docker-images` run, so
// this always parses it as JSON into cupData regardless of source -
// otherwise a post-refresh render would fail to find the fields it needs.
func renderImagesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data string, lastModified string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Image Updates")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pageData := ImagesPageData{LayoutData: layout, LastModified: lastModified}

	var parsed cupData
	if data != "" && json.Unmarshal([]byte(data), &parsed) == nil {
		pageData.Metrics = parsed.Metrics
		pageData.Rows = make([]ImageRow, 0, len(parsed.Images))
		for _, img := range parsed.Images {
			row := ImageRow{
				Repository: img.Parts.Repository,
				Tag:        img.Parts.Tag,
				ImageRef:   img.Parts.Repository + ":" + img.Parts.Tag,
				URL:        img.URL,
				Error:      img.Result.Error,
			}
			switch {
			case img.Result.HasUpdate == nil:
				row.UpdateStatus = "unknown"
			case *img.Result.HasUpdate:
				row.UpdateStatus = "available"
			default:
				row.UpdateStatus = "uptodate"
			}
			if img.Result.Info != nil {
				row.InfoType = img.Result.Info.Type
				row.NewTag = img.Result.Info.NewTag
				row.CurrentVersion = img.Result.Info.CurrentVersion
				row.NewVersion = img.Result.Info.NewVersion
				row.VersionUpdateType = img.Result.Info.VersionUpdateType
			}
			pageData.Rows = append(pageData.Rows, row)
		}
	}

	if err := imagesPage.Render(w, http.StatusOK, pageData); err != nil {
		log.Printf("DOCKER - images template render error: %v", err)
	}
}

var changeImagePage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"docker/change_images.html",
)

// ChangeImagePageData is change_images.html's template context - either
// the single-service tag-change form (Service set) or the service picker
// (Service empty, SelectableServices populated).
type ChangeImagePageData struct {
	web.LayoutData
	Service            string
	CurrentVersion     string
	SelectableServices []string
}

// renderChangeImagePage renders the single-service image tag-change form.
func renderChangeImagePage(a *appctx.App, w http.ResponseWriter, r *http.Request, service, currentVersion string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Change image tag for "+service)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ChangeImagePageData{LayoutData: layout, Service: service, CurrentVersion: currentVersion}
	if err := changeImagePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOCKER - change_images template render error: %v", err)
	}
}

// renderChangeImageSelectPage renders the service picker for changing an
// image tag when no specific service was requested.
func renderChangeImageSelectPage(a *appctx.App, w http.ResponseWriter, r *http.Request, composeData map[string]any) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Change docker image tag")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var selectable []string
	if services, ok := composeData["services"].(map[string]any); ok {
		for name, raw := range services {
			if strings.HasPrefix(name, "php-") {
				continue
			}
			display := name
			if details, ok := raw.(map[string]any); ok {
				if cn, ok := details["container_name"].(string); ok && cn != "" {
					display = cn
				}
			}
			selectable = append(selectable, display)
		}
		sort.Strings(selectable)
	}

	data := ChangeImagePageData{LayoutData: layout, SelectableServices: selectable}
	if err := changeImagePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOCKER - change_images template render error: %v", err)
	}
}

var logsPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"docker/logs.html",
)

type LogsPageData struct {
	web.LayoutData
	Services []string
}

// renderLogsPage renders the service picker for logs when no specific
// container name was requested.
func renderLogsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, serviceNames []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Logs")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sorted := append([]string(nil), serviceNames...)
	sort.Strings(sorted)
	data := LogsPageData{LayoutData: layout, Services: sorted}
	if err := logsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOCKER - logs template render error: %v", err)
	}
}

var terminalPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"docker/terminal.html",
)

type TerminalPageData struct {
	web.LayoutData
	ContainerName          string
	TerminalTimeoutSeconds int
	ActiveServices         []string
}

// renderTerminalPage renders either the terminal itself (containerName
// set) or the service-picker (containerName empty, activeServiceNames
// populated).
func renderTerminalPage(a *appctx.App, w http.ResponseWriter, r *http.Request, terminalTimeout time.Duration, title, containerName string, activeServiceNames []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sorted := append([]string(nil), activeServiceNames...)
	sort.Strings(sorted)
	data := TerminalPageData{
		LayoutData: layout, ContainerName: containerName,
		TerminalTimeoutSeconds: int(terminalTimeout / time.Second),
		ActiveServices:         sorted,
	}
	if err := terminalPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOCKER - terminal template render error: %v", err)
	}
}
