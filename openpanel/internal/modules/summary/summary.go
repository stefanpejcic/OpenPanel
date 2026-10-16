// Package summary builds a daily report for an account: what needs fixing first, with a link to the page that fixes it, then the day's numbers. opencli generates and saves one per day, the Statistics > Summary page shows the saved ones.
package summary

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/errorlogs"
)

// Item is one finding, Link is a panel path and stays empty when the user can't fix it from the panel
type Item struct {
	Title  string `json:"title"`
	Detail string `json:"detail,omitempty"`
	Action string `json:"action,omitempty"`
	Link   string `json:"link,omitempty"`
}

// Website is one domain's or subfolder site's traffic for the day
type Website struct {
	Name         string   `json:"name"`
	Requests     int      `json:"requests"`
	UniqueIPs    int      `json:"unique_ips"`
	Bytes        int64    `json:"bytes"`
	NotFound     int      `json:"not_found"`
	ServerErrors int      `json:"server_errors"`
	RateLimited  int      `json:"rate_limited"`
	WAFBlocked   int      `json:"waf_blocked"`
	TopPages     []string `json:"top_pages,omitempty"`
	TopPageViews int      `json:"top_page_views,omitempty"`
	// Share is this site\'s requests against the busiest site, the bar list\'s width
	Share int `json:"share"`
	// requests the day before, -1 when there's no summary for that day to compare with
	PrevRequests int `json:"prev_requests"`

	ips          map[string]int
	limitedIPs   map[string]int
	wafIPs       map[string]int
	pages        map[string]int
	errorCount   int
	errorExample string
}

// Login is one panel login from the activity log, Country is the 2 letter code the login saved in .lastlogin
type Login struct {
	Time    time.Time `json:"time"`
	IP      string    `json:"ip"`
	Country string    `json:"country,omitempty"`
	New     bool      `json:"new"`
}

// UsageBar is disk or inode use, in the shape OpenAdmin's email template draws as a bar
type UsageBar struct {
	Title   string `json:"title"`
	Used    string `json:"used"`
	Total   string `json:"total"`
	Percent int    `json:"percent"`
	Limit   int    `json:"limit"`
}

// Flag is the panel's /static/flags image name, empty when the country is unknown
func (l Login) Flag() string {
	if len(l.Country) != 2 {
		return ""
	}
	return strings.ToLower(l.Country)
}

// Level colors a usage bar like the email does: full, over the alert limit, or fine
func (u UsageBar) Level() string {
	switch {
	case u.Percent >= 100:
		return "full"
	case u.Limit > 0 && u.Percent >= u.Limit:
		return "warning"
	}
	return "ok"
}

// Width keeps the bar inside its track
func (u UsageBar) Width() int { return min(max(u.Percent, 0), 100) }

