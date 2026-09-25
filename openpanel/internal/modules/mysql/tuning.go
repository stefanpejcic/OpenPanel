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
	"key_buffer_size":         "Key Buffer Size",
	"innodb_buffer_pool_size": "InnoDB Buffer Pool Size",
	"innodb_log_file_size":    "InnoDB Log File Size",
	"innodb_log_buffer_size":  "InnoDB Log Buffer Size",
	"max_heap_table_size":     "Max Heap Table Size",
	"tmp_table_size":          "Temporary Table Size",
	"max_connections":         "Max Connections",
	"thread_cache_size":       "Thread Cache Size",
	"performance_schema":      "Performance Schema",
	"wait_timeout":            "Wait Timeout",
	"interactive_timeout":     "Interactive Timeout",
	"max_allowed_packet":      "Max Allowed Packet",
	"sort_buffer_size":        "Sort Buffer Size",
	"join_buffer_size":        "Join Buffer Size",
	"read_buffer_size":        "Read Buffer Size",
	"read_rnd_buffer_size":    "Read Random Buffer Size",
	"long_query_time":         "Long Query Time",
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
		need := max(innodbUsed*3/2, 128*mb)
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

	// innodb_log_file_size, not used when the server sizes redo logs with innodb_redo_log_capacity
	if cur, ok := s.varSize("innodb_log_file_size"); ok && s.usable("innodb_log_file_size") && s.Vars["innodb_redo_log_capacity"] == "" {
		target := dbtuning.Clamp(dbtuning.RoundUp(effectivePool/4, mb), 48*mb, 2*gb)
		tooSmall := s.logCount("greater than 10% of redo log size") > 0
		if tooSmall {
			target = max(target, dbtuning.Clamp(cur*2, 48*mb, 2*gb))
		}
		if target > cur*6/5 {
			reason := fmt.Sprintf("To reduce disk I/O caused by flushing checkpoint activity, increase InnoDB Log File Size. This is based on the value %s for InnoDB Buffer Pool Size.", formatSize(effectivePool))
			if tooSmall {
				reason = "The error log shows transactions that are too large for the redo log. " + reason
			}
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
	if ok1 && ok2 && ram > 0 {
		target := dbtuning.Clamp(dbtuning.RoundUp(ram/100, mb), 16*mb, 256*mb)
		reasonTmp := "We recommend raising this setting's value to 1% of the available memory. It prevents temporary tables from being created on disk, which is slower."
		reasonHeap := "We recommend raising this setting's value to 1% of the available memory. It will better accommodate MEMORY tables as they become necessary."
		if trustCounters {
			disk, all := s.status("Created_tmp_disk_tables"), s.status("Created_tmp_tables")
			if all > 100 && float64(disk)/float64(all) > 0.25 {
				target = dbtuning.Clamp(max(target, min(tmpCur, heapCur)*2), 16*mb, min(256*mb, ram/16))
				pct := float64(disk) * 100 / float64(all)
				reasonTmp = fmt.Sprintf("%.0f%% of temporary tables were written to disk because they didn't fit in memory. Raising this lets more of them stay in memory.", pct)
				reasonHeap = "Temporary tables are limited by the lower of Temporary Table Size and Max Heap Table Size, so both need to be raised together."
			}
		}
		if target > tmpCur*11/10 && s.usable("tmp_table_size") {
			add("tmp_table_size", formatSize(target), reasonTmp)
		}
		if target > heapCur*11/10 && s.usable("max_heap_table_size") {
			add("max_heap_table_size", formatSize(target), reasonHeap)
		}
	}

	// max_connections
	if curStr, ok := s.Vars["max_connections"]; ok && s.usable("max_connections") {
		cur, _ := strconv.ParseInt(curStr, 10, 64)
		maxUsed := s.status("Max_used_connections")
		refused := s.status("Connection_errors_max_connections") + int64(s.logCount("too many connections"))
		perConn := s.perConnectionMemory()
		maxByRAM := int64(0)
		if ram > 0 && perConn > 0 {
			maxByRAM = (ram - s.globalMemory(effectivePool) - 64*mb) / perConn
		}
		switch {
		case cur > 0 && (refused > 0 || (trustCounters && maxUsed*100 >= cur*85)):
			target := dbtuning.RoundUp(max(maxUsed*3/2, cur+50), 10)
			if maxByRAM > 0 {
				target = min(target, dbtuning.RoundUp(maxByRAM, 10)-10)
			}
			if target > cur {
				reason := fmt.Sprintf("Up to %d of the %d allowed connections were in use at the same time.", maxUsed, cur)
				if refused > 0 {
					reason = fmt.Sprintf("%d connections were refused because the limit of %d was reached.", refused, cur)
				}
				add("max_connections", strconv.FormatInt(target, 10), reason+" Raising the limit prevents \"Too many connections\" errors on your websites.")
			}
		case maxByRAM > 0 && cur > maxByRAM && trustCounters && maxUsed < maxByRAM:
			target := max(dbtuning.RoundUp(maxByRAM*9/10, 10)-10, dbtuning.RoundUp(maxUsed*3/2, 10), 20)
			if target < cur {
				add("max_connections", strconv.FormatInt(target, 10), fmt.Sprintf("If all %d connections were busy at once they could use about %s, more than the %s memory limit. At most %d connections were used at the same time, so a lower limit keeps the service from running out of memory.", cur, dbtuning.HumanSize(s.globalMemory(effectivePool)+cur*perConn), dbtuning.HumanSize(ram), maxUsed))
			}
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

	return report
}
