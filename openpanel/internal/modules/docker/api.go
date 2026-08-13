package docker

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterAPI wires the Docker REST endpoints onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "docker", "GET /api/containers", func(w http.ResponseWriter, r *http.Request) { apiContainersList(a, w, r) })
	apiregistry.Handle(mux, a, "docker", "GET /api/containers/status", func(w http.ResponseWriter, r *http.Request) { apiContainersStatus(a, w, r) })
	apiregistry.Handle(mux, a, "docker", "GET /api/containers/{service}/status", func(w http.ResponseWriter, r *http.Request) { apiContainerServiceStatus(a, w, r) })
	apiregistry.Handle(mux, a, "docker", "POST /api/containers/{service}/start", func(w http.ResponseWriter, r *http.Request) { apiContainerStart(a, w, r) })
	apiregistry.Handle(mux, a, "docker", "POST /api/containers/{service}/stop", func(w http.ResponseWriter, r *http.Request) { apiContainerStop(a, w, r) })
	apiregistry.Handle(mux, a, "docker", "POST /api/containers/{service}/restart", func(w http.ResponseWriter, r *http.Request) { apiContainerRestart(a, w, r) })
	apiregistry.Handle(mux, a, "docker", "PATCH /api/containers/{service}/resources", func(w http.ResponseWriter, r *http.Request) { apiContainerResources(a, w, r) })
	apiregistry.Handle(mux, a, "docker", "GET /api/containers/{service}/logs", func(w http.ResponseWriter, r *http.Request) { apiContainerLogs(a, w, r) })
}

func writeAPIDockerJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiWebserverGroups maps an active webserver to the set of sibling
// webserver service-name substrings to hide.
var apiWebserverGroups = map[string][]string{
	"apache":        {"nginx", "openresty", "openlitespeed", "litespeed"},
	"nginx":         {"apache", "openresty", "openlitespeed", "litespeed"},
	"openresty":     {"apache", "nginx", "openlitespeed", "litespeed"},
	"openlitespeed": {"apache", "openresty", "nginx", "litespeed"},
	"litespeed":     {"apache", "openresty", "nginx", "openlitespeed"},
}

// apiFilterServices is deliberately a separate, differently-shaped filter
// than services.FilterServices: that one hides only the exact alternate
// service name, while this one hides any service whose name merely
// contains an alternate webserver's substring, plus a MySQL/MariaDB
// substring exclusion the other filter doesn't have.
func apiFilterServices(services map[string]any, webserver, mysqlType string) map[string]any {
	hidden := apiWebserverGroups[webserver]
	result := map[string]any{}
	for k, v := range services {
		lk := strings.ToLower(k)
		keep := true
		for _, h := range hidden {
			if strings.Contains(lk, h) {
				keep = false
				break
			}
		}
		if keep {
			result[k] = v
		}
	}
	if mysqlType == "mysql" {
		for k := range result {
			if strings.Contains(strings.ToLower(k), "mariadb") {
				delete(result, k)
			}
		}
	} else if mysqlType == "mariadb" {
		for k := range result {
			if strings.Contains(strings.ToLower(k), "mysql") {
				delete(result, k)
			}
		}
	}
	return result
}

// apiContainersList serves the compose service list, filtered to hide
// inactive webserver/database alternatives.
func apiContainersList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	webserver, _ := GetEnvValue(userContext, "WEB_SERVER")
	mysqlType, _ := GetEnvValue(userContext, "MYSQL_TYPE")

	dockerData, loadErr := podmanmanager.LoadComposeConfig(ctx, userContext)
	if loadErr != nil {
		writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch container data: " + loadErr.Error()})
		return
	}

	if services, ok := dockerData["services"].(map[string]any); ok {
		dockerData["services"] = apiFilterServices(services, webserver, mysqlType)
	}

	writeAPIDockerJSON(w, http.StatusOK, dockerData)
}

// apiContainersStatus reports which containers are currently running.
func apiContainersStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, _, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	running, runErr := GetRunningContainers(r.Context(), currentUsername)
	if runErr != nil {
		writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not retrieve container status"})
		return
	}
	writeAPIDockerJSON(w, http.StatusOK, map[string]any{"running_containers": running})
}

// apiContainerServiceStatus reports the state and health of one service.
func apiContainerServiceStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	service := r.PathValue("service")
	status := GetContainerStatus(r.Context(), userContext, service)
	writeAPIDockerJSON(w, http.StatusOK, map[string]string{"service": service, "state": status.State, "health": status.Health})
}

