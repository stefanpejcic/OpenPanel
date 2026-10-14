package aiassistant

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var usersDir = "/etc/openpanel/openpanel/core/users"

var usageMu sync.Mutex

type usage struct {
	Month    string `json:"month"`
	Requests int    `json:"requests"`
}

func usagePath(username string) string {
	return filepath.Join(usersDir, username, "ai_usage.json")
}

func currentMonth() string { return time.Now().UTC().Format("2006-01") }

func readUsage(username string) usage {
	u := usage{Month: currentMonth()}
	raw, err := os.ReadFile(usagePath(username))
	if err != nil {
		return u
	}
	var stored usage
	if json.Unmarshal(raw, &stored) == nil && stored.Month == u.Month {
		u.Requests = stored.Requests
	}
	return u
}

// takeRequest counts one request against the user's monthly limit, false when they're already out
func takeRequest(username string, limit int) (usage, bool) {
	usageMu.Lock()
	defer usageMu.Unlock()
	u := readUsage(username)
	if limit > 0 && u.Requests >= limit {
		return u, false
	}
	u.Requests++
	if raw, err := json.Marshal(u); err == nil {
		_ = os.MkdirAll(filepath.Dir(usagePath(username)), 0o755)
		_ = os.WriteFile(usagePath(username), raw, 0o600)
	}
	return u, true
}
