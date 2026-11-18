package serverinfo

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// the thresholds opencli docker-collect_stats marks a sample's warning with
const (
	memWarnPct = 85
	cpuWarnPct = 85
	bwWarnPct  = 90
)

// parseUsageTime reads a sample's timestamp, collect_stats writes RFC3339 in UTC and older files used a local "2006-01-02 15:04:05"
func parseUsageTime(s string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// cpuOfLimit is CPU use against the plan's cores, the number collect_stats compares with its threshold
func cpuOfLimit(l ResourceUsageLine) int {
	if l.CPU.Total.Pct <= 0 {
		return 0
	}
	return int(l.CPU.Usage.Pct / l.CPU.Total.Pct * 100)
}

// warningFor is the sample's warning, rebuilt from collect_stats' thresholds for lines written without one
func warningFor(l ResourceUsageLine) string {
	if l.Warning != "" {
		return l.Warning
	}
	var parts []string
	if mem := int(l.Memory.UsagePct); mem >= memWarnPct {
		parts = append(parts, fmt.Sprintf("Memory at %d%%", mem))
	}
	if cpu := cpuOfLimit(l); cpu >= cpuWarnPct {
		parts = append(parts, fmt.Sprintf("CPU at %d%%", cpu))
	}
	if bw := int(l.Bandwidth.UsagePct); bw >= bwWarnPct {
		parts = append(parts, fmt.Sprintf("Bandwidth at %d%%", bw))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ") + " — above threshold"
}

// HistorySample is one hourly sample on the usage history page, T is unix milliseconds for the charts
type HistorySample struct {
	T      int64   `json:"t"`
	CPU    float64 `json:"cpu"`
	CPUPct int     `json:"cpuPct"`
	Mem    int64   `json:"mem"`
	MemPct float64 `json:"memPct"`
	Tasks  int     `json:"tasks"`
	BW     int64   `json:"bw"`
	Warn   string  `json:"warn,omitempty"`
}

// UsageLimits are the plan's limits from the newest sample, zero when the file doesn't say
type UsageLimits struct {
	MemBytes int64   `json:"mem"`
	Cores    float64 `json:"cores"`
	Tasks    int     `json:"tasks"`
	BWBits   int64   `json:"bw"`
}

// historySamples turns the file's lines into samples, oldest first, skipping lines without a readable time
func historySamples(lines []ResourceUsageLine) ([]HistorySample, UsageLimits) {
	out := make([]HistorySample, 0, len(lines))
	var limits UsageLimits
	for _, l := range lines {
		ts, ok := parseUsageTime(l.Timestamp)
		if !ok {
			continue
		}
		out = append(out, HistorySample{
			T: ts.UnixMilli(), CPU: l.CPU.Usage.Pct / 100, CPUPct: cpuOfLimit(l),
			Mem: l.Memory.Used.Bytes, MemPct: l.Memory.UsagePct, Tasks: l.Tasks.Current,
			BW: l.Bandwidth.TotalSent.Bytes, Warn: warningFor(l),
		})
		limits = UsageLimits{MemBytes: l.Memory.Total.Bytes, Cores: l.CPU.Total.Pct / 100, Tasks: l.Tasks.Limit, BWBits: l.Bandwidth.Limit.Bits}
	}
	// collect_stats appends, sort anyway in case lines were merged by hand
	sort.SliceStable(out, func(i, j int) bool { return out[i].T < out[j].T })
	return out, limits
}

// historyRanges are the range buttons the retention leaves room for, 0 keeps every sample
func historyRanges(retention int) []string {
	ranges := []string{"24h"}
	if retention == 0 || retention > 24 {
		ranges = append(ranges, "3d")
	}
	if retention == 0 || retention > 72 {
		ranges = append(ranges, "7d")
	}
	return ranges
}

// Time is the sample's time in UTC, which is how collect_stats writes it
func (s HistorySample) Time() time.Time { return time.UnixMilli(s.T).UTC() }

// Cores prints cores in use like the charts do, 2 decimals under one core
func (s HistorySample) Cores() string { return fmtCores(s.CPU) }

// GB prints memory used in GB, 1 decimal from 10 GB up
func (s HistorySample) GB() string {
	g := float64(s.Mem) / (1 << 30)
	if g >= 10 {
		return strconv.FormatFloat(g, 'f', 1, 64)
	}
	return strconv.FormatFloat(g, 'f', 2, 64)
}

// MemHot is memory at collect_stats' warning level
func (s HistorySample) MemHot() bool { return int(s.MemPct) >= memWarnPct }

// MemBar is the memory bar's width, at least a sliver so zero still shows
func (s HistorySample) MemBar() float64 { return math.Min(math.Max(s.MemPct, 1), 100) }

// Bandwidth prints the bytes sent in binary units
func (s HistorySample) Bandwidth() string { return humanBinary(s.BW) }

func fmtCores(c float64) string {
	if c >= 1 {
		return strconv.FormatFloat(c, 'f', 1, 64)
	}
	return strconv.FormatFloat(c, 'f', 2, 64)
}

func humanBinary(b int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	v, i := float64(b), 0
	for v >= 1024 && i < len(units)-1 {
		v /= 1024
		i++
	}
	if i == 0 {
		return strconv.FormatInt(b, 10) + " B"
	}
	return strconv.FormatFloat(v, 'f', 1, 64) + " " + units[i]
}
