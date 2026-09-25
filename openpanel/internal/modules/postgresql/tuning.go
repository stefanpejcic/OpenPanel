package postgresql

import (
	"fmt"
	"math"
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
	"shared_buffers":               "Shared Buffers",
	"effective_cache_size":         "Effective Cache Size",
	"work_mem":                     "Work Memory",
	"maintenance_work_mem":         "Maintenance Work Memory",
	"max_connections":              "Max Connections",
	"max_worker_processes":         "Max Worker Processes",
	"max_parallel_workers":         "Max Parallel Workers",
	"wal_level":                    "WAL Level",
	"checkpoint_completion_target": "Checkpoint Completion Target",
	"max_wal_size":                 "Max WAL Size",
	"log_min_duration_statement":   "Log Min Duration Statement",
	"log_statement":                "Log Statement",
	"autovacuum":                   "Autovacuum",
}

// Setting is one row of pg_settings
type Setting struct {
	Value string
	Unit  string
}

// PgTuningStats is everything the PostgreSQL rules look at, gathered live from the server and its container
type PgTuningStats struct {
	Version           string
	RAMLimit          int64
	MemUsage          int64
	CPUs              float64
	OOMKilled         bool
	Settings          map[string]Setting
	UptimeSecs        int64
	Databases         int
	DataSize          int64
	BlksHit           int64
	BlksRead          int64
	TempFiles         int64
	TempBytes         int64
	CheckpointsTimed  int64
	CheckpointsReq    int64
	Connections       int
	IdleInTransaction int
	ReplicationSlots  int
	LogLines          []string
	AvailableKeys     []string
}

// unitBytes is how many bytes one unit of a pg_settings memory value is
func unitBytes(unit string) int64 {
	switch unit {
	case "B":
		return 1
	case "kB":
		return kb
	case "8kB":
		return 8 * kb
	case "MB":
		return mb
	case "16MB":
		return 16 * mb
	case "GB":
		return gb
	}
	return 0
}

// bytes returns a memory setting in bytes, false for non memory settings
func (s *PgTuningStats) bytes(key string) (int64, bool) {
	st, ok := s.Settings[key]
	if !ok {
		return 0, false
	}
	mult := unitBytes(st.Unit)
	n, err := strconv.ParseInt(st.Value, 10, 64)
	if mult == 0 || err != nil {
		return 0, false
	}
	return n * mult, true
}

func (s *PgTuningStats) number(key string) (int64, bool) {
	st, ok := s.Settings[key]
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(st.Value, 10, 64)
	return n, err == nil
}

// usable is true when the key is shown on the page and this server actually has it
func (s *PgTuningStats) usable(key string) bool {
	_, onServer := s.Settings[key]
	return onServer && slices.Contains(s.AvailableKeys, key)
}

// display writes a setting the way postgresql.conf would, e.g. 128MB or 1000ms
func (s *PgTuningStats) display(key string) string {
	if b, ok := s.bytes(key); ok {
		return formatPgSize(b)
	}
	st := s.Settings[key]
	if st.Unit != "" && st.Value != "-1" {
		return st.Value + st.Unit
	}
	return st.Value
}

func (s *PgTuningStats) logCount(pattern string) int {
	n := 0
	for _, l := range s.LogLines {
		if strings.Contains(strings.ToLower(l), pattern) {
			n++
		}
	}
	return n
}

// formatPgSize writes bytes in postgresql.conf units, which are case sensitive
func formatPgSize(b int64) string {
	switch {
	case b >= gb && b%gb == 0:
		return strconv.FormatInt(b/gb, 10) + "GB"
	case b >= mb && b%mb == 0:
		return strconv.FormatInt(b/mb, 10) + "MB"
	}
	return strconv.FormatInt(b/kb, 10) + "kB"
}

