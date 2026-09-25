package postgresql

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/dbtuning"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// errorLogTail is how many container log lines are scanned for known problems
const errorLogTail = 500

func int64Cell(v any) int64 {
	return int64(postgresmanager.ToInt(v))
}

// gatherPgTuningStats collects everything BuildPgRecommendations needs from the running server
func gatherPgTuningStats(ctx context.Context, a *appctx.App, userContext string) (PgTuningStats, error) {
	s := PgTuningStats{Settings: map[string]Setting{}, AvailableKeys: availableConfKeys}

	rows, err := postgresmanager.Exec(ctx, userContext, "SELECT name, setting, COALESCE(unit, '') FROM pg_settings", "postgres")
	if err != nil {
		return s, err
	}
	for _, row := range rows {
		s.Settings[genericCellString(row[0])] = Setting{Value: genericCellString(row[1]), Unit: genericCellString(row[2])}
	}
	s.Version = s.Settings["server_version"].Value

	if r, err := postgresmanager.Exec(ctx, userContext, "SELECT EXTRACT(EPOCH FROM now() - pg_postmaster_start_time())::bigint", "postgres"); err == nil && len(r) > 0 {
		s.UptimeSecs = int64Cell(r[0][0])
	}
	if r, err := postgresmanager.Exec(ctx, userContext, `
		SELECT COUNT(*) FILTER (WHERE NOT datistemplate AND datname <> 'postgres'),
		       COALESCE(SUM(pg_database_size(datname)) FILTER (WHERE NOT datistemplate), 0)::bigint
		FROM pg_database`, "postgres"); err == nil && len(r) > 0 {
		s.Databases = postgresmanager.ToInt(r[0][0])
		s.DataSize = int64Cell(r[0][1])
	}
	if r, err := postgresmanager.Exec(ctx, userContext, `
		SELECT COALESCE(SUM(blks_hit),0)::bigint, COALESCE(SUM(blks_read),0)::bigint,
		       COALESCE(SUM(temp_files),0)::bigint, COALESCE(SUM(temp_bytes),0)::bigint
		FROM pg_stat_database`, "postgres"); err == nil && len(r) > 0 {
		s.BlksHit, s.BlksRead, s.TempFiles, s.TempBytes = int64Cell(r[0][0]), int64Cell(r[0][1]), int64Cell(r[0][2]), int64Cell(r[0][3])
	}
	// PostgreSQL 17 moved checkpoint counters from pg_stat_bgwriter to pg_stat_checkpointer
	if r, err := postgresmanager.Exec(ctx, userContext, "SELECT num_timed, num_requested FROM pg_stat_checkpointer", "postgres"); err == nil && len(r) > 0 {
		s.CheckpointsTimed, s.CheckpointsReq = int64Cell(r[0][0]), int64Cell(r[0][1])
	} else if r, err := postgresmanager.Exec(ctx, userContext, "SELECT checkpoints_timed, checkpoints_req FROM pg_stat_bgwriter", "postgres"); err == nil && len(r) > 0 {
		s.CheckpointsTimed, s.CheckpointsReq = int64Cell(r[0][0]), int64Cell(r[0][1])
	}
	if r, err := postgresmanager.Exec(ctx, userContext, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE state = 'idle in transaction' AND now() - state_change > interval '5 minutes')
		FROM pg_stat_activity WHERE backend_type = 'client backend' AND pid <> pg_backend_pid()`, "postgres"); err == nil && len(r) > 0 {
		s.Connections = postgresmanager.ToInt(r[0][0])
		s.IdleInTransaction = postgresmanager.ToInt(r[0][1])
	}
	if r, err := postgresmanager.Exec(ctx, userContext, "SELECT COUNT(*) FROM pg_replication_slots", "postgres"); err == nil && len(r) > 0 {
		s.ReplicationSlots = postgresmanager.ToInt(r[0][0])
	}

	c := dbtuning.InspectContainer(ctx, userContext, "postgres")
	s.RAMLimit, s.MemUsage, s.CPUs, s.OOMKilled = c.MemLimit, c.MemUsage, c.CPUs, c.OOMKilled
	if s.RAMLimit <= 0 {
		if env, ok := parseEnvSize(webserver.GetEnvFileValue(userContext, "POSTGRES_RAM")); ok {
			s.RAMLimit = env
		} else {
			s.RAMLimit = dbtuning.HostMemory()
		}
	}
	if body, status := docker.FetchContainerLog(ctx, a, userContext, "postgres", errorLogTail); status == http.StatusOK {
		for _, line := range strings.Split(body, "\n") {
			l := strings.ToLower(line)
			if strings.Contains(l, "error") || strings.Contains(l, "fatal") || strings.Contains(l, "log:") {
				s.LogLines = append(s.LogLines, line)
			}
		}
	}
	return s, nil
}

// parseEnvSize reads compose style memory limits like 1.0G or 512M
func parseEnvSize(v string) (int64, bool) {
	v = strings.ToUpper(strings.TrimSpace(v))
	mult := int64(1)
	switch {
	case strings.HasSuffix(v, "G"):
		mult, v = gb, strings.TrimSuffix(v, "G")
	case strings.HasSuffix(v, "M"):
		mult, v = mb, strings.TrimSuffix(v, "M")
	case strings.HasSuffix(v, "K"):
		mult, v = kb, strings.TrimSuffix(v, "K")
	}
	var f float64
	if _, err := fmt.Sscan(v, &f); err != nil || f <= 0 {
		return 0, false
	}
	return int64(f * float64(mult)), true
}

// handlePgConfigRecommendations returns tuning suggestions for the configuration page, loaded with ajax so the page itself stays fast
func handlePgConfigRecommendations(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	stats, err := gatherPgTuningStats(r.Context(), a, userContext)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "Could not read the database status: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, BuildPgRecommendations(stats))
}
