package errorlogs

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var errorLogsPage = web.LoadPage("system/error_logs.html")

// ErrorLogsPageData is system/error_logs.html's template context, the page filters, groups and pages Rows in the browser
type ErrorLogsPageData struct {
	web.LayoutData
	Query     Query
	Rows      [][]any
	Total     int
	Truncated bool
	Sources   []Source
	Targets   []Target
	Error     string
}

func renderErrorLogsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, q Query, res Result, targets []Target, status int, errMsg string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Error Logs")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := buildPageData(layout, q, res, targets, errMsg)
	if err := errorLogsPage.Render(w, status, data); err != nil {
		log.Printf("ERROR_LOGS - template render error: %v", err)
	}
}

// buildPageData packs each entry as [unix ms, level, service, source, website, message] to keep the embedded json small
func buildPageData(layout web.LayoutData, q Query, res Result, targets []Target, errMsg string) ErrorLogsPageData {
	rows := make([][]any, len(res.Entries))
	for i, e := range res.Entries {
		rows[i] = []any{e.Time.UnixMilli(), e.Level, e.Service, e.Source, e.Domain, e.Message}
	}
	if targets == nil {
		targets = []Target{}
	}
	sources := res.Sources
	if sources == nil {
		sources = []Source{}
	}
	return ErrorLogsPageData{
		LayoutData: layout, Query: q, Rows: rows, Total: res.Total, Truncated: res.Truncated,
		Sources: sources, Targets: targets, Error: errMsg,
	}
}
