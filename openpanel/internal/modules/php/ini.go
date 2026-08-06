package php

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// HealthIssue is a {id, severity, message} toast issue, rendered
// client-side via reportHealthIssues().
type HealthIssue struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// handlePHPIniEditor edits the php.ini file for one PHP version (or, with
// no version, just renders the version-picker form). versionSeg is "" for
// the bare /php/php_ini_editor route.
func handlePHPIniEditor(a *appctx.App, w http.ResponseWriter, r *http.Request, versionSeg string) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	version := phpVersionFromIniSegment(versionSeg)
	title := "PHP.INI Editor"
	var phpIniContent, displayFilePath string
	installedVersions := FetchPHPVersions(ctx, a, userContext)
	service := ""

	if version != "" {
		phpIniFilePath := "/home/" + userContext + "/php.ini/" + version + ".ini"
		webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")

		var containerToRestart, text string
		if strings.Contains(strings.ToLower(webServer), "litespeed") {
			service = strings.ToLower(webServer)
			displayFilePath = "/usr/local/lsws/lsphp" + strings.ReplaceAll(version, ".", "") + "/etc/php/" + version + "/litespeed/php.ini"
			containerToRestart = webServer
			text = "LSPHP"
		} else {
			displayFilePath = "/etc/php/" + version + "/fpm/php.ini"
			containerToRestart = "php-fpm-" + version
			service = containerToRestart
			text = "PHP-FPM"
		}
		title = "PHP " + version + " INI Editor"

		content, readErr := os.ReadFile(phpIniFilePath)
		if readErr != nil {
			writeJSONError(w, http.StatusInternalServerError, readErr.Error())
			return
		}
		phpIniContent = string(content)

		if r.Method == http.MethodPost {
			_ = r.ParseForm()
			if newContent := r.Form.Get("editor_content"); newContent != "" {
				if writeErr := os.WriteFile(phpIniFilePath, []byte(newContent), 0o644); writeErr != nil {
					flashAndRedirect(a, w, r, "error", "Error saving "+text+" php.ini file!", "/php/php"+version+".ini/editor")
					return
				}

				var message string
				if docker.ComposeContainer(ctx, userContext, containerToRestart, "status") {
					message = fmt.Sprintf("PHP.INI file for %s version %s edited successfully, and %s service restarted to apply.", text, version, containerToRestart)
					docker.ComposeContainer(ctx, userContext, containerToRestart, "stop")
					docker.ComposeContainer(ctx, userContext, containerToRestart, "start")
				} else {
					message = fmt.Sprintf("PHP.INI file for %s version %s edited successfully.", text, version)
				}

				flashSess(a, w, r, "success", message)
				ipAddress := reqip.ClientIP(r)
				_ = logger.RecordUserAction(a.Config, currentUsername, "edited php.ini for PHP "+version, ipAddress)
				http.Redirect(w, r, "/php/php"+version+".ini/editor", http.StatusFound)
				return
			}
		}
	}

	var phpIniIssues []HealthIssue
	if version != "" {
		if syntaxError := checkPHPIniSyntax(ctx, userContext, version); syntaxError != "" {
			phpIniIssues = append(phpIniIssues, HealthIssue{
				ID:       "php-ini-syntax:" + version,
				Severity: "error",
				Message:  fmt.Sprintf("Syntax error in the PHP %s ini file: %s", version, syntaxError),
			})
		}
	}

	renderPHPIniEditorPage(a, w, r, version, service, title, displayFilePath, phpIniContent, installedVersions, phpIniIssues)
}
