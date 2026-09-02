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

// parseWAFRemovals extracts the currently-removed rule IDs and tags from
// a domain conf's Coraza directives block, stripping the hidden
// excludedRuleID/excludedTag sentinel that's always written first.
//
// SecRuleRemoveById accepts multiple space-separated IDs on one line, so
// that line is parsed as-is. SecRuleRemoveByTag only accepts a single tag
// argument per directive, so removed tags are written one per line;
// parsing collects every consecutive SecRuleRemoveByTag line it finds
// (also tolerating legacy files where multiple tags were incorrectly
// joined onto one line).
func parseWAFRemovals(contentStr string) (removedRules, removedTags []string) {
	foundRule := false
	tagBlockStarted, tagBlockEnded := false, false

	for _, line := range strings.Split(contentStr, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case !foundRule && strings.HasPrefix(line, "SecRuleRemoveById"):
			ids := strings.Fields(line)[1:]
			if len(ids) > 0 && ids[0] == excludedRuleID {
				ids = ids[1:]
			}
			removedRules = append(removedRules, ids...)
			foundRule = true
		case !tagBlockEnded && strings.HasPrefix(line, "SecRuleRemoveByTag"):
			tagBlockStarted = true
			for _, tok := range strings.Fields(strings.TrimPrefix(line, "SecRuleRemoveByTag")) {
				tag := strings.Trim(tok, `"`)
				if tag != "" && !strings.EqualFold(tag, excludedTag) {
					removedTags = append(removedTags, tag)
				}
			}
		case tagBlockStarted && !tagBlockEnded:
			tagBlockEnded = true
		}
		if foundRule && tagBlockEnded {
			break
		}
	}
	return removedRules, removedTags
}

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
