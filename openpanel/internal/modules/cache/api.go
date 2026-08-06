// Package cache (this file) implements the JSON status/action endpoints
// for the five generic single-container cache services.
package cache

import (
	"encoding/json"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

var validCacheActions = map[string]bool{"enable": true, "disable": true, "restart": true}

// RegisterRedisAPI, RegisterMemcachedAPI, RegisterElasticsearchAPI,
// RegisterOpensearchAPI and RegisterValkeyAPI each wire up one generic
// cache service's GET+POST pair.
func RegisterRedisAPI(mux *http.ServeMux, a *appctx.App)     { registerCacheAPI(mux, a, redisDef) }
func RegisterMemcachedAPI(mux *http.ServeMux, a *appctx.App) { registerCacheAPI(mux, a, memcachedDef) }
func RegisterElasticsearchAPI(mux *http.ServeMux, a *appctx.App) {
	registerCacheAPI(mux, a, elasticsearchDef)
}
func RegisterOpensearchAPI(mux *http.ServeMux, a *appctx.App) {
	registerCacheAPI(mux, a, opensearchDef)
}
func RegisterValkeyAPI(mux *http.ServeMux, a *appctx.App) { registerCacheAPI(mux, a, valkeyDef) }

func registerCacheAPI(mux *http.ServeMux, a *appctx.App, def serviceDef) {
	path := "/api/cache/" + def.Name
	apiregistry.Handle(mux, a, def.Name, "GET "+path, func(w http.ResponseWriter, r *http.Request) {
		apiCacheStatus(a, w, r, def)
	})
	apiregistry.Handle(mux, a, def.Name, "POST "+path, func(w http.ResponseWriter, r *http.Request) {
		apiCacheAction(a, w, r, def)
	})
}

// apiCacheStatus returns a generic cache service's container state, health,
// and the actions currently available for it.
func apiCacheStatus(a *appctx.App, w http.ResponseWriter, r *http.Request, def serviceDef) {
	_, userContext, err := cacheInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	status := docker.GetContainerStatus(r.Context(), userContext, def.Name)
	actions := []string{"enable", "restart"}
	if status.State == "running" {
		actions = []string{"disable"}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"service": def.Name, "port": def.Port, "description": def.Description,
		"container_state": status.State, "health_status": status.Health, "actions": actions,
	})
}

// apiCacheAction enables, disables, or restarts a generic cache service.
func apiCacheAction(a *appctx.App, w http.ResponseWriter, r *http.Request, def serviceDef) {
	currentUsername, userContext, err := cacheInjected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var body struct {
		Action string `json:"action"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	action := strings.ToLower(strings.TrimSpace(body.Action))
	if !validCacheActions[action] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid action. Valid: disable, enable, restart"})
		return
	}

	ctx := r.Context()
	ip := reqip.ClientIP(r)

	if action == "restart" {
		docker.RestartContainer(ctx, userContext, def.Name)
		_ = logger.RecordUserAction(a.Config, currentUsername, "restarted service "+def.Name, ip)
		writeJSON(w, http.StatusOK, map[string]string{"message": def.Name + " restarted"})
		return
	}

	dockerAction := "deactivate"
	if action == "enable" {
		dockerAction = "activate"
	}
	docker.StartOrStopContainer(ctx, userContext, def.Name, dockerAction, "")
	_ = logger.RecordUserAction(a.Config, currentUsername, action+"d service "+def.Name, ip)
	writeJSON(w, http.StatusOK, map[string]string{"message": def.Name + " " + action + "d"})
}
