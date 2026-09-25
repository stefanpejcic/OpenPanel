package webserverconf

import (
	"net/http"
	"os"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const redirectPath = "/server/webserver_conf"

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	if message != "" {
		flash.Add(sess, category, message)
	}
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, redirectPath, http.StatusFound)
}

// handleWebserverConf serves the in-browser editor for a user's main webserver configuration file and handles saving/restoring it
func handleWebserverConf(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	entry := lookupWebserverConf(webServer)

	var configFilePath string
	if entry.ConfFile != "" {
		configFilePath = "/home/" + userContext + "/" + entry.ConfFile
	}

	if r.Method == http.MethodPost && configFilePath != "" {
		if !handleSaveOrRestore(a, w, r, username, userContext, webServer, entry, configFilePath) {
			return
		}
	}

	existingContent := ""
	if configFilePath != "" {
		content, readErr := os.ReadFile(configFilePath)
		if readErr != nil {
			renderWebserverConfPage(a, w, r, entry.PageTitle, webServer, entry.ConfFile, false, "")
			return
		}
		existingContent = string(content)
	}

	defaultConfPath, hasDefault := defaultConfTemplates[entry.ServiceName]
	canRestoreDefault := hasDefault && fileExists(defaultConfPath)

	renderWebserverConfPage(a, w, r, entry.PageTitle, webServer, entry.ConfFile, canRestoreDefault, existingContent)
}

// handleSaveOrRestore performs the POST action - either "restore_default" or the normal save-and-validate path, followed by a service restart - returns false once it has written a response (redirect), so the caller should stop
func handleSaveOrRestore(a *appctx.App, w http.ResponseWriter, r *http.Request, username, userContext, webServer string, entry webserverConfEntry, configFilePath string) bool {
	_ = r.ParseForm()
	action := r.Form.Get("action")
	if action == "" {
		action = "save"
	}

	var actionDescription string

	if action == "restore_default" {
		defaultConfPath, ok := defaultConfTemplates[entry.ServiceName]
		if !ok || !fileExists(defaultConfPath) {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "No default configuration is available for %(service_name)s.", "service_name", entry.ServiceName))
			return false
		}

		content, readErr := os.ReadFile(defaultConfPath)
		if readErr != nil {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Error restoring %(conf_file)s: %(error)s", "conf_file", entry.ConfFile, "error", readErr.Error()))
			return false
		}
		if writeErr := os.WriteFile(configFilePath, content, 0o644); writeErr != nil {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Error restoring %(conf_file)s: %(error)s", "conf_file", entry.ConfFile, "error", writeErr.Error()))
			return false
		}
		actionDescription = "restored " + entry.ConfFile + " to default"
	} else {
		newContent := r.Form.Get("editor_content")
		if newContent == "" {
			flashAndRedirect(a, w, r, "", "")
			return false
		}

		previousContent, prevErr := os.ReadFile(configFilePath)

		if writeErr := os.WriteFile(configFilePath, []byte(newContent), 0o644); writeErr != nil {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Error writing to %(conf_file)s: %(error)s", "conf_file", entry.ConfFile, "error", writeErr.Error()))
			return false
		}

		if docker.IsServiceRunning(r.Context(), userContext, entry.ServiceName) {
			ok, testOutput := webserver.TestWebserverConfig(r.Context(), userContext, entry.ServiceName)
			if !ok {
				if prevErr == nil {
					_ = os.WriteFile(configFilePath, previousContent, 0o644)
				}
				flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Configuration was not saved - it failed the %(service_name)s syntax check: %(test_output)s", "service_name", entry.ServiceName, "test_output", testOutput))
				return false
			}
		}
		actionDescription = "edited " + entry.ConfFile
	}

	// restart service - unlike the early-exit error branches above, every path from here on falls through to re-read the file and render the page directly, it does not redirect
	sess, _ := a.Sessions.Get(r, session.CookieName)
	if docker.IsServiceRunning(r.Context(), userContext, entry.ServiceName) {
		argv := podmanmanager.PodmanArgv(userContext, "restart", entry.ServiceName)
		if runErr := podmanmanager.Command(r.Context(), userContext, argv).Run(); runErr != nil {
			flash.Add(sess, "error", web.Tr(a, r, "Error restarting %(service_name)s container.", "service_name", entry.ServiceName))
		} else {
			var message string
			if action == "restore_default" {
				message = web.Tr(a, r, "Restored %(conf_file)s to the default configuration and %(service_name)s container restarted successfully.", "conf_file", entry.ConfFile, "service_name", entry.ServiceName)
			} else {
				message = web.Tr(a, r, "Changes saved to %(conf_file)s and %(service_name)s container restarted successfully.", "conf_file", entry.ConfFile, "service_name", entry.ServiceName)
			}
			flash.Add(sess, "success", message)
			_ = logger.RecordUserAction(a.Config, username, actionDescription+" and restarted "+entry.ServiceName, reqip.ClientIP(r))
		}
	} else {
		var message string
		if action == "restore_default" {
			message = web.Tr(a, r, "Restored %(conf_file)s to the default configuration - the %(service_name)s container is not running.", "conf_file", entry.ConfFile, "service_name", entry.ServiceName)
		} else {
			message = web.Tr(a, r, "Changes saved to %(conf_file)s - the %(service_name)s container is not running.", "conf_file", entry.ConfFile, "service_name", entry.ServiceName)
		}
		flash.Add(sess, "success", message)
		_ = logger.RecordUserAction(a.Config, username, actionDescription, reqip.ClientIP(r))
	}
	_ = a.Sessions.Save(r, w, sess)
	return true
}
