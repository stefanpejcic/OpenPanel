package crons

import (
	"log"
	"net/http"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/php"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var pageFiles = []string{
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
}

func loadPage(files ...string) *web.Page {
	return web.MustLoadPage(append(append([]string{}, pageFiles...), files...)...)
}

var (
	cronjobsPage    = loadPage("system/cronjobs.html")
	cronjobsNewPage = loadPage("system/cronjobs_new.html")
)

// CronjobsPageData is system/cronjobs.html's template context, covering both view=table and view=code
type CronjobsPageData struct {
	web.LayoutData
	LogJob         string
	View           string
	Service        string
	CrontabContent string
	Services       []string
	CronJobs       []CronJob
	ScheduleIssues []ScheduleIssue
	Summary        CronSummary
	Timezones      []string
}

func renderCronjobsCodePage(a *appctx.App, w http.ResponseWriter, r *http.Request, crontabContent string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "CronJobs File Editor")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := CronjobsPageData{LayoutData: layout, View: "code", Service: "cron", CrontabContent: crontabContent}
	if err := cronjobsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("CRONS - code view template render error: %v", err)
	}
}

// renderCronjobsLogsPage renders the Logs tab, job preselects the job filter
func renderCronjobsLogsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, cronJobs []CronJob, job string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "CronJobs Logs")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := CronjobsPageData{LayoutData: layout, View: "logs", Service: "cron", CronJobs: cronJobs, LogJob: job}
	if err := cronjobsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("CRONS - logs view template render error: %v", err)
	}
}

func renderCronjobsTablePage(a *appctx.App, w http.ResponseWriter, r *http.Request, services []string, cronJobs []CronJob, scheduleIssues []ScheduleIssue, loc *time.Location) {
	layout, _, err := web.BuildLayoutData(a, w, r, "CronJobs")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.BulkActions = cronBulkActions(layout.T, services)
	data := CronjobsPageData{
		LayoutData: layout, View: "table", Service: "cron", Services: services, CronJobs: cronJobs,
		ScheduleIssues: scheduleIssues, Summary: cronSummary(cronJobs, time.Now(), loc), Timezones: php.AvailableTimezones(),
	}
	if err := cronjobsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("CRONS - table view template render error: %v", err)
	}
}

// CronjobsNewPageData is system/cronjobs_new.html's template context.
type CronjobsNewPageData struct {
	web.LayoutData
	Service    string
	Containers []string
	// the form picks these when the command looks like PHP or a MySQL/MariaDB client, "" when the user doesn't have them
	PHPContainer string
	DBContainer  string
}

func renderCronjobsNewPage(a *appctx.App, w http.ResponseWriter, r *http.Request, containers []string, phpContainer, dbContainer string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "New CronJob")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := CronjobsNewPageData{LayoutData: layout, Service: "cron", Containers: containers, PHPContainer: phpContainer, DBContainer: dbContainer}
	if err := cronjobsNewPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("CRONS - new template render error: %v", err)
	}
}
