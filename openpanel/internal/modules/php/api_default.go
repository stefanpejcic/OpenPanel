package php

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterDefaultAPI wires the PHP default-version API route onto mux.
func RegisterDefaultAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "php", "GET /api/php/default", func(w http.ResponseWriter, r *http.Request) { apiPHPDefault(a, w, r) })
	apiregistry.Handle(mux, a, "php", "PUT /api/php/default", func(w http.ResponseWriter, r *http.Request) { apiPHPDefault(a, w, r) })
}

// apiPHPDefault gets or updates the server-wide default PHP version applied
// to newly created domains that don't have an explicit per-domain override.
func apiPHPDefault(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")

	if r.Method == http.MethodPut {
		var body struct {
			Version string `json:"version"`
		}
		if decErr := json.NewDecoder(r.Body).Decode(&body); decErr != nil {
			_ = r.ParseForm()
			body.Version = r.Form.Get("version")
		}
		newVersion := strings.TrimSpace(body.Version)
		if newVersion == "" {
			writeJSONError(w, http.StatusBadRequest, "version is required")
			return
		}

		if isLitespeed {
			versionFloat, parseErr := strconv.ParseFloat(newVersion, 64)
			if parseErr != nil || !litespeedDefaultVersions[versionFloat] {
				writeJSONError(w, http.StatusBadRequest, "Default PHP version for OpenLitespeed can not be set to "+newVersion+" - only available tags are: 8.5 8.4 8.3 8.2")
				return
			}
		}

		previousVersion, _ := computeDefaultPHPVersionAndService(ctx, userContext, webServer, isLitespeed)

		if !updatePHPVersionPreference(userContext, newVersion) {
			writeJSONError(w, http.StatusNotFound, "Configuration file not found")
			return
		}

		message := "PHP version " + newVersion + " set as default for new domains"
		if isLitespeed {
			if docker.ComposeContainer(ctx, userContext, webServer, "status") {
				message = "PHP version " + newVersion + " saved, and Litespeed restarted to apply the version."
				docker.ComposeContainer(ctx, userContext, webServer, "stop")
				docker.ComposeContainer(ctx, userContext, webServer, "start")
			} else {
				message = "PHP version " + newVersion + " saved and will be used by Litespeed for new domains."
			}
		} else if previousVersion != "" && previousVersion != newVersion {
			stopPHPServiceIfRunningAndUnused(ctx, userContext, previousVersion)
		}

		_ = logger.RecordUserAction(a.Config, currentUsername, "changed default PHP version for new domains to "+newVersion, reqip.ClientIP(r))
		writeJSON(w, http.StatusOK, map[string]any{"message": message, "version": newVersion})
		return
	}

	installedVersions := FetchPHPVersions(ctx, a, userContext)
	version, service := computeDefaultPHPVersionAndService(ctx, userContext, webServer, isLitespeed)
	writeJSON(w, http.StatusOK, map[string]any{
		"version": version, "service": service, "is_litespeed": isLitespeed, "installed_versions": installedVersions,
	})
}
