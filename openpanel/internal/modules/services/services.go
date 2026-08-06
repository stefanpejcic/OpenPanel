// Package services provides a simplified, filtered view of the
// docker-compose services list (excluding all but the currently-active
// webserver/mysql engine) with enable/disable/restart actions and a
// real-time status/log widget, built entirely on top of
// internal/modules/docker's existing podman CLI layer.
package services

import (
	"encoding/json"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// Register wires the services routes onto mux, gated behind the
// "services" feature flag.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "services")(h)
	}
	mux.Handle("/services/", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleManageService(a, w, r)
	}))
	mux.Handle("/services/{service}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleManageService(a, w, r)
	}))
	// The services page's log widget (_service_log.html) hard-depends on
	// this route, so it's registered alongside the rest of this package
	// rather than living with the other JSON helper endpoints.
	mux.Handle("GET /json/containers/log/{service_name}", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleViewLog(a, w, r)
	}))
}

var webserverServices = map[string]bool{
	"apache": true, "nginx": true, "openresty": true, "openlitespeed": true, "litespeed": true,
}
var mysqlServices = map[string]bool{"mysql": true, "mariadb": true}

// FilterServices shows only the currently-active webserver and mysql
// engine, hides every other webserver/mysql service, and passes everything
// else through unchanged.
func FilterServices(all []string, webserver, mysqlType string) []string {
	var result []string
	for _, s := range all {
		if s == webserver || s == mysqlType || (!webserverServices[s] && !mysqlServices[s]) {
			result = append(result, s)
		}
	}
	return result
}

// handleManageService renders the service picker (no service selected) or
// a single service's detail/status page, and applies enable/disable/restart
// actions posted from that page.
func handleManageService(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	service := r.PathValue("service")

	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	// A YAML parse error returns a plain 500 with no flash message, since
	// there's no template render afterward to pop one. A missing compose
	// file isn't distinguished from "no services yet" here - LoadCompose
	// already treats it as an empty (not error) config, matching every
	// other caller in internal/modules/docker.
	composeData, err := docker.LoadCompose(userContext)
	if err != nil {
		http.Error(w, "Error reading services.", http.StatusInternalServerError)
		return
	}
	allServices := composeServiceNames(composeData)

	webserver, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
	mysqlType, _ := docker.GetEnvValue(userContext, "MYSQL_TYPE")
	allowedServices := FilterServices(allServices, webserver, mysqlType)

	if service == "" {
		renderServicesListPage(a, w, r, allowedServices)
		return
	}

	if !containsString(allowedServices, service) {
		// A plain 403 with no flash message, since there's no template
		// render afterward to pop one.
		http.Error(w, "Service '"+service+"' does not exist.", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		HandleServiceAction(a, r, r.Form.Get("action"), service, userContext, username)
	}

	status := docker.GetContainerStatus(ctx, userContext, service)

	if r.URL.Query().Get("output") == "json" {
		actions := []string{"enable", "restart"}
		if status.State == "running" {
			actions = []string{"disable"}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service": service, "actions": actions,
			"container_state": status.State, "health_status": status.Health,
		})
		return
	}

	renderServiceDetailPage(a, w, r, service, userContext, status)
}

// HandleServiceAction applies an enable/disable/restart action and queues a
// flash message on the request's session object, but does not persist it -
// unlike every other handler in this codebase, this route doesn't redirect
// after a POST, it falls through to rendering the same page in the same
// response. gorilla/sessions caches one *sessions.Session per *http.Request,
// so the later web.BuildLayoutData -> flash.Pop call during that render
// reuses this exact session object (including the flash just added here)
// and performs the one real Save/Set-Cookie write for the response.
func HandleServiceAction(a *appctx.App, r *http.Request, action, service, userContext, username string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)

	switch action {
	case "restart":
		result := docker.RestartContainer(r.Context(), userContext, service)
		if result.Success {
			_ = logger.RecordUserAction(a.Config, username, "restarted service "+service, reqip.ClientIP(r))
			flash.Add(sess, "success", service+" service is now restarted.")
		} else {
			msg := result.Message
			if msg == "" {
				msg = "Failed to restart the " + service + " service."
			}
			flash.Add(sess, "error", msg)
		}
	case "enable", "disable":
		dockerAction := "deactivate"
		successMsg := service + " service is now disabled."
		failMsg := "Failed to stop the " + service + " service."
		if action == "enable" {
			dockerAction = "activate"
			successMsg = service + " service is now enabled."
			failMsg = "Failed to start the " + service + " service."
		}
		result := docker.StartOrStopContainer(r.Context(), userContext, service, dockerAction, "")
		if result.Success {
			_ = logger.RecordUserAction(a.Config, username, action+"d service "+service, reqip.ClientIP(r))
			flash.Add(sess, "success", successMsg)
		} else {
			flash.Add(sess, "error", failMsg+" Try restart.")
		}
	}
}

func handleViewLog(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serviceName := r.PathValue("service_name")

	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := injected["context"].(string)
	if userContext == "" {
		userContext = "default"
	}

	body, status := docker.FetchContainerLog(ctx, a, userContext, serviceName, 100)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func composeServiceNames(composeData map[string]any) []string {
	services, _ := composeData["services"].(map[string]any)
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	return names
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
