package php

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterOptionsAPI wires the PHP options API route onto mux.
func RegisterOptionsAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "php_options", "GET /api/php/{version}/options", func(w http.ResponseWriter, r *http.Request) { apiPHPOptions(a, w, r) })
	apiregistry.Handle(mux, a, "php_options", "PUT /api/php/{version}/options", func(w http.ResponseWriter, r *http.Request) { apiPHPOptions(a, w, r) })
}

func apiPHPContainer(userContext, version string) string {
	ws := strings.ToLower(webserver.GetEnvFileValue(userContext, "WEB_SERVER"))
	if strings.Contains(ws, "litespeed") {
		return ws
	}
	return "php-fpm-" + version
}

// apiPHPOptions gets or updates the php.ini options for one PHP version.
func apiPHPOptions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	version := r.PathValue("version")
	availableKeys := loadKeysFromFile(userContext)

	if r.Method == http.MethodGet {
		content, readErr := os.ReadFile("/home/" + userContext + "/php.ini/" + version + ".ini")
		if readErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
			return
		}
		currentConfig := configEntriesToMap(parseConfigContent(string(content)))
		writeJSON(w, http.StatusOK, map[string]any{
			"version": version, "available_keys": availableKeys, "current_config": currentConfig,
		})
		return
	}

	// PUT
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	newConfig := map[string]string{}
	for _, key := range availableKeys {
		if v, ok := body[key]; ok && v != nil {
			newConfig[key] = strings.TrimSpace(toStringValue(v))
		}
	}
	if len(newConfig) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Provide at least one key from: " + strings.Join(availableKeys, ", ")})
		return
	}

	updatePHPConfigFile(ctx, userContext, version, availableKeys, newConfig)
	_ = logger.RecordUserAction(a.Config, currentUsername, "edited PHP "+version+" configuration using PHP Selector", reqip.ClientIP(r))
	container := apiPHPContainer(userContext, version)
	docker.StartComposeServiceIfNotRunning(ctx, userContext, container)

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Options updated and " + container + " restarted", "version": version, "updated": newConfig,
	})
}

func toStringValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}
