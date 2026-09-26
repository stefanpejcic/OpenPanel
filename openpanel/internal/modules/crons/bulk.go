package crons

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// Key identifies a job for bulk selection, a job has no ID of its own
func (j CronJob) Key() string {
	sum := sha1.Sum([]byte(j.Comment + "\x00" + j.Schedule + "\x00" + j.Container + "\x00" + j.Command))
	return hex.EncodeToString(sum[:8])
}

func cronBulkActions(t i18n.Translator, containers []string) []web.BulkAction {
	options := make([]web.BulkOption, 0, len(containers))
	for _, c := range containers {
		options = append(options, web.BulkOption{Value: c, Label: c})
	}
	return []web.BulkAction{
		{Key: "run", Label: t.Get("Run now"), Confirm: t.Get("Run the selected cron jobs now, one after another?")},
		{Key: "schedule", Label: t.Get("Change schedule"), Confirm: t.Get("New schedule for the selected cron jobs:"),
			Input: &web.BulkInput{Type: "text", Placeholder: "0 */5 * * * *", Hint: t.Get("6 fields, seconds first")}},
		{Key: "container", Label: t.Get("Change container"), Confirm: t.Get("Run the selected cron jobs in:"), Input: &web.BulkInput{Type: "select", Options: options}},
		{Key: "overlap_on", Label: t.Get("No overlap on"), Confirm: t.Get("Skip a run of the selected cron jobs while the previous one is still running?")},
		{Key: "overlap_off", Label: t.Get("No overlap off"), Confirm: t.Get("Allow the selected cron jobs to run while the previous run is still going?")},
		{Key: "delete", Label: t.Get("Delete"), Confirm: t.Get("Delete the selected cron jobs?"), Danger: true},
	}
}

// runCronJobOnce runs one job to the end like the Run button, but without streaming the output
func runCronJobOnce(a *appctx.App, r *http.Request, userContext string, job CronJob) web.BulkResult {
	ctx, cancel := context.WithTimeout(context.Background(), cronRunTimeout(a))
	defer cancel()
	argv := podmanmanager.PodmanArgv(userContext, "exec", job.Container, "sh", "-c", job.Command)
	out, err := podmanmanager.Command(ctx, userContext, argv).CombinedOutput()
	lastLine := ""
	if lines := strings.Split(strings.TrimSpace(string(out)), "\n"); len(lines) > 0 {
		lastLine = strings.TrimSpace(lines[len(lines)-1])
	}
	if err != nil {
		code := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		}
		msg := web.Tr(a, r, "Exit code %(code)s.", "code", code)
		if lastLine != "" {
			msg = web.Tr(a, r, "Exit code %(code)s: %(output)s", "code", code, "output", lastLine)
		}
		return web.BulkResult{Item: job.Comment, Message: msg}
	}
	return web.BulkResult{Item: job.Comment, OK: true, Message: web.Tr(a, r, "Exit code %(code)s.", "code", 0)}
}

func handleCronjobsBulk(a *appctx.App, mux http.Handler, w http.ResponseWriter, r *http.Request) {
	username, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	content, _ := os.ReadFile(cronFilePath(userContext))
	jobs := map[string]CronJob{}
	for _, j := range ParseCronFile(string(content)) {
		jobs[j.Key()] = j
	}
	var containers []string
	if compose, composeErr := podmanmanager.LoadComposeConfig(r.Context(), userContext); composeErr == nil {
		containers, _ = serviceNamesFromCompose(compose)
	}

	web.ServeBulkDispatch(a, mux, w, r, cronBulkActions(web.RequestTranslator(a, r), containers), func(action, value, key string) (*web.BulkCall, web.BulkResult) {
		job, ok := jobs[key]
		if !ok {
			return web.Skip("Cron job not found, reload the page.")
		}
		if action == "run" {
			_ = logger.RecordUserAction(a.Config, username, "manually ran cron job "+job.Comment, reqip.ClientIP(r))
			return nil, runCronJobOnce(a, r, userContext, job)
		}
		if action == "delete" {
			return web.Call(web.BulkCall{Label: job.Comment, Method: http.MethodPost, Path: "/cronjobs/delete", Form: url.Values{
				"comment": {job.Comment}, "schedule": {job.Schedule}, "container": {job.Container}, "command": {job.Command},
			}})
		}

		updated := job
		switch action {
		case "schedule":
			if msg := validateCronSchedule(value); msg != "" {
				return nil, web.BulkResult{Item: job.Comment, Message: msg}
			}
			updated.Schedule = value
		case "container":
			if !containsString(containers, value) {
				return nil, web.BulkResult{Item: job.Comment, Message: "Unknown container."}
			}
			updated.Container = value
		case "overlap_on", "overlap_off":
			updated.NoOverlap = action == "overlap_on"
		}
		form := url.Values{
			"original_comment": {job.Comment}, "original_schedule": {job.Schedule}, "original_container": {job.Container}, "original_command": {job.Command},
			"comment": {updated.Comment}, "schedule": {updated.Schedule}, "container": {updated.Container}, "command": {updated.Command},
		}
		if updated.NoOverlap {
			form.Set("no_overlap", "on")
		}
		return web.Call(web.BulkCall{Label: job.Comment, Method: http.MethodPost, Path: "/cronjobs/edit", Form: form})
	})
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
