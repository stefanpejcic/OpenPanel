// Package crons manages the per-user crons.ini file (dodo/go-cron "job-exec" blocks executed inside a container), its table and raw-file editor views, and the log viewer for the shared cron container.
package crons

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
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

// cronMaxFileSizeBytes is derived from cron_max_file_size_kb (default 100 KB); a.Config is loaded once at process startup
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

// validateCronSchedule mirrors validate_cron_schedule(): the Go cron implementation used (robfig/cron, WithSeconds()) needs 6 numeric fields, not the usual 5. @every/@daily-style descriptors pass through as-is. Returns "" when valid.
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

// forbiddenPatterns/execPatterns mirror save_cronjob()'s forbidden_patterns and exec_patterns (case-insensitive, word-boundary matches)
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
	Disabled  bool // the block is commented out with #, cron skips it
}

// ScheduleIssue is one entry of cronjobs.html's health_toast() issues list (one per invalid-schedule cron job, matching cron_schedule_issues)
type ScheduleIssue struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

var cronJobHeaderRE = regexp.MustCompile(`\[job-exec\s+"([^"]*)"\]`)

// cronBlockSplitRE splits crons.ini content on one-or-more blank lines (the "empty row between them" separator the GUI requires between [job-exec] blocks)
var cronBlockSplitRE = regexp.MustCompile(`\r?\n[ \t]*\r?\n[ \t\r\n]*`)

// splitCronBlocks splits raw crons.ini content into trimmed, non-empty [job-exec] blocks separated by blank lines
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

// ParseCronFile mirrors parse_cron_file(). Intentionally tolerant of key order and extra whitespace so existing/legacy crons.ini files keep rendering even if they don't match the stricter format ValidateCronFileFormat enforces for new saves.
func ParseCronFile(content string) []CronJob {
	var jobs []CronJob
	for _, block := range splitCronBlocks(content) {
		disabled := false
		if plain, ok := uncommentBlock(block); ok {
			block, disabled = plain, true
		}
		headerMatch := cronJobHeaderRE.FindStringSubmatch(block)
		if headerMatch == nil {
			continue
		}
		job := CronJob{Comment: strings.TrimSpace(headerMatch[1]), Disabled: disabled}
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

// uncommentBlock strips the # from every line of a disabled job's block, ok is false when any line isn't commented
func uncommentBlock(block string) (string, bool) {
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			return block, false
		}
		lines[i] = strings.TrimSpace(strings.TrimPrefix(line, "#"))
	}
	return strings.Join(lines, "\n"), true
}

// cronBlock is the crons.ini block for j, every line commented out when it's disabled
func cronBlock(j CronJob) string {
	lines := []string{`[job-exec "` + j.Comment + `"]`, "schedule = " + j.Schedule, "container = " + j.Container, "command = " + j.Command}
	if j.NoOverlap {
		lines = append(lines, "no-overlap")
	}
	if j.Disabled {
		for i := range lines {
			lines[i] = "#" + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}

// rewriteCronJob runs change on each job until it reports a match, keep=false deletes that job, blocks that aren't jobs are left as they are
func rewriteCronJob(content string, change func(j *CronJob) (matched, keep bool)) (string, bool) {
	var out []string
	done := false
	for _, block := range splitCronBlocks(content) {
		if jobs := ParseCronFile(block); !done && len(jobs) == 1 {
			j := jobs[0]
			if matched, keep := change(&j); matched {
				done = true
				if keep {
					out = append(out, cronBlock(j))
				}
				continue
			}
		}
		out = append(out, block)
	}
	if len(out) == 0 {
		return "", done
	}
	return strings.Join(out, "\n\n") + "\n\n", done
}

// sameCronJob matches the original_* fields the edit and delete forms send
func sameCronJob(j CronJob, comment, schedule, container, command string) bool {
	return j.Comment == comment && j.Schedule == schedule && j.Container == container && j.Command == command
}

// uniqueCronComment returns base unchanged if no job in existing already uses it, otherwise appends "-1", "-2", etc. until it finds one that's free - used when a new job's comment defaults to the container name, so several jobs on the same container don't collide (e.g. "apache", "apache-1", "apache-2")
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

// ValidateCronFileFormat enforces the format required by the table/GUI view for content saved from the raw file editor (?view=code):
//
//	[job-exec "name"]
//	schedule = ...
//	container = ...
//	command = ...
//	no-overlap   (optional)
//
// with exactly one blank line separating each block. Returns "" if well-formed (or empty), otherwise a message describing the first problem found.
func ValidateCronFileFormat(content string) string {
	for i, block := range splitCronBlocks(content) {
		// a disabled job is checked like an enabled one
		if plain, ok := uncommentBlock(block); ok {
			block = plain
		}
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

// serviceNamesFromCompose lists the compose services a job can run in, filter trims them to the user's current web server and database type
func serviceNamesFromCompose(compose map[string]any, filter func(map[string]any) map[string]any) ([]string, bool) {
	servicesRaw, ok := compose["services"]
	if !ok {
		return nil, false
	}
	services, ok := servicesRaw.(map[string]any)
	if !ok {
		return nil, false
	}
	if filter != nil {
		services = filter(services)
	}
	names := make([]string, 0, len(services))
	for name := range services {
		if !excludedServicesForCrons[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, true
}

// cronContainers is the container list for the new and edit forms and the bulk bar
func cronContainers(ctx context.Context, userContext string) []string {
	compose, err := podmanmanager.LoadComposeConfig(ctx, userContext)
	if err != nil {
		return nil
	}
	names, _ := serviceNamesFromCompose(compose, func(services map[string]any) map[string]any {
		return docker.FilterUserServices(userContext, services)
	})
	return names
}
