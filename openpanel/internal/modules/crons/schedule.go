package crons

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// cronParser reads schedules the way ofelia (robfig/cron v3) does: seconds first, or an @descriptor
var cronParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// nextCronRuns lists the next n run times of schedule after from
func nextCronRuns(schedule string, from time.Time, n int, loc *time.Location) ([]time.Time, error) {
	if msg := validateCronSchedule(schedule); msg != "" {
		return nil, errString(msg)
	}
	sched, err := cronParser.Parse(strings.TrimSpace(schedule))
	if err != nil {
		return nil, err
	}
	runs := make([]time.Time, 0, n)
	t := from.In(loc)
	for len(runs) < n {
		t = sched.Next(t)
		if t.IsZero() {
			break
		}
		runs = append(runs, t)
	}
	return runs, nil
}

type errString string

func (e errString) Error() string { return string(e) }

// CronSummary is the stats row above the cron jobs table
type CronSummary struct {
	Total    int
	Active   int
	NextRun  string // RFC3339, "" when nothing is scheduled
	NextJob  string
	TimeZone string
}

func cronSummary(jobs []CronJob, now time.Time, loc *time.Location) CronSummary {
	sum := CronSummary{Total: len(jobs), TimeZone: loc.String()}
	var soonest time.Time
	for _, j := range jobs {
		if j.Disabled {
			continue
		}
		sum.Active++
		runs, err := nextCronRuns(j.Schedule, now, 1, loc)
		if err != nil || len(runs) == 0 {
			continue
		}
		if soonest.IsZero() || runs[0].Before(soonest) {
			soonest, sum.NextJob = runs[0], j.Comment
		}
	}
	if !soonest.IsZero() {
		sum.NextRun = soonest.Format(time.RFC3339)
	}
	return sum
}

// cronWeekdays are the day names describeSchedule uses, index 0 is Sunday like cron's weekday field
var cronWeekdays = []string{"Sundays", "Mondays", "Tuesdays", "Wednesdays", "Thursdays", "Fridays", "Saturdays"}

// describeSchedule says a schedule in words for the jobs table, "" when there's no short way to say it
func describeSchedule(t i18n.Translator, schedule string) string {
	s := strings.TrimSpace(schedule)
	switch s {
	case "@yearly", "@annually":
		return t.Get("Once a year")
	case "@monthly":
		return t.Get("Once a month")
	case "@weekly":
		return t.Get("Once a week")
	case "@daily", "@midnight":
		return t.Get("Every day at %(time)s", "time", "00:00")
	case "@hourly":
		return t.Get("Every hour")
	}
	if rest, ok := strings.CutPrefix(s, "@every "); ok {
		return t.Get("Every %(interval)s", "interval", strings.TrimSpace(rest))
	}
	f := strings.Fields(s)
	if len(f) != 6 || f[4] != "*" {
		return ""
	}
	sec, minute, hour, dom, dow := f[0], f[1], f[2], f[3], f[5]
	num := func(v string, max int) (int, bool) {
		n, err := strconv.Atoi(v)
		return n, err == nil && n >= 0 && n <= max
	}
	step := func(v string, max int) string {
		if rest, ok := strings.CutPrefix(v, "*/"); ok {
			if n, ok := num(rest, max); ok && n > 1 {
				return rest
			}
		}
		return ""
	}
	everyDay := dom == "*" && dow == "*"
	if everyDay {
		if n := step(sec, 59); n != "" && minute == "*" && hour == "*" {
			return t.Get("Every %(n)s seconds", "n", n)
		}
		if _, ok := num(sec, 59); !ok {
			return ""
		}
		if minute == "*" && hour == "*" {
			return t.Get("Every minute")
		}
		if n := step(minute, 59); n != "" && hour == "*" {
			return t.Get("Every %(n)s minutes", "n", n)
		}
		if m, ok := num(minute, 59); ok && hour == "*" {
			return t.Get("Every hour at minute %(minute)s", "minute", strconv.Itoa(m))
		}
		if n := step(hour, 23); n != "" {
			if _, ok := num(minute, 59); ok {
				return t.Get("Every %(n)s hours", "n", n)
			}
		}
	}
	sn, ok1 := num(sec, 59)
	mn, ok2 := num(minute, 59)
	hn, ok3 := num(hour, 23)
	if !ok1 || !ok2 || !ok3 {
		return ""
	}
	at := fmt.Sprintf("%02d:%02d", hn, mn)
	if sn != 0 {
		at += fmt.Sprintf(":%02d", sn)
	}
	if everyDay {
		return t.Get("Every day at %(time)s", "time", at)
	}
	if d, ok := num(dow, 7); ok && dom == "*" {
		return t.Get("%(day)s at %(time)s", "day", t.Get(cronWeekdays[d%7]), "time", at)
	}
	if d, ok := num(dom, 31); ok && d > 0 && dow == "*" {
		return t.Get("Monthly on day %(day)s at %(time)s", "day", strconv.Itoa(d), "time", at)
	}
	return ""
}

