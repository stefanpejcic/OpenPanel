package crons

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
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

// handleToggleCronjob turns one job on or off by commenting its block out, then restarts cron to pick it up
func handleToggleCronjob(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	key := r.Form.Get("key")
	enable := r.Form.Get("enabled") == "1"

	path := cronFilePath(userContext)
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Error saving cron job. Please try again."), "/cronjobs")
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
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Cron job not found."), "/cronjobs")
		return
	}
	if writeErr := os.WriteFile(path, []byte(newContent), 0o644); writeErr != nil {
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Error saving cron job. Please try again."), "/cronjobs")
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
	flashAndRedirect(a, w, r, "success", web.Tr(a, r, msg, "name", name), "/cronjobs")
}

func init() {
	docker.ServiceSwitchHooks = append(docker.ServiceSwitchHooks, moveCronJobsToService)
}

// moveCronJobsToService points the jobs that ran in the old web server or MySQL container at the new one after a type switch
func moveCronJobsToService(ctx context.Context, userContext, from, to string) {
	path := cronFilePath(userContext)
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	newContent, moved, restart := moveCronJobs(string(content), from, to)
	if moved == 0 {
		return
	}
	if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
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
