package php

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterExtensionsAPI wires the PHP extensions API routes onto mux.
func RegisterExtensionsAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "php_extensions", "GET /api/php/{version}/extensions", func(w http.ResponseWriter, r *http.Request) { apiPHPExtensionsList(a, w, r) })
	apiregistry.Handle(mux, a, "php_extensions", "POST /api/php/{version}/extensions", func(w http.ResponseWriter, r *http.Request) { apiPHPExtensionToggle(a, w, r) })
	apiregistry.Handle(mux, a, "php_extensions", "GET /api/php/{version}/extensions/available", func(w http.ResponseWriter, r *http.Request) { apiPHPExtensionsAvailable(a, w, r) })
	apiregistry.Handle(mux, a, "php_extensions", "POST /api/php/{version}/extensions/install", func(w http.ResponseWriter, r *http.Request) { apiPHPExtensionsInstall(a, w, r) })
	apiregistry.Handle(mux, a, "php_extensions", "GET /api/php/{version}/extensions/install/status", func(w http.ResponseWriter, r *http.Request) { apiPHPExtensionsInstallStatus(a, w, r) })
}

// apiCheckLitespeed returns true (and writes the 400 JSON error) if
// extension management isn't available for this webserver.
func apiCheckLitespeed(w http.ResponseWriter, userContext string) bool {
	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	if strings.Contains(strings.ToLower(webServer), "litespeed") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "PHP extension management not available for LiteSpeed"})
		return true
	}
	return false
}

// apiPHPExtensionsList returns the supported extensions for a PHP version
// along with their active/disabled/not-installed state and install history.
func apiPHPExtensionsList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if apiCheckLitespeed(w, userContext) {
		return
	}

	version := r.PathValue("version")
	service := "php-fpm-" + version
	_, _ = ensurePHPServiceRunning(ctx, userContext, service)

	active, disabled := getActiveAndDisabledExtensions(ctx, userContext, service)
	supportedNames := extensionsSupportedForVersion(ctx, version)

	extensions := make([]ExtensionRow, 0, len(supportedNames))
	for _, name := range supportedNames {
		lname := strings.ToLower(name)
		state := "not_installed"
		if active[lname] {
			state = "active"
		} else if disabled[lname] {
			state = "disabled"
		}
		extensions = append(extensions, ExtensionRow{Name: name, State: state})
	}

	history := loadExtensionsHistory(userContext, version)
	writeJSON(w, http.StatusOK, map[string]any{
		"version": version, "service": service, "extensions": extensions, "history": history,
	})
}

// apiPHPExtensionToggle enables or disables one already-installed PHP
// extension and restarts the PHP service to apply the change.
func apiPHPExtensionToggle(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if apiCheckLitespeed(w, userContext) {
		return
	}

	var body struct {
		Extension string `json:"extension"`
		Enable    any    `json:"enable"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		_ = r.ParseForm()
		body.Extension = r.Form.Get("extension")
		body.Enable = r.Form.Get("enable")
	}
	extension := strings.TrimSpace(body.Extension)
	enable := true
	switch v := body.Enable.(type) {
	case nil:
		enable = true
	case string:
		lv := strings.ToLower(v)
		enable = lv == "" || lv == "1" || lv == "true" || lv == "yes"
	case bool:
		enable = v
	case float64:
		enable = v != 0
	}

	if !extensionNameRE.MatchString(extension) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid extension name"})
		return
	}

	version := r.PathValue("version")
	service := "php-fpm-" + version
	if ok, errMsg := ensurePHPServiceRunning(ctx, userContext, service); !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to start PHP " + version + ": " + errMsg})
		return
	}

	ok, errMsg := toggleExtension(ctx, userContext, service, extension, enable)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not change " + extension + ": " + errMsg})
		return
	}

	docker.ComposeContainer(ctx, userContext, service, "restart")

	action := "disabled"
	if enable {
		action = "enabled"
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, action+" PHP extension "+extension+" for PHP "+version, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]any{
		"message":   "Extension " + extension + " " + action + ", PHP " + version + " restarted",
		"extension": extension, "enabled": enable,
	})
}

// apiPHPExtensionsAvailable lists all extensions supported for a PHP
// version, flagging which ones are already installed.
func apiPHPExtensionsAvailable(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if apiCheckLitespeed(w, userContext) {
		return
	}

	version := r.PathValue("version")
	service := "php-fpm-" + version
	active, disabled := getActiveAndDisabledExtensions(ctx, userContext, service)
	supportedNames := extensionsSupportedForVersion(ctx, version)

	type availableExt struct {
		Name      string `json:"name"`
		Installed bool   `json:"installed"`
	}
	extensions := make([]availableExt, 0, len(supportedNames))
	for _, name := range supportedNames {
		lname := strings.ToLower(name)
		extensions = append(extensions, availableExt{Name: name, Installed: active[lname] || disabled[lname]})
	}

	writeJSON(w, http.StatusOK, map[string]any{"version": version, "extensions": extensions})
}

var apiExtensionsCommaSplitRE = regexp.MustCompile(`\s*,\s*`)

// apiPHPExtensionsInstall queues an asynchronous install of one or more
// PHP extensions and returns an install ID for polling progress.
func apiPHPExtensionsInstall(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if apiCheckLitespeed(w, userContext) {
		return
	}

	var body struct {
		Extensions any `json:"extensions"`
	}
	var raw []string
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
		switch v := body.Extensions.(type) {
		case string:
			raw = apiExtensionsCommaSplitRE.Split(strings.TrimSpace(v), -1)
		case []any:
			for _, e := range v {
				if s, ok := e.(string); ok {
					raw = append(raw, s)
				}
			}
		}
	} else {
		_ = r.ParseMultipartForm(1 << 20)
		raw = r.Form["extensions[]"]
		if len(raw) == 0 {
			raw = r.Form["extensions"]
		}
	}

	var extensions []string
	for _, e := range raw {
		e = strings.TrimSpace(e)
		if extensionNameRE.MatchString(e) {
			extensions = append(extensions, e)
		}
	}
	if len(extensions) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No valid extensions provided"})
		return
	}

	version := r.PathValue("version")
	service := "php-fpm-" + version
	if ok, errMsg := ensurePHPServiceRunning(ctx, userContext, service); !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to start PHP " + version + ": " + errMsg})
		return
	}

	installID := uuid.NewString()
	info := installState{
		Status: "queued", Message: "Install queued...", Extensions: extensions,
		Version: version, Context: userContext, CurrentUsername: currentUsername, Service: service,
	}
	if err := saveInstallState(installID, info); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	go runExtensionInstall(installID)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"install_id": installID, "extensions": extensions, "message": "Install started",
	})
}

// apiPHPExtensionsInstallStatus reports the progress of a queued extension
// install, logging the user action once it completes.
func apiPHPExtensionsInstallStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	version := r.PathValue("version")
	service := "php-fpm-" + version
	installID := r.URL.Query().Get("install_id")

	containerBusy := isInstallRunningInContainer(ctx, userContext, service)
	info, ok := loadInstallState(installID)

	if ok {
		if info.Status == "done" && !info.Logged {
			ipAddress := reqip.ClientIP(r)
			_ = logger.RecordUserAction(a.Config, info.CurrentUsername, "installed PHP extensions "+strings.Join(info.Extensions, ", ")+" for PHP "+info.Version, ipAddress)
			info.Logged = true
			_ = saveInstallState(installID, info)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status": info.Status, "message": info.Message, "extensions": info.Extensions, "container_busy": containerBusy,
		})
		return
	}

	status := "idle"
	if containerBusy {
		status = "busy"
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": status, "container_busy": containerBusy})
}
