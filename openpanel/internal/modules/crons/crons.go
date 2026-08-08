// Package crons manages the per-user crons.ini file (dodo/go-cron
// "job-exec" blocks executed inside a container), its table and raw-file
// editor views, and the log viewer for the shared cron container.
package crons

import (
	"fmt"
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
	NoOverlap bool
}

// ScheduleIssue is one entry of cronjobs.html's health_toast() issues list
// (one per invalid-schedule cron job, matching cron_schedule_issues).
type ScheduleIssue struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

var cronJobHeaderRE = regexp.MustCompile(`\[job-exec\s+"([^"]*)"\]`)

// cronBlockSplitRE splits crons.ini content on one-or-more blank lines
// (the "empty row between them" separator the GUI requires between
// [job-exec] blocks).
var cronBlockSplitRE = regexp.MustCompile(`\r?\n[ \t]*\r?\n[ \t\r\n]*`)

// splitCronBlocks splits raw crons.ini content into trimmed, non-empty
// [job-exec] blocks separated by blank lines.
func splitCronBlocks(content string) []string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil
	}
	var blocks []string
	for _, b := range cronBlockSplitRE.Split(trimmed, -1) {
		b = strings.TrimSpace(b)
		if b != "" {
			blocks = append(blocks, b)
		}
	}
	return blocks
}

// splitKV splits a "key = value" (or "key=value") line on its first "=".
func splitKV(line string) (key, val string, ok bool) {
	idx := strings.Index(line, "=")
	if idx == -1 {
		return "", "", false
	}
	return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
}

// ParseCronFile mirrors parse_cron_file(). It's intentionally tolerant of
// key order and extra whitespace so existing/legacy crons.ini files keep
// rendering even if they don't match the stricter format enforced by
// ValidateCronFileFormat for new saves from the raw editor.
func ParseCronFile(content string) []CronJob {
	var jobs []CronJob
	for _, block := range splitCronBlocks(content) {
		headerMatch := cronJobHeaderRE.FindStringSubmatch(block)
		if headerMatch == nil {
			continue
		}
		job := CronJob{Comment: strings.TrimSpace(headerMatch[1])}
		for _, rawLine := range strings.Split(block, "\n") {
			line := strings.TrimSpace(rawLine)
			if line == "no-overlap" {
				job.NoOverlap = true
				continue
			}
			key, val, ok := splitKV(line)
			if !ok {
				continue
			}
			switch key {
			case "schedule":
				job.Schedule = val
			case "container":
				job.Container = val
			case "command":
				job.Command = val
			}
		}
		jobs = append(jobs, job)
	}
	return jobs
}

// uniqueCronComment returns base unchanged if no job in existing already
// uses it as its comment, otherwise it appends "-1", "-2", etc. until it
// finds a name that isn't taken. Used when a new job's comment is left
// empty and defaults to the container name, so scheduling several jobs
// against the same container doesn't silently collide (e.g. "apache",
// "apache-1", "apache-2").
func uniqueCronComment(existing []CronJob, base string) string {
	taken := make(map[string]bool, len(existing))
	for _, j := range existing {
		taken[j.Comment] = true
	}
	if !taken[base] {
		return base
	}
	for n := 1; ; n++ {
		candidate := fmt.Sprintf("%s-%d", base, n)
		if !taken[candidate] {
			return candidate
		}
	}
}

// ValidateCronFileFormat enforces the format required by the table/GUI view
// for content saved from the raw file editor (?view=code):
//
//	[job-exec "name"]
//	schedule = ...
//	container = ...
//	command = ...
//	no-overlap   (optional)
//
// with exactly one blank line separating each block. Returns "" when the
// content is well-formed (or empty), otherwise a human-readable message
// describing the first problem found.
func ValidateCronFileFormat(content string) string {
	for i, block := range splitCronBlocks(content) {
		lines := strings.Split(block, "\n")
		for li := range lines {
			lines[li] = strings.TrimSpace(lines[li])
		}

		jobLabel := fmt.Sprintf("Job #%d", i+1)

		headerMatch := cronJobHeaderRE.FindStringSubmatch(lines[0])
		if headerMatch == nil || !strings.HasPrefix(lines[0], "[job-exec") || !strings.HasSuffix(lines[0], "]") {
			return fmt.Sprintf(`%s: expected a [job-exec "name"] header, got %q.`, jobLabel, lines[0])
		}
		name := strings.TrimSpace(headerMatch[1])
		if name == "" {
			return fmt.Sprintf(`%s: [job-exec "..."] name cannot be empty.`, jobLabel)
		}
		jobLabel = fmt.Sprintf(`Job "%s"`, name)

		if len(lines) < 4 {
			return fmt.Sprintf(`%s: expected schedule, container, and command lines after the header.`, jobLabel)
		}

		for idx, expectedKey := range []string{"schedule", "container", "command"} {
			key, val, ok := splitKV(lines[idx+1])
			if !ok || key != expectedKey || val == "" {
				return fmt.Sprintf(`%s: expected "%s = ..." on line %d, got %q.`, jobLabel, expectedKey, idx+2, lines[idx+1])
			}
		}

		switch {
		case len(lines) == 5 && lines[4] != "no-overlap":
			return fmt.Sprintf(`%s: unexpected line %q — only "no-overlap" is allowed after command.`, jobLabel, lines[4])
		case len(lines) > 5:
			return fmt.Sprintf(`%s: too many lines in this block — make sure exactly one empty line separates each [job-exec] block.`, jobLabel)
		}
	}
	return ""
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
