package crons

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// RegisterAPI wires the crons API routes onto mux.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "crons", "GET /api/crons", func(w http.ResponseWriter, r *http.Request) { apiCronsList(a, w, r) })
	apiregistry.Handle(mux, a, "crons", "POST /api/crons", func(w http.ResponseWriter, r *http.Request) { apiCronsCreate(a, w, r) })
	apiregistry.Handle(mux, a, "crons", "PATCH /api/crons", func(w http.ResponseWriter, r *http.Request) { apiCronsEdit(a, w, r) })
	apiregistry.Handle(mux, a, "crons", "DELETE /api/crons", func(w http.ResponseWriter, r *http.Request) { apiCronsDelete(a, w, r) })
	apiregistry.Handle(mux, a, "crons", "GET /api/crons/raw", func(w http.ResponseWriter, r *http.Request) { apiCronsRawGet(a, w, r) })
	apiregistry.Handle(mux, a, "crons", "PUT /api/crons/raw", func(w http.ResponseWriter, r *http.Request) { apiCronsRawPut(a, w, r) })
	apiregistry.Handle(mux, a, "crons", "GET /api/crons/log", func(w http.ResponseWriter, r *http.Request) { apiCronsLog(a, w, r) })
}

func writeAPICronsJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiCronPath resolves and validates the per-user crons.ini path, mirroring
// the repeated `Path(f'/home/{context}/crons.ini').resolve()` +
// base-dir-prefix guard shared by several routes.
func apiCronPath(userContext string) (path string, ok bool) {
	baseDir, baseErr := filepath.Abs("/home/" + userContext)
	resolvedPath, resolveErr := filepath.Abs(cronFilePath(userContext))
	if baseErr != nil || resolveErr != nil || !strings.HasPrefix(resolvedPath, baseDir) {
		return "", false
	}
	return resolvedPath, true
}

// apiCronsList mirrors api_crons_list().
func apiCronsList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	path := cronFilePath(userContext)
	info, statErr := os.Stat(path)
	if statErr != nil {
		var serviceNames []string
		if compose, composeErr := podmanmanager.LoadComposeConfig(ctx, userContext); composeErr == nil {
			serviceNames, _ = serviceNamesFromCompose(compose)
		}
		writeAPICronsJSON(w, http.StatusOK, map[string]any{"jobs": []crJob{}, "containers": serviceNames})
		return
	}

	if info.Size() > int64(cronMaxFileSizeBytes(a)) {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "Cron file exceeds the " + cronMaxFileSizeKB(a) + " KB limit"})
		return
	}

	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeAPICronsJSON(w, http.StatusOK, map[string]any{"jobs": []crJob{}, "containers": []string{}})
		return
	}
	jobs := ParseCronFile(string(content))

	var serviceNames []string
	if compose, composeErr := podmanmanager.LoadComposeConfig(ctx, userContext); composeErr == nil {
		serviceNames, _ = serviceNamesFromCompose(compose)
	}

	type scheduleIssue struct {
		Comment string `json:"comment"`
		Error   string `json:"error"`
	}
	issues := []scheduleIssue{}
	for _, job := range jobs {
		if errMsg := validateCronSchedule(job.Schedule); errMsg != "" {
			issues = append(issues, scheduleIssue{Comment: job.Comment, Error: errMsg})
		}
	}

	writeAPICronsJSON(w, http.StatusOK, map[string]any{
		"jobs": toAPIJobs(jobs), "containers": serviceNames, "schedule_issues": issues,
	})
}

// crJob is the API's job JSON shape (matches _parse_cron_file()'s dict).
type crJob struct {
	Comment   string `json:"comment"`
	Schedule  string `json:"schedule"`
	Container string `json:"container"`
	Command   string `json:"command"`
	NoOverlap bool   `json:"no_overlap"`
}

func toAPIJobs(jobs []CronJob) []crJob {
	out := make([]crJob, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, crJob{Comment: j.Comment, Schedule: j.Schedule, Container: j.Container, Command: j.Command, NoOverlap: j.NoOverlap})
	}
	return out
}

// apiCronsRawGet mirrors api_crons_raw()'s GET branch.
func apiCronsRawGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	path, ok := apiCronPath(userContext)
	if !ok {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid cron file path"})
		return
	}

	content := ""
	if data, readErr := os.ReadFile(path); readErr == nil {
		content = string(data)
	}
	writeAPICronsJSON(w, http.StatusOK, map[string]string{"content": content})
}

