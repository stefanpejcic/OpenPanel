// Package websites (this file) implements a live/recent-window
// unique-visitor-IP counter read straight from Caddy's JSON access log,
// backing the "Live Visitors" widget on the wp/single, websitebuilder and
// app-runtime pages.
package websites

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

type visitorsLogRequest struct {
	ClientIP string              `json:"client_ip"`
	Headers  map[string][]string `json:"headers"`
}

type visitorsLogEntry struct {
	TS      float64            `json:"ts"`
	Request visitorsLogRequest `json:"request"`
}

// readRecentVisitorIPs reads the log file backwards in 1KB blocks (newest
// entries first) so a short recent window on a huge log file doesn't
// require reading the whole thing, stopping as soon as an entry older than
// the window is seen.
func readRecentVisitorIPs(logFile string, seconds int) []string {
	f, err := os.Open(logFile)
	if err != nil {
		return nil
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil
	}

	timeWindow := float64(time.Now().Unix()) - float64(seconds)
	const blockSize = 1024
	uniqueIPs := map[string]bool{}

	var data []byte
	pos := info.Size()

outer:
	for pos > 0 {
		readSize := int64(blockSize)
		if pos < readSize {
			readSize = pos
		}
		pos -= readSize

		buf := make([]byte, readSize)
		if _, err := f.ReadAt(buf, pos); err != nil {
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
			var entry visitorsLogEntry
			if err := json.Unmarshal(line, &entry); err != nil {
				continue
			}
			if entry.TS < timeWindow {
				break outer
			}
			ip := ""
			if xff, ok := entry.Request.Headers["X-Forwarded-For"]; ok && len(xff) > 0 {
				ip = xff[0]
			}
			if ip == "" {
				ip = entry.Request.ClientIP
			}
			if ip != "" {
				uniqueIPs[ip] = true
			}
		}
	}

	ips := make([]string, 0, len(uniqueIPs))
	for ip := range uniqueIPs {
		ips = append(ips, ip)
	}
	return ips
}

// handleVisitors returns the recent unique-visitor-IP count/list for a
// domain. Notably, this doesn't check domain ownership - any logged-in
// user can query any domain's recent visitor count/IPs.
func handleVisitors(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	domain, _ = splitDomainAndFolder(domain)

	seconds := 60
	if v := r.URL.Query().Get("seconds"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			seconds = parsed
		}
	}

	logFile := "/var/log/caddy/domlogs/" + domain + "/access.log"
	ips := readRecentVisitorIPs(logFile, seconds)

	writeJSON(w, http.StatusOK, map[string]any{
		"domain": domain, "seconds": seconds, "count": len(ips), "ips": ips,
	})
}
