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

// coreServices mirrors containers.html's core_services set: services that never show Edit/Delete (built-in, not user-added)
var coreServices = map[string]bool{
	"elasticsearch": true, "redis": true, "valkey": true, "postgres": true,
	"mysql": true, "mariadb": true, "phpmyadmin": true, "mongodb": true,
	"opensearch": true, "memcached": true, "openresty": true, "nginx": true,
	"apache": true, "openlitespeed": true, "litespeed": true, "varnish": true,
	"cron": true, "backup": true, "tor": true, "docker-proxy": true,
}

// imageTrustKeywords mirrors containers.html's keywords list: image references containing one of these get the "verified" badge
var imageTrustKeywords = []string{
	"httpd", "openlitespeed", "litespeed", "offen/docker-volume-backup", "mcuadros/ofelia", "openresty", "postgres",
	"elasticsearch", "mariadb", "memcached", "mysql", "redis", "valkey", "opensearchproject/opensearch",
	"nginx", "-fpm", "openpanel", "phpmyadmin", "varnish", "docker-socket-proxy", "mongo",
}

// ContainerRow is one containers.html table row, pre-resolved from the (already env-substituted, via `podman-compose config`) service details
type ContainerRow struct {
	Service       string
	DisplayName   string
	Image         string
	ImageTrusted  bool
	CPUUnlimited  bool
	CPUValue      string
	RAMUnlimited  bool
	RAMGB         string
	PIDsUnlimited bool
	PIDsValue     string
	ShowManage    bool
	// Ports lists the container-side ports, PortsTitle the full host:container mappings
	Ports           string
	PortsTitle      string
	Icon            string
	Hue             string
	ImageChangeable bool
}

// ContainersPageData is containers.html's full template context.
type ContainersPageData struct {
	web.LayoutData
	TotalCPU int
	TotalRAM int
	Rows     []ContainerRow
}

// serviceIcon picks the row's kit icon and tile hue from the service name
func serviceIcon(name string) (icon, hue string) {
	switch {
	case strings.HasPrefix(name, "php-fpm"):
		return "code", "indigo"
	case name == "mysql" || name == "mariadb" || name == "postgres" || name == "mongodb" || name == "phpmyadmin":
		return "db", "purple"
	case name == "redis" || name == "valkey" || name == "memcached":
		return "layout", "red"
	case name == "elasticsearch" || name == "opensearch":
		return "search", "orange"
	case name == "nginx" || name == "apache" || name == "openresty" || name == "openlitespeed" || name == "litespeed" || name == "varnish":
		return "browser", "cyan"
	case name == "cron" || name == "backup" || name == "tor" || name == "docker-proxy":
		return "shield", "green"
	}
	return "apps", "blue"
}

// servicePorts reads compose ports in short ("8080:80/tcp") or long ({published, target}) syntax
func servicePorts(details map[string]any) (targets, mappings []string) {
	ports, _ := details["ports"].([]any)
	for _, p := range ports {
		switch v := p.(type) {
		case string:
			mappings = append(mappings, v)
			parts := strings.Split(v, ":")
			targets = append(targets, strings.SplitN(parts[len(parts)-1], "/", 2)[0])
		case map[string]any:
			target := toStr(v["target"])
			if target == "" {
				continue
			}
			targets = append(targets, target)
			if pub := toStr(v["published"]); pub != "" {
				mappings = append(mappings, pub+":"+target)
			} else {
				mappings = append(mappings, target)
			}
		}
	}
	return targets, mappings
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

// parseMemoryBytes parses a compose-spec memory limit value into bytes. `podman-compose config` resolves deploy.resources.limits.memory to a bare compose-style string like "0.5G" or "512M", not a raw byte count, so the unit suffix has to be parsed explicitly for the real value to display.
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

// buildContainerRows converts the podman-compose "services" map into sorted, display-ready rows
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
		row.ImageChangeable = imageChangeable(name)
		row.Icon, row.Hue = serviceIcon(name)
		if details != nil {
			targets, mappings := servicePorts(details)
			row.Ports = strings.Join(targets, ", ")
			row.PortsTitle = strings.Join(mappings, ", ")
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Service < rows[j].Service })
	return rows
}

// renderContainersPage renders the containers list page. mysqlType/webserver aren't referenced by containers.html itself - they were only used earlier to filter dockerData's services - but are kept as parameters so the caller's intent is visible at the call site.
func renderContainersPage(a *appctx.App, w http.ResponseWriter, r *http.Request, totalCPU, totalRAM int, mysqlType, webserver string, dockerData map[string]any) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Containers")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	layout.BulkActions = containerBulkActions(layout.T, totalCPU, totalRAM)
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

// prefilledContainerForm holds the edit-form prefill values, read from the existing compose-file service definition
type prefilledContainerForm struct {
	ServiceName string
	Image       string
	Environment string
	CPU         string
	RAM         string
	PIDs        string
	Volumes     []VolumeEntry
	AddSocket   bool
	Networks    []string
	Healthcheck string
	Domain      string
	Port        string
}

