package mysql

import (
	"strconv"
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

func TestBigServerGrowsBufferPoolFromData(t *testing.T) {
	s := baseStats()
	s.RAMLimit = 8 * gb
	s.MariaDB = true
	s.Vars["innodb_log_file_size"] = "100663296"
	s.EngineData["InnoDB"] = 2 * gb
	s.EngineIndex["InnoDB"] = 512 * mb
	r := BuildRecommendations(s)

	// 2.5G of data plus 30%, rounded up to the 128M chunk
	if pool := recFor(r, "innodb_buffer_pool_size"); pool == nil || pool.Recommended != "3328M" {
		t.Fatalf("pool: got %+v", pool)
	}
	if rec := recFor(r, "innodb_log_file_size"); rec != nil {
		t.Errorf("log file must follow redo writes, not the pool size: got %+v", rec)
	}
	for _, k := range []string{"tmp_table_size", "max_heap_table_size"} {
		if rec := recFor(r, k); rec != nil {
			t.Errorf("%s: no temp tables on disk, nothing to suggest, got %+v", k, rec)
		}
	}
}

func TestBufferPoolShrinksWhenFarBiggerThanData(t *testing.T) {
	s := baseStats()
	s.Vars["innodb_buffer_pool_size"] = "536870912"
	rec := recFor(BuildRecommendations(s), "innodb_buffer_pool_size")
	if rec == nil || rec.Recommended != "128M" || !strings.Contains(rec.Reason, "only 12 MB") {
		t.Fatalf("got %+v", rec)
	}
	// lots of reads from disk means the pool is in use, leave it
	s.Status["Innodb_buffer_pool_reads"] = "5000"
	s.Status["Innodb_buffer_pool_read_requests"] = "100000"
	if rec := recFor(BuildRecommendations(s), "innodb_buffer_pool_size"); rec != nil && rec.Recommended == "128M" {
		t.Errorf("should not shrink a pool that misses: got %+v", rec)
	}
}

func TestRedoLogSizedFromWrites(t *testing.T) {
	s := baseStats()
	s.Vars["innodb_log_file_size"] = "50331648"
	// 24h uptime, 12G written is 512M per hour
	s.Status["Innodb_os_log_written"] = "12884901888"
	rec := recFor(BuildRecommendations(s), "innodb_log_file_size")
	if rec == nil || rec.Recommended != "512M" || !strings.Contains(rec.Reason, "512 MB of redo log per hour") || !strings.Contains(rec.Reason, "5 minutes") {
		t.Fatalf("got %+v", rec)
	}
	// MySQL before 8.0.30 splits it over two files
	s.Vars["innodb_log_files_in_group"] = "2"
	if rec := recFor(BuildRecommendations(s), "innodb_log_file_size"); rec == nil || rec.Recommended != "256M" {
		t.Errorf("two files: got %+v", rec)
	}
	s.Status["Innodb_os_log_written"] = "1073741824"
	if rec := recFor(BuildRecommendations(s), "innodb_log_file_size"); rec != nil {
		t.Errorf("about 43M per hour fits in 96M, got %+v", rec)
	}
}

func TestMaxConnectionsCoversPHPWorkers(t *testing.T) {
	s := baseStats()
	s.RAMLimit = 4 * gb
	s.PHPWorkers = 200
	rec := recFor(BuildRecommendations(s), "max_connections")
	if rec == nil || rec.Recommended != "210" || !strings.Contains(rec.Reason, "200 workers") {
		t.Fatalf("got %+v", rec)
	}
	s.PHPWorkers = 40
	if rec := recFor(BuildRecommendations(s), "max_connections"); rec != nil {
		t.Errorf("151 already covers 40 workers, got %+v", rec)
	}
}

func TestMaxConnectionsNotLoweredBelowPHPWorkers(t *testing.T) {
	s := baseStats()
	s.RAMLimit = 512 * mb
	s.Vars["max_connections"] = "500"
	s.Vars["sort_buffer_size"] = "2097152"
	s.Status["Max_used_connections"] = "10"
	s.PHPWorkers = 120
	rec := recFor(BuildRecommendations(s), "max_connections")
	if rec == nil {
		t.Fatal("expected max_connections to be lowered")
	}
	if n, _ := strconv.Atoi(rec.Recommended); n < 130 {
		t.Errorf("lowered below the PHP workers: got %+v", rec)
	}
}

func TestTableCaches(t *testing.T) {
	s := baseStats()
	s.EngineTables["InnoDB"] = 3000
	s.Vars["table_open_cache"] = "2000"
	s.Vars["table_definition_cache"] = "1400"
	s.Vars["open_files_limit"] = "32768"
	s.Status["Open_tables"] = "2000"
	s.Status["Opened_tables"] = "48000"
	s.Status["Open_table_definitions"] = "1400"
	r := BuildRecommendations(s)
	if rec := recFor(r, "table_open_cache"); rec == nil || rec.Recommended != "6000" || !strings.Contains(rec.Reason, "2000 times per hour") {
		t.Errorf("table_open_cache: got %+v", rec)
	}
	if rec := recFor(r, "table_definition_cache"); rec == nil || rec.Recommended != "3400" {
		t.Errorf("table_definition_cache: got %+v", rec)
	}
	s.Status["Open_tables"] = "500"
	if rec := recFor(BuildRecommendations(s), "table_open_cache"); rec != nil {
		t.Errorf("cache isn't full, got %+v", rec)
	}
}

func TestFlushAtCommitIsOptional(t *testing.T) {
	s := baseStats()
	s.Vars["innodb_flush_log_at_trx_commit"] = "1"
	s.Status["Innodb_os_log_fsyncs"] = "1728000"
	rec := recFor(BuildRecommendations(s), "innodb_flush_log_at_trx_commit")
	if rec == nil || rec.Recommended != "2" || !rec.Optional || !strings.Contains(rec.Reason, "last second") {
		t.Fatalf("got %+v", rec)
	}
	s.Status["Innodb_os_log_fsyncs"] = "1000"
	if recFor(BuildRecommendations(s), "innodb_flush_log_at_trx_commit") != nil {
		t.Error("few writes, nothing to gain")
	}
}

func TestInsights(t *testing.T) {
	s := baseStats()
	if r := BuildRecommendations(s); len(r.Insights) != 0 {
		t.Errorf("healthy server: got %+v", r.Insights)
	}
	s.EngineTables["MyISAM"] = 14
	s.MyISAMTables = []string{"a.t1", "a.t2", "a.t3", "a.t4", "a.t5", "a.t6", "a.t7", "a.t8", "a.t9", "a.t10"}
	s.NoPKTables = []string{"b.log"}
	s.WPAutoload = []string{"wp.wp_options: 2.1 MB autoloaded"}
	s.Status["Select_full_join"] = "4800"
	s.Status["Slow_queries"] = "3"
	s.Vars["long_query_time"] = "10.000000"
	titles := map[string]dbtuning.Insight{}
	for _, in := range BuildRecommendations(s).Insights {
		titles[in.Title] = in
	}
	if in, ok := titles["14 MyISAM tables"]; !ok || len(in.Items) != 11 || in.Items[10] != "and 4 more" {
		t.Errorf("myisam: got %+v", in)
	}
	for _, want := range []string{"1 table without a primary key", "Large WordPress autoloaded options", "Queries joining tables without an index", "Slow queries"} {
		if _, ok := titles[want]; !ok {
			t.Errorf("missing insight %q in %v", want, titles)
		}
	}
	if !strings.Contains(titles["Slow queries"].Detail, "longer than 10 seconds") || !strings.Contains(titles["Slow queries"].Detail, "Turn on slow_query_log") {
		t.Errorf("slow: got %+v", titles["Slow queries"])
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
