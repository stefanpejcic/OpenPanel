package docker

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// UndeletableServices are the built-in services that can never be
// removed through the delete-container flow.
var UndeletableServices = map[string]bool{
	"elasticsearch": true, "redis": true, "valkey": true, "postgres": true,
	"mysql": true, "mariadb": true, "phpmyadmin": true,
	"opensearch": true, "memcached": true, "openresty": true, "nginx": true,
	"apache": true, "openlitespeed": true, "litespeed": true, "varnish": true,
	"cron": true, "backup": true, "tor": true,
}

// webserverHideFilters maps the active webserver to the service-name
// substrings of every OTHER webserver, so the containers list only shows
// the one actually in use.
var webserverHideFilters = map[string][]string{
	"apache":        {"nginx", "openresty", "openlitespeed", "litespeed"},
	"nginx":         {"apache", "openresty", "openlitespeed", "litespeed"},
	"openlitespeed": {"apache", "openresty", "nginx", "litespeed"},
	"litespeed":     {"apache", "openresty", "nginx", "openlitespeed"},
	"openresty":     {"apache", "nginx", "openlitespeed", "litespeed"},
}

// filterContainerServices drops every OTHER webserver's service (per
// webserverHideFilters) and the inactive MySQL/MariaDB variant, using exact
// name matches - NOT strings.Contains, which would make "openlitespeed"
// hide itself since it contains "litespeed" as a substring.
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

// handleContainersList serves the containers page, with services filtered
// to the active webserver and MySQL/MariaDB variant.
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

// containerFormValidation runs the shared service_name/cpu/ram/pids
// validation add_container() and edit_container() both do, returning ""
// if all valid.
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

// handleAddContainer shows the add-service form and, on POST, validates
// and appends a new service to docker-compose.yml.
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
	network := strings.TrimSpace(r.Form.Get("network"))
	healthcheck := strings.TrimSpace(r.Form.Get("healthcheck"))

	formView := containerFormView{
		Volumes: availableVolumes, Networks: availableNetworks,
		ExistingServices: existingServices, FormData: r.Form, Title: "Add service", Editing: false,
	}

	if serviceName == "" || !IsValidServiceName(serviceName) {
		formView.Error = "Invalid service name. Must start with a letter, contain only lowercase letters and digits, and be at least 3 characters long."
		renderContainerFormPage(a, w, r, formView)
		return
	}
	if containsString(existingServices, serviceName) {
		formView.Error = fmt.Sprintf("Service '%s' already exists in docker-compose.yml.", serviceName)
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
		"networks": []string{network},
	}
	if len(environmentVars) > 0 {
		newService["environment"] = environmentVars
	}
	if healthcheck != "" {
		hc, err := parseYAMLString(healthcheck)
		if err != nil {
			formView.Error = "Invalid healthcheck YAML: " + err.Error()
			renderContainerFormPage(a, w, r, formView)
			return
		}
		newService["healthcheck"] = hc
	}

	services := servicesRaw(composeData)
	services[serviceName] = newService
	composeData["services"] = services
	if err := SaveCompose(userContext, composeData); err != nil {
		formView.Error = "Failed to save configuration: " + err.Error()
		renderContainerFormPage(a, w, r, formView)
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "added container "+serviceName, reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", fmt.Sprintf("Container %s created successfully!", serviceName), "/containers")
}

// handleEditContainer shows the edit-service form prefilled from the
// existing compose definition and, on POST, saves the updated service.
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
		network := r.Form.Get("network")
		healthcheck := strings.TrimSpace(r.Form.Get("healthcheck"))

		formView := containerFormView{
			Volumes: availableVolumes, Networks: availableNetworks,
			ExistingServices: existingServices, FormData: r.Form, Title: title, Editing: true,
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
			"volumes": mappedVolumes, "networks": []string{network},
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
	envVars := map[string]string{}
	if envRaw, ok := svc["environment"].(map[string]any); ok {
		for k := range envRaw {
			envVars[k] = ""
		}
	}
	var envLines []string
	for k := range envVars {
		envLines = append(envLines, k+": ")
	}

	env := LoadEnvFile(userContext)
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

	network := ""
	if netsRaw, ok := svc["networks"].([]any); ok && len(netsRaw) > 0 {
		network, _ = netsRaw[0].(string)
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
			Network: network, Healthcheck: healthcheck,
		},
	})
}

// handleDeleteContainer stops and removes a service and its image, and on
// GET shows a confirmation page first.
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
