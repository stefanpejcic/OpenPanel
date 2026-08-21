package docker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterContainerManageAPI wires the docker-compose service management
// REST endpoints onto mux (create/edit/delete a compose service). Kept in
// its own file/Register function, separate from RegisterAPI in api.go,
// since it covers a distinct slice of docker.go's non-API routes
// (/containers/new, /containers/edit/{service}, /containers/delete/{service})
// rather than the status/start/stop/logs endpoints RegisterAPI already
// covers. Gated behind the same "docker" feature flag as RegisterAPI.
//
// The switch-MySQL/webserver and change-image-tag API twins live in
// RegisterChangeDBAPI/RegisterChangeWSAPI/RegisterChangeImageAPI instead,
// gated behind those routes' own feature flags rather than "docker".
func RegisterContainerManageAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "docker", "POST /api/containers", func(w http.ResponseWriter, r *http.Request) { apiContainerCreate(a, w, r) })
	apiregistry.Handle(mux, a, "docker", "PATCH /api/containers/{service}", func(w http.ResponseWriter, r *http.Request) { apiContainerEdit(a, w, r) })
	apiregistry.Handle(mux, a, "docker", "DELETE /api/containers/{service}", func(w http.ResponseWriter, r *http.Request) { apiContainerDelete(a, w, r) })
}

// RegisterChangeDBAPI wires the MySQL/MariaDB-swap REST endpoint onto mux,
// gated behind its own "change_db" feature flag (same flag as
// docker.RegisterChangeDB's web route).
func RegisterChangeDBAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "change_db", "POST /api/containers/mysql", func(w http.ResponseWriter, r *http.Request) { apiContainerSwitchMySQL(a, w, r) })
}

// RegisterChangeWSAPI wires the webserver-swap REST endpoint onto mux,
// gated behind its own "change_ws" feature flag (same flag as
// docker.RegisterChangeWS's web route).
func RegisterChangeWSAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "change_ws", "POST /api/containers/webserver", func(w http.ResponseWriter, r *http.Request) { apiContainerSwitchWebserver(a, w, r) })
}

// RegisterChangeImageAPI wires the change-image-tag REST endpoint onto mux,
// gated behind its own "change_image" feature flag (same flag as
// docker.RegisterChangeImage's web route).
func RegisterChangeImageAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "change_image", "PATCH /api/containers/{service}/image", func(w http.ResponseWriter, r *http.Request) { apiContainerChangeImage(a, w, r) })
}

// containerVolumeInput is one entry of the JSON-body "volumes" array for
// create/edit, mirroring the volume_name[]/volume_mount[]/volume_readonly[]
// parallel form fields the HTML page posts.
type containerVolumeInput struct {
	Name     string `json:"name"`
	Mount    string `json:"mount"`
	ReadOnly bool   `json:"readonly"`
}

// containerServiceBody is the shared request-body shape for creating and
// editing a compose service.
type containerServiceBody struct {
	ServiceName string                 `json:"service_name"`
	Image       string                 `json:"image"`
	CPU         string                 `json:"cpu"`
	RAM         string                 `json:"ram"`
	Network     string                 `json:"network"`
	Healthcheck string                 `json:"healthcheck"`
	Environment string                 `json:"environment"`
	AddSocket   bool                   `json:"add_socket"`
	Volumes     []containerVolumeInput `json:"volumes"`
}

// decodeContainerServiceBody decodes a JSON containerServiceBody, falling
// back to form fields (matching the HTML add/edit-service forms) if the
// body isn't valid JSON. It also returns the volume fields as the parallel
// slices ParseVolumeEntries expects.
func decodeContainerServiceBody(r *http.Request) (body containerServiceBody, volNames, volMounts, volReadonly []string) {
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr == nil {
		for _, v := range body.Volumes {
			volNames = append(volNames, v.Name)
			volMounts = append(volMounts, v.Mount)
			readonly := ""
			if v.ReadOnly {
				readonly = "on"
			}
			volReadonly = append(volReadonly, readonly)
		}
		return body, volNames, volMounts, volReadonly
	}
	_ = r.ParseForm()
	body.ServiceName = strings.TrimSpace(r.Form.Get("service_name"))
	body.Image = strings.TrimSpace(r.Form.Get("image"))
	body.CPU = strings.TrimSpace(r.Form.Get("cpu"))
	body.RAM = strings.TrimSpace(r.Form.Get("ram"))
	body.Network = strings.TrimSpace(r.Form.Get("network"))
	body.Healthcheck = strings.TrimSpace(r.Form.Get("healthcheck"))
	body.Environment = r.Form.Get("environment")
	body.AddSocket = r.Form.Get("add_socket") != ""
	return body, r.Form["volume_name"], r.Form["volume_mount"], r.Form["volume_readonly"]
}

