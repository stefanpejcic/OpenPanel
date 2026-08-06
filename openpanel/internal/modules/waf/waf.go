// Package waf handles per-domain OWASP Coraza WAF enable/disable
// toggling, excluded rule ID/tag management, and the Coraza access log
// viewer.
package waf

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

// excludedRuleID and excludedTag are always the first (hidden) entry in
// SecRuleRemoveById/SecRuleRemoveByTag, stripped from what's shown/edited
// in the UI.
const (
	excludedRuleID = "007"
	excludedTag    = "example"
)

func injected(a *appctx.App, r *http.Request) (username string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", err
	}
	username, _ = data["current_username"].(string)
	return username, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// firstPathSegment drops an accidental subfolder suffix from a path
// parameter, keeping only the leading domain component.
func firstPathSegment(s string) string {
	if idx := strings.Index(s, "/"); idx != -1 {
		return s[:idx]
	}
	return s
}

func domainConfigPath(domainName string) string {
	return "/etc/openpanel/caddy/domains/" + domainName + ".conf"
}

func wafLogPath(domainName string) string {
	return "/var/log/caddy/coraza_waf/" + domainName + ".log"
}

// wafLogStats is the {"checks", "blocks"} summary returned by readWAFLogs.
type wafLogStats struct {
	Checks int `json:"checks"`
	Blocks int `json:"blocks"`
}

type wafLogEntry struct {
	Transaction struct {
		Timestamp     string `json:"timestamp"`
		IsInterrupted bool   `json:"is_interrupted"`
	} `json:"transaction"`
}

// readWAFLogs scans a Coraza JSON-lines log file
// backwards in 1KB blocks, counting checks/blocks within the last
// `seconds`, and stops as soon as it reaches a line older than the
// window (the log is chronological, so anything before that is older
// still).
func readWAFLogs(path string, seconds int) wafLogStats {
	stats := wafLogStats{}

	f, err := os.Open(path)
	if err != nil {
		return stats
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return stats
	}

	timeWindow := time.Now().Unix() - int64(seconds)

	const blockSize = 1024
	pos := info.Size()
	var data []byte

	for pos > 0 {
		readSize := int64(blockSize)
		if readSize > pos {
			readSize = pos
		}
		pos -= readSize
		buf := make([]byte, readSize)
		if _, readErr := f.ReadAt(buf, pos); readErr != nil {
			break
		}
		data = append(buf, data...)

		lines := bytes.Split(data, []byte("\n"))
		data = lines[0]
		lines = lines[1:]

		for i := len(lines) - 1; i >= 0; i-- {
			line := bytes.TrimSpace(lines[i])
			if len(line) == 0 {
				continue
			}
			var entry wafLogEntry
			if unmarshalErr := json.Unmarshal(line, &entry); unmarshalErr != nil {
				continue
			}
			var ts int64
			if entry.Transaction.Timestamp != "" {
				if t, parseErr := time.ParseInLocation("2006/01/02 15:04:05", entry.Transaction.Timestamp, time.Local); parseErr == nil {
					ts = t.Unix()
				}
			}
			if ts < timeWindow {
				return stats
			}
			stats.Checks++
			if entry.Transaction.IsInterrupted {
				stats.Blocks++
			}
		}
	}

	return stats
}
