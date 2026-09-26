package dns

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// BulkKey identifies a record by line number plus a hash of its content, so a zone edited since the page loaded can't hit the wrong line
func (z ZoneRow) BulkKey() string {
	sum := sha1.Sum([]byte(z.RawLine))
	return strconv.Itoa(z.LineNumber) + ":" + hex.EncodeToString(sum[:4])
}

func dnsBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "ttl", Label: t.Get("Update TTL"), Confirm: t.Get("New TTL in seconds for the selected records:"),
			Input: &web.BulkInput{Type: "number", Min: "60", Step: "1", Default: "14400", Placeholder: "14400"}},
		{Key: "delete", Label: t.Get("Delete"), Confirm: t.Get("Delete the selected DNS records?"), Danger: true},
	}
}

// ttlFieldRE is the name and TTL at the start of a record's first line
var ttlFieldRE = regexp.MustCompile(`^(\S+\s+)(\d+)(\s)`)

// handleDNSBulk edits the zone file once for the whole selection: deletes run bottom-up so earlier line numbers stay valid, then the serial is bumped and the zone reloaded once
func handleDNSBulk(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	username, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domain := r.PathValue("domain")
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	req, ok := web.DecodeBulkRequest(w, r)
	if !ok {
		return
	}
	act, found := web.FindBulkAction(dnsBulkActions(web.RequestTranslator(a, r)), req.Action)
	if !found {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Unknown bulk action"})
		return
	}
	ttl := strings.TrimSpace(req.Value)
	if req.Action == "ttl" {
		if n, convErr := strconv.Atoi(ttl); convErr != nil || n < 60 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "TTL must be a whole number of at least 60 seconds."})
			return
		}
	}

	path := zoneFilePath(domain)
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		writeJSON(w, http.StatusOK, map[string]string{"error": "Zone file not found."})
		return
	}
	lines := readLinesKeepEnds(string(content))
	rows := map[string]ZoneRow{}
	for _, row := range buildZoneRows(parseZoneWithLineNumbers(string(content))) {
		rows[row.BulkKey()] = row
	}

	results := make([]web.BulkResult, len(req.Items))
	var targets []int
	for i, key := range req.Items {
		row, exists := rows[key]
		label := key
		if exists {
			label = row.Name + " " + row.Type
		}
		results[i].Item = label
		switch {
		case !exists:
			results[i].Message = "Record changed since the page loaded, reload and try again."
		case row.EndLineNumber > len(lines):
			results[i].Message = "Record not found."
		case req.Action == "ttl" && !ttlFieldRE.MatchString(lines[row.LineNumber-1]):
			results[i].Message = "This record has no TTL field to change."
		default:
			targets = append(targets, i)
		}
	}

	// bottom-up so removing one record doesn't shift the lines of the next
	sort.Slice(targets, func(x, y int) bool { return rows[req.Items[targets[x]]].LineNumber > rows[req.Items[targets[y]]].LineNumber })
	for _, i := range targets {
		row := rows[req.Items[i]]
		if req.Action == "delete" {
			lines = append(lines[:row.LineNumber-1], lines[row.EndLineNumber:]...)
		} else {
			lines[row.LineNumber-1] = ttlFieldRE.ReplaceAllString(lines[row.LineNumber-1], "${1}"+ttl+"${3}")
		}
	}

	if len(targets) > 0 {
		if writeErr := os.WriteFile(path, []byte(strings.Join(lines, "")), 0o644); writeErr != nil {
			for _, i := range targets {
				results[i].Message = "Error saving the zone file."
			}
			targets = nil
		} else {
			RestartDNSService(domain)
		}
	}
	for _, i := range targets {
		results[i].OK = true
		results[i].Message = "Done."
	}
	if len(targets) > 0 {
		_ = logger.RecordUserAction(a.Config, username, "bulk "+req.Action+" on "+strconv.Itoa(len(targets))+" DNS record(s) for domain "+domain, reqip.ClientIP(r))
	}
	web.FinishBulk(a, w, r, act.Label, results)
}
