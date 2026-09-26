package crons

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// handleCronjobsLog mirrors cronjobs_log().
func handleCronjobsLog(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	lines := 100
	if v := r.URL.Query().Get("lines"); v != "" {
		if n, parseErr := strconv.Atoi(v); parseErr == nil {
			lines = n
		}
	}
	jobName := r.URL.Query().Get("job")

	if len(jobName) > 50 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Job name too long"})
		return
	}

	body, status := docker.FetchContainerLog(ctx, a, userContext, "cron", lines)
	if status != http.StatusOK {
		writeJSON(w, status, map[string]any{"error": "Failed to fetch container log"})
		return
	}

	var pattern *regexp.Regexp
	if jobName != "" {
		pattern = regexp.MustCompile(`\[Job "` + regexp.QuoteMeta(jobName) + `"`)
	}

	var matched []map[string]string
	for _, line := range strings.Split(body, "\n") {
		if pattern == nil || pattern.MatchString(line) {
			matched = append(matched, map[string]string{"log": line})
			if lines > 0 && len(matched) > lines {
				matched = matched[len(matched)-lines:]
			}
		}
	}
	if matched == nil {
		matched = []map[string]string{}
	}

	writeJSON(w, http.StatusOK, matched)
}

// handleCronjobs is the Cron Jobs tab, ?view=code still works for API callers and old links
func handleCronjobs(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("view") == "code" {
		if r.URL.Query().Get("output") != "json" {
			http.Redirect(w, r, "/cronjobs/editor", http.StatusFound)
			return
		}
		handleCronjobsView(a, w, r, "code")
		return
	}
	handleCronjobsView(a, w, r, "table")
}

// handleCronjobsView renders one of the three tabs: table, code (File Editor) or logs
func handleCronjobsView(a *appctx.App, w http.ResponseWriter, r *http.Request, view string) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	path := cronFilePath(userContext)
	if info, statErr := os.Stat(path); statErr == nil && info.Size() > int64(cronMaxFileSizeBytes(a)) {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Cron file exceeds the " + cronMaxFileSizeKB(a) + " KB limit. Please contact the administrator.",
		})
		return
	}

	if data, readErr := os.ReadFile(path); readErr != nil || !hasCronJobs(strings.Split(string(data), "\n")) {
		stopIdleCron(ctx, userContext)
	}

	switch view {
	case "code":
		content := ""
		if data, readErr := os.ReadFile(path); readErr == nil {
			content = string(data)
		}
		if r.URL.Query().Get("output") == "json" {
			writeJSON(w, http.StatusOK, content)
			return
		}
		renderCronjobsCodePage(a, w, r, content)

	case "table":
		content := ""
		if data, readErr := os.ReadFile(path); readErr == nil {
			content = string(data)
		}
		cronJobs := ParseCronFile(content)

		serviceNames := cronContainers(ctx, userContext)

		if r.URL.Query().Get("output") == "json" {
			writeJSON(w, http.StatusOK, cronJobs)
			return
		}

		var scheduleIssues []ScheduleIssue
		for _, job := range cronJobs {
			if errMsg := validateCronSchedule(job.Schedule); errMsg != "" {
				id := job.Comment
				if id == "" {
					id = job.Schedule
				}
				scheduleIssues = append(scheduleIssues, ScheduleIssue{ID: "cron-schedule:" + id, Severity: "error", Message: errMsg})
			}
		}

		renderCronjobsTablePage(a, w, r, serviceNames, cronJobs, scheduleIssues, cronTimeZone(userContext))

	case "logs":
		content := ""
		if data, readErr := os.ReadFile(path); readErr == nil {
			content = string(data)
		}
		renderCronjobsLogsPage(a, w, r, ParseCronFile(content), r.URL.Query().Get("job"))

	default:
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// handleCronjobsNew mirrors cronjobs_new().
func handleCronjobsNew(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	serviceNames := cronContainers(ctx, userContext)

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, serviceNames)
		return
	}

	phpContainer, dbContainer := "", ""
	if v, ok := docker.GetEnvValue(userContext, "DEFAULT_PHP_VERSION"); ok && containsString(serviceNames, "php-fpm-"+v) {
		phpContainer = "php-fpm-" + v
	}
	if v, ok := docker.GetEnvValue(userContext, "MYSQL_TYPE"); ok && containsString(serviceNames, v) {
		dbContainer = v
	}
	renderCronjobsNewPage(a, w, r, serviceNames, phpContainer, dbContainer)
}

// writeCronFile mirrors write_cron_file().
func writeCronFile(path, content string, truncate bool) error {
	flags := os.O_CREATE | os.O_WRONLY
	if truncate {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}
	f, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(strings.TrimSpace(content) + "\n\n")
	return err
}