// containerFormView is what the add/edit container handlers pass to renderContainerFormPage: either a fresh form (GET add, neither FormData nor PrefilledForm set), a failed-validation redisplay (FormData set to the submitted values), or a GET-edit prefill (PrefilledForm set)
type containerFormView struct {
	Volumes          []string
	Networks         []string
	ExistingServices []string
	Title            string
	Editing          bool
	Error            string
	FormData         url.Values
	PrefilledForm    *prefilledContainerForm
	Domains          []string
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
	Networks             []string
	Healthcheck          string
	AddSocket            bool
	VolumeEntries        []VolumeEntry
	Domains              []string
	Domain               string
	Port                 string
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

// renderContainerFormPage renders the add/edit container form, choosing field values by priority: PrefilledForm wins (GET-edit), then FormData as-submitted with no defaulting (POST validation failure - an empty field stays empty rather than falling back to a default), then hardcoded defaults (GET-add)
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
		Domains:              v.Domains,
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
		data.Networks = p.Networks
		data.Healthcheck = p.Healthcheck
		data.AddSocket = p.AddSocket
		data.VolumeEntries = p.Volumes
		data.Domain = p.Domain
		data.Port = p.Port
	case v.FormData != nil:
		data.ServiceName = v.FormData.Get("service_name")
		data.Image = v.FormData.Get("image")
		data.Environment = v.FormData.Get("environment")
		data.CPU = v.FormData.Get("cpu")
		data.RAM = v.FormData.Get("ram")
		data.PIDs = v.FormData.Get("pids")
		data.Networks = v.FormData["network"]
		data.Healthcheck = v.FormData.Get("healthcheck")
		data.AddSocket = v.FormData.Get("add_socket") != ""
		data.Domain = v.FormData.Get("domain")
		data.Port = v.FormData.Get("port")
		names, mounts, ro := v.FormData["volume_name"], v.FormData["volume_mount"], v.FormData["volume_readonly"]
		for i := range names {
			e := VolumeEntry{Name: names[i]}
			if i < len(mounts) {
				e.Mount = mounts[i]
			}
			e.ReadOnly = i < len(ro) && ro[i] == "on"
			data.VolumeEntries = append(data.VolumeEntries, e)
		}
	default:
		data.CPU = "0.5"
		data.RAM = "1G"
		data.PIDs = "500"
		if len(v.Networks) > 0 {
			data.Networks = []string{v.Networks[0]}
		}
	}
	sort.Strings(data.AvailableNetworks)
	sort.Strings(data.AvailableVolumes)

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
	Choices   []ServerChoice
	Blocked   bool
	DBCount   int
}

// databaseChoices are the cards shown on /containers/mysql, same names and descriptions as the onboarding modal
var databaseChoices = []ServerChoice{
	{Value: "mysql", Label: "MySQL", Description: "Can be better fine-tuned for any type of application.", Icon: "mysql"},
	{Value: "mariadb", Label: "MariaDB", Description: "Recommended for WordPress sites - smaller image, lower resource usage, and faster default query execution.", Icon: "mariadb"},
	{Value: "percona", Label: "Percona", Description: "MySQL drop-in with extra performance and monitoring features, geared toward high-traffic sites.", Icon: "percona"},
}

// databaseService is the compose service behind a database choice, Percona runs as the mysql service with another image
func databaseService(choice string) string {
	if choice == "percona" {
		return "mysql"
	}
	return choice
}

// renderChangeMySQLPage renders the database server switch page, blocked while user databases still exist
func renderChangeMySQLPage(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext, flavor string, blocked bool, dbCount int) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Database Server")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ChangeMySQLPageData{
		LayoutData: layout, Title: "Database Server", MySQLType: flavor, Blocked: blocked, DBCount: dbCount,
		Choices: serverChoicesFor(userContext, databaseChoices, flavor, databaseService),
	}
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

// ServerChoice is one option card on the webserver and database switch pages
type ServerChoice struct {
	Value       string
	Label       string
	Description string
	Icon        string
	Current     bool
}

// webserverChoices are the cards shown on /containers/webserver, same names and descriptions as the onboarding modal
var webserverChoices = []ServerChoice{
	{Value: "apache", Label: "Apache", Description: "Recommended for websites that use .htaccess for rewrite rules.", Icon: "apache"},
	{Value: "nginx", Label: "Nginx", Description: "Recommended for NodeJS and Python applications, and high-traffic sites.", Icon: "nginx"},
	{Value: "openlitespeed", Label: "OpenLiteSpeed", Description: "Recommended for WordPress websites, but only one PHP version.", Icon: "openlitespeed"},
	{Value: "openresty", Label: "OpenResty", Description: "Nginx core bundled with Lua scripting for custom logic.", Icon: "openresty"},
	{Value: "litespeed", Label: "LiteSpeed", Description: "LiteSpeed Enterprise, needs a license.", Icon: "openlitespeed"},
}

type ChangeWebserverPageData struct {
	web.LayoutData
	Title       string
	Webserver   string
	Choices     []ServerChoice
	DomainCount int
}

