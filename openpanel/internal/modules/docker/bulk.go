package docker

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// containerBulkActions are the buttons in the /containers bulk bar, keys match handleContainersBulk
func containerBulkActions(t i18n.Translator, totalCPU, totalRAM int) []web.BulkAction {
	cpuMax, ramMax := "", ""
	if totalCPU > 0 {
		cpuMax = strconv.Itoa(totalCPU)
	}
	if totalRAM > 0 {
		ramMax = strconv.Itoa(totalRAM)
	}
	return []web.BulkAction{
		{Key: "start", Label: t.Get("Start"), Confirm: t.Get("Start the selected containers?")},
		{Key: "stop", Label: t.Get("Stop"), Confirm: t.Get("Stop the selected containers?")},
		{Key: "restart", Label: t.Get("Restart"), Confirm: t.Get("Restart the selected containers?")},
		{Key: "cpu", Label: t.Get("Edit CPU"), Confirm: t.Get("Set the CPU limit for the selected containers:"),
			Input: &web.BulkInput{Type: "number", Min: "0", Max: cpuMax, Step: ".01", Placeholder: t.Get("cores"), Hint: t.Get("0 = unlimited")}},
		{Key: "ram", Label: t.Get("Edit RAM"), Confirm: t.Get("Set the memory limit (GB) for the selected containers:"),
			Input: &web.BulkInput{Type: "number", Min: "0", Max: ramMax, Step: ".01", Placeholder: "GB", Hint: t.Get("0 = unlimited")}},
		{Key: "pids", Label: t.Get("Edit PIDs"), Confirm: t.Get("Set the max processes for the selected containers:"),
			Input: &web.BulkInput{Type: "number", Min: "0", Step: "1", Placeholder: t.Get("PIDs"), Hint: t.Get("0 = unlimited")}},
		{Key: "delete", Label: t.Get("Delete"), Confirm: t.Get("Permanently delete the selected containers and their images? This cannot be undone."), Danger: true},
	}
}

// canDeleteService matches the row's ShowManage: only services the user added, never the built-in ones
func canDeleteService(service string) bool {
	return !coreServices[service] && !UndeletableServices[service] && !strings.HasPrefix(service, "php-fpm-")
}

// validateBulkLimit checks the cpu/ram/pids value once, before touching any container, returning a msgid and its placeholder values
func validateBulkLimit(action, value string, totalCPU, totalRAM int) (string, []any) {
	switch action {
	case "cpu", "ram":
		val, err := strconv.ParseFloat(value, 64)
		if err != nil || val < 0 {
			return "Limit must be 0 or a positive number.", nil
		}
		if action == "cpu" && totalCPU > 0 && val > float64(totalCPU) {
			return "CPU limit can't be more than the %(cores)s cores on your plan.", []any{"cores", totalCPU}
		}
		if action == "ram" && totalRAM > 0 && val > float64(totalRAM) {
			return "Memory limit can't be more than the %(gb)s GB on your plan.", []any{"gb", totalRAM}
		}
	case "pids":
		val, err := strconv.Atoi(value)
		if err != nil || val < 0 {
			return "PIDs limit must be 0 or a positive whole number.", nil
		}
	}
	return "", nil
}

// handleContainersBulk runs start/stop/restart/cpu/ram/pids/delete over the selected services, one by one
func handleContainersBulk(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)
	planID, _ := injected["hosting_plan"].(int)
	if username == "" {
		writeJSONError(w, http.StatusUnauthorized, web.Tr(a, r, "User not authenticated"))
		return
	}

	req, ok := web.DecodeBulkRequest(a, w, r)
	if !ok {
		return
	}

	totalCPU, totalRAM := getCPUAndRAMForPlanID(a, ctx, planID)
	var label string
	for _, act := range containerBulkActions(web.RequestTranslator(a, r), totalCPU, totalRAM) {
		if act.Key == req.Action {
			label = act.Label
		}
	}
	if label == "" {
		web.BulkError(a, w, r, "Unknown bulk action.")
		return
	}
	value := strings.TrimSpace(req.Value)
	if msg, kv := validateBulkLimit(req.Action, value, totalCPU, totalRAM); msg != "" {
		web.BulkError(a, w, r, msg, kv...)
		return
	}

	composeData, loadErr := LoadCompose(userContext)
	if loadErr != nil {
		writeJSONError(w, http.StatusInternalServerError, web.Tr(a, r, "Could not read docker-compose.yml."))
		return
	}

	results := web.RunBulk(req.Items, func(service string) web.BulkResult {
		svcRaw, exists := servicesRaw(composeData)[service]
		if !exists {
			return web.BulkResult{Message: "Container not found."}
		}

		switch req.Action {
		case "start", "stop":
			upDown := map[string]string{"start": "activate", "stop": "deactivate"}[req.Action]
			res := StartOrStopContainer(ctx, userContext, service, upDown, "")
			return web.BulkResult{OK: res.Success, Message: res.Message}
		case "restart":
			res := RestartContainer(ctx, userContext, service)
			return web.BulkResult{OK: res.Success, Message: res.Message}
		case "cpu", "ram":
			msg, ok := updateContainerRAMOrCPU(a, ctx, userContext, planID, service, req.Action, value)
			return web.BulkResult{OK: ok, Message: msg}
		case "pids":
			msg, ok := updateContainerPIDs(ctx, userContext, service, value)
			return web.BulkResult{OK: ok, Message: msg}
		case "delete":
			if !canDeleteService(service) {
				return web.BulkResult{Message: "Built-in containers can't be deleted."}
			}
			if err := removeService(ctx, userContext, composeData, service, svcRaw); err != nil {
				return web.BulkResult{Message: "Failed to update docker-compose.yml."}
			}
			return web.BulkResult{OK: true, Message: "Deleted."}
		}
		return web.BulkResult{Message: "Unknown bulk action."}
	})

	logMsg := fmt.Sprintf("ran bulk action '%s' on containers: %s", req.Action, strings.Join(req.Items, ", "))
	if value != "" {
		logMsg += " (value: " + value + ")"
	}
	_ = logger.RecordUserAction(a.Config, username, logMsg, reqip.ClientIP(r))
	web.FinishBulk(a, w, r, label, results)
}