// Summary is one account's day, saved as JSON and turned into the page and the email
type Summary struct {
	Username    string    `json:"username"`
	Date        string    `json:"date"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
	GeneratedAt time.Time `json:"generated_at"`

	// Headline is the one-sentence verdict the page and the email open with
	Headline string `json:"headline"`
	// PrevRequests is all websites\' requests the day before, -1 when that day has no summary
	PrevRequests int        `json:"prev_requests"`
	Attention    []Item     `json:"attention"`
	Look         []Item     `json:"look"`
	Good         []string   `json:"good"`
	Websites     []Website  `json:"websites"`
	Logins       []Login    `json:"logins"`
	Changes      []string   `json:"changes"`
	Usage        []UsageBar `json:"usage"`
	Facts        []string   `json:"facts"`
}

// builder carries what every check needs, Allowed is the user's sidebar features
type builder struct {
	a           *appctx.App
	username    string
	userID      int
	userContext string
	allowed     map[string]bool
	domains     []appctx.Domain
	sites       []errorlogs.Site
	from, to    time.Time
	s           *Summary
	buckets     map[string]*Website
	prevTotal   int
}

func (b *builder) can(feature string) bool { return b.allowed[feature] }

func (b *builder) attention(it Item) { b.s.Attention = append(b.s.Attention, it) }
func (b *builder) look(it Item)      { b.s.Look = append(b.s.Look, it) }
func (b *builder) good(line string)  { b.s.Good = append(b.s.Good, line) }

// website returns the stats bucket for a domain or subfolder site, creating it on first use
func (b *builder) website(name string) *Website {
	if w, ok := b.buckets[name]; ok {
		return w
	}
	w := &Website{Name: name, ips: map[string]int{}, limitedIPs: map[string]int{}, wafIPs: map[string]int{}, pages: map[string]int{}}
	b.buckets[name] = w
	return w
}

// Allowed reports whether the account can use the summary module, same rule as the sidebar
func Allowed(ctx context.Context, a *appctx.App, username, userContext string) (map[string]bool, bool) {
	features, _ := a.LoadUserFeatures(ctx, username, userContext)
	allowed := map[string]bool{}
	for _, f := range features {
		if a.ModuleEnabled(f) {
			allowed[f] = true
		}
	}
	return allowed, a.ModuleEnabled("summary") && slices.Contains(features, "summary")
}

// Build collects the account's summary for the local day that starts at day
func Build(ctx context.Context, a *appctx.App, username string, day time.Time) (Summary, bool, error) {
	var userID int
	if err := a.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE username = ?", username).Scan(&userID); err != nil {
		return Summary{}, false, fmt.Errorf("user %s not found", username)
	}
	userContext, err := a.QueryContextByUsername(ctx, username)
	if err != nil || userContext == "" {
		return Summary{}, false, fmt.Errorf("no context for user %s", username)
	}
	allowed, ok := Allowed(ctx, a, username, userContext)
	if !ok {
		return Summary{}, false, nil
	}

	from := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	s := &Summary{
		Username: username, Date: from.Format("2006-01-02"), From: from, To: from.AddDate(0, 0, 1), GeneratedAt: time.Now(),
		Attention: []Item{}, Look: []Item{}, Good: []string{}, Websites: []Website{}, Logins: []Login{}, Changes: []string{}, Facts: []string{}, Usage: []UsageBar{},
	}
	b := &builder{a: a, username: username, userID: userID, userContext: userContext, allowed: allowed, from: s.From, to: s.To, s: s, buckets: map[string]*Website{}}
	b.domains, _ = a.AllDomainsForUser(ctx, userID)
	b.sites = errorlogs.LoadSites(ctx, a, userID)

	// order here is the order things show up in each section
	b.checkVulnerabilities()
	b.checkMalware()
	b.collectTraffic()
	b.collectWAF()
	b.checkErrors(ctx)
	b.checkSSL()
	b.checkCrons(ctx)
	b.checkServices(ctx)
	b.checkDisk(ctx)
	b.checkResources()
	b.checkBackups()
	b.checkMailboxes(ctx)
	b.collectActivity()
	b.finishWebsites()
	b.compareWithDayBefore()
	s.PrevRequests = b.prevTotal
	b.busiestPage()
	s.Headline = b.headline()
	return *s, true, nil
}

// finishWebsites turns the traffic buckets into the sorted table, busiest first
func (b *builder) finishWebsites() {
	for _, w := range b.buckets {
		w.UniqueIPs = len(w.ips)
		w.TopPages = topKeys(w.pages, 3)
		if len(w.TopPages) > 0 {
			w.TopPageViews = w.pages[w.TopPages[0]]
		}
		if w.Requests > 0 || w.WAFBlocked > 0 {
			b.s.Websites = append(b.s.Websites, *w)
		}
	}
	sort.Slice(b.s.Websites, func(i, j int) bool {
		if b.s.Websites[i].Requests != b.s.Websites[j].Requests {
			return b.s.Websites[i].Requests > b.s.Websites[j].Requests
		}
		return b.s.Websites[i].Name < b.s.Websites[j].Name
	})
	if len(b.s.Websites) > 0 && b.s.Websites[0].Requests > 0 {
		for i := range b.s.Websites {
			b.s.Websites[i].Share = max(b.s.Websites[i].Requests*100/b.s.Websites[0].Requests, 1)
		}
	}
}

// compareWithDayBefore fills in each website's requests from the previous saved summary
func (b *builder) compareWithDayBefore() {
	prev, ok := Load(b.username, b.from.AddDate(0, 0, -1).Format("2006-01-02"))
	before := map[string]int{}
	for _, w := range prev.Websites {
		before[w.Name] = w.Requests
	}
	for i := range b.s.Websites {
		b.s.Websites[i].PrevRequests = -1
		if ok {
			b.s.Websites[i].PrevRequests = before[b.s.Websites[i].Name]
		}
	}
	b.prevTotal = -1
	if ok {
		b.prevTotal = 0
		for _, w := range prev.Websites {
			b.prevTotal += w.Requests
		}
	}
}

// busiestPage is good news worth a line, the page people came for most
func (b *builder) busiestPage() {
	var top *Website
	for i := range b.s.Websites {
		w := &b.s.Websites[i]
		if w.TopPageViews > 0 && (top == nil || w.TopPageViews > top.TopPageViews) {
			top = w
		}
	}
	if top != nil && top.TopPageViews >= 10 {
		page := top.TopPages[0]
		if page == "/" {
			b.good(fmt.Sprintf("Your busiest page was the homepage of %s, with %s.", top.Name, plural(top.TopPageViews, "view", "views")))
		} else {
			b.good(fmt.Sprintf("Your busiest page was %s on %s, with %s.", page, top.Name, plural(top.TopPageViews, "view", "views")))
		}
	}
}

// change describes now against before in words, empty when there's nothing to compare with
func change(now, before int) string {
	if before <= 0 {
		return ""
	}
	pct := (now - before) * 100 / before
	switch {
	case pct >= 5:
		return fmt.Sprintf("%d%% more than the day before", pct)
	case pct <= -5:
		return fmt.Sprintf("%d%% fewer than the day before", -pct)
	}
	return "about the same as the day before"
}

// headline sums the day up in one sentence: traffic first, then whether anything needs the user
func (b *builder) headline() string {
	requests, visitors := 0, 0
	for _, w := range b.s.Websites {
		requests += w.Requests
		visitors += w.UniqueIPs
	}
	var out string
	if requests == 0 {
		out = "Your websites had no visitors."
	} else {
		out = fmt.Sprintf("Your %s handled %s requests from %s", plural(len(b.s.Websites), "website", "websites"), num(requests), plural(visitors, "visitor", "visitors"))
		if c := change(requests, b.prevTotal); c != "" {
			out += ", " + c
		}
		out += "."
	}
	switch {
	case len(b.s.Attention) > 0:
		out += " " + plural(len(b.s.Attention), "thing needs", "things need") + " your attention."
	case len(b.s.Look) > 0:
		out += " Nothing urgent, but " + plural(len(b.s.Look), "thing is", "things are") + " worth a look."
	default:
		out += " Nothing needs your attention."
	}
	return out
}

// topKeys returns the n keys with the highest counts, ties broken by name so the output is stable
func topKeys(m map[string]int, n int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if m[keys[i]] != m[keys[j]] {
			return m[keys[i]] > m[keys[j]]
		}
		return keys[i] < keys[j]
	})
	if len(keys) > n {
		keys = keys[:n]
	}
	return keys
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return num(n) + " " + many
}

// Sent is the website's traffic for templates, which have no byte formatter
func (w Website) Sent() string { return humanBytes(w.Bytes) }

// num writes 4812 as 4,812
func num(n int) string {
	s := fmt.Sprint(n)
	if n < 0 {
		return "-" + num(-n)
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

// joinAnd writes a list the way people do: "A", "A and B", "A, B and C"
func joinAnd(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

// vulnerableNames turns "Contact Form 7 [contact-form-7] < 5.3.2" into "Contact Form 7", each plugin or theme once
func vulnerableNames(names []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, n := range names {
		if i := strings.Index(n, " ["); i > 0 {
			n = n[:i]
		} else if i := strings.Index(n, " <"); i > 0 {
			n = n[:i]
		}
		n = strings.TrimSpace(n)
		if n != "" && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

// totals for the KPI strip at the top of the page

func (s Summary) TotalRequests() int {
	n := 0
	for _, w := range s.Websites {
		n += w.Requests
	}
	return n
}

func (s Summary) TotalVisitors() int {
	n := 0
	for _, w := range s.Websites {
		n += w.UniqueIPs
	}
	return n
}

func (s Summary) TotalSent() string {
	var n int64
	for _, w := range s.Websites {
		n += w.Bytes
	}
	return humanBytes(n)
}

func (s Summary) TotalErrors() int {
	n := 0
	for _, w := range s.Websites {
		n += w.ServerErrors
	}
	return n
}

func (s Summary) TotalBlocked() int {
	n := 0
	for _, w := range s.Websites {
		n += w.WAFBlocked
	}
	return n
}

// Delta is the change in requests against the day before like "+12%", empty with nothing to compare
func (s Summary) Delta() string {
	if s.PrevRequests <= 0 {
		return ""
	}
	pct := (s.TotalRequests() - s.PrevRequests) * 100 / s.PrevRequests
	if pct > 0 {
		return fmt.Sprintf("+%d%%", pct)
	}
	return fmt.Sprintf("%d%%", pct)
}

func (s Summary) DeltaUp() bool { return s.TotalRequests() >= s.PrevRequests }

// Num is num for templates
func (s Summary) Num(n int) string { return num(n) }
