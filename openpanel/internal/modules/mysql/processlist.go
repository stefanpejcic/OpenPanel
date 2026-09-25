package mysql

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
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
		flashSess(a, w, r, "error", web.Tr(a, r, "Error fetching process list: %(error)s", "error", execErr.Error()))
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

var (
	errInvalidProcessID = errors.New("invalid process ID")
	errProcessNotFound  = errors.New("process not found or no longer running a query")
	errProcessProtected = errors.New("system threads can not be killed")
)

// killQuery runs KILL QUERY on id, but only if it's a running query in the current processlist and not a system thread
func killQuery(ctx context.Context, userContext, id string) error {
	pid, err := strconv.ParseUint(id, 10, 64)
	if err != nil || pid == 0 {
		return errInvalidProcessID
	}
	rows, err := mysqlmanager.Exec(ctx, userContext, "SHOW FULL PROCESSLIST", "")
	if err != nil {
		return err
	}
	for _, row := range rows {
		if toStringCell(row[0]) != strconv.FormatUint(pid, 10) {
			continue
		}
		switch toStringCell(row[1]) {
		case "system user", "event_scheduler":
			return errProcessProtected
		}
		if toStringCell(row[4]) != "Query" {
			return errProcessNotFound
		}
		_, err := mysqlmanager.Exec(ctx, userContext, "KILL QUERY "+strconv.FormatUint(pid, 10), "")
		return err
	}
	return errProcessNotFound
}

// handleMySQLKillQuery kills one running query from the processlist page
func handleMySQLKillQuery(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	id := r.Form.Get("id")
	if err := killQuery(r.Context(), userContext, id); err != nil {
		flashAndRedirect(a, w, r, "error", web.Tr(a, r, "Error killing query %(id)s: %(error)s", "id", id, "error", err.Error()), "/mysql/processlist")
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "killed MySQL query "+id, reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", web.Tr(a, r, "Query %(id)s killed.", "id", id), "/mysql/processlist")
}
