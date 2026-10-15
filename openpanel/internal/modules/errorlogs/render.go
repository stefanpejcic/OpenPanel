package errorlogs

import (
	"log"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var errorLogsPage = web.LoadPage("system/error_logs.html")

// errorLogsBulkActions are both browser-only, the page script copies or saves the selected rows' lines
func errorLogsBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "copy", Label: t.Get("Copy"), Icon: "bi-clipboard", Client: true},
		{Key: "download", Label: t.Get("Download"), Icon: "bi-download", Client: true},
	}
}

// RowData is one table row, with the export line precomputed for the copy button
type RowData struct {
	Entry
	Unix int64
	Rank int
	Line string
}

// ErrorLogsPageData is system/error_logs.html's template context.
type ErrorLogsPageData struct {
	web.LayoutData
	Query     Query
	Rows      []RowData
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
	layout.BulkActions = errorLogsBulkActions(layout.T)
	data := buildPageData(layout, q, res, targets, errMsg)
	if err := errorLogsPage.Render(w, status, data); err != nil {
		log.Printf("ERROR_LOGS - template render error: %v", err)
	}
}

func buildPageData(layout web.LayoutData, q Query, res Result, targets []Target, errMsg string) ErrorLogsPageData {
	rows := make([]RowData, len(res.Entries))
	for i, e := range res.Entries {
		rows[i] = RowData{Entry: e, Unix: e.Time.Unix(), Rank: levelRank[e.Level], Line: exportLine(e)}
	}
	data := ErrorLogsPageData{
		LayoutData: layout, Query: q, Rows: rows, Total: res.Total, Truncated: res.Truncated,
		Sources: res.Sources, Targets: targets, Error: errMsg,
	}
	return data
}
