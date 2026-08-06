// Package serverinfo implements the server information page,
// current/historical resource usage pages, and the /json/system/hosting/*
// endpoints those pages fetch from.
package serverinfo

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

func injected(a *appctx.App, r *http.Request) (data map[string]any, err error) {
	userID, _ := auth.UserID(r)
	return a.InjectData(r.Context(), userID)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// platformInfo holds the uname-derived fields shown on the server info
// page.
type platformInfo struct {
	System, Node, Release, Version, Machine, Processor string
}

func unameToString(b [65]int8) string {
	buf := make([]byte, 0, len(b))
	for _, c := range b {
		if c == 0 {
			break
		}
		buf = append(buf, byte(c))
	}
	return string(buf)
}

// getPlatformInfo reads system/node/release/version/machine via
// syscall.Uname (no subprocess needed), and shells out to `uname -p` for
// the processor field, which has no equivalent in the uname struct.
func getPlatformInfo() platformInfo {
	var info platformInfo
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err == nil {
		info.System = unameToString(uts.Sysname)
		info.Node = unameToString(uts.Nodename)
		info.Release = unameToString(uts.Release)
		info.Version = unameToString(uts.Version)
		info.Machine = unameToString(uts.Machine)
	}
	if out, err := exec.Command("uname", "-p").Output(); err == nil {
		info.Processor = strings.TrimSpace(string(out))
	}
	return info
}

var (
	uptimeRE  = regexp.MustCompile(`up (.*?),`)
	loadAvgRE = regexp.MustCompile(`load average: (.*)$`)
)

// getUptimeAndLoad shells out to `uptime` and regex-extracts the two
// pieces callers actually want.
func getUptimeAndLoad() (uptime, loadAvg string) {
	out, err := exec.Command("uptime").Output()
	if err != nil {
		return err.Error(), ""
	}
	output := strings.TrimSpace(string(out))
	if m := uptimeRE.FindStringSubmatch(output); m != nil {
		uptime = strings.TrimSpace(m[1])
	}
	if m := loadAvgRE.FindStringSubmatch(output); m != nil {
		loadAvg = strings.TrimSpace(m[1])
	}
	return uptime, loadAvg
}

// humanValue is the shared {pct, human} shape used throughout
// resource_usage.txt's JSON lines.
type humanValue struct {
	Pct   float64 `json:"pct"`
	Human string  `json:"human"`
}

type humanOnly struct {
	Human string `json:"human"`
}

// ResourceUsageLine is the shared per-line schema of
// /home/<context>/resource_usage.txt, used by both the current-usage page
// (last line only) and the usage-history page (every line).
type ResourceUsageLine struct {
	Timestamp string `json:"timestamp"`
	CPU       struct {
		Usage humanValue `json:"usage"`
		Total humanValue `json:"total"`
	} `json:"cpu"`
	Memory struct {
		UsagePct  float64   `json:"usage_pct"`
		Used      humanOnly `json:"used"`
		Total     humanOnly `json:"total"`
		Available humanOnly `json:"available"`
	} `json:"memory"`
	Bandwidth struct {
		TotalSent humanOnly `json:"total_sent"`
		Limit     humanOnly `json:"limit"`
		UsagePct  float64   `json:"usage_pct"`
	} `json:"bandwidth"`
	Warning string `json:"warning"`
}
