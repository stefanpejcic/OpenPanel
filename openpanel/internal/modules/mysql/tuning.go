package mysql

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/dbtuning"
)

const (
	kb = dbtuning.KB
	mb = dbtuning.MB
	gb = dbtuning.GB
)

// confLabels are the human names shown in the confirm dialog
var confLabels = map[string]string{
	"key_buffer_size":                "Key Buffer Size",
	"innodb_buffer_pool_size":        "InnoDB Buffer Pool Size",
	"innodb_log_file_size":           "InnoDB Log File Size",
	"innodb_log_buffer_size":         "InnoDB Log Buffer Size",
	"max_heap_table_size":            "Max Heap Table Size",
	"tmp_table_size":                 "Temporary Table Size",
	"max_connections":                "Max Connections",
	"thread_cache_size":              "Thread Cache Size",
	"performance_schema":             "Performance Schema",
	"wait_timeout":                   "Wait Timeout",
	"interactive_timeout":            "Interactive Timeout",
	"max_allowed_packet":             "Max Allowed Packet",
	"sort_buffer_size":               "Sort Buffer Size",
	"join_buffer_size":               "Join Buffer Size",
	"read_buffer_size":               "Read Buffer Size",
	"read_rnd_buffer_size":           "Read Random Buffer Size",
	"long_query_time":                "Long Query Time",
	"table_open_cache":               "Table Open Cache",
	"table_definition_cache":         "Table Definition Cache",
	"innodb_flush_log_at_trx_commit": "InnoDB Flush Log at Commit",
}

// sizeKeys hold byte sizes, the rest are counts, seconds or ON/OFF
var sizeKeys = map[string]bool{
	"key_buffer_size": true, "innodb_buffer_pool_size": true, "innodb_log_file_size": true, "innodb_log_buffer_size": true,
	"max_heap_table_size": true, "tmp_table_size": true, "max_allowed_packet": true, "sort_buffer_size": true,
	"join_buffer_size": true, "read_buffer_size": true, "read_rnd_buffer_size": true,
}

// TuningStats is everything the recommendation rules look at, gathered live from the server and its container
type TuningStats struct {
	MariaDB       bool
	Version       string
	RAMLimit      int64
	MemUsage      int64
	OOMKilled     bool
	Vars          map[string]string
	Status        map[string]string
	EngineData    map[string]int64
	EngineIndex   map[string]int64
	EngineTables  map[string]int64
	Databases     int
	SleepingConns int
	PerfSchemaMem int64
	LogLines      []string
	AvailableKeys []string
	PHPWorkers    int

	// findings for the insights list, not settings
	MyISAMTables []string
	NoPKTables   []string
	Fragmented   []string
	WPAutoload   []string
}

