package mysql

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
)

// ProcessRow is one row of mysql/processlist.html's table.
type ProcessRow struct {
	ID      string
	User    string
	Host    string
	DB      string
	Command string
	Time    string
	State   string
	Info    string
}

// handleMySQLProcessList shows the live MySQL processlist.
func handleMySQLProcessList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var processList []ProcessRow
	rows, execErr := mysqlmanager.Exec(ctx, userContext, "SHOW FULL PROCESSLIST", "")
	if execErr != nil {
		flashSess(a, w, r, "error", "Error fetching process list: "+execErr.Error())
	} else {
		for _, row := range rows {
			processList = append(processList, ProcessRow{
				ID: toStringCell(row[0]), User: toStringCell(row[1]), Host: toStringCell(row[2]),
				DB: toStringCell(row[3]), Command: toStringCell(row[4]), Time: toStringCell(row[5]),
				State: toStringCell(row[6]), Info: toStringCell(row[7]),
			})
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{"processlist": processList})
		return
	}

	renderProcessListPage(a, w, r, processList)
}
