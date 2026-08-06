package crons

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
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

// CronjobsPageData is system/cronjobs.html's template context, covering
// both view=table and view=code.
type CronjobsPageData struct {
	web.LayoutData
	View           string
	Service        string
	CrontabContent string
	Services       []string
	CronJobs       []CronJob
	ScheduleIssues []ScheduleIssue
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

func renderCronjobsTablePage(a *appctx.App, w http.ResponseWriter, r *http.Request, services []string, cronJobs []CronJob, scheduleIssues []ScheduleIssue) {
	layout, _, err := web.BuildLayoutData(a, w, r, "CronJobs")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := CronjobsPageData{
		LayoutData: layout, View: "table", Service: "cron", Services: services, CronJobs: cronJobs,
		ScheduleIssues: scheduleIssues,
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
}

func renderCronjobsNewPage(a *appctx.App, w http.ResponseWriter, r *http.Request, containers []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "New CronJob")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := CronjobsNewPageData{LayoutData: layout, Service: "cron", Containers: containers}
	if err := cronjobsNewPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("CRONS - new template render error: %v", err)
	}
}
