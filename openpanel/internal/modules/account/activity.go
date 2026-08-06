package account

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
)

func activityLogPath(username string) string {
	return "/etc/openpanel/openpanel/core/users/" + username + "/activity.log"
}

// readActivityLog returns the user's activity log with the newest entry
// first (the log file itself is append-only chronological, so this just
// reverses it), or nil if the user has no log yet.
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
}

// parseActivityLine splits on a literal single space (not
// whitespace-collapsing - logger.go's entry format has an intentional
// double space between the timestamp and IP, which must land as an empty
// token to keep every later field at its expected index). Lines with
// fewer than 6 tokens are skipped.
//
// Note: parts[4] is always the literal word "User" from the log format
// itself (`... User <username> <action>`), so there's no per-role
// distinction to render - every row is treated the same way regardless of
// who performed the action.
func parseActivityLine(line string) (ActivityLogRow, bool) {
	parts := strings.Split(line, " ")
	if len(parts) < 6 {
		return ActivityLogRow{}, false
	}
	return ActivityLogRow{
		Timestamp: parts[0] + " " + parts[1] + " " + parts[2],
		IP:        parts[3],
		User:      parts[5],
		Action:    strings.Join(parts[6:], " "),
	}, true
}

// PageEntry is one rendered pagination control: either a page number link
// or an ellipsis.
type PageEntry struct {
	Number     int
	IsEllipsis bool
}

// buildPageEntries builds the pagination control list: current page
// (active), first/last page, and current+-2 render as links; page 2 and
// total_pages-1 render as an ellipsis when they don't already qualify
// above; every other page renders nothing.
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
}

// paginateActivityLog filters the log lines by searchTerm (if any) and
// slices out the requested page.
func paginateActivityLog(a *appctx.App, lines []string, searchTerm string, showAll bool, page int) ActivityPageResult {
	filtered := lines
	if searchTerm != "" {
		lower := strings.ToLower(searchTerm)
		filtered = nil
		for _, l := range lines {
			if strings.Contains(strings.ToLower(l), lower) {
				filtered = append(filtered, l)
			}
		}
		showAll = true
	}

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
	var pageLines []string
	if startIdx < endIdx {
		pageLines = filtered[startIdx:endIdx]
	}

	rows := make([]ActivityLogRow, 0, len(pageLines))
	for _, line := range pageLines {
		if row, ok := parseActivityLine(line); ok {
			rows = append(rows, row)
		}
	}

	return ActivityPageResult{
		Rows: rows, Page: page, ItemsPerPage: itemsPerPage, TotalPages: totalPages,
		TotalLines: totalLines, ShowAll: showAll, SearchTerm: searchTerm,
		PageEntries: buildPageEntries(page, totalPages),
	}
}

// handleViewActivityPage renders the account activity log page, applying
// any search/pagination query parameters.
func handleViewActivityPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	searchTerm := r.URL.Query().Get("search")
	showAll := r.URL.Query().Get("show_all") == "true"
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	logContent := readActivityLog(username)
	result := paginateActivityLog(a, logContent, searchTerm, showAll, page)

	renderActivityPage(a, w, r, result)
}

// RegisterActivity wires the /account/activity route onto mux, gated
// behind the "activity" feature flag.
func RegisterActivity(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "activity")(h)
	}
	mux.Handle("/account/activity", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleViewActivityPage(a, w, r) }))
}
