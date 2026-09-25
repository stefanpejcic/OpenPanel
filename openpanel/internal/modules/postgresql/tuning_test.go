package postgresql

import (
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/dbtuning"
)

// basePgStats is a healthy PostgreSQL 17 with a 1GB container and default settings adjusted to fit it
func basePgStats() PgTuningStats {
	return PgTuningStats{
		Version: "17.4", RAMLimit: gb, CPUs: 1, UptimeSecs: 86400,
		Settings: map[string]Setting{
			"shared_buffers":               {"32768", "8kB"}, // 256MB
			"effective_cache_size":         {"65536", "8kB"}, // 512MB
			"work_mem":                     {"4096", "kB"},
			"maintenance_work_mem":         {"65536", "kB"},
			"max_connections":              {"100", ""},
			"max_worker_processes":         {"8", ""},
			"max_parallel_workers":         {"1", ""},
			"wal_level":                    {"replica", ""},
			"checkpoint_completion_target": {"0.9", ""},
			"log_min_duration_statement":   {"-1", "ms"},
			"log_statement":                {"none", ""},
			"autovacuum":                   {"on", ""},
		},
		Databases: 2, DataSize: 50 * mb, BlksHit: 99900, BlksRead: 100,
		AvailableKeys: defaultConfKeys,
	}
}

func pgRec(r dbtuning.Report, key string) *dbtuning.Recommendation {
	for i := range r.Recommendations {
		if r.Recommendations[i].Key == key {
			return &r.Recommendations[i]
		}
	}
	return nil
}

func TestPgHealthyServerHasNoRecommendations(t *testing.T) {
	r := BuildPgRecommendations(basePgStats())
	if len(r.Recommendations) != 0 {
		t.Errorf("expected none, got %+v", r.Recommendations)
	}
}

func TestPgDefaultsOnSmallContainer(t *testing.T) {
	s := basePgStats()
	s.Settings["shared_buffers"] = Setting{"16384", "8kB"}       // 128MB default
	s.Settings["effective_cache_size"] = Setting{"524288", "8kB"} // 4GB default
	s.Settings["max_parallel_workers"] = Setting{"8", ""}
	s.Settings["checkpoint_completion_target"] = Setting{"0.5", ""}
	r := BuildPgRecommendations(s)

	if rec := pgRec(r, "effective_cache_size"); rec == nil || rec.Current != "4GB" || rec.Recommended != "512MB" {
		t.Errorf("effective_cache_size: %+v", rec)
	}
	if rec := pgRec(r, "max_parallel_workers"); rec == nil || rec.Recommended != "1" {
		t.Errorf("max_parallel_workers: %+v", rec)
	}
	if rec := pgRec(r, "checkpoint_completion_target"); rec == nil || rec.Recommended != "0.9" {
		t.Errorf("checkpoint_completion_target: %+v", rec)
	}
	if pgRec(r, "shared_buffers") != nil {
		t.Error("128MB shared buffers already holds 50MB of data")
	}
}

func TestPgSharedBuffersGrowWithData(t *testing.T) {
	s := basePgStats()
	s.RAMLimit = 4 * gb
	s.Settings["effective_cache_size"] = Setting{"262144", "8kB"} // 2GB
	s.DataSize = 600 * mb
	rec := pgRec(BuildPgRecommendations(s), "shared_buffers")
	if rec == nil || rec.Current != "256MB" || rec.Recommended != "912MB" {
		t.Errorf("got %+v", rec)
	}
}

func TestPgSharedBuffersTooBigAfterOOM(t *testing.T) {
	s := basePgStats()
	s.Settings["shared_buffers"] = Setting{"65536", "8kB"} // 512MB
	s.OOMKilled = true
	rec := pgRec(BuildPgRecommendations(s), "shared_buffers")
	if rec == nil || rec.Recommended != "154MB" || !strings.Contains(rec.Reason, "running out of memory") {
		t.Errorf("got %+v", rec)
	}
}

func TestPgWorkMemFromTempFiles(t *testing.T) {
	s := basePgStats()
	s.RAMLimit = 8 * gb
	s.Settings["effective_cache_size"] = Setting{"524288", "8kB"}
	s.TempFiles = 40
	s.TempBytes = 40 * 20 * mb
	rec := pgRec(BuildPgRecommendations(s), "work_mem")
	if rec == nil || rec.Recommended != "20MB" || !strings.Contains(rec.Reason, "40 queries") {
		t.Errorf("got %+v", rec)
	}
}

func TestPgWorkMemTooBigForRAM(t *testing.T) {
	s := basePgStats()
	s.Settings["work_mem"] = Setting{"65536", "kB"} // 64MB x 100 connections
	rec := pgRec(BuildPgRecommendations(s), "work_mem")
	if rec == nil || rec.Recommended != "4MB" {
		t.Errorf("got %+v", rec)
	}
}

func TestPgMaxConnectionsFromLog(t *testing.T) {
	s := basePgStats()
	s.RAMLimit = 4 * gb
	s.Settings["effective_cache_size"] = Setting{"262144", "8kB"}
	s.LogLines = []string{"FATAL:  sorry, too many clients already"}
	rec := pgRec(BuildPgRecommendations(s), "max_connections")
	if rec == nil || rec.Recommended != "150" {
		t.Errorf("got %+v", rec)
	}
}

func TestPgLoggingAndWal(t *testing.T) {
	s := basePgStats()
	s.Settings["log_min_duration_statement"] = Setting{"0", "ms"}
	s.Settings["log_statement"] = Setting{"all", ""}
	s.Settings["wal_level"] = Setting{"logical", ""}
	s.Settings["autovacuum"] = Setting{"off", ""}
	r := BuildPgRecommendations(s)
	for key, want := range map[string]string{"log_min_duration_statement": "1000ms", "log_statement": "ddl", "wal_level": "replica", "autovacuum": "on"} {
		if rec := pgRec(r, key); rec == nil || rec.Recommended != want {
			t.Errorf("%s: %+v", key, rec)
		}
	}
	s.ReplicationSlots = 1
	if pgRec(BuildPgRecommendations(s), "wal_level") != nil {
		t.Error("logical WAL is needed while a replication slot exists")
	}
}

func TestPgMaxWalSizeOnlyWhenShown(t *testing.T) {
	s := basePgStats()
	s.Settings["max_wal_size"] = Setting{"1024", "MB"}
	s.CheckpointsTimed, s.CheckpointsReq = 20, 30
	if pgRec(BuildPgRecommendations(s), "max_wal_size") != nil {
		t.Error("max_wal_size isn't in the default keys")
	}
	s.AvailableKeys = append(append([]string{}, defaultConfKeys...), "max_wal_size")
	if rec := pgRec(BuildPgRecommendations(s), "max_wal_size"); rec == nil || rec.Recommended != "2GB" {
		t.Errorf("got %+v", rec)
	}
}

func TestPgUsageRulesWaitForUptime(t *testing.T) {
	s := basePgStats()
	s.UptimeSecs = 60
	s.BlksHit, s.BlksRead = 100, 900
	r := BuildPgRecommendations(s)
	if pgRec(r, "shared_buffers") != nil {
		t.Error("hit ratio rule should wait for uptime")
	}
	if len(r.Notes) == 0 {
		t.Error("expected the recent restart note")
	}
}

func TestPgDisplay(t *testing.T) {
	s := basePgStats()
	for key, want := range map[string]string{"shared_buffers": "256MB", "work_mem": "4MB", "log_min_duration_statement": "-1", "max_connections": "100"} {
		if got := s.display(key); got != want {
			t.Errorf("display(%s) = %q, want %q", key, got, want)
		}
	}
}
