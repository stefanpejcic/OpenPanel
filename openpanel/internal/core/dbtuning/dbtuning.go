// Package dbtuning holds what the MySQL and PostgreSQL configuration suggestions share: the report types, size helpers and live container limits.
package dbtuning

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
)

const (
	KB = int64(1024)
	MB = 1024 * KB
	GB = 1024 * MB
)

// MinUptime is how long a server must be up before its usage counters are trusted
const MinUptime = 3600

// Recommendation is one suggested change for a configuration page
type Recommendation struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Current     string `json:"current"`
	Recommended string `json:"recommended"`
	Reason      string `json:"reason"`
}

// Report is the recommendations plus the facts they were based on
type Report struct {
	Recommendations []Recommendation  `json:"recommendations"`
	Facts           map[string]string `json:"facts"`
	Notes           []string          `json:"notes"`
}

func NewReport() Report {
	return Report{Recommendations: []Recommendation{}, Facts: map[string]string{}, Notes: []string{}}
}

const (
	NoteNoMemoryLimit = "The memory limit of the database service could not be read, so memory based suggestions are skipped."
	NoteRecentRestart = "The database service was restarted recently. Suggestions based on usage will show up after it has been running for at least an hour."
)

// HumanSize is for reason text, not for config values
func HumanSize(b int64) string {
	switch {
	case b >= GB:
		return strings.TrimSuffix(strconv.FormatFloat(float64(b)/float64(GB), 'f', 1, 64), ".0") + " GB"
	case b >= MB:
		return strconv.FormatInt((b+MB/2)/MB, 10) + " MB"
	case b >= KB:
		return strconv.FormatInt(b/KB, 10) + " KB"
	}
	return strconv.FormatInt(b, 10) + " B"
}

func RoundUp(b, unit int64) int64 {
	if unit <= 0 {
		return b
	}
	return (b + unit - 1) / unit * unit
}

func Clamp(v, lo, hi int64) int64 {
	return max(lo, min(v, hi))
}

// FormatUptime formats seconds as a short uptime like 3d 4h
func FormatUptime(secs int64) string {
	d, h, m := secs/86400, secs%86400/3600, secs%3600/60
	switch {
	case d > 0:
		return fmt.Sprintf("%dd %dh", d, h)
	case h > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// Container is the live resource picture of one database container
type Container struct {
	MemLimit  int64
	MemUsage  int64
	CPUs      float64
	OOMKilled bool
}

// parseMemUsage reads the "232.6MB" half of podman's MemUsage column
func parseMemUsage(s string) int64 {
	s = strings.ToUpper(strings.TrimSpace(s))
	mult := int64(1)
	for _, u := range []struct {
		suffix string
		mult   int64
	}{{"GB", 1000 * 1000 * 1000}, {"MB", 1000 * 1000}, {"KB", 1000}, {"B", 1}} {
		if strings.HasSuffix(s, u.suffix) {
			s, mult = strings.TrimSuffix(s, u.suffix), u.mult
			break
		}
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return int64(f * float64(mult))
}

// InspectContainer reads memory limit, CPU limit, OOM state and current memory use of a service container
func InspectContainer(ctx context.Context, userContext, service string) Container {
	var c Container
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	argv := podmanmanager.PodmanArgv(userContext, "inspect", service, "--format", "{{.HostConfig.Memory}} {{.HostConfig.NanoCpus}} {{.State.OOMKilled}}")
	if out, err := podmanmanager.Command(cctx, userContext, argv).Output(); err == nil {
		if f := strings.Fields(string(out)); len(f) == 3 {
			c.MemLimit, _ = strconv.ParseInt(f[0], 10, 64)
			if nano, _ := strconv.ParseInt(f[1], 10, 64); nano > 0 {
				c.CPUs = float64(nano) / 1e9
			}
			c.OOMKilled = f[2] == "true"
		}
	}
	argv = podmanmanager.PodmanArgv(userContext, "stats", "--no-stream", "--format", "{{.MemUsage}}", service)
	if out, err := podmanmanager.Command(cctx, userContext, argv).Output(); err == nil {
		if used, _, ok := strings.Cut(string(out), "/"); ok {
			c.MemUsage = parseMemUsage(used)
		}
	}
	return c
}

// HostMemory is the fallback when the container has no memory limit
func HostMemory() int64 {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(content), "\n") {
		if f := strings.Fields(line); len(f) >= 2 && f[0] == "MemTotal:" {
			n, _ := strconv.ParseInt(f[1], 10, 64)
			return n * KB
		}
	}
	return 0
}
