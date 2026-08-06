package php

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// handlePHPInfo renders `php -i` output for one PHP version, starting its
// container first if needed.
func handlePHPInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	version := phpVersionFromSegment(r.PathValue("phpversion"))
	title := "PHP INFO"
	var phpInfoContent string
	installedVersions := FetchPHPVersions(ctx, a, userContext)
	service := ""

	if version != "" {
		webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
		if strings.Contains(strings.ToLower(webServer), "litespeed") {
			service = strings.ToLower(webServer)
		} else {
			service = "php-fpm-" + version
		}
		title = "PHP " + version + " Info"

		if !docker.IsServiceRunning(ctx, userContext, service) {
			result := docker.StartOrStopContainer(ctx, userContext, service, "activate", "run")
			if !result.Success || !docker.IsServiceRunning(ctx, userContext, service) {
				flashAndRedirect(a, w, r, "error", fmt.Sprintf("Failed to start PHP %s: %s", service, result.Message), "/php/domains")
				return
			}
		}

		argv := podmanmanager.PodmanArgv(userContext, "exec", service, "php",
			"-d", "memory_limit=-1",
			"-d", "open_basedir=none",
			"-d", "disable_functions=",
			"-d", "display_errors=0",
			"-d", "error_log=/dev/null",
			"-i")
		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		cmd := podmanmanager.Command(cctx, userContext, argv)
		var stderr strings.Builder
		cmd.Stderr = &stderr
		out, runErr := cmd.Output()
		if runErr != nil {
			msg := stderr.String()
			if msg == "" {
				if ee, ok := runErr.(*exec.ExitError); ok {
					msg = string(ee.Stderr)
				} else {
					msg = runErr.Error()
				}
			}
			writeJSONError(w, http.StatusInternalServerError, strings.TrimSpace(msg))
			return
		}
		phpInfoContent = strings.TrimSpace(string(out))
	}

	renderPHPInfoPage(a, w, r, version, service, title, phpInfoContent, installedVersions)
}