// apiContainerCreate adds a new service to docker-compose.yml. Mirrors
// handleAddContainer's POST branch (containers.go) with JSON responses
// instead of the flash+redirect/HTML form the web handler uses.
func apiContainerCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	composeData, _ := LoadCompose(userContext)
	existingServices := serviceNames(composeData)

	body, volNames, volMounts, volReadonly := decodeContainerServiceBody(r)
	serviceName := strings.TrimSpace(body.ServiceName)
	image := strings.TrimSpace(body.Image)
	cpu := strings.TrimSpace(body.CPU)
	ram := strings.TrimSpace(body.RAM)
	network := strings.TrimSpace(body.Network)
	healthcheck := strings.TrimSpace(body.Healthcheck)

	if serviceName == "" || !IsValidServiceName(serviceName) {
		writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid service name. Must start with a letter, contain only lowercase letters and digits, and be at least 3 characters long."})
		return
	}
	if containsString(existingServices, serviceName) {
		writeAPIDockerJSON(w, http.StatusConflict, map[string]string{"error": fmt.Sprintf("Service '%s' already exists in docker-compose.yml.", serviceName)})
		return
	}
	if !IsValidCPULimit(cpu) {
		writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "CPU limit must be a positive number."})
		return
	}
	if !IsValidRAMLimit(ram) {
		writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "Memory limit must be a positive number followed by 'M' or 'G' (e.g., 512M or 1.5G)."})
		return
	}

	servicePrefix := ServiceKeyPrefix(serviceName)
	environmentVars, newEnvVars := ParseEnvVars(strings.Split(body.Environment, "\n"), servicePrefix)

	cpuKey, ramKey := servicePrefix+"_CPU", servicePrefix+"_RAM"
	newEnvVars[cpuKey] = cpu
	newEnvVars[ramKey] = ram
	UpdateEnvFileWithVars(userContext, newEnvVars)

	mappedVolumes := ParseVolumeEntries(volNames, volMounts, volReadonly, body.AddSocket)

	newService := map[string]any{
		"image": image, "container_name": serviceName, "restart": "always",
		"volumes": mappedVolumes,
		"deploy": map[string]any{"resources": map[string]any{"limits": map[string]any{
			"cpus": "${" + cpuKey + ":-" + cpu + "}", "memory": "${" + ramKey + ":-" + ram + "}", "pids": 100,
		}}},
		"networks": []string{network},
	}
	if len(environmentVars) > 0 {
		newService["environment"] = environmentVars
	}
	if healthcheck != "" {
		hc, hcErr := parseYAMLString(healthcheck)
		if hcErr != nil {
			writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid healthcheck YAML: " + hcErr.Error()})
			return
		}
		newService["healthcheck"] = hc
	}

	services := servicesRaw(composeData)
	services[serviceName] = newService
	composeData["services"] = services
	if saveErr := SaveCompose(userContext, composeData); saveErr != nil {
		writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save configuration: " + saveErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "added container "+serviceName, reqip.ClientIP(r))
	writeAPIDockerJSON(w, http.StatusCreated, map[string]any{"message": fmt.Sprintf("Container %s created successfully!", serviceName), "service": serviceName})
}

