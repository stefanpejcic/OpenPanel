package postgresql

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
)

// handleProcessList renders the PostgreSQL active-process list page.
func handleProcessList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var processlistOutput string
	rows, execErr := postgresmanager.Exec(ctx, userContext, `
		SELECT datid, datname, pid, leader_pid, usesysid, usename, application_name,
		       client_addr, client_hostname, client_port, backend_start, xact_start,
		       query_start, state_change, wait_event_type, wait_event, state,
		       backend_xid, backend_xmin, query_id, query, backend_type
		FROM pg_stat_activity
	`, "postgres")
	if execErr == nil {
		lines := make([]string, 0, len(rows))
		for _, row := range rows {
			cells := make([]string, len(row))
			for i, v := range row {
				if v == nil {
					cells[i] = ""
				} else {
					cells[i] = toStringCell(v)
				}
			}
			lines = append(lines, strings.Join(cells, "|"))
		}
		processlistOutput = strings.Join(lines, "\n")
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]string{"processlist_output": processlistOutput})
		return
	}

	renderProcessListPage(a, w, r, processlistOutput)
}
