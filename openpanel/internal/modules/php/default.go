package php

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

var litespeedDefaultVersions = map[float64]bool{8.5: true, 8.4: true, 8.3: true, 8.2: true}

// handleDefaultPHPVersion gets or sets the default PHP version applied to
// newly created domains.
func handleDefaultPHPVersion(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		newPHPVersion := r.Form.Get("new_php_version")
		if newPHPVersion == "" {
			writeJSONError(w, http.StatusBadRequest, "Invalid input data")
			return
		}

		if isLitespeed {
			versionFloat, parseErr := strconv.ParseFloat(newPHPVersion, 64)
			if parseErr != nil {
				flashAndRedirect(a, w, r, "error", "Invalid PHP version format", "/php/default")
				return
			}
			if !litespeedDefaultVersions[versionFloat] {
				flashAndRedirect(a, w, r, "error", "Default PHP version for OpenLitespeed can not be set to "+newPHPVersion+" - only available tags are: 8.5 8.4 8.3 8.2", "/php/default")
				return
			}
		}

		if updatePHPVersionPreference(userContext, newPHPVersion) {
			ipAddress := reqip.ClientIP(r)
			_ = logger.RecordUserAction(a.Config, currentUsername, "changed default PHP version for new domains to "+newPHPVersion, ipAddress)
			message := "PHP version " + newPHPVersion + " set as default for new domains"

			if isLitespeed {
				if docker.ComposeContainer(ctx, userContext, webServer, "status") {
					message = "PHP version " + newPHPVersion + " saved, and Litespeed restarted to apply the version."
					docker.ComposeContainer(ctx, userContext, webServer, "stop")
					docker.ComposeContainer(ctx, userContext, webServer, "start")
				} else {
					message = "PHP version " + newPHPVersion + " saved and will be used by Litespeed for new domains."
				}
			} else if previousVersion := r.Form.Get("previous_version"); previousVersion != "" {
				stopPHPServiceIfRunningAndUnused(ctx, userContext, previousVersion)
			}

			flashSess(a, w, r, "success", message)
		} else {
			flashSess(a, w, r, "error", "Default PHP version could not be changed to "+newPHPVersion)
			writeJSONError(w, http.StatusNotFound, "Configuration file not found")
			return
		}
	}

	installedVersions := FetchPHPVersions(ctx, a, userContext)
	phpDefaultVersion, service := computeDefaultPHPVersionAndService(ctx, userContext, webServer, isLitespeed)

	renderDefaultPage(a, w, r, phpDefaultVersion, service, installedVersions)
}

// computeDefaultPHPVersionAndService resolves the current default PHP
// version and the service name that runs it, querying the LiteSpeed
// container directly when running under LiteSpeed since it has no
// per-domain php-fpm service.
func computeDefaultPHPVersionAndService(ctx context.Context, userContext, webServer string, isLitespeed bool) (version, service string) {
	if !isLitespeed {
		version = webserver.GetEnvFileValue(userContext, "DEFAULT_PHP_VERSION")
		return version, "php-fpm-" + version
	}

	service = webServer
	argv := podmanmanager.PodmanArgv(userContext, "exec", webServer, "php",
		"-d", "memory_limit=-1",
		"-d", "open_basedir=none",
		"-d", "disable_functions=",
		"-d", "display_errors=0",
		"-d", "error_log=/dev/null",
		"-r", "echo PHP_VERSION, PHP_EOL;")
	out, err := podmanmanager.Command(ctx, userContext, argv).Output()
	if err == nil {
		return strings.TrimSpace(string(out)), service
	}

	var tagKey string
	if webServer == "openlitespeed" {
		tagKey = "OPENLITESPEED_VERSION"
	} else if webServer == "litespeed" {
		tagKey = "LITESPEED_VERSION"
	}
	tag := webserver.GetEnvFileValue(userContext, tagKey)
	if tag == "latest" {
		return "8.5", service
	}
	if m := litespeedTagVersionRE.FindStringSubmatch(tag); m != nil {
		return m[1] + "." + m[2], service
	}
	return "8.5", service
}
