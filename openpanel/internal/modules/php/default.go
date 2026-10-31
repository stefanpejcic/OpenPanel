package php

import (
	"context"
	"net/http"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func writeJSONError(w http.ResponseWriter, status int, message string) {
	web.WriteJSON(w, status, map[string]string{"error": message})
}

// litespeedDefaultVersions are compared as exact strings, parsing them as floats let 8.50 through as 8.5 and wrote an image tag that doesn't exist
var litespeedDefaultVersions = map[string]bool{"8.5": true, "8.4": true, "8.3": true, "8.2": true}

// checkDefaultVersion returns why version can't be the default, or "" when it can
func checkDefaultVersion(version string, isLitespeed bool, installed []string) string {
	if !phpVersionFormRE.MatchString(version) {
		return "Invalid PHP version format"
	}
	if isLitespeed {
		if !litespeedDefaultVersions[version] {
			return "Default PHP version for OpenLitespeed can not be set to " + version + " - only available tags are: 8.5 8.4 8.3 8.2"
		}
		return ""
	}
	if !containsString(installed, version) {
		return "PHP " + version + " is not available on this account"
	}
	return ""
}

// handleDefaultPHPVersion gets or sets the default PHP version applied to newly created domains
func handleDefaultPHPVersion(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
	outputJSON := r.URL.Query().Get("output") == "json"

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		newPHPVersion := strings.TrimSpace(r.Form.Get("new_php_version"))
		fail := func(status int, msg string) {
			if outputJSON {
				writeJSONError(w, status, msg)
				return
			}
			web.FlashRedirect(a, w, r, "error", msg, "/php/default")
		}

		// on php-fpm anything used to be accepted, a version that isn't installed broke every new domain
		if msg := checkDefaultVersion(newPHPVersion, isLitespeed, FetchPHPVersions(ctx, a, userContext)); msg != "" {
			fail(http.StatusBadRequest, msg)
			return
		}

		// the restart below must not stop halfway when the browser leaves, litespeed would stay stopped
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Minute)
		defer cancel()

		if prefErr := updatePHPVersionPreference(userContext, newPHPVersion); prefErr == nil {
			ipAddress := reqip.ClientIP(r)
			_ = logger.RecordUserAction(a.Config, currentUsername, "changed default PHP version for new domains to "+newPHPVersion, ipAddress)
			message := web.Tr(a, r, "PHP version %(new_phpversion)s set as default for new domains", "new_phpversion", newPHPVersion)

			if isLitespeed {
				if docker.ComposeContainer(ctx, userContext, webServer, "status") {
					message = web.Tr(a, r, "PHP version %(new_phpversion)s saved, and Litespeed restarted to apply the version.", "new_phpversion", newPHPVersion)
					docker.ComposeContainer(ctx, userContext, webServer, "stop")
					docker.ComposeContainer(ctx, userContext, webServer, "start")
				} else {
					message = web.Tr(a, r, "PHP version %(new_phpversion)s saved and will be used by Litespeed for new domains.", "new_phpversion", newPHPVersion)
				}
			} else if previousVersion := r.Form.Get("previous_version"); previousVersion != "" {
				stopPHPServiceIfRunningAndUnused(ctx, userContext, previousVersion)
			}

			if outputJSON {
				web.WriteJSON(w, http.StatusOK, map[string]any{"message": message, "version": newPHPVersion})
				return
			}
			web.Flash(a, w, r, "success", message)
		} else {
			fail(http.StatusInternalServerError, web.Tr(a, r, "Default PHP version could not be changed to %(new_phpversion)s", "new_phpversion", newPHPVersion)+": "+prefErr.Error())
			return
		}
	}

	installedVersions := FetchPHPVersions(ctx, a, userContext)
	phpDefaultVersion, service := computeDefaultPHPVersionAndService(ctx, userContext, webServer, isLitespeed)

	if outputJSON {
		versionsAPI := fetchPHPVersionsAPI(ctx)
		versionStatus := make(map[string]map[string]any, len(installedVersions))
		for _, v := range installedVersions {
			level, label := classifyPHPVersionLevel(v, versionsAPI)
			versionStatus[v] = map[string]any{"level": level, "label": label, "is_latest": versionsAPI[v].IsLatestVersion}
		}
		web.WriteJSON(w, http.StatusOK, map[string]any{
			"version": phpDefaultVersion, "service": service, "is_litespeed": isLitespeed,
			"installed_versions": installedVersions, "version_status": versionStatus,
		})
		return
	}

	renderDefaultPage(a, w, r, phpDefaultVersion, service, installedVersions, isLitespeed)
}

// computeDefaultPHPVersionAndService resolves the current default PHP version and the service that runs it, querying the LiteSpeed container directly when running under LiteSpeed since it has no per-domain php-fpm service
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
