// Package crons manages the per-user crons.ini file (dodo/go-cron
// "job-exec" blocks executed inside a container), its table and raw-file
// editor views, and the log viewer for the shared cron container.
package crons

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

var excludedServicesForCrons = map[string]bool{"cron": true, "docker-proxy": true}

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

func flashSess(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

// cronMaxFileSizeBytes is derived from cron_max_file_size_kb (default
// 100 KB). a.Config is loaded once at process startup.
func cronMaxFileSizeBytes(a *appctx.App) int {
	kb, err := strconv.Atoi(a.Config.Get("cron_max_file_size_kb", "100"))
	if err != nil {
		kb = 100
	}
	return kb * 1024
}

func cronMaxFileSizeKB(a *appctx.App) string {
	return a.Config.Get("cron_max_file_size_kb", "100")
}

func cronFilePath(userContext string) string {
	return "/home/" + userContext + "/crons.ini"
}

// hasCronJobs mirrors has_cron_jobs().
func hasCronJobs(lines []string) bool {
	for _, line := range lines {
		if strings.HasPrefix(line, "[job-exec") {
			return true
		}
	}
	return false
}

// validateCronSchedule mirrors validate_cron_schedule(): the Go cron
// implementation used (robfig/cron, WithSeconds()) needs 6 numeric fields
// (seconds minutes hours day month weekday), not the usual 5.
// @every/@daily-style descriptors are passed through as-is. Returns "" when
// valid.
func validateCronSchedule(schedule string) string {
	schedule = strings.TrimSpace(schedule)
	if schedule == "" {
		return "Schedule is empty."
	}
	if strings.HasPrefix(schedule, "@") {
		return ""
	}
	parts := strings.Fields(schedule)
	switch len(parts) {
	case 6:
		return ""
	case 5:
		return `Schedule "` + schedule + `" has only 5 fields, but this cron implementation requires 6 (seconds minutes hours day month weekday). Add one more field for seconds at the start.`
	default:
		return `Schedule "` + schedule + `" is invalid - expected 6 fields (seconds minutes hours day month weekday) or an @every/@daily-style descriptor.`
	}
}

// forbiddenPatterns/execPatterns mirror save_cronjob()'s forbidden_patterns
// and exec_patterns (case-insensitive, word-boundary matches).
var (
	forbiddenPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bimage\s*=`),
		regexp.MustCompile(`(?i)\bnetwork\s*=`),
	}
	execPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bjob-run`),
		regexp.MustCompile(`(?i)\bjob-local`),
		regexp.MustCompile(`(?i)\bjob-service-run`),
	}
)

func containsAnyPattern(text string, patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.MatchString(text) {
			return true
		}
	}
	return false
}

// CronJob is one parsed [job-exec] block.
type CronJob struct {
	Comment   string
	Schedule  string
	Container string
	Command   string
}

// ScheduleIssue is one entry of cronjobs.html's health_toast() issues list
// (one per invalid-schedule cron job, matching cron_schedule_issues).
type ScheduleIssue struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

var cronJobBlockRE = regexp.MustCompile(`\[job-exec\s+"([^"]+)"\]\s*schedule\s*=\s*(.*?)\s*container\s*=\s*([^\s]+)\s*command\s*=\s*([^\n]+)\s*`)

// ParseCronFile mirrors parse_cron_file().
func ParseCronFile(content string) []CronJob {
	var jobs []CronJob
	for _, m := range cronJobBlockRE.FindAllStringSubmatch(content, -1) {
		jobs = append(jobs, CronJob{
			Comment:   strings.TrimSpace(m[1]),
			Schedule:  strings.TrimSpace(m[2]),
			Container: strings.TrimSpace(m[3]),
			Command:   strings.TrimSpace(m[4]),
		})
	}
	return jobs
}

// serviceNames mirrors the containers = load_compose_config(context)["services"]
// keys-minus-excluded pattern shared by cronjobs()/cronjobs_new().
func serviceNamesFromCompose(compose map[string]any) ([]string, bool) {
	servicesRaw, ok := compose["services"]
	if !ok {
		return nil, false
	}
	services, ok := servicesRaw.(map[string]any)
	if !ok {
		return nil, false
	}
	names := make([]string, 0, len(services))
	for name := range services {
		if !excludedServicesForCrons[name] {
			names = append(names, name)
		}
	}
	return names, true
}
