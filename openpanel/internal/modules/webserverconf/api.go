package webserverconf

import (
	"encoding/json"
	"net/http"
	"os"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// RegisterAPI wires the webserver-conf JSON API routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "webserver_conf", "GET /api/webserver-conf", func(w http.ResponseWriter, r *http.Request) { apiWebserverConfGet(a, w, r) })
	apiregistry.Handle(mux, a, "webserver_conf", "PUT /api/webserver-conf", func(w http.ResponseWriter, r *http.Request) { apiWebserverConfPut(a, w, r) })
}

// apiWebserverConfLabels gives each web server's short display label - webserverConfEntry.PageTitle is a different, longer string used only for the UI page title, so this stays a separate small lookup
var apiWebserverConfLabels = map[string]string{
	"nginx": "Nginx", "apache": "Apache", "openresty": "OpenResty",
	"openlitespeed": "OpenLiteSpeed", "litespeed": "LiteSpeed",
}

// apiWebserverConfGet returns the current user's webserver config file content
func apiWebserverConfGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	entry, known := webserverConfs[webServer]
	if !known {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Unknown web server: " + webServer})
		return
	}
	label := apiWebserverConfLabels[webServer]
	configFilePath := "/home/" + userContext + "/" + entry.ConfFile

	content, readErr := os.ReadFile(configFilePath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			web.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "Config file not found: " + entry.ConfFile})
		} else {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": readErr.Error()})
		}
		return
	}

	web.WriteJSON(w, http.StatusOK, map[string]any{
		"web_server": webServer, "label": label, "filename": entry.ConfFile,
		"service": entry.ServiceName, "content": string(content),
	})
}

// apiWebserverConfPut saves new webserver config content, syntax-checks it, and restarts the service - rolling back on a failed check
func apiWebserverConfPut(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
	entry, known := webserverConfs[webServer]
	if !known {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Unknown web server: " + webServer})
		return
	}
	configFilePath := "/home/" + userContext + "/" + entry.ConfFile

	var body struct {
		Content *string `json:"content"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Content == nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing \"content\" field"})
		return
	}

	previousContent, prevErr := os.ReadFile(configFilePath)

	if writeErr := os.WriteFile(configFilePath, []byte(*body.Content), 0o644); writeErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	serviceRunning := docker.IsServiceRunning(ctx, userContext, entry.ServiceName)
	if serviceRunning {
		ok, testOutput := webserver.TestWebserverConfig(ctx, userContext, entry.ServiceName)
		if !ok {
			if prevErr == nil {
				_ = os.WriteFile(configFilePath, previousContent, 0o644)
			}
			web.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Config rejected - failed " + entry.ServiceName + " syntax check: " + testOutput,
			})
			return
		}
	}

	restarted := false
	if serviceRunning {
		argv := podmanmanager.PodmanArgv(userContext, "restart", entry.ServiceName)
		var stderr []byte
		cmd := podmanmanager.Command(ctx, userContext, argv)
		out, runErr := cmd.CombinedOutput()
		if runErr != nil {
			stderr = out
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Config saved but service restart failed: " + string(stderr),
			})
			return
		}
		restarted = true
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "edited "+entry.ConfFile+" and restarted "+entry.ServiceName, reqip.ClientIP(r))

	message := "Config saved (service not running)"
	if restarted {
		message = "Config saved and service restarted"
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"message": message, "web_server": webServer, "service": entry.ServiceName, "restarted": restarted,
	})
}
