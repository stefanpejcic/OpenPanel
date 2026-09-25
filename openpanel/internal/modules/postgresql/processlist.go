package postgresql

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// ProcessRow is one row of psql/processlist.html's table.
type ProcessRow struct {
	PID           string `json:"pid"`
	DB            string `json:"datname"`
	User          string `json:"usename"`
	ClientAddr    string `json:"client_addr"`
	XactStart     string `json:"xact_start"`
	QueryStart    string `json:"query_start"`
	WaitEventType string `json:"wait_event_type"`
	WaitEvent     string `json:"wait_event"`
	State         string `json:"state"`
	Query         string `json:"query"`
	BackendType   string `json:"backend_type"`
}

// Killable is true for a user query that's actually running right now
func (p ProcessRow) Killable() bool {
	return p.BackendType == "client backend" && p.State == "active"
}

func timeCell(v any) string {
	if t, ok := v.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}
	return genericCellString(v)
}

// handleProcessList renders the PostgreSQL active-process list page.
func handleProcessList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var processList []ProcessRow
	rows, execErr := postgresmanager.Exec(ctx, userContext, `
		SELECT pid, datname, usename, client_addr, xact_start, query_start,
		       wait_event_type, wait_event, state, query, backend_type
		FROM pg_stat_activity
		WHERE pid <> pg_backend_pid()
		ORDER BY pid
	`, "postgres")
	if execErr != nil {
		flashSess(a, w, r, "error", "Error fetching process list: "+execErr.Error())
	} else {
		for _, row := range rows {
			processList = append(processList, ProcessRow{
				PID: genericCellString(row[0]), DB: genericCellString(row[1]), User: genericCellString(row[2]),
				ClientAddr: genericCellString(row[3]), XactStart: timeCell(row[4]), QueryStart: timeCell(row[5]),
				WaitEventType: genericCellString(row[6]), WaitEvent: genericCellString(row[7]), State: genericCellString(row[8]),
				Query: genericCellString(row[9]), BackendType: genericCellString(row[10]),
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
	errInvalidPID       = errors.New("invalid process ID")
	errProcessNotFound  = errors.New("process not found or no longer running a query")
	errProcessProtected = errors.New("only user queries can be cancelled")
)

// cancelQuery runs pg_cancel_backend on pid, but only if it's an active client query and not the panel's own connection
func cancelQuery(ctx context.Context, userContext, pid string) error {
	n, err := strconv.ParseInt(pid, 10, 32)
	if err != nil || n <= 0 {
		return errInvalidPID
	}
	rows, err := postgresmanager.Exec(ctx, userContext,
		"SELECT backend_type, state FROM pg_stat_activity WHERE pid = $1 AND pid <> pg_backend_pid()", "postgres", n)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return errProcessNotFound
	}
	if genericCellString(rows[0][0]) != "client backend" {
		return errProcessProtected
	}
	if genericCellString(rows[0][1]) != "active" {
		return errProcessNotFound
	}
	res, err := postgresmanager.Exec(ctx, userContext, "SELECT pg_cancel_backend($1)", "postgres", n)
	if err != nil {
		return err
	}
	if len(res) == 0 || res[0][0] != true {
		return errProcessNotFound
	}
	return nil
}

// handleKillQuery cancels one running query from the processlist page
func handleKillQuery(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_ = r.ParseForm()
	pid := r.Form.Get("pid")
	if err := cancelQuery(r.Context(), userContext, pid); err != nil {
		flashAndRedirect(a, w, r, "error", "Error killing query "+pid+": "+err.Error(), "/postgresql/processlist")
		return
	}
	_ = logger.RecordUserAction(a.Config, currentUsername, "killed PostgreSQL query "+pid, reqip.ClientIP(r))
	flashAndRedirect(a, w, r, "success", "Query "+pid+" killed.", "/postgresql/processlist")
}
