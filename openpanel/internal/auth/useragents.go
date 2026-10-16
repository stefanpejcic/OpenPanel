package auth

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

// BlacklistUseragentsPath is the list OpenAdmin > Security > Blocked User Agents edits, one user-agent fragment per line
var BlacklistUseragentsPath = "/etc/openpanel/openpanel/conf/blacklist_useragents.txt"

// UseragentsConfigPath holds blacklist_useragents=yes|no, a var so tests can point it elsewhere
var UseragentsConfigPath = appctx.ConfigPath

// how often the two files are checked for changes, OpenAdmin edits apply without restarting the panel
const useragentsRecheck = 5 * time.Second

type useragentBlocker struct {
	mu        sync.Mutex
	checkedAt time.Time
	cfgMod    time.Time
	listMod   time.Time
	enabled   bool
	entries   []string
}

// BlockUseragents answers 403 to requests whose User-Agent contains an entry of the blacklist, when blacklist_useragents=yes in openpanel.config
func BlockUseragents() func(http.Handler) http.Handler {
	b := &useragentBlocker{}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if b.blocked(r.UserAgent()) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (b *useragentBlocker) blocked(ua string) bool {
	if ua == "" {
		return false
	}
	b.mu.Lock()
	b.refresh()
	enabled, entries := b.enabled, b.entries
	b.mu.Unlock()
	if !enabled {
		return false
	}
	ua = strings.ToLower(ua)
	for _, e := range entries {
		if strings.Contains(ua, e) {
			return true
		}
	}
	return false
}

// refresh re-reads the flag and the list when either file changed, at most every useragentsRecheck
func (b *useragentBlocker) refresh() {
	if !b.checkedAt.IsZero() && time.Since(b.checkedAt) < useragentsRecheck {
		return
	}
	b.checkedAt = time.Now()

	cfgMod, listMod := modTime(UseragentsConfigPath), modTime(BlacklistUseragentsPath)
	if cfgMod.Equal(b.cfgMod) && listMod.Equal(b.listMod) && !b.cfgMod.IsZero() {
		return
	}
	b.cfgMod, b.listMod = cfgMod, listMod

	cfg, err := config.Load(UseragentsConfigPath)
	if err != nil {
		return
	}
	b.enabled = strings.EqualFold(strings.TrimSpace(cfg.Get("blacklist_useragents", "no")), "yes")
	b.entries = parseUseragentList(BlacklistUseragentsPath)
}

// parseUseragentList keeps the non-empty lines that don't start with #, lowercased for a case-insensitive match
func parseUseragentList(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries []string
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.ToLower(strings.TrimSpace(line))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, line)
	}
	return entries
}

func modTime(path string) time.Time {
	st, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return st.ModTime()
}
