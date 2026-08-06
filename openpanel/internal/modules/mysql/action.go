package mysql

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// handleDatabaseAction: GET returns the table list for one database, POST
// runs OPTIMIZE/REPAIR TABLE against every table in it.
func handleDatabaseAction(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	action := r.PathValue("action")
	dbName := r.PathValue("db_name")

	if action != "optimize" && action != "repair" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid action. Must be one of: optimize, repair"})
		return
	}
	if !validators.IsValidIdentifier(dbName) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid database name '" + dbName + "'"})
		return
	}

	checkRows, checkErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT schema_name FROM information_schema.schemata WHERE schema_name NOT IN ("+restricted.dbsSQL+") AND schema_name = '"+dbName+"'", "")
	if checkErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error validating database: " + checkErr.Error()})
		return
	}
	if len(checkRows) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Database '" + dbName + "' not found or not accessible"})
		return
	}

	if r.Method == http.MethodGet {
		tableRows, tblErr := mysqlmanager.Exec(ctx, userContext,
			"SELECT table_name, table_rows, data_length, index_length, data_free FROM information_schema.tables WHERE table_schema = DATABASE() ORDER BY table_name", dbName)
		if tblErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error fetching tables: " + tblErr.Error()})
			return
		}
		tables := make([]map[string]any, 0, len(tableRows))
		for _, row := range tableRows {
			tables = append(tables, map[string]any{
				"table": toStringCell(row[0]), "rows": mysqlmanager.ToInt(row[1]),
				"data_length": mysqlmanager.ToInt(row[2]), "index_length": mysqlmanager.ToInt(row[3]),
				"data_free": mysqlmanager.ToInt(row[4]),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"database": dbName, "action": action, "tables": tables})
		return
	}

	tableRows, tblErr := mysqlmanager.Exec(ctx, userContext,
		"SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() ORDER BY table_name", dbName)
	if tblErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error fetching tables: " + tblErr.Error()})
		return
	}
	var tables []string
	for _, row := range tableRows {
		tables = append(tables, toStringCell(row[0]))
	}
	if len(tables) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"database": dbName, "action": action, "results": []any{}})
		return
	}

	sqlVerb := "OPTIMIZE"
	if action == "repair" {
		sqlVerb = "REPAIR"
	}
	results := make([]map[string]any, 0, len(tables))
	for _, table := range tables {
		opRows, opErr := mysqlmanager.Exec(ctx, userContext, sqlVerb+" TABLE `"+table+"`", dbName)
		if opErr != nil {
			results = append(results, map[string]any{"table": table, "status": "error", "error": opErr.Error()})
			continue
		}
		details := make([]map[string]any, 0, len(opRows))
		for _, opRow := range opRows {
			details = append(details, map[string]any{"op": toStringCell(opRow[1]), "msg_type": toStringCell(opRow[2]), "msg_text": toStringCell(opRow[3])})
		}
		results = append(results, map[string]any{"table": table, "status": "ok", "details": details})
	}
	writeJSON(w, http.StatusOK, map[string]any{"database": dbName, "action": action, "results": results})
}
