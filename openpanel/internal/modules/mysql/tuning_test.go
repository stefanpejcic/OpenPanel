package mysql

import (
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/dbtuning"
)

func baseStats() TuningStats {
	return TuningStats{
		Version:  "8.4.2",
		RAMLimit: gb,
		Vars: map[string]string{
			"innodb_buffer_pool_size": "134217728", "innodb_buffer_pool_chunk_size": "134217728",
			"key_buffer_size": "8388608", "tmp_table_size": "16777216", "max_heap_table_size": "16777216",
			"max_connections": "151", "thread_cache_size": "9", "performance_schema": "OFF",
			"wait_timeout": "28800", "interactive_timeout": "28800", "max_allowed_packet": "67108864",
			"sort_buffer_size": "262144", "join_buffer_size": "262144", "read_buffer_size": "131072",
			"read_rnd_buffer_size": "262144", "thread_stack": "1048576", "innodb_log_buffer_size": "16777216",
		},
		Status:        map[string]string{"Uptime": "86400"},
		EngineData:    map[string]int64{"InnoDB": 10 * mb},
		EngineIndex:   map[string]int64{"InnoDB": 2 * mb},
		EngineTables:  map[string]int64{"InnoDB": 40},
		AvailableKeys: defaultConfKeys,
	}
}

func recFor(r dbtuning.Report, key string) *dbtuning.Recommendation {
	for i := range r.Recommendations {
		if r.Recommendations[i].Key == key {
			return &r.Recommendations[i]
		}
	}
	return nil
}