var sizeRE = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([KMGT]?)B?$`)

// parseSize turns cnf style sizes like 128M or 134217728 into bytes
func parseSize(s string) (int64, bool) {
	m := sizeRE.FindStringSubmatch(strings.ToUpper(strings.TrimSpace(s)))
	if m == nil {
		return 0, false
	}
	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	switch m[2] {
	case "K":
		f *= float64(kb)
	case "M":
		f *= float64(mb)
	case "G":
		f *= float64(gb)
	case "T":
		f *= float64(gb) * 1024
	}
	return int64(f), true
}

// formatSize writes bytes back in the shortest exact cnf unit
func formatSize(b int64) string {
	switch {
	case b >= gb && b%gb == 0:
		return strconv.FormatInt(b/gb, 10) + "G"
	case b >= mb && b%mb == 0:
		return strconv.FormatInt(b/mb, 10) + "M"
	case b >= kb && b%kb == 0:
		return strconv.FormatInt(b/kb, 10) + "K"
	}
	return strconv.FormatInt(b, 10)
}

func (s *TuningStats) varSize(key string) (int64, bool) {
	v, ok := s.Vars[key]
	if !ok {
		return 0, false
	}
	return parseSize(v)
}

func (s *TuningStats) status(key string) int64 {
	n, _ := strconv.ParseInt(s.Status[key], 10, 64)
	return n
}

func (s *TuningStats) logCount(pattern string) int {
	n := 0
	for _, l := range s.LogLines {
		if strings.Contains(strings.ToLower(l), pattern) {
			n++
		}
	}
	return n
}

// usable is true when the key is shown on the page and this server actually has it
func (s *TuningStats) usable(key string) bool {
	_, onServer := s.Vars[key]
	return onServer && slices.Contains(s.AvailableKeys, key)
}

func (s *TuningStats) flavor() string {
	if s.MariaDB {
		return "MariaDB"
	}
	return "MySQL"
}

func (s *TuningStats) totalTables() int64 {
	var n int64
	for _, c := range s.EngineTables {
		n += c
	}
	return n
}

// perConnectionMemory is roughly what one busy connection can allocate on its own
func (s *TuningStats) perConnectionMemory() int64 {
	var total int64
	for _, k := range []string{"sort_buffer_size", "join_buffer_size", "read_buffer_size", "read_rnd_buffer_size", "thread_stack", "binlog_cache_size"} {
		if v, ok := s.varSize(k); ok {
			total += v
		}
	}
	return total
}

// globalMemory is the memory the server allocates once, regardless of connections
func (s *TuningStats) globalMemory(bufferPool int64) int64 {
	total := bufferPool + s.PerfSchemaMem
	for _, k := range []string{"key_buffer_size", "innodb_log_buffer_size", "aria_pagecache_buffer_size", "query_cache_size"} {
		if v, ok := s.varSize(k); ok {
			total += v
		}
	}
	return total
}

// BuildRecommendations applies the tuning rules to the gathered stats
func BuildRecommendations(s TuningStats) dbtuning.Report {
	report := dbtuning.NewReport()
	uptime := s.status("Uptime")
	trustCounters := uptime >= dbtuning.MinUptime

	report.Facts["server"] = s.flavor() + " " + s.Version
	if s.RAMLimit > 0 {
		report.Facts["memory_limit"] = dbtuning.HumanSize(s.RAMLimit)
	}
	if s.MemUsage > 0 {
		report.Facts["memory_used"] = dbtuning.HumanSize(s.MemUsage)
	}
	report.Facts["databases"] = strconv.Itoa(s.Databases)
	report.Facts["innodb_data"] = dbtuning.HumanSize(s.EngineData["InnoDB"] + s.EngineIndex["InnoDB"])
	report.Facts["myisam_data"] = dbtuning.HumanSize(s.EngineData["MyISAM"] + s.EngineIndex["MyISAM"])
	report.Facts["uptime"] = dbtuning.FormatUptime(uptime)
	if s.RAMLimit <= 0 {
		report.Notes = append(report.Notes, dbtuning.NoteNoMemoryLimit)
	}
	if !trustCounters {
		report.Notes = append(report.Notes, dbtuning.NoteRecentRestart)
	}

	add := func(key, recommended, reason string) {
		cur := s.Vars[key]
		if sizeKeys[key] {
			curBytes, ok1 := parseSize(cur)
			recBytes, ok2 := parseSize(recommended)
			if ok1 && ok2 && curBytes == recBytes {
				return
			}
			if ok1 {
				cur = formatSize(curBytes)
			}
		} else if strings.EqualFold(cur, recommended) {
			return
		}
		report.Recommendations = append(report.Recommendations, dbtuning.Recommendation{Key: key, Label: confLabels[key], Current: cur, Recommended: recommended, Reason: reason})
	}

	ram := s.RAMLimit
	innodbUsed := s.EngineData["InnoDB"] + s.EngineIndex["InnoDB"]

	// innodb_buffer_pool_size
	poolTarget := int64(0)
	if cur, ok := s.varSize("innodb_buffer_pool_size"); ok && ram > 0 {
		share := 0.5
		switch {
		case ram >= 4*gb:
			share = 0.7
		case ram >= 3*gb/2:
			share = 0.6
		}
		if s.OOMKilled {
			share = 0.4
		}
		ramShare := int64(float64(ram) * share)
		chunk := int64(128 * mb)
		if c, ok := s.varSize("innodb_buffer_pool_chunk_size"); ok && c > 0 {
			chunk = c
		}
		need := max(dbtuning.RoundUp(innodbUsed*13/10, chunk), 128*mb)
		reads, requests := s.status("Innodb_buffer_pool_reads"), s.status("Innodb_buffer_pool_read_requests")
		missRate := 0.0
		if requests > 0 {
			missRate = float64(reads) / float64(requests)
		}
		reason := ""
		switch {
		case cur > ramShare:
			poolTarget = max(dbtuning.RoundUp(ramShare-chunk+1, chunk), chunk)
			if s.OOMKilled {
				reason = fmt.Sprintf("The database service was stopped for running out of memory. The buffer pool uses %s of the %s memory limit, lowering it leaves room for connections and temporary tables.", dbtuning.HumanSize(cur), dbtuning.HumanSize(ram))
			} else {
				reason = fmt.Sprintf("The buffer pool uses %s of the %s memory limit. That doesn't leave enough memory for connections and temporary tables and risks the service running out of memory.", dbtuning.HumanSize(cur), dbtuning.HumanSize(ram))
			}
		case trustCounters && missRate > 0.01 && cur < ramShare:
			poolTarget = min(dbtuning.RoundUp(max(cur*2, need), chunk), dbtuning.RoundUp(ramShare-chunk+1, chunk))
			reason = fmt.Sprintf("%.1f%% of InnoDB reads have to go to disk because the data doesn't fit in memory. A larger buffer pool keeps more of your %s of InnoDB data and indexes in memory and reduces disk I/O.", missRate*100, dbtuning.HumanSize(innodbUsed))
		case need > cur && cur < ramShare:
			poolTarget = min(dbtuning.RoundUp(need, chunk), dbtuning.RoundUp(ramShare-chunk+1, chunk))
			reason = fmt.Sprintf("Your InnoDB tables use %s for data and indexes, more than the buffer pool can hold. A larger buffer pool allows the database to hold more data in memory and reduces disk I/O.", dbtuning.HumanSize(innodbUsed))
		case cur > need*2 && cur > 256*mb && !(trustCounters && missRate > 0.01):
			poolTarget = need
			reason = fmt.Sprintf("Your InnoDB tables use only %s for data and indexes, but the buffer pool can grow to %s of the %s memory limit. %s leaves room for the data to grow by about a third and frees the rest for connections and temporary tables.", dbtuning.HumanSize(innodbUsed), dbtuning.HumanSize(cur), dbtuning.HumanSize(ram), formatSize(need))
		}
		if poolTarget > 0 && poolTarget != cur && s.usable("innodb_buffer_pool_size") {
			add("innodb_buffer_pool_size", formatSize(poolTarget), reason)
		} else {
			poolTarget = 0
		}
	}
	effectivePool := poolTarget
	if effectivePool == 0 {
		effectivePool, _ = s.varSize("innodb_buffer_pool_size")
	}

	// innodb_log_file_size, sized to hold about an hour of redo writes, not used when the server sizes redo logs with innodb_redo_log_capacity
	if cur, ok := s.varSize("innodb_log_file_size"); ok && s.usable("innodb_log_file_size") && s.Vars["innodb_redo_log_capacity"] == "" {
		files := int64(1)
		if n, err := strconv.ParseInt(s.Vars["innodb_log_files_in_group"], 10, 64); err == nil && n > 0 {
			files = n
		}
		target, reason := int64(0), ""
		if written := s.status("Innodb_os_log_written"); trustCounters && written > 0 {
			perHour := written * 3600 / uptime
			if perHour > cur*files*6/5 {
				target = dbtuning.Clamp(dbtuning.RoundUp(perHour/files, mb), 48*mb, 2*gb)
				reason = fmt.Sprintf("InnoDB writes about %s of redo log per hour, but the redo log only holds %s, so it fills in about %d minutes and forces extra flushing to disk. Sizing it for about an hour of writes reduces disk I/O.", dbtuning.HumanSize(perHour), dbtuning.HumanSize(cur*files), max(cur*files*60/perHour, 1))
			}
		}
		if s.logCount("greater than 10% of redo log size") > 0 {
			target = max(target, dbtuning.Clamp(cur*2, 48*mb, 2*gb))
			reason = strings.TrimSpace("The error log shows transactions that are too large for the redo log. " + reason)
		}
		if target > cur {
			add("innodb_log_file_size", formatSize(target), reason)
		}
	}

	// innodb_log_buffer_size
	if cur, ok := s.varSize("innodb_log_buffer_size"); ok && s.usable("innodb_log_buffer_size") && trustCounters {
		if waits := s.status("Innodb_log_waits"); waits > 0 && cur < 64*mb {
			add("innodb_log_buffer_size", formatSize(min(cur*2, 64*mb)), fmt.Sprintf("InnoDB had to wait for the log buffer to be flushed %d times. A larger log buffer lets bigger transactions run without writing to disk first.", waits))
		}
	}

	// key_buffer_size
	if cur, ok := s.varSize("key_buffer_size"); ok && s.usable("key_buffer_size") {
		myisamIdx := s.EngineIndex["MyISAM"]
		if s.EngineTables["MyISAM"] == 0 {
			if cur > 8*mb {
				add("key_buffer_size", "8M", "Your databases do not appear to use the MyISAM storage engine. We recommend lowering the key buffer size to save memory.")
			}
		} else {
			target := dbtuning.RoundUp(myisamIdx*6/5, mb)
			if ram > 0 {
				target = min(target, ram/4)
			}
			target = max(target, 8*mb)
			if target > cur*6/5 {
				add("key_buffer_size", formatSize(target), fmt.Sprintf("Your MyISAM tables have %s of indexes. A key buffer that fits them avoids reading indexes from disk.", dbtuning.HumanSize(myisamIdx)))
			}
		}
	}

	// tmp_table_size and max_heap_table_size work together, the lower one wins
	tmpCur, ok1 := s.varSize("tmp_table_size")
	heapCur, ok2 := s.varSize("max_heap_table_size")
	if ok1 && ok2 && ram > 0 && trustCounters {
		disk, all := s.status("Created_tmp_disk_tables"), s.status("Created_tmp_tables")
		if all > 100 && float64(disk)/float64(all) > 0.25 {
			target := dbtuning.Clamp(max(min(tmpCur, heapCur)*2, 32*mb), 16*mb, min(256*mb, ram/16))
			pct := float64(disk) * 100 / float64(all)
			reasonTmp := fmt.Sprintf("%.0f%% of temporary tables were written to disk because they didn't fit in memory. Raising this lets more of them stay in memory.", pct)
			reasonHeap := "Temporary tables are limited by the lower of Temporary Table Size and Max Heap Table Size, so both need to be raised together."
			if target > tmpCur && s.usable("tmp_table_size") {
				add("tmp_table_size", formatSize(target), reasonTmp)
			}
			if target > heapCur && s.usable("max_heap_table_size") {
				add("max_heap_table_size", formatSize(target), reasonHeap)
			}
		}
	}

	// max_connections, every PHP worker can hold a connection so they set the floor
	if curStr, ok := s.Vars["max_connections"]; ok && s.usable("max_connections") {
		cur, _ := strconv.ParseInt(curStr, 10, 64)
		maxUsed := s.status("Max_used_connections")
		refused := s.status("Connection_errors_max_connections") + int64(s.logCount("too many connections"))
		perConn := s.perConnectionMemory()
		maxByRAM := int64(0)
		if ram > 0 && perConn > 0 {
			maxByRAM = (ram - s.globalMemory(effectivePool) - 64*mb) / perConn
		}
		floor := int64(0)
		if s.PHPWorkers > 0 {
			floor = dbtuning.RoundUp(int64(s.PHPWorkers)+10, 10)
		}
		capByRAM := func(target int64) int64 {
			if maxByRAM > 0 {
				return min(target, dbtuning.RoundUp(maxByRAM, 10)-10)
			}
			return target
		}
		switch {
		case cur > 0 && (refused > 0 || (trustCounters && maxUsed*100 >= cur*85)):
			target := capByRAM(max(dbtuning.RoundUp(max(maxUsed*3/2, cur+50), 10), floor))
			if target > cur {
				reason := fmt.Sprintf("Up to %d of the %d allowed connections were in use at the same time since the last restart.", maxUsed, cur)
				if refused > 0 {
					reason = fmt.Sprintf("%d connections were refused because the limit of %d was reached.", refused, cur)
				}
				add("max_connections", strconv.FormatInt(target, 10), reason+" Raising the limit prevents \"Too many connections\" errors on your websites.")
			}
		case floor > cur:
			if target := capByRAM(floor); target > cur {
				add("max_connections", strconv.FormatInt(target, 10), fmt.Sprintf("Your PHP-FPM pools can run up to %d workers at once and each one can open a database connection, but only %d connections are allowed. A traffic spike would end in \"Too many connections\" errors on your websites.", s.PHPWorkers, cur))
			}
		case maxByRAM > 0 && cur > maxByRAM && trustCounters && maxUsed < maxByRAM:
			target := max(dbtuning.RoundUp(maxByRAM*9/10, 10)-10, dbtuning.RoundUp(maxUsed*3/2, 10), floor, 20)
			if target < cur {
				add("max_connections", strconv.FormatInt(target, 10), fmt.Sprintf("If all %d connections were busy at once they could use about %s, more than the %s memory limit. At most %d connections were used at the same time since the last restart, so a lower limit keeps the service from running out of memory.", cur, dbtuning.HumanSize(s.globalMemory(effectivePool)+cur*perConn), dbtuning.HumanSize(ram), maxUsed))
			}
		}
	}

	// table_open_cache, when every slot is taken and tables keep getting opened
	if curStr, ok := s.Vars["table_open_cache"]; ok && s.usable("table_open_cache") && trustCounters {
		cur, _ := strconv.ParseInt(curStr, 10, 64)
		open, perHour := s.status("Open_tables"), s.status("Opened_tables")*3600/uptime
		if cur > 0 && open*100 >= cur*95 && perHour > 60 {
			target := dbtuning.Clamp(dbtuning.RoundUp(max(cur*2, s.totalTables()*2), 100), 400, 16384)
			if n, err := strconv.ParseInt(s.Vars["open_files_limit"], 10, 64); err == nil && n > 0 {
				target = min(target, dbtuning.RoundUp(n/2, 100)-100)
			}
			if target > cur {
				add("table_open_cache", strconv.FormatInt(target, 10), fmt.Sprintf("All %d table cache slots are in use and tables are still opened about %d times per hour. Each reopen reads the table from disk again, a bigger cache keeps them open.", cur, perHour))
			}
		}
	}

	// table_definition_cache, should fit every table
	if curStr, ok := s.Vars["table_definition_cache"]; ok && s.usable("table_definition_cache") {
		cur, _ := strconv.ParseInt(curStr, 10, 64)
		tables := s.totalTables()
		if cur > 0 && tables > cur*9/10 && s.status("Open_table_definitions")*100 >= cur*95 {
			target := dbtuning.Clamp(dbtuning.RoundUp(tables+400, 100), 400, 65536)
			if target > cur {
				add("table_definition_cache", strconv.FormatInt(target, 10), fmt.Sprintf("Your databases have %d tables but only %d table definitions can be cached, so definitions are read from disk again and again. Sites with many tables, like WordPress multisite, benefit the most.", tables, cur))
			}
		}
	}

	// innodb_flush_log_at_trx_commit=2 is a speed for safety trade, so only optional
	if v, ok := s.Vars["innodb_flush_log_at_trx_commit"]; ok && v == "1" && s.usable("innodb_flush_log_at_trx_commit") && trustCounters {
		if perSec := s.status("Innodb_os_log_fsyncs") / uptime; perSec >= 5 {
			report.Recommendations = append(report.Recommendations, dbtuning.Recommendation{
				Key: "innodb_flush_log_at_trx_commit", Label: confLabels["innodb_flush_log_at_trx_commit"], Current: v, Recommended: "2", Optional: true,
				Reason: fmt.Sprintf("InnoDB flushes the redo log to disk about %d times per second, once for every commit. With 2 it writes on every commit but flushes once per second, which is much faster for sites that write a lot. If the server crashes or loses power, up to the last second of changes can be lost, so only apply this if that's acceptable.", perSec),
			})
		}
	}

	// thread_cache_size
	if curStr, ok := s.Vars["thread_cache_size"]; ok && s.usable("thread_cache_size") && trustCounters {
		cur, _ := strconv.ParseInt(curStr, 10, 64)
		created, conns := s.status("Threads_created"), s.status("Connections")
		if conns > 1000 && float64(created)/float64(conns) > 0.05 {
			target := dbtuning.Clamp(max(s.status("Max_used_connections"), cur+16), 16, 256)
			if target > cur {
				add("thread_cache_size", strconv.FormatInt(target, 10), fmt.Sprintf("A new thread had to be created for %.0f%% of connections. Caching more threads makes new connections faster.", float64(created)*100/float64(conns)))
			}
		}
	}

	// performance_schema
	if v, ok := s.Vars["performance_schema"]; ok && s.usable("performance_schema") && strings.EqualFold(v, "ON") && ram > 0 && ram <= 2*gb {
		reason := fmt.Sprintf("Performance Schema collects detailed statistics that are mostly useful for debugging. With a %s memory limit, turning it off frees memory for caching your data.", dbtuning.HumanSize(ram))
		if s.PerfSchemaMem > 0 {
			reason = fmt.Sprintf("Performance Schema uses %s of memory to collect detailed statistics that are mostly useful for debugging. With a %s memory limit, turning it off frees that memory for caching your data.", dbtuning.HumanSize(s.PerfSchemaMem), dbtuning.HumanSize(ram))
		}
		add("performance_schema", "OFF", reason)
	}

	// wait_timeout and interactive_timeout
	if curStr, ok := s.Vars["wait_timeout"]; ok && s.usable("wait_timeout") {
		cur, _ := strconv.ParseInt(curStr, 10, 64)
		if cur > 600 && s.SleepingConns >= 10 {
			add("wait_timeout", "600", fmt.Sprintf("%d connections have been idle for more than 5 minutes. Each one holds memory and counts toward Max Connections. Closing idle connections after 10 minutes frees them.", s.SleepingConns))
			if ic, ok := s.Vars["interactive_timeout"]; ok && s.usable("interactive_timeout") {
				if n, _ := strconv.ParseInt(ic, 10, 64); n > 600 {
					add("interactive_timeout", "600", "Matches the lower Wait Timeout so idle interactive connections are closed as well.")
				}
			}
		}
	}

	// max_allowed_packet
	if cur, ok := s.varSize("max_allowed_packet"); ok && s.usable("max_allowed_packet") {
		if n := s.logCount("max_allowed_packet"); n > 0 && cur < gb {
			add("max_allowed_packet", formatSize(min(max(cur*2, 64*mb), gb)), fmt.Sprintf("The error log shows %d queries or connections that failed because a packet was bigger than max_allowed_packet. This usually happens when importing large dumps or saving large rows.", n))
		}
	}

	// per connection buffers set far above the default
	defaults := map[string]int64{"sort_buffer_size": 256 * kb, "join_buffer_size": 256 * kb, "read_buffer_size": 128 * kb, "read_rnd_buffer_size": 256 * kb}
	if s.MariaDB {
		defaults["sort_buffer_size"] = 2 * mb
	}
	for _, key := range []string{"sort_buffer_size", "join_buffer_size", "read_buffer_size", "read_rnd_buffer_size"} {
		cur, ok := s.varSize(key)
		if !ok || !s.usable(key) {
			continue
		}
		def := defaults[key]
		if cur > def*8 && cur > 4*mb {
			add(key, formatSize(def), fmt.Sprintf("This buffer is allocated for every connection that needs it, so %s can add up quickly and even slow queries down. The %s default of %s is best for most workloads.", dbtuning.HumanSize(cur), s.flavor(), formatSize(def)))
		}
	}
	if cur, ok := s.varSize("sort_buffer_size"); ok && s.usable("sort_buffer_size") && trustCounters {
		passes, sorts := s.status("Sort_merge_passes"), s.status("Sort_scan")+s.status("Sort_range")
		if sorts > 1000 && float64(passes)/float64(sorts) > 0.1 && cur < 4*mb {
			add("sort_buffer_size", formatSize(min(cur*2, 4*mb)), fmt.Sprintf("%.0f%% of sorts didn't fit in the sort buffer and had to use temporary files. A slightly larger buffer speeds up ORDER BY and GROUP BY queries.", float64(passes)*100/float64(sorts)))
		}
	}

	// long_query_time only matters when the slow query log is on
	if v, ok := s.Vars["long_query_time"]; ok && s.usable("long_query_time") && strings.EqualFold(s.Vars["slow_query_log"], "ON") {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 10 {
			add("long_query_time", "2", fmt.Sprintf("The slow query log is on but only records queries slower than %s seconds, so most slow website queries are never logged.", strings.TrimRight(strings.TrimRight(v, "0"), ".")))
		}
	}

	report.Insights = buildInsights(s, trustCounters, uptime)
	return report
}

func plural(n int64, what string) string {
	if n == 1 {
		return "1 " + what
	}
	return fmt.Sprintf("%d %ss", n, what)
}

// listed shows the first few names of a finding and how many more there are
func listed(names []string, total int64) []string {
	const show = 10
	if len(names) > show {
		names = names[:show]
	}
	out := append([]string{}, names...)
	if total > int64(len(out)) {
		out = append(out, fmt.Sprintf("and %d more", total-int64(len(out))))
	}
	return out
}

// buildInsights are the findings a setting can't fix, the user or their app has to act on them
func buildInsights(s TuningStats, trustCounters bool, uptime int64) []dbtuning.Insight {
	insights := []dbtuning.Insight{}
	if n := s.EngineTables["MyISAM"]; n > 0 {
		insights = append(insights, dbtuning.Insight{
			Title:  plural(n, "MyISAM table"),
			Detail: "MyISAM locks the whole table on every write and isn't crash safe. Converting these tables to InnoDB (ALTER TABLE name ENGINE=InnoDB) makes them faster under load, and the Key Buffer Size can then be lowered.",
			Items:  listed(s.MyISAMTables, n),
		})
	}
	if len(s.NoPKTables) > 0 {
		insights = append(insights, dbtuning.Insight{
			Title:  plural(int64(len(s.NoPKTables)), "table") + " without a primary key",
			Detail: "InnoDB stores rows in primary key order, without one it uses a hidden key that can't be used in queries, and updates and deletes have to scan the whole table. Add a primary key, usually an auto increment id column.",
			Items:  listed(s.NoPKTables, int64(len(s.NoPKTables))),
		})
	}
	if len(s.WPAutoload) > 0 {
		insights = append(insights, dbtuning.Insight{
			Title:  "Large WordPress autoloaded options",
			Detail: "WordPress loads every autoloaded option on every page view. Above 800 KB this slows down every request, usually because of leftovers from removed plugins or plugins storing caches in the options table.",
			Items:  s.WPAutoload,
		})
	}
	if len(s.Fragmented) > 0 {
		insights = append(insights, dbtuning.Insight{
			Title:  "Fragmented tables",
			Detail: "These tables have a lot of unused space left over from deleted rows. Optimizing them on the Databases page reclaims the disk space and makes full table scans faster.",
			Items:  s.Fragmented,
		})
	}
	if trustCounters {
		if joins := s.status("Select_full_join"); joins*3600/uptime >= 10 {
			insights = append(insights, dbtuning.Insight{
				Title:  "Queries joining tables without an index",
				Detail: fmt.Sprintf("%d queries since the last restart (about %d per hour) joined tables without using an index, which reads every row of the joined table. This usually means a missing index in a plugin or app.", joins, joins*3600/uptime),
			})
		}
		if slow := s.status("Slow_queries"); slow > 0 {
			detail := fmt.Sprintf("%d queries since the last restart took longer than %s seconds.", slow, strings.TrimRight(strings.TrimRight(s.Vars["long_query_time"], "0"), "."))
			if strings.EqualFold(s.Vars["slow_query_log"], "ON") {
				detail += " They are recorded in the slow query log."
			} else {
				detail += " Turn on slow_query_log on this page to record which queries they are."
			}
			insights = append(insights, dbtuning.Insight{Title: "Slow queries", Detail: detail})
		}
	}
	return insights
}
