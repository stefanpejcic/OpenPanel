package services

import (
	"encoding/json"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterAPI wires the services API routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "services", "GET /api/services", func(w http.ResponseWriter, r *http.Request) { apiServicesList(a, w, r) })
	apiregistry.Handle(mux, a, "services", "GET /api/services/{service}", func(w http.ResponseWriter, r *http.Request) { apiServiceGet(a, w, r) })
	apiregistry.Handle(mux, a, "services", "POST /api/services/{service}", func(w http.ResponseWriter, r *http.Request) { apiServiceAction(a, w, r) })
}

func writeAPIServicesJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiAllowedServices is the compose-load + FilterServices prelude shared by
// all three routes. Returns ok=false once it has written an error response.
func apiAllowedServices(w http.ResponseWriter, userContext string) (allowed []string, ok bool) {
	composeData, err := docker.LoadCompose(userContext)
	if err != nil {
		writeAPIServicesJSON(w, http.StatusNotFound, map[string]string{"error": "docker-compose.yml not found"})
		return nil, false
	}
	allServices := composeServiceNames(composeData)
	webserver, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
	mysqlType, _ := docker.GetEnvValue(userContext, "MYSQL_TYPE")
	return FilterServices(allServices, webserver, mysqlType), true
}

// apiServicesList returns every allowed service's current container status.
func apiServicesList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)

	allowed, ok := apiAllowedServices(w, userContext)
	if !ok {
		return
	}

	type serviceEntry struct {
		Service        string `json:"service"`
		ContainerState string `json:"container_state"`
		HealthStatus   string `json:"health_status"`
	}
	list := make([]serviceEntry, 0, len(allowed))
	for _, service := range allowed {
		status := docker.GetContainerStatus(ctx, userContext, service)
		list = append(list, serviceEntry{Service: service, ContainerState: status.State, HealthStatus: status.Health})
	}
	writeAPIServicesJSON(w, http.StatusOK, map[string]any{"services": list})
}

// apiServiceGet returns one service's current container status and the
// actions available on it.
func apiServiceGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)
	service := r.PathValue("service")

	allowed, ok := apiAllowedServices(w, userContext)
	if !ok {
		return
	}
	if !containsString(allowed, service) {
		writeAPIServicesJSON(w, http.StatusNotFound, map[string]string{"error": "Service '" + service + "' not found or not allowed"})
		return
	}

	status := docker.GetContainerStatus(ctx, userContext, service)
	actions := []string{"enable", "restart"}
	if status.State == "running" {
		actions = []string{"disable"}
	}
	writeAPIServicesJSON(w, http.StatusOK, map[string]any{
		"service": service, "container_state": status.State, "health_status": status.Health,
		"available_actions": actions,
	})
}

// apiServiceAction applies an enable/disable/restart action to one service.
func apiServiceAction(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)
	currentUsername, _ := injected["current_username"].(string)
	service := r.PathValue("service")

	allowed, ok := apiAllowedServices(w, userContext)
	if !ok {
		return
	}
	if !containsString(allowed, service) {
		writeAPIServicesJSON(w, http.StatusNotFound, map[string]string{"error": "Service '" + service + "' not found or not allowed"})
		return
	}

	var body struct {
		Action string `json:"action"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	action := strings.ToLower(strings.TrimSpace(body.Action))

	if action != "enable" && action != "disable" && action != "restart" {
		writeAPIServicesJSON(w, http.StatusBadRequest, map[string]string{"error": "action must be 'enable', 'disable', or 'restart'"})
		return
	}

	var success bool
	if action == "restart" {
		docker.StartOrStopContainer(ctx, userContext, service, "deactivate", "")
		success = docker.StartOrStopContainer(ctx, userContext, service, "activate", "").Success
	} else {
		dockerAction := "deactivate"
		if action == "enable" {
			dockerAction = "activate"
		}
		success = docker.StartOrStopContainer(ctx, userContext, service, dockerAction, "").Success
	}

	if !success {
		writeAPIServicesJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to " + action + " service " + service})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, action+"d service "+service, reqip.ClientIP(r))
	status := docker.GetContainerStatus(ctx, userContext, service)
	writeAPIServicesJSON(w, http.StatusOK, map[string]any{
		"message": "Service " + service + " " + action + "d successfully",
		"service": service, "container_state": status.State, "health_status": status.Health,
	})
}
