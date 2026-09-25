package mysql

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/dbtuning"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// errorLogTail is how many container log lines are scanned for known problems
const errorLogTail = 500

func rowsToMap(rows [][]any) map[string]string {
	m := make(map[string]string, len(rows))
	for _, row := range rows {
		if len(row) >= 2 {
			m[toStringCell(row[0])] = toStringCell(row[1])
		}
	}
	return m
}

// gatherTuningStats collects everything BuildRecommendations needs from the running server
func gatherTuningStats(ctx context.Context, a *appctx.App, userContext string) (TuningStats, error) {
	s := TuningStats{
		EngineData: map[string]int64{}, EngineIndex: map[string]int64{}, EngineTables: map[string]int64{},
		AvailableKeys: availableConfKeys,
	}

	varRows, err := mysqlmanager.Exec(ctx, userContext, "SHOW GLOBAL VARIABLES", "")
	if err != nil {
		return s, err
	}
	s.Vars = rowsToMap(varRows)
	statusRows, err := mysqlmanager.Exec(ctx, userContext, "SHOW GLOBAL STATUS", "")
	if err != nil {
		return s, err
	}
	s.Status = rowsToMap(statusRows)
	s.Version = s.Vars["version"]
	s.MariaDB = strings.Contains(strings.ToLower(s.Vars["version"]+" "+s.Vars["version_comment"]), "mariadb")

	engineRows, err := mysqlmanager.Exec(ctx, userContext,
		"SELECT ENGINE, COUNT(*), COALESCE(SUM(DATA_LENGTH),0), COALESCE(SUM(INDEX_LENGTH),0) FROM information_schema.TABLES WHERE TABLE_SCHEMA NOT IN ("+restricted.dbsSQL+") AND TABLE_TYPE = 'BASE TABLE' AND ENGINE IS NOT NULL GROUP BY ENGINE", "")
	if err == nil {
		for _, row := range engineRows {
			engine := toStringCell(row[0])
			s.EngineTables[engine] = int64(mysqlmanager.ToInt(row[1]))
			s.EngineData[engine] = int64(toFloatCell(row[2]))
			s.EngineIndex[engine] = int64(toFloatCell(row[3]))
		}
	}
	if dbRows, err := mysqlmanager.Exec(ctx, userContext, "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME NOT IN ("+restricted.dbsSQL+")", ""); err == nil && len(dbRows) > 0 {
		s.Databases = mysqlmanager.ToInt(dbRows[0][0])
	}
	if procRows, err := mysqlmanager.Exec(ctx, userContext, "SHOW FULL PROCESSLIST", ""); err == nil {
		for _, row := range procRows {
			if toStringCell(row[4]) == "Sleep" && mysqlmanager.ToInt(row[5]) > 300 {
				s.SleepingConns++
			}
		}
	}
	if strings.EqualFold(s.Vars["performance_schema"], "ON") {
		if psRows, err := mysqlmanager.Exec(ctx, userContext, "SHOW ENGINE PERFORMANCE_SCHEMA STATUS", ""); err == nil {
			for _, row := range psRows {
				if len(row) >= 3 && toStringCell(row[1]) == "performance_schema.memory" {
					s.PerfSchemaMem, _ = strconv.ParseInt(toStringCell(row[2]), 10, 64)
				}
			}
		}
	}

	service := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")
	if service == "mysql" || service == "mariadb" {
		c := dbtuning.InspectContainer(ctx, userContext, service)
		s.RAMLimit, s.MemUsage, s.OOMKilled = c.MemLimit, c.MemUsage, c.OOMKilled
		if body, status := docker.FetchContainerLog(ctx, a, userContext, service, errorLogTail); status == http.StatusOK {
			for _, line := range strings.Split(body, "\n") {
				l := strings.ToLower(line)
				if strings.Contains(l, "error") || strings.Contains(l, "warning") || strings.Contains(l, "too many connections") || strings.Contains(l, "max_allowed_packet") {
					s.LogLines = append(s.LogLines, line)
				}
			}
		}
	}
	if s.RAMLimit <= 0 {
		if env, ok := parseSize(webserver.GetEnvFileValue(userContext, "MYSQL_RAM")); ok && env > 0 {
			s.RAMLimit = env
		} else {
			s.RAMLimit = dbtuning.HostMemory()
		}
	}
	return s, nil
}

// handleMySQLConfigRecommendations returns tuning suggestions for the configuration page, loaded with ajax so the page itself stays fast
func handleMySQLConfigRecommendations(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	stats, err := gatherTuningStats(r.Context(), a, userContext)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "Could not read the database status: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, BuildRecommendations(stats))
}
