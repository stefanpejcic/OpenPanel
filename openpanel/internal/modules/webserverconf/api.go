package webserverconf

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// RegisterAPI wires the webserver-conf JSON API routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "webserver_conf", "GET /api/webserver-conf", func(w http.ResponseWriter, r *http.Request) { apiWebserverConfGet(a, w, r) })
	apiregistry.Handle(mux, a, "webserver_conf", "PUT /api/webserver-conf", func(w http.ResponseWriter, r *http.Request) { apiWebserverConfPut(a, w, r) })
	apiregistry.Handle(mux, a, "webserver_conf", "GET /api/webserver-conf/litespeed-license", func(w http.ResponseWriter, r *http.Request) { apiLitespeedLicense(a, w, r) })
	apiregistry.Handle(mux, a, "webserver_conf", "POST /api/webserver-conf/litespeed-license", func(w http.ResponseWriter, r *http.Request) { apiLitespeedLicense(a, w, r) })
}

var litespeedSerialRE = regexp.MustCompile(`^[A-Za-z0-9+/=-]{8,64}$`)

// litespeedLicenseCmd is a var so tests don't run opencli
var litespeedLicenseCmd = func(r *http.Request, args ...string) ([]byte, error) {
	return exec.CommandContext(r.Context(), "opencli", append([]string{"user-litespeed"}, args...)...).Output()
}

// apiLitespeedLicense shows the LiteSpeed Enterprise license, a POST sets a serial ({"serial": "..."}) or goes back to a trial ({"trial": true})
func apiLitespeedLicense(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if webserver.GetEnvFileValue(userContext, "WEB_SERVER") != "litespeed" {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "LiteSpeed is not your web server."})
		return
	}
	args := []string{userContext, "status", "--json"}
	if r.Method == http.MethodPost {
		var body struct {
			Serial string `json:"serial"`
			Trial  bool   `json:"trial"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		body.Serial = strings.TrimSpace(body.Serial)
		switch {
		case body.Trial:
			args = []string{userContext, "trial", "--json"}
		case litespeedSerialRE.MatchString(body.Serial):
			args = []string{userContext, "serial", body.Serial, "--json"}
		default:
			web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid serial number."})
			return
		}
	}
	out, _ := litespeedLicenseCmd(r, args...)
	var result map[string]any
	if json.Unmarshal(out, &result) != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not read the LiteSpeed license."})
		return
	}
	if r.Method == http.MethodPost {
		_ = logger.RecordUserAction(a.Config, currentUsername, "changed the LiteSpeed license ("+args[1]+")", reqip.ClientIP(r))
	}
	status := http.StatusOK
	if ok, _ := result["success"].(bool); !ok {
		status = http.StatusBadRequest
	}
	web.WriteJSON(w, status, result)
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

	content, readErr := tenantfs.ReadFile(configFilePath)
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

// apiWebserverConfPut saves new webserver config content, syntax-checks it, and reloads the service - rolling back on a failed check
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

	previousContent, prevErr := tenantfs.ReadFile(configFilePath)

	if writeErr := tenantfs.WriteFile(configFilePath, []byte(*body.Content), 0o644); writeErr != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	if ok, testOutput := webserver.TestWebserverConfig(ctx, userContext, entry.ServiceName); !ok {
		if prevErr == nil {
			_ = tenantfs.WriteFile(configFilePath, previousContent, 0o644)
		}
		failure := describeTestFailure(requestTr(a, r), entry, testOutput)
		web.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error":    "Config rejected - failed " + entry.ServiceName + " syntax check: " + strings.Join(failure.Problems, "; "),
			"problems": failure.Problems, "hints": failure.Hints, "line": failure.Line, "test_output": testOutput,
		})
		return
	}

	serviceRunning := docker.IsServiceRunning(ctx, userContext, entry.ServiceName)
	reloaded := false
	if serviceRunning {
		if reloadErr := webserver.ReloadWebserver(ctx, userContext, entry.ServiceName); reloadErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Config saved but service reload failed: " + reloadErr.Error(),
			})
			return
		}
		reloaded = true
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "edited "+entry.ConfFile+" and reloaded "+entry.ServiceName, reqip.ClientIP(r))

	message := "Config saved (service not running)"
	if reloaded {
		message = "Config saved and service reloaded"
	}
	// "restarted" stays in the response for existing API clients
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"message": message, "web_server": webServer, "service": entry.ServiceName, "reloaded": reloaded, "restarted": reloaded,
	})
}
