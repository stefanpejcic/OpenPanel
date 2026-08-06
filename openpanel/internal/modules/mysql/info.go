package mysql

import (
	"context"
	"fmt"
	"net/http"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// DatabasesInfoPayload is databases_info()'s JSON shape - the shared
// database/user/assignment lookup several forms (assign.html, import.html,
// remove.html) populate their <select> dropdowns from via fetch().
type DatabasesInfoPayload struct {
	Databases         []string              `json:"databases"`
	Users             []string              `json:"users"`
	AssignedDatabases []AssignedDatabaseRow `json:"assigned_databases"`
}

type AssignedDatabaseRow struct {
	Database string `json:"database"`
	Users    string `json:"users"`
}

func ComputeDatabasesInfo(ctx context.Context, userContext string) (DatabasesInfoPayload, error) {
	query := fmt.Sprintf(`
		SELECT 'db' AS type, schema_name AS col1, NULL AS col2
			FROM information_schema.schemata
			WHERE schema_name NOT IN (%s)
		UNION ALL
		SELECT 'user', User, NULL
			FROM mysql.user
			WHERE User NOT IN (%s)
		UNION ALL
		SELECT 'assigned', Db, GROUP_CONCAT(User SEPARATOR ', ')
			FROM mysql.db
			WHERE User NOT LIKE 'mysql.s%%' AND User NOT IN (%s)
			GROUP BY Db
	`, restricted.dbsSQL, restricted.usersSQL, restricted.usersSQL)

	rows, err := mysqlmanager.Exec(ctx, userContext, query, "")
	if err != nil {
		return DatabasesInfoPayload{}, err
	}

	var payload DatabasesInfoPayload
	for _, row := range rows {
		switch toStringCell(row[0]) {
		case "db":
			payload.Databases = append(payload.Databases, toStringCell(row[1]))
		case "user":
			payload.Users = append(payload.Users, toStringCell(row[1]))
		case "assigned":
			payload.AssignedDatabases = append(payload.AssignedDatabases, AssignedDatabaseRow{Database: toStringCell(row[1]), Users: toStringCell(row[2])})
		}
	}
	return payload, nil
}

// handleDatabasesInfo returns databases, users, and assigned-databases
// summary info, cached 300s per userContext (never a single cache entry
// shared across accounts). The MySQL reachability check runs outside the
// cache (always fresh) rather than letting a stale "MySQL unreachable"
// redirect get cached for 300s past recovery.
func handleDatabasesInfo(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")

	if !CheckMySQLInsideContainer(ctx, userContext, false) {
		http.Redirect(w, r, "/mysql", http.StatusFound)
		return
	}

	payload, cacheErr := cache.Memoize(ctx, a.Cache, "databases_info:"+userContext, 300*time.Second, func() (DatabasesInfoPayload, error) {
		return ComputeDatabasesInfo(ctx, userContext)
	})
	if cacheErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": cacheErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, payload)
}
