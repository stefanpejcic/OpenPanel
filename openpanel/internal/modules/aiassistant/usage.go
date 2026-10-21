package aiassistant

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

var usersDir = "/etc/openpanel/openpanel/core/users"

var usageMu sync.Mutex

const keepMonths = 12

// monthUsage is one month of a user's AI usage, OpenAdmin reads the same file to show it per user
type monthUsage struct {
	Requests     int   `json:"requests"`
	Actions      int   `json:"actions"`
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

type usageFile struct {
	Months   map[string]*monthUsage `json:"months"`
	LastUsed string                 `json:"last_used,omitempty"`
}

func usagePath(username string) string {
	return filepath.Join(usersDir, username, "ai_usage.json")
}

func currentMonth() string { return time.Now().UTC().Format("2006-01") }

func loadUsage(username string) usageFile {
	f := usageFile{Months: map[string]*monthUsage{}}
	raw, err := tenantfs.ReadFile(usagePath(username))
	if err != nil {
		return f
	}
	_ = json.Unmarshal(raw, &f)
	if f.Months == nil {
		f.Months = map[string]*monthUsage{}
	}
	// files from before per-month history only had the current month's request count
	var old struct {
		Month    string `json:"month"`
		Requests int    `json:"requests"`
	}
	if json.Unmarshal(raw, &old) == nil && old.Month != "" && f.Months[old.Month] == nil {
		f.Months[old.Month] = &monthUsage{Requests: old.Requests}
	}
	return f
}

func (f usageFile) month() *monthUsage {
	m := f.Months[currentMonth()]
	if m == nil {
		m = &monthUsage{}
		f.Months[currentMonth()] = m
	}
	return m
}

func saveUsage(username string, f usageFile) {
	keys := make([]string, 0, len(f.Months))
	for k := range f.Months {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for len(keys) > keepMonths {
		delete(f.Months, keys[0])
		keys = keys[1:]
	}
	f.LastUsed = time.Now().UTC().Format(time.RFC3339)
	if raw, err := json.Marshal(f); err == nil {
		_ = tenantfs.MkdirAll(filepath.Dir(usagePath(username)), 0o755)
		_ = tenantfs.WriteFile(usagePath(username), raw, 0o600)
	}
}

func readUsage(username string) monthUsage {
	usageMu.Lock()
	defer usageMu.Unlock()
	return *loadUsage(username).month()
}

// takeRequest counts one question against the user's monthly limit, false when they're already out
func takeRequest(username string, limit int) (monthUsage, bool) {
	usageMu.Lock()
	defer usageMu.Unlock()
	f := loadUsage(username)
	m := f.month()
	if limit > 0 && m.Requests >= limit {
		return *m, false
	}
	m.Requests++
	saveUsage(username, f)
	return *m, true
}

func addUsage(username string, inTokens, outTokens int64, actions int) {
	usageMu.Lock()
	defer usageMu.Unlock()
	f := loadUsage(username)
	m := f.month()
	m.InputTokens += inTokens
	m.OutputTokens += outTokens
	m.Actions += actions
	saveUsage(username, f)
}