func TestParseAndFormatSize(t *testing.T) {
	for in, want := range map[string]int64{"128M": 128 * mb, "134217728": 128 * mb, "64K": 64 * kb, "1G": gb, "1.5G": 3 * gb / 2, "16m": 16 * mb} {
		if got, ok := parseSize(in); !ok || got != want {
			t.Errorf("parseSize(%q) = %d, %v", in, got, ok)
		}
	}
	for in, want := range map[int64]string{128 * mb: "128M", gb: "1G", 64 * kb: "64K", 1536 * mb: "1536M", 1000: "1000"} {
		if got := formatSize(in); got != want {
			t.Errorf("formatSize(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestHealthySmallServerHasNoRecommendations(t *testing.T) {
	r := BuildRecommendations(baseStats())
	if len(r.Recommendations) != 0 {
		t.Errorf("expected no recommendations, got %+v", r.Recommendations)
	}
}

func TestPerformanceSchemaOffOnSmallContainer(t *testing.T) {
	s := baseStats()
	s.Vars["performance_schema"] = "ON"
	s.PerfSchemaMem = 200 * mb
	rec := recFor(BuildRecommendations(s), "performance_schema")
	if rec == nil || rec.Recommended != "OFF" || !strings.Contains(rec.Reason, "200 MB") {
		t.Errorf("got %+v", rec)
	}
	s.RAMLimit = 8 * gb
	if recFor(BuildRecommendations(s), "performance_schema") != nil {
		t.Error("performance_schema should be left alone with plenty of memory")
	}
}

func TestKeyBufferLoweredWithoutMyISAM(t *testing.T) {
	s := baseStats()
	s.Vars["key_buffer_size"] = "134217728"
	rec := recFor(BuildRecommendations(s), "key_buffer_size")
	if rec == nil || rec.Recommended != "8M" || rec.Current != "128M" {
		t.Errorf("got %+v", rec)
	}
	s.EngineTables["MyISAM"] = 3
	s.EngineIndex["MyISAM"] = 300 * mb
	rec = recFor(BuildRecommendations(s), "key_buffer_size")
	if rec == nil || rec.Recommended != "256M" {
		t.Errorf("with MyISAM, capped at a quarter of RAM: got %+v", rec)
	}
}

func TestBigServerGrowsBufferPoolAndRelatedSettings(t *testing.T) {
	s := baseStats()
	s.RAMLimit = 8 * gb
	s.MariaDB = true
	s.Vars["innodb_log_file_size"] = "100663296"
	s.EngineData["InnoDB"] = 2 * gb
	s.EngineIndex["InnoDB"] = 512 * mb
	r := BuildRecommendations(s)

	pool := recFor(r, "innodb_buffer_pool_size")
	if pool == nil || pool.Recommended != "3840M" {
		t.Fatalf("pool: got %+v", pool)
	}
	logf := recFor(r, "innodb_log_file_size")
	if logf == nil || logf.Recommended != "960M" || !strings.Contains(logf.Reason, "3840M") {
		t.Errorf("log file should be a quarter of the new pool: got %+v", logf)
	}
	for _, k := range []string{"tmp_table_size", "max_heap_table_size"} {
		if rec := recFor(r, k); rec == nil || rec.Recommended != "82M" {
			t.Errorf("%s: got %+v", k, rec)
		}
	}
}

func TestBufferPoolTooBigForContainer(t *testing.T) {
	s := baseStats()
	s.Vars["innodb_buffer_pool_size"] = "1073741824"
	s.OOMKilled = true
	rec := recFor(BuildRecommendations(s), "innodb_buffer_pool_size")
	if rec == nil || rec.Recommended != "384M" || !strings.Contains(rec.Reason, "running out of memory") {
		t.Errorf("got %+v", rec)
	}
}

func TestLogFileSkippedWhenRedoCapacityIsUsed(t *testing.T) {
	s := baseStats()
	s.RAMLimit = 8 * gb
	s.Vars["innodb_log_file_size"] = "50331648"
	s.Vars["innodb_redo_log_capacity"] = "104857600"
	s.EngineData["InnoDB"] = 2 * gb
	if recFor(BuildRecommendations(s), "innodb_log_file_size") != nil {
		t.Error("innodb_log_file_size is ignored by servers using innodb_redo_log_capacity")
	}
}

func TestMaxConnectionsRaisedWhenRefused(t *testing.T) {
	s := baseStats()
	s.RAMLimit = 4 * gb
	s.Status["Max_used_connections"] = "151"
	s.Status["Connection_errors_max_connections"] = "12"
	rec := recFor(BuildRecommendations(s), "max_connections")
	if rec == nil || rec.Recommended != "230" || !strings.Contains(rec.Reason, "12 connections were refused") {
		t.Errorf("got %+v", rec)
	}
}

func TestMaxConnectionsFromErrorLog(t *testing.T) {
	s := baseStats()
	s.RAMLimit = 4 * gb
	s.Status["Max_used_connections"] = "20"
	s.LogLines = []string{"2026-09-25 [Warning] Too many connections"}
	if recFor(BuildRecommendations(s), "max_connections") == nil {
		t.Error("expected a max_connections recommendation from the error log")
	}
}

func TestUsageRulesWaitForUptime(t *testing.T) {
	s := baseStats()
	s.RAMLimit = 4 * gb
	s.Status = map[string]string{"Uptime": "120", "Max_used_connections": "150", "Created_tmp_disk_tables": "900", "Created_tmp_tables": "1000"}
	r := BuildRecommendations(s)
	if recFor(r, "max_connections") != nil {
		t.Error("counter based max_connections rule should wait for uptime")
	}
	if len(r.Notes) == 0 {
		t.Error("expected a note about the recent restart")
	}
}

func TestTmpTablesOnDisk(t *testing.T) {
	s := baseStats()
	s.RAMLimit = 2 * gb
	s.Status["Created_tmp_disk_tables"] = "600"
	s.Status["Created_tmp_tables"] = "1000"
	r := BuildRecommendations(s)
	tmp, heap := recFor(r, "tmp_table_size"), recFor(r, "max_heap_table_size")
	if tmp == nil || heap == nil || tmp.Recommended != heap.Recommended || tmp.Recommended != "32M" || !strings.Contains(tmp.Reason, "60%") {
		t.Errorf("got tmp=%+v heap=%+v", tmp, heap)
	}
}

func TestHugePerConnectionBuffersUseFlavorDefault(t *testing.T) {
	s := baseStats()
	s.MariaDB = true
	s.Vars["sort_buffer_size"] = "67108864"
	rec := recFor(BuildRecommendations(s), "sort_buffer_size")
	if rec == nil || rec.Recommended != "2M" || !strings.Contains(rec.Reason, "MariaDB") {
		t.Errorf("got %+v", rec)
	}
}

func TestOnlyKeysShownOnThePage(t *testing.T) {
	s := baseStats()
	s.Vars["key_buffer_size"] = "134217728"
	s.AvailableKeys = []string{"max_connections"}
	if recFor(BuildRecommendations(s), "key_buffer_size") != nil {
		t.Error("key_buffer_size isn't on the page, it must not be recommended")
	}
	delete(s.Vars, "key_buffer_size")
	s.AvailableKeys = defaultConfKeys
	if recFor(BuildRecommendations(s), "key_buffer_size") != nil {
		t.Error("the server doesn't have key_buffer_size, it must not be recommended")
	}
}

func TestIdleConnectionsLowerTimeouts(t *testing.T) {
	s := baseStats()
	s.SleepingConns = 25
	r := BuildRecommendations(s)
	if rec := recFor(r, "wait_timeout"); rec == nil || rec.Recommended != "600" {
		t.Errorf("wait_timeout: got %+v", rec)
	}
	if recFor(r, "interactive_timeout") == nil {
		t.Error("expected interactive_timeout to follow wait_timeout")
	}
}

func TestPacketErrorsRaiseMaxAllowedPacket(t *testing.T) {
	s := baseStats()
	s.LogLines = []string{"[Warning] Aborted connection 12 (Got a packet bigger than 'max_allowed_packet' bytes)"}
	if rec := recFor(BuildRecommendations(s), "max_allowed_packet"); rec == nil || rec.Recommended != "128M" {
		t.Errorf("got %+v", rec)
	}
}
