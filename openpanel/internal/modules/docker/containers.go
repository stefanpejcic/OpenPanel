package docker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// UndeletableServices are the built-in services that can never be removed through the delete-container flow
var UndeletableServices = map[string]bool{
	"elasticsearch": true, "redis": true, "valkey": true, "postgres": true,
	"mysql": true, "mariadb": true, "phpmyadmin": true, "mongodb": true,
	"opensearch": true, "memcached": true, "openresty": true, "nginx": true,
	"apache": true, "openlitespeed": true, "litespeed": true, "varnish": true,
	"cron": true, "backup": true, "tor": true,
}

// webserverHideFilters maps the active webserver to the service-name substrings of every OTHER webserver, so the containers list only shows the one actually in use
var webserverHideFilters = map[string][]string{
	"apache":        {"nginx", "openresty", "openlitespeed", "litespeed"},
	"nginx":         {"apache", "openresty", "openlitespeed", "litespeed"},
	"openlitespeed": {"apache", "openresty", "nginx", "litespeed"},
	"litespeed":     {"apache", "openresty", "nginx", "openlitespeed"},
	"openresty":     {"apache", "nginx", "openlitespeed", "litespeed"},
}

// filterContainerServices drops every OTHER webserver's service (per webserverHideFilters) and the inactive MySQL/MariaDB variant, using exact name matches - not strings.Contains, which would make "openlitespeed" hide itself since it contains "litespeed"
func filterContainerServices(services map[string]any, webserver, mysqlType string) map[string]any {
	filtered := map[string]any{}
	hide := webserverHideFilters[webserver]
	for name, details := range services {
		lower := strings.ToLower(name)
		skip := false
		for _, h := range hide {
			if lower == h {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		filtered[name] = details
	}

	if mysqlType == "mysql" {
		for name := range filtered {
			if strings.Contains(strings.ToLower(name), "mariadb") {
				delete(filtered, name)
			}
		}
	} else if mysqlType == "mariadb" {
		for name := range filtered {
			if strings.Contains(strings.ToLower(name), "mysql") {
				delete(filtered, name)
			}
		}
	}

	return filtered
}

// handleContainersList serves the containers page, with services filtered to the active webserver and MySQL/MariaDB variant
func handleContainersList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)
	planID, _ := injected["hosting_plan"].(int)

	webserver, _ := GetEnvValue(userContext, "WEB_SERVER")
	mysqlType, _ := GetEnvValue(userContext, "MYSQL_TYPE")
	totalCPU, totalRAM := getCPUAndRAMForPlanID(a, ctx, planID)

	if webserver == "" || mysqlType == "" {
		flashAndRedirect(a, w, r, "error", "Missing environment variables. Please check the configuration or restore from backup.", "/dashboard")
		return
	}

	dockerData, dataErr := podmanmanager.LoadComposeConfig(ctx, userContext)
	if dataErr != nil {
		log.Printf("DOCKER - user %s: failed to load compose config, showing 0 containers: %v", userContext, dataErr)
		dockerData = map[string]any{"error": "Failed to fetch container data", "details": dataErr.Error()}
	} else if services, ok := dockerData["services"].(map[string]any); ok {
		dockerData["services"] = filterContainerServices(services, webserver, mysqlType)
	} else {
		dockerData = map[string]any{"error": "Invalid data format", "details": "docker_data does not contain 'services'."}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, dockerData)
		return
	}

	renderContainersPage(a, w, r, totalCPU, totalRAM, mysqlType, webserver, dockerData)
}

// containerFormValidation runs the shared service_name/cpu/ram/pids validation add_container() and edit_container() both do, returning "" if all valid
func validateServiceForm(serviceName, cpu, ram, pids string) string {
	if !IsValidServiceName(serviceName) {
		return "Invalid service name. Must start with a letter, contain only lowercase letters and digits, and be at least 3 characters long."
	}
	if !IsValidCPULimit(cpu) {
		return "CPU limit must be a positive number."
	}
	if !IsValidRAMLimit(ram) {
		return "Memory limit must be a positive number followed by 'M' or 'G' (e.g., 512M or 1.5G)."
	}
	if !IsValidPIDsLimit(pids) {
		return "PIDs limit must be a positive whole number."
	}
	return ""
}

var imageRefRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/:@-]{0,254}$`)

// selectedNetworks keeps the submitted networks that exist in the compose file, in the order given, without duplicates
func selectedNetworks(submitted, available []string) []string {
	var out []string
	for _, n := range submitted {
		n = strings.TrimSpace(n)
		if containsString(available, n) && !containsString(out, n) {
			out = append(out, n)
		}
	}
	return out
}

// handleAddContainer shows the add-service form and, on POST, validates and appends a new service to docker-compose.yml
func handleAddContainer(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	composeData, _ := LoadCompose(userContext)
	availableVolumes := GetAvailableVolumes(composeData)
	availableNetworks := GetAvailableNetworks(composeData)
	existingServices := serviceNames(composeData)

	if r.Method != http.MethodPost {
		if r.URL.Query().Get("output") == "json" {
			writeJSON(w, map[string]any{
				"volumes": availableVolumes, "networks": availableNetworks,
				"error": nil, "existing_services": existingServices,
				"form_data": map[string]any{}, "editing": false,
			})
			return
		}
		renderContainerFormPage(a, w, r, containerFormView{
			Volumes: availableVolumes, Networks: availableNetworks,
			ExistingServices: existingServices, Title: "Add service", Editing: false,
		})
		return
	}

	_ = r.ParseForm()
	serviceName := strings.TrimSpace(r.Form.Get("service_name"))
	image := strings.TrimSpace(r.Form.Get("image"))
	cpu := strings.TrimSpace(r.Form.Get("cpu"))
	ram := strings.TrimSpace(r.Form.Get("ram"))
	pids := strings.TrimSpace(r.Form.Get("pids"))
	networks := selectedNetworks(r.Form["network"], availableNetworks)
	healthcheck := strings.TrimSpace(r.Form.Get("healthcheck"))

	formView := containerFormView{
		Volumes: availableVolumes, Networks: availableNetworks,
		ExistingServices: existingServices, FormData: r.Form, Title: "Add service", Editing: false,
	}

	// the page's JS asks for a stream so it can show each step, a plain form post still gets the redirect
	stream := r.Form.Get("stream") == "1"
	emit := func(map[string]any) {}
	if stream {
		w.Header().Set("Content-Type", "application/x-ndjson")
		flusher, canFlush := w.(http.Flusher)
		emit = func(v map[string]any) {
			b, _ := json.Marshal(v)
			_, _ = w.Write(append(b, '\n'))
			if canFlush {
				flusher.Flush()
			}
		}
	}
	fail := func(msg string) {
		if stream {
			emit(map[string]any{"error": msg})
			return
		}
		formView.Error = msg
		renderContainerFormPage(a, w, r, formView)
	}

	emit(map[string]any{"step": "validate", "status": "Validating container details"})

	if !imageRefRE.MatchString(image) {
		fail("Enter a valid image, for example nginx:latest or ghcr.io/owner/app:1.0.")
		return
	}
	if len(networks) == 0 {
		fail("Choose at least one network.")
		return
	}

	if serviceName == "" || !IsValidServiceName(serviceName) {
		fail("Invalid service name. Must start with a letter, contain only lowercase letters and digits, and be at least 3 characters long.")
		return
	}
	if containsString(existingServices, serviceName) {
		fail(fmt.Sprintf("Service '%s' already exists in docker-compose.yml.", serviceName))
		return
	}
	if !IsValidCPULimit(cpu) {
		fail("CPU limit must be a positive number.")
		return
	}
	if !IsValidRAMLimit(ram) {
		fail("Memory limit must be a positive number followed by 'M' or 'G' (e.g., 512M or 1.5G).")
		return
	}
	if !IsValidPIDsLimit(pids) {
		fail("PIDs limit must be a positive whole number.")
		return
	}

	emit(map[string]any{"step": "save", "status": "Saving container to docker-compose.yml"})

	servicePrefix := ServiceKeyPrefix(serviceName)
	environmentVars, newEnvVars := ParseEnvVars(strings.Split(r.Form.Get("environment"), "\n"), servicePrefix)

	cpuKey, ramKey, pidsKey := servicePrefix+"_CPU", servicePrefix+"_RAM", servicePrefix+"_PIDS"
	newEnvVars[cpuKey] = cpu
	newEnvVars[ramKey] = ram
	newEnvVars[pidsKey] = pids
	UpdateEnvFileWithVars(userContext, newEnvVars)

	mappedVolumes := ParseVolumeEntries(
		r.Form["volume_name"], r.Form["volume_mount"], r.Form["volume_readonly"],
		r.Form.Get("add_socket") != "")

	newService := map[string]any{
		"image": image, "container_name": serviceName, "restart": "always",
		"volumes": mappedVolumes,
		"deploy": map[string]any{"resources": map[string]any{"limits": map[string]any{
			"cpus": "${" + cpuKey + ":-" + cpu + "}", "memory": "${" + ramKey + ":-" + ram + "}", "pids": "${" + pidsKey + ":-" + pids + "}",
		}}},
		"networks": networks,
	}
	if len(environmentVars) > 0 {
		newService["environment"] = environmentVars
	}
	if healthcheck != "" {
		hc, err := parseYAMLString(healthcheck)
		if err != nil {
			fail("Invalid healthcheck YAML: " + err.Error())
			return
		}
		newService["healthcheck"] = hc
	}

	services := servicesRaw(composeData)
	services[serviceName] = newService
	composeData["services"] = services
	if err := SaveCompose(userContext, composeData); err != nil {
		fail("Failed to save configuration: " + err.Error())
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "added container "+serviceName, reqip.ClientIP(r))
	if !stream {
		flashAndRedirect(a, w, r, "success", fmt.Sprintf("Container %s created successfully!", serviceName), "/containers")
		return
	}

	emit(map[string]any{"step": "pull", "status": "Downloading image " + image})
	out, err := podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "pull", image)).CombinedOutput()
	if err != nil {
		emit(map[string]any{"error": "Container saved, but the image could not be downloaded: " + lastLines(string(out), 5), "saved": true})
		return
	}
	emit(map[string]any{"done": true, "status": fmt.Sprintf("Container %s created successfully!", serviceName)})
}

// handleEditContainer shows the edit-service form prefilled from the existing compose definition and, on POST, saves the updated service
func handleEditContainer(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	service := r.PathValue("service")

	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	composeData, _ := LoadCompose(userContext)
	services := servicesRaw(composeData)
	availableVolumes := GetAvailableVolumes(composeData)
	availableNetworks := GetAvailableNetworks(composeData)
	existingServices := serviceNames(composeData)
	title := fmt.Sprintf("Edit service %s", service)

	svcRaw, ok := services[service]
	if !ok {
		http.NotFound(w, r)
		return
	}
	svc, _ := svcRaw.(map[string]any)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		serviceName := strings.TrimSpace(r.Form.Get("service_name"))
		image := strings.TrimSpace(r.Form.Get("image"))
		cpu := strings.TrimSpace(r.Form.Get("cpu"))
		ram := strings.TrimSpace(r.Form.Get("ram"))
		pids := strings.TrimSpace(r.Form.Get("pids"))
		networks := selectedNetworks(r.Form["network"], availableNetworks)
		healthcheck := strings.TrimSpace(r.Form.Get("healthcheck"))

		formView := containerFormView{
			Volumes: availableVolumes, Networks: availableNetworks,
			ExistingServices: existingServices, FormData: r.Form, Title: title, Editing: true,
		}

		if !imageRefRE.MatchString(image) {
			formView.Error = "Enter a valid image, for example nginx:latest or ghcr.io/owner/app:1.0."
			renderContainerFormPage(a, w, r, formView)
			return
		}
		if len(networks) == 0 {
			formView.Error = "Choose at least one network."
			renderContainerFormPage(a, w, r, formView)
			return
		}

		if serviceName != service {
			formView.Error = "Service name cannot be changed."
			renderContainerFormPage(a, w, r, formView)
			return
		}
		if !IsValidServiceName(serviceName) {
			formView.Error = "Invalid service name. Must start with a letter, contain only lowercase letters and digits, and be at least 3 characters long."
			renderContainerFormPage(a, w, r, formView)
			return
		}
		if !IsValidCPULimit(cpu) {
			formView.Error = "CPU limit must be a positive number."
			renderContainerFormPage(a, w, r, formView)
			return
		}
		if !IsValidRAMLimit(ram) {
			formView.Error = "Memory limit must be a positive number followed by 'M' or 'G' (e.g., 512M or 1.5G)."
			renderContainerFormPage(a, w, r, formView)
			return
		}
		if !IsValidPIDsLimit(pids) {
			formView.Error = "PIDs limit must be a positive whole number."
			renderContainerFormPage(a, w, r, formView)
			return
		}

		servicePrefix := ServiceKeyPrefix(serviceName)
		environmentVars, newEnvVars := ParseEnvVars(strings.Split(r.Form.Get("environment"), "\n"), servicePrefix)

		cpuKey, ramKey, pidsKey := servicePrefix+"_CPU", servicePrefix+"_RAM", servicePrefix+"_PIDS"
		newEnvVars[cpuKey] = cpu
		newEnvVars[ramKey] = ram
		newEnvVars[pidsKey] = pids

		currentEnv := LoadEnvFile(userContext)
		prefix := servicePrefix + "_"
		for k := range currentEnv {
			if strings.HasPrefix(k, prefix) {
				delete(currentEnv, k)
			}
		}
		for k, v := range newEnvVars {
			currentEnv[k] = v
		}
		_ = SaveEnvFile(userContext, currentEnv)

		mappedVolumes := ParseVolumeEntries(
			r.Form["volume_name"], r.Form["volume_mount"], r.Form["volume_readonly"],
			r.Form.Get("add_socket") != "")

		updatedService := map[string]any{
			"image": image, "container_name": serviceName, "restart": "always",
			"volumes": mappedVolumes, "networks": networks,
			"deploy": map[string]any{"resources": map[string]any{"limits": map[string]any{
				"cpus": "${" + cpuKey + "}", "memory": "${" + ramKey + "}", "pids": "${" + pidsKey + "}",
			}}},
		}
		if len(environmentVars) > 0 {
			updatedService["environment"] = environmentVars
		}
		if healthcheck != "" {
			hc, err := parseYAMLString(healthcheck)
			if err != nil {
				formView.Error = "Invalid healthcheck YAML: " + err.Error()
				renderContainerFormPage(a, w, r, formView)
				return
			}
			updatedService["healthcheck"] = hc
		}

		services[service] = updatedService
		composeData["services"] = services
		_ = SaveCompose(userContext, composeData)

		_ = logger.RecordUserAction(a.Config, username, "edited container "+service, reqip.ClientIP(r))
		flashAndRedirect(a, w, r, "success", "Container updated successfully.", "/containers/edit/"+service)
		return
	}

	// GET: populate the form from the existing service definition.
	// values live in .env behind ${SERVICE_KEY} placeholders, resolve them so saving the edit doesn't wipe them
	env := LoadEnvFile(userContext)
	var envLines []string
	if envRaw, ok := svc["environment"].(map[string]any); ok {
		keys := make([]string, 0, len(envRaw))
		for k := range envRaw {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			envLines = append(envLines, k+": "+ResolveEnvPlaceholder(toStr(envRaw[k]), env))
		}
	}

	var cpu, ram, pids string
	if deploy, ok := svc["deploy"].(map[string]any); ok {
		if resources, ok := deploy["resources"].(map[string]any); ok {
			if limits, ok := resources["limits"].(map[string]any); ok {
				cpu = ResolveEnvPlaceholder(toStr(limits["cpus"]), env)
				ram = ResolveEnvPlaceholder(toStr(limits["memory"]), env)
				pids = ResolveEnvPlaceholder(toStr(limits["pids"]), env)
			}
		}
	}
	if pids == "" {
		pids = "100"
	}

	var volumeEntries []VolumeEntry
	addSocket := false
	if volsRaw, ok := svc["volumes"].([]any); ok {
		for _, v := range volsRaw {
			vs, _ := v.(string)
			parts := strings.SplitN(vs, ":", 3)
			if len(parts) >= 2 {
				volumeEntries = append(volumeEntries, VolumeEntry{
					Name: parts[0], Mount: parts[1], ReadOnly: len(parts) == 3 && parts[2] == "ro",
				})
			}
			if strings.Contains(vs, "docker.sock") {
				addSocket = true
			}
		}
	}

	var networks []string
	switch nets := svc["networks"].(type) {
	case []any:
		for _, n := range nets {
			if name, ok := n.(string); ok {
				networks = append(networks, name)
			}
		}
	case map[string]any:
		for name := range nets {
			networks = append(networks, name)
		}
		sort.Strings(networks)
	}

	healthcheck := ""
	if hc, ok := svc["healthcheck"]; ok && hc != nil {
		healthcheck = strings.TrimSpace(dumpYAML(hc))
	}

	renderContainerFormPage(a, w, r, containerFormView{
		Volumes: availableVolumes, Networks: availableNetworks,
		ExistingServices: existingServices, Title: title, Editing: true,
		PrefilledForm: &prefilledContainerForm{
			ServiceName: service, Image: toStr(svc["image"]), Environment: strings.Join(envLines, "\n"),
			CPU: cpu, RAM: ram, PIDs: pids, Volumes: volumeEntries, AddSocket: addSocket,
			Networks: networks, Healthcheck: healthcheck,
		},
	})
}

// handleDeleteContainer stops and removes a service and its image, and on GET shows a confirmation page first
func handleDeleteContainer(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	service := r.PathValue("service")

	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	if UndeletableServices[service] || strings.HasPrefix(service, "php-fpm-") {
		http.Error(w, fmt.Sprintf("Service '%s' cannot be deleted.", service), http.StatusForbidden)
		return
	}

	composeData, _ := LoadCompose(userContext)
	services := servicesRaw(composeData)
	svcRaw, ok := services[service]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		svc, _ := svcRaw.(map[string]any)
		containerName := service
		if cn, ok := svc["container_name"].(string); ok && cn != "" {
			containerName = cn
		}
		imageName, _ := svc["image"].(string)

		StartOrStopContainer(ctx, userContext, containerName, "deactivate", "")
		if imageName != "" {
			removeImage(ctx, userContext, imageName)
		}

		env := LoadEnvFile(userContext)
		prefix := strings.ToUpper(service) + "_"
		for k := range env {
			if strings.HasPrefix(k, prefix) {
				delete(env, k)
			}
		}
		_ = SaveEnvFile(userContext, env)

		delete(services, service)
		composeData["services"] = services
		_ = SaveCompose(userContext, composeData)

		_ = logger.RecordUserAction(a.Config, username, "deleted container "+service, reqip.ClientIP(r))
		http.Redirect(w, r, "/containers/new", http.StatusFound)
		return
	}

	renderDeleteConfirmPage(a, w, r, service)
}

func serviceNames(composeData map[string]any) []string {
	names := make([]string, 0)
	for k := range servicesRaw(composeData) {
		names = append(names, k)
	}
	return names
}

func servicesRaw(composeData map[string]any) map[string]any {
	if s, ok := composeData["services"].(map[string]any); ok {
		return s
	}
	m := map[string]any{}
	composeData["services"] = m
	return m
}

func toStr(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// lastLines keeps the end of a command's output, that's where the actual error is
func lastLines(out string, n int) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