// apiContainerEdit updates an existing service's compose definition.
// Mirrors handleEditContainer's POST branch (containers.go). The
// service_name body field, if present, must match the {service} path
// value - service renaming isn't supported, same as the web form.
func apiContainerEdit(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	service := r.PathValue("service")

	currentUsername, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	composeData, _ := LoadCompose(userContext)
	services := servicesRaw(composeData)
	if _, ok := services[service]; !ok {
		writeAPIDockerJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Service '%s' not found.", service)})
		return
	}

	body, volNames, volMounts, volReadonly := decodeContainerServiceBody(r)
	serviceName := strings.TrimSpace(body.ServiceName)
	if serviceName == "" {
		serviceName = service
	}
	image := strings.TrimSpace(body.Image)
	cpu := strings.TrimSpace(body.CPU)
	ram := strings.TrimSpace(body.RAM)
	network := body.Network
	healthcheck := strings.TrimSpace(body.Healthcheck)

	if serviceName != service {
		writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "Service name cannot be changed."})
		return
	}
	if !IsValidCPULimit(cpu) {
		writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "CPU limit must be a positive number."})
		return
	}
	if !IsValidRAMLimit(ram) {
		writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "Memory limit must be a positive number followed by 'M' or 'G' (e.g., 512M or 1.5G)."})
		return
	}

	servicePrefix := ServiceKeyPrefix(serviceName)
	environmentVars, newEnvVars := ParseEnvVars(strings.Split(body.Environment, "\n"), servicePrefix)

	cpuKey, ramKey := servicePrefix+"_CPU", servicePrefix+"_RAM"
	newEnvVars[cpuKey] = cpu
	newEnvVars[ramKey] = ram

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

	mappedVolumes := ParseVolumeEntries(volNames, volMounts, volReadonly, body.AddSocket)

	updatedService := map[string]any{
		"image": image, "container_name": serviceName, "restart": "always",
		"volumes": mappedVolumes, "networks": []string{network},
		"deploy": map[string]any{"resources": map[string]any{"limits": map[string]any{
			"cpus": "${" + cpuKey + "}", "memory": "${" + ramKey + "}", "pids": 100,
		}}},
	}
	if len(environmentVars) > 0 {
		updatedService["environment"] = environmentVars
	}
	if healthcheck != "" {
		hc, hcErr := parseYAMLString(healthcheck)
		if hcErr != nil {
			writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid healthcheck YAML: " + hcErr.Error()})
			return
		}
		updatedService["healthcheck"] = hc
	}

	services[service] = updatedService
	composeData["services"] = services
	if saveErr := SaveCompose(userContext, composeData); saveErr != nil {
		writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save configuration: " + saveErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "edited container "+service, reqip.ClientIP(r))
	writeAPIDockerJSON(w, http.StatusOK, map[string]any{"message": "Container updated successfully.", "service": service})
}

// apiContainerDelete stops and removes a service, its image, and its
// per-service env vars. Mirrors handleDeleteContainer's POST branch
// (containers.go) - the API has no confirmation step, matching the rest
// of this REST surface (DELETE itself is the confirmation).
func apiContainerDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	service := r.PathValue("service")

	currentUsername, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if UndeletableServices[service] || strings.HasPrefix(service, "php-fpm-") {
		writeAPIDockerJSON(w, http.StatusForbidden, map[string]string{"error": fmt.Sprintf("Service '%s' cannot be deleted.", service)})
		return
	}

	composeData, _ := LoadCompose(userContext)
	services := servicesRaw(composeData)
	svcRaw, ok := services[service]
	if !ok {
		writeAPIDockerJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Service '%s' not found.", service)})
		return
	}

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

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted container "+service, reqip.ClientIP(r))
	writeAPIDockerJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Service '%s' deleted successfully.", service)})
}