// stopIdleCron stops the cron container when there's no enabled job, ofelia exits on an empty schedule and the restart policy would keep bringing it back
func stopIdleCron(ctx context.Context, userContext string) {
	if docker.IsServiceRunning(ctx, userContext, "cron") {
		docker.StartOrStopContainer(ctx, userContext, "cron", "deactivate", "")
	}
}

// restartOrActivateCron mirrors the repeated "activate if not running, else restart" block in save_cronjob(), edit_cronjob(), and delete_cronjob()
func restartOrActivateCron(ctx context.Context, userContext string) {
	if !docker.IsServiceRunning(ctx, userContext, "cron") {
		docker.StartOrStopContainer(ctx, userContext, "cron", "activate", "detached")
	} else {
		docker.ComposeContainer(ctx, userContext, "cron", "restart")
	}
}

// handleSaveCronjob mirrors save_cronjob().
func handleSaveCronjob(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	baseDir, baseErr := filepath.Abs("/home/" + userContext)
	path := cronFilePath(userContext)
	resolvedPath, resolveErr := filepath.Abs(path)
	if baseErr != nil || resolveErr != nil || !strings.HasPrefix(resolvedPath, baseDir) {
		flashAndRedirect(a, w, r, "error", "Invalid cron file path", "/cronjobs/editor")
		return
	}

	_ = r.ParseForm()
	schedule := r.Form.Get("schedule")
	command := r.Form.Get("command")
	container := r.Form.Get("container")
	comment := r.Form.Get("comment")
	autoComment := comment == ""
	if autoComment {
		comment = container
	}
	crontabContent := r.Form.Get("crontab_content")

	if crontabContent == "" {
		if schedule == "" || command == "" || container == "" {
			flashAndRedirect(a, w, r, "error", "Missing one or more required fields (schedule, command, container).", "/cronjobs/new")
			return
		}
		if strings.Contains(schedule, "\n") || strings.Contains(command, "\n") || strings.Contains(container, "\n") || strings.Contains(comment, "\n") {
			flashAndRedirect(a, w, r, "error", "Invalid characters in input.", "/cronjobs/new")
			return
		}
		if _, parseErr := nextCronRuns(schedule, time.Now(), 1, time.UTC); parseErr != nil {
			flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Invalid schedule: %(error)s", "error", parseErr.Error()), "/cronjobs/new")
			return
		}
		if containsAnyPattern(command, forbiddenPatterns) {
			flashAndRedirect(a, w, r, "error", "image= or network= are not allowed in the command.", "/cronjobs/new")
			return
		}
		if containsAnyPattern(command, execPatterns) {
			flashAndRedirect(a, w, r, "error", "job-run, job-local, and job-service-run are not allowed in the command.", "/cronjobs/new")
			return
		}

		truncatedComment := comment
		if len(truncatedComment) > 30 {
			truncatedComment = truncatedComment[:30]
		}
		truncatedComment = strings.ReplaceAll(truncatedComment, `"`, `\"`)

		if autoComment {
			if existingContent, readErr := os.ReadFile(resolvedPath); readErr == nil {
				truncatedComment = uniqueCronComment(ParseCronFile(string(existingContent)), truncatedComment)
			}
		}

		cronJobBlock := "[job-exec \"" + truncatedComment + "\"]\n" +
			"schedule = " + schedule + "\n" +
			"container = " + container + "\n" +
			"command = " + command
		if r.Form.Get("no_overlap") != "" {
			cronJobBlock += "\nno-overlap"
		}

		if writeErr := writeCronFile(resolvedPath, cronJobBlock, false); writeErr != nil {
			flashAndRedirect(a, w, r, "error", "Error saving cron job. Please try again.", "/cronjobs/new")
			return
		}
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "added a new cron job", ipAddress)
		flashSess(a, w, r, "success", "Cron job created and saved successfully!")

		restartOrActivateCron(ctx, userContext)

		http.Redirect(w, r, "/cronjobs", http.StatusFound)
		return
	}

	if containsAnyPattern(crontabContent, forbiddenPatterns) {
		flashAndRedirect(a, w, r, "error", "image= or network= are not allowed in the crons file.", "/cronjobs/editor")
		return
	}
	if containsAnyPattern(crontabContent, execPatterns) {
		flashAndRedirect(a, w, r, "error", "job-run, job-local, and job-service-run are not allowed in the crons file.", "/cronjobs/editor")
		return
	}
	if errMsg := ValidateCronFileFormat(crontabContent); errMsg != "" {
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Invalid crons file format: %(err_msg)s", "err_msg", errMsg), "/cronjobs/editor")
		return
	}

	if writeErr := writeCronFile(resolvedPath, crontabContent, true); writeErr != nil {
		flashAndRedirect(a, w, r, "error", "Error saving cron job. Please try again.", "/cronjobs/editor")
		return
	}
	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "edited cron file", ipAddress)

	if hasCronJobs(strings.Split(crontabContent, "\n")) {
		restartOrActivateCron(ctx, userContext)
	} else {
		docker.StartOrStopContainer(ctx, userContext, "cron", "deactivate", "")
	}

	flashAndRedirect(a, w, r, "success", "Crons file saved successfully!", "/cronjobs/editor")
}

