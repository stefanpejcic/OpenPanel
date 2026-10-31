package postgresql

import (
	"context"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// ServerInfo is the server version and open client connections for the service tile
type ServerInfo struct {
	ServerVersion string
	Connections   int
	HasInfo       bool
}

// serverInfoSQL counts client connections other than this one
const serverInfoSQL = "SELECT current_setting('server_version'), (SELECT count(*) FROM pg_stat_activity WHERE backend_type = 'client backend' AND pid <> pg_backend_pid())"

// fetchServerInfo asks postgres for its version and connection count, zero value when it can't
func fetchServerInfo(ctx context.Context, userContext string) ServerInfo {
	rows, err := postgresmanager.Exec(ctx, userContext, serverInfoSQL, "postgres")
	if err != nil || len(rows) == 0 || len(rows[0]) < 2 {
		return ServerInfo{}
	}
	return ServerInfo{ServerVersion: shortPGVersion(toStringCell(rows[0][0])), Connections: postgresmanager.ToInt(rows[0][1]), HasInfo: true}
}

// shortPGVersion drops the build suffix from a server_version like "17.4 (Debian 17.4-1.pgdg120+2)"
func shortPGVersion(v string) string {
	v, _, _ = strings.Cut(strings.TrimSpace(v), " ")
	return v
}

// planDBLimit is the plan's database limit, 0 when unlimited or unknown
func planDBLimit(ctx context.Context, a *appctx.App, injectedData map[string]any) int {
	planID, _ := injectedData["hosting_plan"].(int)
	plan, err := a.QueryPlanDetailsByID(ctx, planID)
	if err != nil {
		return 0
	}
	return web.AtoiDefault(plan.DBLimit, 0)
}
