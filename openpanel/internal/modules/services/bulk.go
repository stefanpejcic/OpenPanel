package services

import (
	"fmt"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// servicesBulkActions are the buttons in the services bulk bar, keys match handleServicesBulk
func servicesBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "restart", Label: t.Get("Restart"), Confirm: t.Get("Restart the selected services?")},
		{Key: "start", Label: t.Get("Start"), Confirm: t.Get("Start the selected services?")},
		{Key: "stop", Label: t.Get("Stop"), Confirm: t.Get("Stop the selected services?"), Danger: true},
	}
}

// runServiceBulkAction does what the row menu's restart/enable/disable does for one service
func runServiceBulkAction(r *http.Request, action, service, userContext string, allowed []string) web.BulkResult {
	if !containsString(allowed, service) {
		return web.BulkResult{Message: "Service not found."}
	}
	var res docker.StartStopResult
	switch action {
	case "restart":
		res = docker.RestartContainer(r.Context(), userContext, service)
	case "start":
		res = docker.StartOrStopContainer(r.Context(), userContext, service, "activate", "")
	case "stop":
		res = docker.StartOrStopContainer(r.Context(), userContext, service, "deactivate", "")
	default:
		return web.BulkResult{Message: "Unknown bulk action."}
	}
	return web.BulkResult{OK: res.Success, Message: res.Message}
}

// handleServicesBulk runs start/stop/restart over the selected services one by one, the same way /containers/bulk does
func handleServicesBulk(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	req, ok := web.DecodeBulkRequest(a, w, r)
	if !ok {
		return
	}
	var label string
	for _, act := range servicesBulkActions(web.RequestTranslator(a, r)) {
		if act.Key == req.Action {
			label = act.Label
		}
	}
	if label == "" {
		web.BulkError(a, w, r, "Unknown bulk action.")
		return
	}

	composeData, err := docker.LoadCompose(userContext)
	if err != nil {
		web.BulkError(a, w, r, "Error reading services.")
		return
	}
	webserver, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
	mysqlType, _ := docker.GetEnvValue(userContext, "MYSQL_TYPE")
	allowed := FilterServices(composeServiceNames(composeData), webserver, mysqlType)

	results := web.RunBulk(req.Items, func(service string) web.BulkResult {
		return runServiceBulkAction(r, req.Action, service, userContext, allowed)
	})

	_ = logger.RecordUserAction(a.Config, username, fmt.Sprintf("ran bulk action '%s' on services: %s", req.Action, strings.Join(req.Items, ", ")), reqip.ClientIP(r))
	web.FinishBulk(a, w, r, label, results)
}