// BuildPgRecommendations applies the PostgreSQL tuning rules to the gathered stats
func BuildPgRecommendations(s PgTuningStats) dbtuning.Report {
	report := dbtuning.NewReport()
	trustCounters := s.UptimeSecs >= dbtuning.MinUptime
	ram := s.RAMLimit

	report.Facts["server"] = "PostgreSQL " + s.Version
	if ram > 0 {
		report.Facts["memory_limit"] = dbtuning.HumanSize(ram)
	}
	if s.MemUsage > 0 {
		report.Facts["memory_used"] = dbtuning.HumanSize(s.MemUsage)
	}
	report.Facts["databases"] = strconv.Itoa(s.Databases)
	report.Facts["data_size"] = dbtuning.HumanSize(s.DataSize)
	report.Facts["uptime"] = dbtuning.FormatUptime(s.UptimeSecs)
	if ram <= 0 {
		report.Notes = append(report.Notes, dbtuning.NoteNoMemoryLimit)
	}
	if !trustCounters {
		report.Notes = append(report.Notes, dbtuning.NoteRecentRestart)
	}

	add := func(key, recommended, reason string) {
		if !s.usable(key) {
			return
		}
		cur := s.display(key)
		if cur == recommended {
			return
		}
		report.Recommendations = append(report.Recommendations, dbtuning.Recommendation{Key: key, Label: confLabels[key], Current: cur, Recommended: recommended, Reason: reason})
	}

	// shared_buffers
	sharedTarget := int64(0)
	if cur, ok := s.bytes("shared_buffers"); ok && ram > 0 {
		quarter := dbtuning.RoundUp(ram/4, mb)
		if s.OOMKilled {
			quarter = dbtuning.RoundUp(ram*3/20, mb)
		}
		need := max(dbtuning.RoundUp(s.DataSize*3/2, 16*mb), 128*mb)
		hitRatio := 1.0
		if total := s.BlksHit + s.BlksRead; total > 0 {
			hitRatio = float64(s.BlksHit) / float64(total)
		}
		switch {
		case cur > ram*2/5 || (s.OOMKilled && cur > quarter):
			sharedTarget = quarter
			reason := fmt.Sprintf("Shared buffers use %s of the %s memory limit. PostgreSQL also relies on the operating system cache and needs memory for connections, so about a quarter of the memory is recommended.", dbtuning.HumanSize(cur), dbtuning.HumanSize(ram))
			if s.OOMKilled {
				reason = fmt.Sprintf("The database service was stopped for running out of memory. Lowering shared buffers from %s leaves more room for connections and queries.", dbtuning.HumanSize(cur))
			}
			add("shared_buffers", formatPgSize(sharedTarget), reason)
		case trustCounters && hitRatio < 0.99 && cur < quarter:
			sharedTarget = quarter
			add("shared_buffers", formatPgSize(sharedTarget), fmt.Sprintf("Only %.1f%% of reads were served from shared buffers, the rest went to disk. Raising shared buffers to a quarter of the %s memory limit keeps more of your %s of data in memory.", hitRatio*100, dbtuning.HumanSize(ram), dbtuning.HumanSize(s.DataSize)))
		case need > cur*6/5 && cur < quarter:
			sharedTarget = min(need, quarter)
			add("shared_buffers", formatPgSize(sharedTarget), fmt.Sprintf("Your databases use %s, more than shared buffers can hold. A larger value keeps more of your data in memory and reduces disk reads.", dbtuning.HumanSize(s.DataSize)))
		}
	}
	effectiveShared := sharedTarget
	if effectiveShared == 0 {
		effectiveShared, _ = s.bytes("shared_buffers")
	}

	// effective_cache_size is only a planner hint, it doesn't allocate anything
	if cur, ok := s.bytes("effective_cache_size"); ok && ram > 0 {
		target := dbtuning.RoundUp(ram/2, mb)
		if cur > ram || cur*4 < target*3 || cur > target*3/2 {
			add("effective_cache_size", formatPgSize(target), fmt.Sprintf("This tells the query planner how much memory is available for caching data. With a %s memory limit about half of it is realistic, a wrong value makes PostgreSQL pick slower query plans.", dbtuning.HumanSize(ram)))
		}
	}

	maxConn, _ := s.number("max_connections")

	// work_mem
	if cur, ok := s.bytes("work_mem"); ok && ram > 0 && maxConn > 0 {
		ceiling := dbtuning.Clamp(dbtuning.RoundUp((ram-effectiveShared)/(maxConn*3), mb)-mb, 4*mb, 64*mb)
		switch {
		case cur*maxConn > ram:
			add("work_mem", formatPgSize(ceiling), fmt.Sprintf("Every sort or hash in every connection can use up to %s, so with %d connections queries could need more than the %s memory limit. %s is safer for this server.", dbtuning.HumanSize(cur), maxConn, dbtuning.HumanSize(ram), formatPgSize(ceiling)))
		case trustCounters && s.TempFiles > 0 && cur < ceiling:
			target := min(max(cur*2, dbtuning.RoundUp(s.TempBytes/max(s.TempFiles, 1), mb)), ceiling)
			if target > cur {
				add("work_mem", formatPgSize(target), fmt.Sprintf("%d queries wrote %s of temporary files to disk because their sorts didn't fit in memory. More work memory lets them sort in memory, which is much faster.", s.TempFiles, dbtuning.HumanSize(s.TempBytes)))
			}
		}
	}

	// maintenance_work_mem speeds up VACUUM, CREATE INDEX and restores
	if cur, ok := s.bytes("maintenance_work_mem"); ok && ram > 0 {
		target := dbtuning.Clamp(dbtuning.RoundUp(ram/16, mb), 64*mb, gb)
		if target > cur*3/2 || cur > ram/4 {
			add("maintenance_work_mem", formatPgSize(target), fmt.Sprintf("Used by VACUUM, CREATE INDEX and database imports. About 1/16 of the %s memory limit makes them faster without taking memory from your queries.", dbtuning.HumanSize(ram)))
		}
	}

	// max_connections
	if maxConn > 0 {
		refused := s.logCount("too many clients")
		if refused > 0 || (s.Connections > 0 && int64(s.Connections)*100 >= maxConn*85) {
			target := dbtuning.RoundUp(max(maxConn+50, int64(s.Connections)*3/2), 10)
			if ram > 0 {
				workMem, _ := s.bytes("work_mem")
				perConn := 10*mb + workMem
				target = min(target, max((ram-effectiveShared)/perConn/10*10, maxConn))
			}
			if target > maxConn {
				reason := fmt.Sprintf("%d of the %d allowed connections are in use.", s.Connections, maxConn)
				if refused > 0 {
					reason = fmt.Sprintf("The log shows %d connections refused with \"too many clients\".", refused)
				}
				add("max_connections", strconv.FormatInt(target, 10), reason+" Raising the limit prevents connection errors on your websites.")
			}
		}
	}

	// parallel workers can't use more CPUs than the container has
	if s.CPUs > 0 {
		cpus := int64(math.Ceil(s.CPUs))
		if cur, ok := s.number("max_parallel_workers"); ok && cur > cpus {
			add("max_parallel_workers", strconv.FormatInt(cpus, 10), fmt.Sprintf("The database service is limited to %s CPU. Running more parallel workers than CPUs makes queries wait for each other instead of running faster.", strconv.FormatFloat(s.CPUs, 'f', -1, 64)))
		}
		if cur, ok := s.number("max_worker_processes"); ok && cpus > cur {
			add("max_worker_processes", strconv.FormatInt(cpus, 10), fmt.Sprintf("The database service has %d CPUs, raising this lets PostgreSQL use all of them for background and parallel work.", cpus))
		}
	}

	// wal_level logical writes extra WAL only needed for logical replication
	if v := s.Settings["wal_level"].Value; v == "logical" && s.ReplicationSlots == 0 {
		add("wal_level", "replica", "WAL level is set to logical but there are no replication slots. The replica level writes less to disk and still supports backups and streaming replication.")
	}

	// checkpoints
	if v := s.Settings["checkpoint_completion_target"].Value; v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f < 0.9 {
			add("checkpoint_completion_target", "0.9", "Spreading checkpoint writes over 90% of the checkpoint interval avoids sudden bursts of disk I/O that slow down queries.")
		}
	}
	if cur, ok := s.bytes("max_wal_size"); ok && trustCounters {
		total := s.CheckpointsTimed + s.CheckpointsReq
		frequent := s.logCount("checkpoints are occurring too frequently")
		if (total > 10 && s.CheckpointsReq*10 > total*3) || frequent > 0 {
			add("max_wal_size", formatPgSize(min(cur*2, 4*gb)), fmt.Sprintf("%d of %d checkpoints were forced because the WAL filled up before the checkpoint timeout. A larger max WAL size means fewer, calmer checkpoints.", s.CheckpointsReq, total))
		}
	}

	// logging every statement costs disk and time
	if v := s.Settings["log_min_duration_statement"].Value; v == "0" {
		add("log_min_duration_statement", "1000ms", "Every statement is being logged, which slows the server down and fills the disk. Logging only statements slower than 1 second still shows the slow queries.")
	}
	if v := s.Settings["log_statement"].Value; v == "all" || v == "mod" {
		add("log_statement", "ddl", "Logging every "+map[string]string{"all": "statement", "mod": "data change"}[v]+" slows the server down and fills the disk. Logging only schema changes (ddl) is enough for most sites.")
	}

	if v := s.Settings["autovacuum"].Value; v == "off" {
		add("autovacuum", "on", "Autovacuum is off, so deleted and updated rows are never cleaned up. Tables and indexes keep growing and queries get slower over time.")
	}

	if s.IdleInTransaction >= 5 {
		report.Notes = append(report.Notes, fmt.Sprintf("%d connections are idle in a transaction. These hold locks and block autovacuum, check the application that opens them.", s.IdleInTransaction))
	}

	return report
}
