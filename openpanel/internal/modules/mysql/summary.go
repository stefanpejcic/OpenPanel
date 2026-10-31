package mysql

import (
	"context"
	"regexp"
	"slices"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/dbtuning"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var serverVersionRE = regexp.MustCompile(`^\d+(\.\d+)*`)

// shortServerVersion keeps the number of a VERSION() string like 13.0.2-MariaDB-ubu2604
func shortServerVersion(v string) string {
	return serverVersionRE.FindString(strings.TrimSpace(v))
}

// uptimeSince is "2d 14h" for a container start time, "" when it isn't known
func uptimeSince(started time.Time) string {
	if started.IsZero() {
		return ""
	}
	return dbtuning.FormatUptime(int64(time.Since(started).Seconds()))
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

// remoteMySQLPort is the host port MySQL is published on, "" when MYSQL_PORT isn't set the usual way
func remoteMySQLPort(userContext string) string {
	if m := mysqlPortRE.FindStringSubmatch(webserver.GetEnvFileValue(userContext, "MYSQL_PORT")); m != nil {
		return m[2]
	}
	return ""
}

// UserGrant is one database a user has privileges on, for the users table
type UserGrant struct {
	Database   string
	Privileges []string
}

// All reports whether the grant holds every privilege the assign page offers
func (g UserGrant) All() bool {
	for _, p := range mysqlPrivileges {
		if p != "ALL PRIVILEGES" && !slices.Contains(g.Privileges, p) {
			return false
		}
	}
	return true
}

// Summary is "ALL" or the first few privileges, the rest behind an ellipsis
func (g UserGrant) Summary() string {
	if g.All() {
		return "ALL"
	}
	if len(g.Privileges) > 3 {
		return strings.Join(g.Privileges[:3], ", ") + "…"
	}
	return strings.Join(g.Privileges, ", ")
}

// schemaPrivilegesSQL lists every account's database-level privileges in one query
const schemaPrivilegesSQL = "SELECT GRANTEE, TABLE_SCHEMA, GROUP_CONCAT(PRIVILEGE_TYPE ORDER BY PRIVILEGE_TYPE SEPARATOR ',') FROM information_schema.SCHEMA_PRIVILEGES GROUP BY GRANTEE, TABLE_SCHEMA ORDER BY TABLE_SCHEMA"

// fetchUserGrants maps each user to its database grants, nil when the lookup fails
func fetchUserGrants(ctx context.Context, userContext string) map[string][]UserGrant {
	rows, err := mysqlmanager.Exec(ctx, userContext, schemaPrivilegesSQL, "")
	if err != nil {
		return nil
	}
	cells := make([][3]string, 0, len(rows))
	for _, row := range rows {
		if len(row) == 3 {
			cells = append(cells, [3]string{toStringCell(row[0]), toStringCell(row[1]), toStringCell(row[2])})
		}
	}
	return groupUserGrants(cells)
}

// groupUserGrants folds (grantee, database, privileges) rows into per-user grants, merging the same user on several hosts
func groupUserGrants(rows [][3]string) map[string][]UserGrant {
	out := map[string][]UserGrant{}
	for _, row := range rows {
		user := granteeUser(row[0])
		if user == "" || row[1] == "" {
			continue
		}
		privs := strings.Split(row[2], ",")
		grants := out[user]
		i := slices.IndexFunc(grants, func(g UserGrant) bool { return g.Database == row[1] })
		if i < 0 {
			grants = append(grants, UserGrant{Database: row[1]})
			i = len(grants) - 1
		}
		for _, p := range privs {
			if p = strings.TrimSpace(p); p != "" && !slices.Contains(grants[i].Privileges, p) {
				grants[i].Privileges = append(grants[i].Privileges, p)
			}
		}
		out[user] = grants
	}
	return out
}

// granteeUser takes the user out of a GRANTEE like 'app'@'%'
func granteeUser(grantee string) string {
	at := strings.LastIndex(grantee, "@")
	if at < 0 {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(grantee[:at], "'"), "'")
}
