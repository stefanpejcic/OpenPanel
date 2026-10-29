package webserverconf

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
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
	_, username, userContext, err := auth.Injected(a, r)
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
		content, readErr := tenantfs.ReadFile(configFilePath)
		if readErr != nil {
			renderWebserverConfPage(a, w, r, entry.PageTitle, webServer, entry.ConfFile, false, "", nil)
			return
		}
		existingContent = string(content)
	}

	renderWebserverConfPage(a, w, r, entry.PageTitle, webServer, entry.ConfFile, canRestoreDefault(entry), existingContent, nil)
}

// browsers post textarea content with CRLF, without this every save rewrote each line of the config
func editorContent(r *http.Request) string {
	return strings.ReplaceAll(r.Form.Get("editor_content"), "\r\n", "\n")
}

func canRestoreDefault(entry webserverConfEntry) bool {
	defaultConfPath, hasDefault := defaultConfTemplates[entry.ServiceName]
	return hasDefault && fileExists(defaultConfPath)
}

// handleSaveOrRestore performs the POST action - either "restore_default" or the normal save-and-validate path, followed by a service reload - returns false once it has written a response (redirect or rejected edit), so the caller should stop
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

		content, readErr := tenantfs.ReadFile(defaultConfPath)
		if readErr != nil {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Error restoring %(conf_file)s: %(error)s", "conf_file", entry.ConfFile, "error", readErr.Error()))
			return false
		}
		if writeErr := tenantfs.WriteFile(configFilePath, content, 0o644); writeErr != nil {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Error restoring %(conf_file)s: %(error)s", "conf_file", entry.ConfFile, "error", writeErr.Error()))
			return false
		}
		actionDescription = "restored " + entry.ConfFile + " to default"
	} else {
		newContent := editorContent(r)
		if newContent == "" {
			flashAndRedirect(a, w, r, "", "")
			return false
		}

		previousContent, prevErr := tenantfs.ReadFile(configFilePath)

		if writeErr := tenantfs.WriteFile(configFilePath, []byte(newContent), 0o644); writeErr != nil {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Error writing to %(conf_file)s: %(error)s", "conf_file", entry.ConfFile, "error", writeErr.Error()))
			return false
		}

		if ok, testOutput := webserver.TestWebserverConfig(r.Context(), userContext, entry.ServiceName); !ok {
			if prevErr == nil {
				_ = tenantfs.WriteFile(configFilePath, previousContent, 0o644)
			}
			// rendered here instead of a redirect so the editor keeps what the user typed
			failure := describeTestFailure(requestTr(a, r), entry, testOutput)
			sess, _ := a.Sessions.Get(r, session.CookieName)
			flash.Add(sess, "error", web.Tr(a, r, "Configuration was not saved, the %(service_name)s config test found a problem. Your changes are still in the editor.", "service_name", entry.ServiceName))
			_ = a.Sessions.Save(r, w, sess)
			renderWebserverConfPage(a, w, r, entry.PageTitle, webServer, entry.ConfFile, canRestoreDefault(entry), newContent, &failure)
			return false
		}
		actionDescription = "edited " + entry.ConfFile
	}

	// reload service - unlike the early-exit error branches above, every path from here on falls through to re-read the file and render the page directly, it does not redirect
	sess, _ := a.Sessions.Get(r, session.CookieName)
	if docker.IsServiceRunning(r.Context(), userContext, entry.ServiceName) {
		if reloadErr := webserver.ReloadWebserver(r.Context(), userContext, entry.ServiceName); reloadErr != nil {
			flash.Add(sess, "error", web.Tr(a, r, "Changes were saved to %(conf_file)s, but reloading the %(service_name)s container failed: %(error)s", "conf_file", entry.ConfFile, "service_name", entry.ServiceName, "error", reloadErr.Error()))
		} else {
			var message string
			if action == "restore_default" {
				message = web.Tr(a, r, "Restored %(conf_file)s to the default configuration and %(service_name)s container reloaded successfully.", "conf_file", entry.ConfFile, "service_name", entry.ServiceName)
			} else {
				message = web.Tr(a, r, "Changes saved to %(conf_file)s and %(service_name)s container reloaded successfully.", "conf_file", entry.ConfFile, "service_name", entry.ServiceName)
			}
			flash.Add(sess, "success", message)
			_ = logger.RecordUserAction(a.Config, username, actionDescription+" and reloaded "+entry.ServiceName, reqip.ClientIP(r))
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