// apiCronsRawPut mirrors api_crons_raw()'s PUT branch.
func apiCronsRawPut(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	path, ok := apiCronPath(userContext)
	if !ok {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid cron file path"})
		return
	}

	var body struct {
		Content *string `json:"content"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Content == nil {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "content is required"})
		return
	}
	content := *body.Content

	if containsAnyPattern(content, forbiddenPatterns) {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "image= or network= are not allowed in crontab"})
		return
	}
	if containsAnyPattern(content, execPatterns) {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "job-run, job-local, and job-service-run are not allowed in crontab"})
		return
	}

	if writeErr := writeCronFile(path, content, true); writeErr != nil {
		writeAPICronsJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "edited cron file (raw)", reqip.ClientIP(r))
	if hasCronJobs(strings.Split(content, "\n")) {
		restartOrActivateCron(ctx, userContext)
	} else {
		docker.StartOrStopContainer(ctx, userContext, "cron", "deactivate", "")
	}

	writeAPICronsJSON(w, http.StatusOK, map[string]string{"message": "Cron file saved"})
}

// apiCronsCreate mirrors api_crons_create().
func apiCronsCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	path, ok := apiCronPath(userContext)
	if !ok {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid cron file path"})
		return
	}

	var body struct {
		Schedule  string `json:"schedule"`
		Command   string `json:"command"`
		Container string `json:"container"`
		Comment   string `json:"comment"`
		NoOverlap bool   `json:"no_overlap"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	schedule := strings.TrimSpace(body.Schedule)
	command := strings.TrimSpace(body.Command)
	container := strings.TrimSpace(body.Container)
	comment := strings.TrimSpace(body.Comment)
	autoComment := comment == ""
	if autoComment {
		comment = container
	}

	if schedule == "" || command == "" || container == "" {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "schedule, command, and container are required"})
		return
	}
	if errMsg := validateCronSchedule(schedule); errMsg != "" {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": errMsg})
		return
	}
	if containsAnyPattern(command, forbiddenPatterns) {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "image= or network= are not allowed in command"})
		return
	}
	if strings.Contains(schedule, "\n") || strings.Contains(command, "\n") || strings.Contains(container, "\n") || strings.Contains(comment, "\n") {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "Newlines are not allowed in fields"})
		return
	}

	if autoComment {
		if existingContent, readErr := os.ReadFile(path); readErr == nil {
			comment = uniqueCronComment(ParseCronFile(string(existingContent)), comment)
		}
	}

	truncatedComment := comment
	if len(truncatedComment) > 30 {
		truncatedComment = truncatedComment[:30]
	}
	truncatedComment = strings.ReplaceAll(truncatedComment, `"`, `\"`)

	cronJobBlock := "[job-exec \"" + truncatedComment + "\"]\n" +
		"schedule = " + schedule + "\n" +
		"container = " + container + "\n" +
		"command = " + command
	if body.NoOverlap {
		cronJobBlock += "\nno-overlap"
	}

	if writeErr := writeCronFile(path, cronJobBlock, false); writeErr != nil {
		writeAPICronsJSON(w, http.StatusInternalServerError, map[string]string{"error": writeErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "added cron job: "+comment, reqip.ClientIP(r))
	restartOrActivateCron(ctx, userContext)

	writeAPICronsJSON(w, http.StatusCreated, map[string]any{
		"message": "Cron job created",
		"job":     crJob{Comment: comment, Schedule: schedule, Container: container, Command: command, NoOverlap: body.NoOverlap},
	})
}

// apiCronsEdit mirrors api_crons_edit().
func apiCronsEdit(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	path, ok := apiCronPath(userContext)
	if !ok {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid cron file path"})
		return
	}

	var body struct {
		Schedule          string `json:"schedule"`
		Command           string `json:"command"`
		Container         string `json:"container"`
		Comment           string `json:"comment"`
		NoOverlap         bool   `json:"no_overlap"`
		OriginalSchedule  string `json:"original_schedule"`
		OriginalCommand   string `json:"original_command"`
		OriginalContainer string `json:"original_container"`
		OriginalComment   string `json:"original_comment"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	schedule := strings.TrimSpace(body.Schedule)
	command := strings.TrimSpace(body.Command)
	container := strings.TrimSpace(body.Container)
	comment := strings.TrimSpace(body.Comment)
	originalSchedule := strings.TrimSpace(body.OriginalSchedule)
	originalCommand := strings.TrimSpace(body.OriginalCommand)
	originalContainer := strings.TrimSpace(body.OriginalContainer)
	originalComment := strings.TrimSpace(body.OriginalComment)

	if schedule == "" || command == "" || container == "" {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "schedule, command, and container are required"})
		return
	}
	if originalSchedule == "" || originalCommand == "" || originalContainer == "" || originalComment == "" {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "original_schedule, original_command, original_container, original_comment are required"})
		return
	}
	if errMsg := validateCronSchedule(schedule); errMsg != "" {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": errMsg})
		return
	}
	if strings.Contains(container, "\n") || strings.Contains(comment, "\n") || strings.Contains(schedule, "\n") {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "Newlines are not allowed in fields"})
		return
	}

	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeAPICronsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error saving cron job: " + readErr.Error()})
		return
	}

	sections := splitJobExecSections(string(content))
	updatedComment := comment
	if updatedComment == "" {
		updatedComment = originalComment
	}
	updatedSection := "[job-exec \"" + updatedComment + "\"]\n" +
		"schedule = " + schedule + "\n" +
		"container = " + container + "\n" +
		"command = " + command
	if body.NoOverlap {
		updatedSection += "\nno-overlap"
	}
	updatedSection += "\n\n"

	updated := false
	for i, section := range sections {
		if strings.Contains(section, `[job-exec "`+originalComment+`"]`) &&
			strings.Contains(section, "schedule = "+originalSchedule) &&
			strings.Contains(section, "container = "+originalContainer) &&
			strings.Contains(section, "command = "+originalCommand) {
			sections[i] = updatedSection
			updated = true
			break
		}
	}
	if !updated {
		writeAPICronsJSON(w, http.StatusNotFound, map[string]string{"error": "Cron job not found — check original_* fields match exactly"})
		return
	}

	if writeErr := os.WriteFile(path, []byte(strings.Join(sections, "")), 0o644); writeErr != nil {
		writeAPICronsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error saving cron job: " + writeErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "edited cron job: "+originalComment, reqip.ClientIP(r))
	restartOrActivateCron(ctx, userContext)
	writeAPICronsJSON(w, http.StatusOK, map[string]string{"message": "Cron job updated"})
}

// apiCronsDelete mirrors api_crons_delete().
func apiCronsDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	path, ok := apiCronPath(userContext)
	if !ok {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid cron file path"})
		return
	}

	var body struct {
		Schedule  string `json:"schedule"`
		Command   string `json:"command"`
		Container string `json:"container"`
		Comment   string `json:"comment"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	schedule := strings.TrimSpace(body.Schedule)
	command := strings.TrimSpace(body.Command)
	container := strings.TrimSpace(body.Container)
	comment := strings.TrimSpace(body.Comment)

	if schedule == "" || command == "" || container == "" || comment == "" {
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "schedule, command, container, and comment are required"})
		return
	}

	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeAPICronsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error deleting cron job: " + readErr.Error()})
		return
	}
	lines := readLinesKeepEnds(string(content))

	sectionMatches := func(section []string) bool {
		if len(section) == 0 || !strings.HasPrefix(section[0], "[job-exec") {
			return false
		}
		headerComment, ok := sectionHeaderComment(section[0])
		if !ok {
			return false
		}
		sectionStr := strings.Join(section, "")
		return headerComment == comment &&
			strings.Contains(sectionStr, "schedule = "+schedule) &&
			strings.Contains(sectionStr, "command = "+command) &&
			strings.Contains(sectionStr, "container = "+container)
	}

	var newLines []string
	var currentSection []string
	isMatchingSection := false
	for _, line := range lines {
		if strings.HasPrefix(line, "[job-exec") {
			if len(currentSection) > 0 {
				if !isMatchingSection {
					newLines = append(newLines, currentSection...)
				}
				currentSection = nil
				isMatchingSection = false
			}
		}
		currentSection = append(currentSection, line)
		if strings.TrimSpace(line) == "" {
			isMatchingSection = sectionMatches(currentSection)
		}
	}
	if len(currentSection) > 0 && !sectionMatches(currentSection) {
		newLines = append(newLines, currentSection...)
	}

	if writeErr := os.WriteFile(path, []byte(strings.Join(newLines, "")), 0o644); writeErr != nil {
		writeAPICronsJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error deleting cron job: " + writeErr.Error()})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "deleted cron job: "+comment, reqip.ClientIP(r))
	if hasCronJobs(newLines) {
		restartOrActivateCron(ctx, userContext)
	} else {
		docker.StartOrStopContainer(ctx, userContext, "cron", "deactivate", "")
	}

	writeAPICronsJSON(w, http.StatusOK, map[string]string{"message": "Cron job \"" + comment + "\" deleted"})
}

// apiCronsLog mirrors api_crons_log().
func apiCronsLog(a *appctx.App, w http.ResponseWriter, r *http.Request) {
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
		writeAPICronsJSON(w, http.StatusBadRequest, map[string]string{"error": "Job name too long"})
		return
	}

	body, status := docker.FetchContainerLog(ctx, a, userContext, "cron", lines)
	if status != http.StatusOK {
		writeAPICronsJSON(w, status, map[string]string{"error": "Failed to fetch log: " + body})
		return
	}

	needle := ""
	if jobName != "" {
		needle = `[Job "` + jobName + `"`
	}

	var matched []map[string]string
	for _, line := range strings.Split(body, "\n") {
		if needle == "" || strings.Contains(line, needle) {
			matched = append(matched, map[string]string{"log": line})
			if len(matched) > lines {
				matched = matched[len(matched)-lines:]
			}
		}
	}
	if matched == nil {
		matched = []map[string]string{}
	}

	writeAPICronsJSON(w, http.StatusOK, map[string]any{"entries": matched})
}