// handleToggleCronjob turns one job on or off by commenting its block out, then restarts cron to pick it up
func handleToggleCronjob(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	key := r.Form.Get("key")
	enable := r.Form.Get("enabled") == "1"

	path := cronFilePath(userContext)
	content, readErr := tenantfs.ReadFile(path)
	if readErr != nil {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Error saving cron job. Please try again."), "/cronjobs")
		return
	}
	var name string
	newContent, found := rewriteCronJob(string(content), func(j *CronJob) (bool, bool) {
		if j.Key() != key {
			return false, true
		}
		name, j.Disabled = j.Comment, !enable
		return true, true
	})
	if !found {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Cron job not found."), "/cronjobs")
		return
	}
	if writeErr := tenantfs.WriteFile(path, []byte(newContent), 0o644); writeErr != nil {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Error saving cron job. Please try again."), "/cronjobs")
		return
	}

	msg, logMsg := "Cron job %(name)s disabled.", "disabled cron job "+name
	if enable {
		msg, logMsg = "Cron job %(name)s enabled.", "enabled cron job "+name
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, logMsg, reqip.ClientIP(r))
	if hasCronJobs(strings.Split(newContent, "\n")) {
		restartOrActivateCron(r.Context(), userContext)
	} else {
		docker.StartOrStopContainer(r.Context(), userContext, "cron", "deactivate", "")
	}
	web.FlashRedirect(a, w, r, "success", web.Tr(a, r, msg, "name", name), "/cronjobs")
}

func init() {
	docker.ServiceSwitchHooks = append(docker.ServiceSwitchHooks, moveCronJobsToService)
}

// moveCronJobsToService points the jobs that ran in the old web server or MySQL container at the new one after a type switch
func moveCronJobsToService(ctx context.Context, userContext, from, to string) {
	path := cronFilePath(userContext)
	content, err := tenantfs.ReadFile(path)
	if err != nil {
		return
	}
	newContent, moved, restart := moveCronJobs(string(content), from, to)
	if moved == 0 {
		return
	}
	if err := tenantfs.WriteFile(path, []byte(newContent), 0o644); err != nil {
		log.Printf("CRONS - moving jobs from %s to %s: %v", from, to, err)
		return
	}
	if restart && docker.IsServiceRunning(ctx, userContext, "cron") {
		docker.ComposeContainer(ctx, userContext, "cron", "restart")
	}
}

// moveCronJobs rewrites every job's container from -> to, restart is true when an enabled job moved
func moveCronJobs(content, from, to string) (newContent string, moved int, restart bool) {
	var out []string
	for _, block := range splitCronBlocks(content) {
		if jobs := ParseCronFile(block); len(jobs) == 1 && jobs[0].Container == from {
			j := jobs[0]
			j.Container = to
			block = cronBlock(j)
			moved++
			restart = restart || !j.Disabled
		}
		out = append(out, block)
	}
	return strings.Join(out, "\n\n") + "\n\n", moved, restart
}
