package account

import (
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

func activityLogPath(username string) string {
	return "/etc/openpanel/openpanel/core/users/" + username + "/activity.log"
}

// readActivityLog returns the user's activity log newest-first (the file itself is append-only chronological, so this just reverses it), or nil if there's no log yet
func readActivityLog(username string) []string {
	content, err := os.ReadFile(activityLogPath(username))
	if err != nil {
		return nil
	}
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return lines
}

// ActivityLogRow is one parsed row of the activity table.
type ActivityLogRow struct {
	Timestamp string
	IP        string
	User      string
	Action    string
	Kind      string
}

// Activity kinds, used to pick the icon and color for each row
const (
	ActivityKindSecurity = "security"
	ActivityKindCommand  = "command"
	ActivityKindDanger   = "danger"
	ActivityKindCreate   = "create"
	ActivityKindEdit     = "edit"
	ActivityKindInfo     = "info"
)

var (
	securityKeywords = []string{"logged in", "logged out", "password", "2fa", "passkey", "mcp token", "session", "ip blocker", "blocked", "waf", "malware", "quarantine", "interactive terminal"}
	commandPrefixes  = []string{"executed command", "executed a command", "ran ", "manually ran", "terminated process"}
	dangerPrefixes   = []string{"deleted", "removed", "uninstalled", "permanently deleted", "emptied", "revoked", "reset", "detached", "suspended", "terminated"}
	createPrefixes   = []string{"created", "added", "installed", "generated", "registered", "uploaded", "cloned", "imported", "granted", "assigned", "used wizard", "started a backup", "set ", "executed "}
	editPrefixes     = []string{"edited", "changed", "updated", "switched", "renamed", "moved", "copied", "enabled", "disabled", "capitalized", "configured", "saved", "restored", "marked", "unsuspended", "rebuilt", "cleared", "flushed", "fixed", "extracted", "pulled", "activated", "deactivated", "restarted", "started", "stopped", "optimized", "repaired"}
)

// classifyActivity guesses the kind of an action from its wording, order matters since e.g. "changed password" is security, not a plain edit
func classifyActivity(action string) string {
	lower := strings.ToLower(action)
	for _, k := range securityKeywords {
		if strings.Contains(lower, k) {
			return ActivityKindSecurity
		}
	}
	if strings.HasPrefix(lower, "moved ") && strings.Contains(lower, " to trash") {
		return ActivityKindDanger
	}
	for _, groups := range []struct {
		kind     string
		prefixes []string
	}{
		{ActivityKindCommand, commandPrefixes},
		{ActivityKindDanger, dangerPrefixes},
		{ActivityKindCreate, createPrefixes},
		{ActivityKindEdit, editPrefixes},
	} {
		for _, p := range groups.prefixes {
			if strings.HasPrefix(lower, p) {
				return groups.kind
			}
		}
	}
	return ActivityKindInfo
}

// parseActivityLine splits on a literal single space, not whitespace-collapsing, since logger.go's format has an intentional double space between timestamp and IP that needs to land as an empty token to keep later fields at the right index - skips lines with fewer than 6 tokens; parts[4] is always the literal word "User", nothing to branch on there
func parseActivityLine(line string) (ActivityLogRow, bool) {
	parts := strings.Split(line, " ")
	if len(parts) < 6 {
		return ActivityLogRow{}, false
	}
	action := strings.Join(parts[6:], " ")
	return ActivityLogRow{
		Timestamp: parts[0] + " " + parts[1] + " " + parts[2],
		IP:        parts[3],
		User:      parts[5],
		Action:    action,
		Kind:      classifyActivity(action),
	}, true
}

// PageEntry is one rendered pagination control: either a page number link or an ellipsis
type PageEntry struct {
	Number     int
	IsEllipsis bool
}

// buildPageEntries builds the pagination list: current page, first/last, and current+-2 render as links; page 2 and total_pages-1 render as an ellipsis if they don't already qualify; everything else is skipped
func buildPageEntries(current, total int) []PageEntry {
	var entries []PageEntry
	for p := 1; p <= total; p++ {
		switch {
		case p == current:
			entries = append(entries, PageEntry{Number: p})
		case p == 1 || p == total || (p >= current-2 && p <= current+2):
			entries = append(entries, PageEntry{Number: p})
		case p == 2 || p == total-1:
			entries = append(entries, PageEntry{IsEllipsis: true})
		}
	}
	return entries
}

// ActivityFilter is everything the activity page can be narrowed by, all of it lives in the URL so filtered views can be shared
type ActivityFilter struct {
	Search  string
	Kind    string
	From    string
	To      string
	ShowAll bool
}

var activityKindOrder = []string{ActivityKindDanger, ActivityKindCreate, ActivityKindEdit, ActivityKindSecurity, ActivityKindCommand, ActivityKindInfo}

var activityKindLabels = map[string]string{
	ActivityKindDanger:   "Destructive",
	ActivityKindCreate:   "Created",
	ActivityKindEdit:     "Changed",
	ActivityKindSecurity: "Security",
	ActivityKindCommand:  "Command",
	ActivityKindInfo:     "Info",
}

// parseActivityFilter reads the filter from query params, dropping anything invalid so a bad link just shows everything
func parseActivityFilter(q url.Values) ActivityFilter {
	f := ActivityFilter{Search: q.Get("search"), ShowAll: q.Get("show_all") == "true"}
	if _, ok := activityKindLabels[q.Get("type")]; ok {
		f.Kind = q.Get("type")
	}
	if _, err := time.Parse("2006-01-02", q.Get("from")); err == nil {
		f.From = q.Get("from")
	}
	if _, err := time.Parse("2006-01-02", q.Get("to")); err == nil {
		f.To = q.Get("to")
	}
	if f.From != "" && f.To != "" && f.From > f.To {
		f.From, f.To = f.To, f.From
	}
	return f
}

// URL builds the activity page link for this filter, page <= 1 is left out to keep links short
func (f ActivityFilter) URL(page int) template.URL {
	q := url.Values{}
	for k, v := range map[string]string{"search": f.Search, "type": f.Kind, "from": f.From, "to": f.To} {
		if v != "" {
			q.Set(k, v)
		}
	}
	if f.ShowAll {
		q.Set("show_all", "true")
	}
	if page > 1 {
		q.Set("page", strconv.Itoa(page))
	}
	if len(q) == 0 {
		return "/account/activity"
	}
	return template.URL("/account/activity?" + q.Encode())
}

func (f ActivityFilter) inDateRange(row ActivityLogRow) bool {
	if len(row.Timestamp) < 10 {
		return f.From == "" && f.To == ""
	}
	day := row.Timestamp[:10]
	return (f.From == "" || day >= f.From) && (f.To == "" || day <= f.To)
}

// ActivityKindTab is one of the type filter buttons above the table
type ActivityKindTab struct {
	Kind   string
	Label  string
	Count  int
	Active bool
	URL    template.URL
}

// ActivityPageResult is the paginated/filtered view of the activity log.
type ActivityPageResult struct {
	Rows         []ActivityLogRow
	Page         int
	ItemsPerPage int
	TotalPages   int
	TotalLines   int
	ShowAll      bool
	SearchTerm   string
	PageEntries  []PageEntry
	Filter       ActivityFilter
	KindTabs     []ActivityKindTab
	AllCount     int
}

// PageURL keeps the current filters when moving between pages
func (r ActivityPageResult) PageURL(page int) template.URL {
	return r.Filter.URL(page)
}

// AllURL is the "All" type button, same filters minus the type
func (r ActivityPageResult) AllURL() template.URL {
	f := r.Filter
	f.Kind = ""
	return f.URL(1)
}

// paginateActivityLog applies the filter and slices out the requested page, type counts ignore the type filter so the buttons show what you'd get
func paginateActivityLog(a *appctx.App, lines []string, f ActivityFilter, page int) ActivityPageResult {
	lowerSearch := strings.ToLower(f.Search)
	counts := map[string]int{}
	allCount := 0
	var filtered []ActivityLogRow
	for _, l := range lines {
		if lowerSearch != "" && !strings.Contains(strings.ToLower(l), lowerSearch) {
			continue
		}
		row, ok := parseActivityLine(l)
		if !ok || !f.inDateRange(row) {
			continue
		}
		counts[row.Kind]++
		allCount++
		if f.Kind == "" || row.Kind == f.Kind {
			filtered = append(filtered, row)
		}
	}

	showAll := f.ShowAll || f.Search != ""
	totalLines := len(filtered)

	var itemsPerPage, totalPages int
	if showAll {
		itemsPerPage = totalLines
		totalPages = 1
	} else {
		itemsPerPage, _ = strconv.Atoi(a.Config.Get("activity_items_per_page", "100"))
		if itemsPerPage <= 0 {
			itemsPerPage = 100
		}
		totalPages = totalLines / itemsPerPage
		if totalLines%itemsPerPage != 0 {
			totalPages++
		}
	}
	if totalPages < 1 {
		totalPages = 1
	}
	if page < 1 {
		page = 1
	}

	startIdx := (page - 1) * itemsPerPage
	endIdx := startIdx + itemsPerPage
	if startIdx > totalLines {
		startIdx = totalLines
	}
	if endIdx > totalLines {
		endIdx = totalLines
	}
	rows := []ActivityLogRow{}
	if startIdx < endIdx {
		rows = filtered[startIdx:endIdx]
	}

	tabs := make([]ActivityKindTab, 0, len(activityKindOrder))
	for _, k := range activityKindOrder {
		tf := f
		tf.Kind = k
		tabs = append(tabs, ActivityKindTab{Kind: k, Label: activityKindLabels[k], Count: counts[k], Active: f.Kind == k, URL: tf.URL(1)})
	}

	return ActivityPageResult{
		Rows: rows, Page: page, ItemsPerPage: itemsPerPage, TotalPages: totalPages,
		TotalLines: totalLines, ShowAll: showAll, SearchTerm: f.Search,
		PageEntries: buildPageEntries(page, totalPages),
		Filter:      f, KindTabs: tabs, AllCount: allCount,
	}
}

// handleViewActivityPage renders the account activity log page, applying any search/pagination query parameters
func handleViewActivityPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	logContent := readActivityLog(username)
	result := paginateActivityLog(a, logContent, parseActivityFilter(r.URL.Query()), page)

	renderActivityPage(a, w, r, result)
}

// RegisterActivity wires the /account/activity route onto mux, gated behind the "activity" feature flag
func RegisterActivity(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "activity")(h)
	}
	mux.Handle("/account/activity", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleViewActivityPage(a, w, r) }))
}