// apiContainerSwitchMySQL swaps between mysql/mariadb, wiping the old data
// volume. Mirrors handleContainersMySQL's POST branch (switch.go).
func apiContainerSwitchMySQL(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	mysqlType, _ := GetEnvValue(userContext, "MYSQL_TYPE")
	if IsServiceRunning(ctx, userContext, mysqlType) {
		writeAPIDockerJSON(w, http.StatusConflict, map[string]string{"error": fmt.Sprintf("Existing databases must first be deleted and %s container stopped in order to change mysql type.", mysqlType)})
		return
	}

	var body struct {
		NewSQL string `json:"new_sql"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.NewSQL = r.Form.Get("new_sql")
	}
	newSQL := body.NewSQL
	if newSQL != "mysql" && newSQL != "mariadb" {
		writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid mysql server selected"})
		return
	}

	stopResp := StartOrStopContainer(ctx, userContext, mysqlType, "deactivate", "")
	if !stopResp.Success {
		writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": stopResp.Message})
		return
	}

	var warnings []string
	if stopResp.Message != "" {
		warnings = append(warnings, stopResp.Message)
	}

	deleteDockerVolume(ctx, userContext, userContext+"_mysql_data")
	StartOrStopContainer(ctx, userContext, "phpmyadmin", "deactivate", "")
	startResp := StartOrStopContainer(ctx, userContext, newSQL, "activate", "")
	if !startResp.Success {
		writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]any{"error": fmt.Sprintf("Failed to start %s.", newSQL), "warnings": warnings})
		return
	}

	SetEnvValue(userContext, "MYSQL_TYPE", newSQL)
	removeImage(ctx, userContext, mysqlType)
	mysqlmanager.InvalidatePool(userContext)

	_ = logger.RecordUserAction(a.Config, currentUsername, "switched mysql type to: "+newSQL, reqip.ClientIP(r))
	writeAPIDockerJSON(w, http.StatusOK, map[string]any{"message": fmt.Sprintf("Successfully switched to %s!", newSQL), "warnings": warnings})
}

// apiContainerSwitchWebserver swaps the active webserver container.
// Mirrors handleContainersWebserver's POST branch (switch.go).
func apiContainerSwitchWebserver(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webserver, _ := GetEnvValue(userContext, "WEB_SERVER")
	var available []string
	for _, ws := range webserverOptions {
		if ws != webserver {
			available = append(available, ws)
		}
	}

	userDomains, domainsErr := a.AllDomainsForUser(ctx, userID)
	if domainsErr == nil && len(userDomains) > 0 {
		writeAPIDockerJSON(w, http.StatusConflict, map[string]string{"error": fmt.Sprintf("Existing domains (%d) must first be removed in order to change webserver.", len(userDomains))})
		return
	}

	var body struct {
		NewWebserver string `json:"new_ws"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.NewWebserver = r.Form.Get("new_ws")
	}
	newWebserver := body.NewWebserver
	if newWebserver == "" || !containsString(available, newWebserver) {
		writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid web server selected"})
		return
	}

	stopResp := StartOrStopContainer(ctx, userContext, webserver, "deactivate", "")
	if !stopResp.Success {
		writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": stopResp.Message})
		return
	}

	deleteDockerVolume(ctx, userContext, userContext+"_webserver_data")
	startResp := StartOrStopContainer(ctx, userContext, newWebserver, "activate", "")
	if !startResp.Success {
		writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to start %s.", newWebserver)})
		return
	}

	SetEnvValue(userContext, "WEB_SERVER", newWebserver)
	removeImage(ctx, userContext, webserver)
	_ = logger.RecordUserAction(a.Config, currentUsername, "switched webserver type to: "+newWebserver, reqip.ClientIP(r))
	writeAPIDockerJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Successfully switched to %s!", newWebserver)})
}

// apiContainerChangeImage changes a service's image tag/version (stored as
// the SERVICE_VERSION env var), stopping the container first so its old
// image can later be pruned. Mirrors handleContainersChangeImage's POST
// branch (images.go).
func apiContainerChangeImage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	service := r.PathValue("service")

	currentUsername, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		NewTag string `json:"new_tag"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.NewTag = r.Form.Get("new_tag")
	}
	value := strings.TrimSpace(body.NewTag)
	if value == "" {
		writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "new_tag is required"})
		return
	}

	result := StartOrStopContainer(ctx, userContext, service, "deactivate", "")
	if !result.Success {
		writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to stop the service in order to delete old image."})
		return
	}

	SetEnvValue(userContext, service+"_VERSION", value)
	_ = logger.RecordUserAction(a.Config, currentUsername, fmt.Sprintf("changed image tag for %s to %s", service, value), reqip.ClientIP(r))
	writeAPIDockerJSON(w, http.StatusOK, map[string]string{"message": fmt.Sprintf("Successfully changed image tag for %s to %s!", service, value), "service": service})
}
