package php

import (
	"context"
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

// RegisterIniAPI wires the PHP ini API routes onto mux.
func RegisterIniAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "php_ini", "GET /api/php/versions", func(w http.ResponseWriter, r *http.Request) { apiPHPVersions(a, w, r) })
	apiregistry.Handle(mux, a, "php_ini", "GET /api/php/{version}/ini", func(w http.ResponseWriter, r *http.Request) { apiPHPIni(a, w, r) })
	apiregistry.Handle(mux, a, "php_ini", "PUT /api/php/{version}/ini", func(w http.ResponseWriter, r *http.Request) { apiPHPIni(a, w, r) })
}

func apiIniPath(userContext, version string) string {
	return "/home/" + userContext + "/php.ini/" + version + ".ini"
}

// apiRestartPHP restarts the PHP container for a version (or the LiteSpeed
// container, if that's what's serving PHP), if it's currently running.
func apiRestartPHP(ctx context.Context, userContext, version string) string {
	ws := strings.ToLower(webserver.GetEnvFileValue(userContext, "WEB_SERVER"))
	container := "php-fpm-" + version
	if strings.Contains(ws, "litespeed") {
		container = ws
	}
	if docker.ComposeContainer(ctx, userContext, container, "status") {
		docker.ComposeContainer(ctx, userContext, container, "stop")
		docker.ComposeContainer(ctx, userContext, container, "start")
	}
	return container
}

// apiPHPVersions returns the list of installed PHP versions.
func apiPHPVersions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	versions := FetchPHPVersions(r.Context(), a, userContext)
	writeJSON(w, http.StatusOK, map[string]any{"versions": versions})
}

// apiPHPIni gets or updates the raw php.ini content for one PHP version.
func apiPHPIni(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	version := r.PathValue("version")
	iniPath := apiIniPath(userContext, version)

	if r.Method == http.MethodGet {
		content, readErr := os.ReadFile(iniPath)
		if readErr != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "php.ini not found for version " + version})
			return
		}
		syntaxError := checkPHPIniSyntax(ctx, userContext, version)
		writeJSON(w, http.StatusOK, map[string]any{"version": version, "content": string(content), "syntax_error": syntaxError})
		return
	}

	// PUT
	var body struct {
		Content *string `json:"content"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Content == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}

	if _, statErr := os.Stat(iniPath); statErr != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "php.ini not found for version " + version})
		return
	}

	if writeErr := os.WriteFile(iniPath, []byte(*body.Content), 0o644); writeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error writing file: " + writeErr.Error()})
		return
	}

	container := apiRestartPHP(ctx, userContext, version)
	_ = logger.RecordUserAction(a.Config, currentUsername, "edited php.ini for PHP "+version, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]any{"message": "php.ini updated and " + container + " restarted", "version": version})
}