// handleEditCronjob mirrors edit_cronjob().
func handleEditCronjob(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	schedule := r.Form.Get("schedule")
	command := r.Form.Get("command")
	container := r.Form.Get("container")
	comment := r.Form.Get("comment")

	originalSchedule := r.Form.Get("original_schedule")
	originalCommand := r.Form.Get("original_command")
	originalContainer := r.Form.Get("original_container")
	originalComment := r.Form.Get("original_comment")

	if schedule == "" || command == "" || container == "" {
		flashAndRedirect(a, w, r, "error", "Missing one or more required fields.", "/cronjobs")
		return
	}
	if originalSchedule == "" || originalCommand == "" || originalContainer == "" || originalComment == "" {
		flashAndRedirect(a, w, r, "error", "Missing original cron job fields.", "/cronjobs")
		return
	}
	if strings.Contains(container, "\n") || strings.Contains(comment, "\n") || strings.Contains(schedule, "\n") {
		flashAndRedirect(a, w, r, "error", "Invalid characters in input.", "/cronjobs")
		return
	}

	baseDir, baseErr := filepath.Abs("/home/" + userContext)
	resolvedPath, resolveErr := filepath.Abs(cronFilePath(userContext))
	if baseErr != nil || resolveErr != nil || !strings.HasPrefix(resolvedPath, baseDir) {
		flashAndRedirect(a, w, r, "error", "Invalid cron file path.", "/cronjobs")
		return
	}

	content, readErr := os.ReadFile(resolvedPath)
	if readErr != nil {
		flashAndRedirect(a, w, r, "error", "Error saving cron job. Please try again.", "/cronjobs")
		return
	}

	updatedComment := comment
	if updatedComment == "" {
		updatedComment = originalComment
	}
	// a disabled job stays disabled after an edit
	newContent, updated := rewriteCronJob(string(content), func(j *CronJob) (bool, bool) {
		if !sameCronJob(*j, originalComment, originalSchedule, originalContainer, originalCommand) {
			return false, true
		}
		*j = CronJob{Comment: updatedComment, Schedule: schedule, Container: container, Command: command, NoOverlap: r.Form.Get("no_overlap") != "", Disabled: j.Disabled}
		return true, true
	})

	if !updated {
		flashAndRedirect(a, w, r, "error", "Cron job not found.", "/cronjobs")
		return
	}

	if writeErr := os.WriteFile(resolvedPath, []byte(newContent), 0o644); writeErr != nil {
		flashAndRedirect(a, w, r, "error", "Error saving cron job. Please try again.", "/cronjobs")
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "edited cron job", ipAddress)
	restartOrActivateCron(ctx, userContext)

	flashAndRedirect(a, w, r, "success", "Cron job was successfully edited.", "/cronjobs")
}

// handleDeleteCronjob mirrors delete_cronjob().
func handleDeleteCronjob(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	command := r.Form.Get("command")
	schedule := r.Form.Get("schedule")
	container := r.Form.Get("container")
	comment := r.Form.Get("comment")

	if schedule == "" || command == "" || container == "" || comment == "" {
		flashAndRedirect(a, w, r, "error", "Missing one or more required fields (schedule, command, container).", "/cronjobs")
		return
	}

	path := cronFilePath(userContext)
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		flashAndRedirect(a, w, r, "error", "Error deleting cron job.", "/cronjobs")
		return
	}
	newContent, _ := rewriteCronJob(string(content), func(j *CronJob) (bool, bool) {
		return sameCronJob(*j, comment, schedule, container, command), false
	})
	newLines := strings.Split(newContent, "\n")

	if writeErr := os.WriteFile(path, []byte(newContent), 0o644); writeErr != nil {
		flashAndRedirect(a, w, r, "error", "Error deleting cron job.", "/cronjobs")
		return
	}

	ipAddress := reqip.ClientIP(r)
	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted a cron job", ipAddress)

	if hasCronJobs(newLines) {
		restartOrActivateCron(ctx, userContext)
	} else {
		docker.StartOrStopContainer(ctx, userContext, "cron", "deactivate", "")
	}

	flashAndRedirect(a, w, r, "success", "Cron job was successfully deleted.", "/cronjobs")
}