// serverChoicesFor keeps the choices whose service exists in the user's compose file and marks the current one
func serverChoicesFor(userContext string, all []ServerChoice, current string, service func(string) string) []ServerChoice {
	services := map[string]map[string]any{}
	if composeData, err := LoadCompose(userContext); err == nil {
		services = servicesOf(composeData)
	}
	var out []ServerChoice
	for _, c := range all {
		_, exists := services[service(c.Value)]
		if !exists && c.Value != current {
			continue
		}
		c.Current = c.Value == current
		out = append(out, c)
	}
	return out
}

// renderChangeWebserverPage renders the webserver-switch page.
func renderChangeWebserverPage(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext, webserver string, userDomains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Webserver")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ChangeWebserverPageData{
		LayoutData: layout, Title: "Webserver", Webserver: webserver,
		Choices:     webserverChoicesFor(userContext, webserver),
		DomainCount: len(userDomains),
	}
	if err := changeWebserverPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("DOCKER - change_webserver template render error: %v", err)
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

// ChangeImagePageData is change_images.html's template context: every service whose tag can change, Open is the one whose sheet opens on load
type ChangeImagePageData struct {
	web.LayoutData
	Rows  []ImageRow
	Open  string
	Error string

	// the classic theme's picker and single-service form read these
	Service            string
	CurrentVersion     string
	HubURL             string
	SelectableServices []string
}

// ImageRow is one service on the Software Versions page
type ImageRow struct {
	Service  string `json:"service"`
	Repo     string `json:"repo"`
	Tag      string `json:"tag"`
	HubURL   string `json:"hub"`
	Floating bool   `json:"floating"` // a tag like latest that moves to a new version on the next pull
	Data     bool   `json:"data"`     // keeps data files a newer major version upgrades
	Pinned   bool   `json:"pinned"`   // the tag is written in docker-compose.yml, not read from .env
	Icon     string `json:"-"`
	Hue      string `json:"-"`
}

// dataServices keep data files that a major version upgrade rewrites
var dataServices = map[string]bool{"mysql": true, "mariadb": true, "postgres": true, "mongodb": true, "redis": true, "valkey": true, "elasticsearch": true, "opensearch": true}

// imageRows lists the changeable services of a compose file with the tag each one runs
func imageRows(userContext string, composeData map[string]any, env map[string]string) []ImageRow {
	var rows []ImageRow
	for name, details := range servicesOf(composeData) {
		if !imageChangeable(name) {
			continue
		}
		image, _ := details["image"].(string)
		img, parsed := parseImage(image)
		_, fixed := fixedTagVar(userContext, name)
		envVar, tag := tagVarFor(userContext, name, img, parsed)
		if v := strings.Trim(env[envVar], `'"`); v != "" {
			tag = v
		}
		repo := img.Repo
		if !parsed {
			repo, tag = splitImage(image, tag)
		}
		row := ImageRow{Service: name, Repo: repo, Tag: tag, HubURL: hubTagsPage(repo), Data: dataServices[name], Pinned: !parsed && !fixed}
		row.Floating = tag != "" && !strings.ContainsAny(tag, "0123456789")
		row.Icon, row.Hue = serviceIcon(name)
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Service < rows[j].Service })
	return rows
}

// currentImage is the tag and Docker Hub page of one service, read directly when it isn't in rows
func currentImage(userContext, service string, rows []ImageRow) (tag, hub string) {
	for _, row := range rows {
		if row.Service == service {
			return row.Tag, row.HubURL
		}
	}
	envVar, tag := imageTagVar(userContext, service)
	if v, _ := GetEnvValue(userContext, envVar); v != "" {
		tag = v
	}
	return tag, dockerHubPage(userContext, service)
}

// splitImage cuts "repo:tag" for images whose tag isn't a variable, fallback when there's no tag
func splitImage(image, fallback string) (repo, tag string) {
	image = strings.TrimSpace(image)
	if i := strings.LastIndex(image, ":"); i > strings.LastIndex(image, "/") {
		return image[:i], image[i+1:]
	}
	return image, fallback
}

// renderChangeImagePage renders the Software Versions table, open names the service whose change sheet opens on load
func renderChangeImagePage(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext, open string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Software Versions")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ChangeImagePageData{LayoutData: layout, Open: open}
	if composeData, err := LoadCompose(userContext); err != nil {
		data.Error = web.Tr(a, r, "Could not read your services.")
	} else {
		data.Rows = imageRows(userContext, composeData, LoadEnvFile(userContext))
	}
	for _, row := range data.Rows {
		data.SelectableServices = append(data.SelectableServices, row.Service)
	}
	if open != "" {
		data.Service = open
		data.CurrentVersion, data.HubURL = currentImage(userContext, open, data.Rows)
	}
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

// renderLogsPage renders the service picker for logs when no specific container name was requested
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

// renderTerminalPage renders either the terminal itself (containerName set) or the service-picker (containerName empty, activeServiceNames populated)
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
