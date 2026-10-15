package errorlogs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

const (
	digestEntries    = 20
	digestMessageMax = 300
)

// Digest is what `openpanel error-logs-digest` prints for opencli's daily error report email
type Digest struct {
	Username string         `json:"username"`
	Allowed  bool           `json:"allowed"`
	Hours    int            `json:"hours"`
	Total    int            `json:"total"`
	Websites []WebsiteCount `json:"websites"`
	Entries  []DigestEntry  `json:"entries"`
}

// WebsiteCount is how many errors one website had, "" is errors that couldn't be tied to a website
type WebsiteCount struct {
	Website string `json:"website"`
	Count   int    `json:"count"`
}

// DigestEntry is one error shortened to its first line so the email stays readable
type DigestEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Service string    `json:"service"`
	Website string    `json:"website,omitempty"`
	Message string    `json:"message"`
}

// RunDigestCLI handles `openpanel error-logs-digest <username> [--hours N]`, only errors and critical entries count
func RunDigestCLI(ctx context.Context, a *appctx.App, args []string, w io.Writer) int {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		fmt.Fprintln(w, "usage: openpanel error-logs-digest <username> [--hours N]")
		return 2
	}
	username, hours := args[0], 24
	for i := 1; i < len(args); i++ {
		if args[i] == "--hours" && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil && n > 0 && n <= 24*30 {
				hours = n
			}
			i++
		}
	}

	d, err := buildDigest(ctx, a, username, hours, time.Now())
	if err != nil {
		fmt.Fprintln(w, "error:", err)
		return 1
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if enc.Encode(d) != nil {
		return 1
	}
	return 0
}

func buildDigest(ctx context.Context, a *appctx.App, username string, hours int, now time.Time) (Digest, error) {
	d := Digest{Username: username, Hours: hours, Websites: []WebsiteCount{}, Entries: []DigestEntry{}}

	var userID int
	if err := a.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE username = ?", username).Scan(&userID); err != nil {
		return d, fmt.Errorf("user %s not found", username)
	}
	userContext, err := a.QueryContextByUsername(ctx, username)
	if err != nil || userContext == "" {
		return d, fmt.Errorf("no context for user %s", username)
	}

	// same rule as the sidebar: the admin has the module on and the user's plan has the feature
	features, _ := a.LoadUserFeatures(ctx, username, userContext)
	d.Allowed = a.ModuleEnabled("error_logs") && slices.Contains(features, "error_logs")
	if !d.Allowed {
		return d, nil
	}

	domains, _ := a.AllDomainsForUser(ctx, userID)
	sites := loadSites(ctx, a, userID)
	res := Collect(ctx, a, userContext, domains, sites, Query{Level: "error", From: now.Add(-time.Duration(hours) * time.Hour).UTC()})
	return summarize(d, res.Entries), nil
}

// summarize counts errors per website and keeps the newest ones for the email
func summarize(d Digest, entries []Entry) Digest {
	d.Total = len(entries)
	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Domain]++
	}
	for site, n := range counts {
		d.Websites = append(d.Websites, WebsiteCount{Website: site, Count: n})
	}
	// busiest first, errors without a website last
	sort.Slice(d.Websites, func(i, j int) bool {
		wi, wj := d.Websites[i], d.Websites[j]
		if (wi.Website == "") != (wj.Website == "") {
			return wj.Website == ""
		}
		if wi.Count != wj.Count {
			return wi.Count > wj.Count
		}
		return wi.Website < wj.Website
	})

	for _, e := range entries {
		if len(d.Entries) == digestEntries {
			break
		}
		msg, _, _ := strings.Cut(e.Message, "\n")
		if len(msg) > digestMessageMax {
			msg = msg[:digestMessageMax] + "..."
		}
		d.Entries = append(d.Entries, DigestEntry{Time: e.Time, Level: e.Level, Service: e.Service, Website: e.Domain, Message: msg})
	}
	return d
}
