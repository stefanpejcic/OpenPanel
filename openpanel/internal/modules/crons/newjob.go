package crons

import (
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

// cronNameMax is how much of a job name the save keeps, longer names used to be cut silently
const cronNameMax = 30

// newJobProblem is the create form's checks on top of the schedule/command ones, "" when the job can be added; msg may use %(name)s and %(container)s
func newJobProblem(container, comment string, autoComment bool, containers []string, existing []CronJob) string {
	if !containsString(containers, container) {
		return "Container %(container)s isn't one of your containers."
	}
	if autoComment {
		return ""
	}
	if len(comment) > cronNameMax {
		return "Keep the name to 30 characters or fewer."
	}
	if strings.ContainsAny(comment, "\"\\") {
		return "The name can't contain double quotes or backslashes."
	}
	for _, j := range existing {
		if j.Comment == comment {
			return "A cron job named \"%(name)s\" already exists. Pick another name."
		}
	}
	return ""
}

// SchedulePreview is the create form's live answer for a schedule
type SchedulePreview struct {
	Error string   `json:"error,omitempty"`
	Text  string   `json:"text,omitempty"`
	Next  []string `json:"next,omitempty"` // RFC3339 in the cron time zone
}

func schedulePreview(t i18n.Translator, schedule string, now time.Time, loc *time.Location) SchedulePreview {
	runs, err := nextCronRuns(schedule, now, 3, loc)
	if err != nil {
		if msg := validateCronSchedule(schedule); msg != "" {
			return SchedulePreview{Error: t.Get(msg)}
		}
		return SchedulePreview{Error: t.Get("Invalid schedule: %(error)s", "error", err.Error())}
	}
	p := SchedulePreview{Text: describeSchedule(t, schedule)}
	for _, r := range runs {
		p.Next = append(p.Next, r.Format(time.RFC3339))
	}
	return p
}