// apiContainerStart starts (activates) a container, optionally pulling
// its image first.
func apiContainerStart(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	service := r.PathValue("service")

	var body struct {
		Pull bool `json:"pull"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	flag := ""
	if body.Pull {
		flag = "pull"
	}

	response := StartOrStopContainer(ctx, userContext, service, "activate", flag)
	_ = logger.RecordUserAction(a.Config, currentUsername, "started container "+service, reqip.ClientIP(r))
	writeAPIDockerJSON(w, http.StatusOK, map[string]any{"message": service + " started", "details": response})
}

// apiContainerStop stops (deactivates) a container.
func apiContainerStop(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	service := r.PathValue("service")

	response := StartOrStopContainer(ctx, userContext, service, "deactivate", "")
	_ = logger.RecordUserAction(a.Config, currentUsername, "stopped container "+service, reqip.ClientIP(r))
	writeAPIDockerJSON(w, http.StatusOK, map[string]any{"message": service + " stopped", "details": response})
}

// apiContainerRestart restarts a container.
func apiContainerRestart(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	service := r.PathValue("service")

	response := RestartContainer(ctx, userContext, service)
	if !response.Success {
		msg := response.Message
		if msg == "" {
			msg = "Failed to restart " + service
		}
		writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": msg})
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "restarted container "+service, reqip.ClientIP(r))
	writeAPIDockerJSON(w, http.StatusOK, map[string]string{"message": service + " restarted"})
}

// apiContainerResources updates a container's CPU and/or RAM limits.
func apiContainerResources(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	service := r.PathValue("service")

	userID, _ := auth.UserID(r)
	injected, injErr := a.InjectData(ctx, userID)
	if injErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	planID, _ := injected["hosting_plan"].(int)

	var body struct {
		CPU  string `json:"cpu"`
		RAM  string `json:"ram"`
		PIDs string `json:"pids"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.CPU = r.Form.Get("cpu")
		body.RAM = r.Form.Get("ram")
		body.PIDs = r.Form.Get("pids")
	}

	results := map[string]any{}
	for _, resource := range []struct{ name, value string }{{"cpu", strings.TrimSpace(body.CPU)}, {"ram", strings.TrimSpace(body.RAM)}} {
		if resource.value == "" {
			continue
		}
		message, success := updateContainerRAMOrCPU(a, ctx, userContext, planID, service, resource.name, resource.value)
		if !success {
			writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": message})
			return
		}
		results[resource.name] = message
		_ = logger.RecordUserAction(a.Config, currentUsername, "updated "+resource.name+" limit for "+service+" to "+resource.value, reqip.ClientIP(r))
	}

	if pids := strings.TrimSpace(body.PIDs); pids != "" {
		message, success := updateContainerPIDs(ctx, userContext, service, pids)
		if !success {
			writeAPIDockerJSON(w, http.StatusInternalServerError, map[string]string{"error": message})
			return
		}
		results["pids"] = message
		_ = logger.RecordUserAction(a.Config, currentUsername, "updated pids limit for "+service+" to "+pids, reqip.ClientIP(r))
	}

	if len(results) == 0 {
		writeAPIDockerJSON(w, http.StatusBadRequest, map[string]string{"error": "Provide cpu, ram and/or pids in request body"})
		return
	}
	writeAPIDockerJSON(w, http.StatusOK, map[string]any{"message": "Resources updated", "results": results})
}

// apiContainerLogs returns the tail of a container's logs as JSON entries.
func apiContainerLogs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := apiInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	service := r.PathValue("service")

	lines := 100
	if v := r.URL.Query().Get("lines"); v != "" {
		if n, convErr := strconv.Atoi(v); convErr == nil {
			lines = n
		}
	}

	body, status := FetchContainerLog(ctx, a, userContext, service, lines)
	if status != http.StatusOK {
		writeAPIDockerJSON(w, status, map[string]string{"error": "Docker error: " + body})
		return
	}

	rawLines := strings.Split(body, "\n")
	entries := make([]map[string]string, 0, len(rawLines))
	for _, line := range rawLines {
		if line == "" {
			continue
		}
		entries = append(entries, map[string]string{"log": line})
	}
	writeAPIDockerJSON(w, http.StatusOK, map[string]any{"service": service, "lines": lines, "entries": entries})
}

func apiInjected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}
