package summary

import (
	"fmt"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// emailKey is the Account > Notifications option that mails the summary each morning
const emailKey = "notify_daily_summary"

// Tile is one KPI with the day before's value, HasPrev is false when that day has no summary
type Tile struct {
	Now, Prev int64
	HasPrev   bool
	// UpIsBad flips the badge color for errors, Plain leaves it grey for firewall blocks
	UpIsBad, Plain bool
}

// Delta is the change against the day before like "+12%", empty when there's nothing to compare with
func (t Tile) Delta() string {
	if !t.HasPrev || t.Prev <= 0 {
		return ""
	}
	pct := (t.Now - t.Prev) * 100 / t.Prev
	if pct > 0 {
		return fmt.Sprintf("+%d%%", pct)
	}
	return fmt.Sprintf("%d%%", pct)
}

// PrevNum prints the day before's value like 4,812
func (t Tile) PrevNum() string { return num(int(t.Prev)) }

// Tone is the badge class, green for good news and red for bad
func (t Tile) Tone() string {
	switch up := t.Now > t.Prev; {
	case t.Plain || t.Now == t.Prev:
		return ""
	case up != t.UpIsBad:
		return "is-success"
	}
	return "is-danger"
}

// Tiles are the five numbers at the top of the page
type Tiles struct {
	Requests, Visitors, Bandwidth, Errors, Blocked Tile
}

// tiles compares the day with the saved day before, requests use the count saved with the summary
func tiles(s Summary, prev Summary, hasPrev bool) Tiles {
	t := Tiles{
		Requests:  Tile{Now: int64(s.TotalRequests()), Prev: int64(prev.TotalRequests()), HasPrev: hasPrev},
		Visitors:  Tile{Now: int64(s.TotalVisitors()), Prev: int64(prev.TotalVisitors()), HasPrev: hasPrev},
		Bandwidth: Tile{Now: s.TotalBytes(), Prev: prev.TotalBytes(), HasPrev: hasPrev},
		Errors:    Tile{Now: int64(s.TotalErrors()), Prev: int64(prev.TotalErrors()), HasPrev: hasPrev, UpIsBad: true},
		Blocked:   Tile{Now: int64(s.TotalBlocked()), Prev: int64(prev.TotalBlocked()), HasPrev: hasPrev, Plain: true},
	}
	if s.PrevRequests >= 0 {
		t.Requests.Prev, t.Requests.HasPrev = int64(s.PrevRequests), true
	}
	return t
}

// dayBefore is the date string one day before d, empty when d isn't a date
func dayBefore(d string) string {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return ""
	}
	return t.AddDate(0, 0, -1).Format("2006-01-02")
}

// DayLabel prints a saved date for the day picker, "Wed 28 Oct"
func (PageData) DayLabel(d string) string {
	t, err := time.Parse("2006-01-02", d)
	if err != nil {
		return d
	}
	return t.Format("Mon 2 Jan")
}

func notificationsPath(username string) string { return usersDir + username + "/notifications.yaml" }

// emailOn reads the user's notify_daily_summary, off when the file or the key is missing like the account page's default
func emailOn(username string) bool {
	raw, err := tenantfs.ReadFile(notificationsPath(username))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(line), "="); ok && strings.TrimSpace(k) == emailKey {
			return strings.TrimSpace(v) == "1"
		}
	}
	return false
}

// setPref sets key=value in a notifications.yaml, keeping every other line as it was
func setPref(content, key, value string) string {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if content == "" {
		lines = nil
	}
	found := false
	for i, line := range lines {
		if k, _, ok := strings.Cut(strings.TrimSpace(line), "="); ok && strings.TrimSpace(k) == key {
			lines[i] = key + "=" + value
			found = true
		}
	}
	if !found {
		lines = append(lines, key+"="+value)
	}
	return strings.Join(lines, "\n") + "\n"
}

// sizeParts splits bytes into a number and binary unit for the page, one decimal from KB up like the theme pages
func sizeParts(b int64) (string, string) {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	v, i := float64(b), 0
	for v >= 1024 && i < len(units)-1 {
		v /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprint(b), "B"
	}
	return fmt.Sprintf("%.1f", v), units[i]
}

// Size is the website's traffic for the table, 1.8 KB rather than the email's 2 KB
func (w Website) Size() string {
	n, u := sizeParts(w.Bytes)
	return n + " " + u
}

// SentNum and SentUnit split the bandwidth tile so the unit can be drawn small
func (s Summary) SentNum() string {
	n, _ := sizeParts(s.TotalBytes())
	return n
}

func (s Summary) SentUnit() string {
	_, u := sizeParts(s.TotalBytes())
	return u
}

// TotalBytes is all websites' traffic for the bandwidth tile
func (s Summary) TotalBytes() int64 {
	var n int64
	for _, w := range s.Websites {
		n += w.Bytes
	}
	return n
}

func (s Summary) TotalNotFound() int {
	n := 0
	for _, w := range s.Websites {
		n += w.NotFound
	}
	return n
}

func (s Summary) TotalRateLimited() int {
	n := 0
	for _, w := range s.Websites {
		n += w.RateLimited
	}
	return n
}

// TopPage is the site's most viewed page for the table's second line, empty without one
func (w Website) TopPage() string {
	if len(w.TopPages) == 0 {
		return ""
	}
	return w.TopPages[0]
}

// Change is the site's requests against the day before like "+40%", empty with nothing to compare
func (w Website) Change() string {
	return Tile{Now: int64(w.Requests), Prev: int64(w.PrevRequests), HasPrev: w.PrevRequests > 0}.Delta()
}

// Bar is the requests bar's width, a sliver for the quiet sites
func (w Website) Bar() int { return max(w.Share, 2) }
