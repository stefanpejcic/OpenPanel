package aiassistant

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

var permMu sync.Mutex

// actions the user ticked "Always allow" on, only non-destructive ones ever get here
type permissions struct {
	AlwaysAllow []string `json:"always_allow"`
}

func permPath(username string) string {
	return filepath.Join(usersDir, username, "ai_permissions.json")
}

func loadPermissions(username string) permissions {
	var p permissions
	if raw, err := os.ReadFile(permPath(username)); err == nil {
		_ = json.Unmarshal(raw, &p)
	}
	return p
}

func isAlwaysAllowed(username, name string) bool {
	permMu.Lock()
	defer permMu.Unlock()
	for _, n := range loadPermissions(username).AlwaysAllow {
		if n == name {
			return true
		}
	}
	return false
}

func setAlwaysAllowed(username, name string, allow bool) error {
	permMu.Lock()
	defer permMu.Unlock()
	p := loadPermissions(username)
	var out []string
	for _, n := range p.AlwaysAllow {
		if n != name {
			out = append(out, n)
		}
	}
	if allow {
		out = append(out, name)
	}
	sort.Strings(out)
	p.AlwaysAllow = out
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(permPath(username)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(permPath(username), raw, 0o600)
}
