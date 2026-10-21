package summary

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

const keepDays = 90

var dateRE = regexp.MustCompile(`^\d{4}-\d\d-\d\d$`)

func summaryDir(username string) string { return usersDir + username + "/summaries/" }

// Save writes the day's summary and drops the ones older than keepDays
func Save(s Summary) error {
	dir := summaryDir(s.Username)
	if err := tenantfs.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := tenantfs.WriteFile(dir+s.Date+".json", raw, 0o644); err != nil {
		return err
	}
	cutoff := time.Now().AddDate(0, 0, -keepDays).Format("2006-01-02")
	for _, d := range Dates(s.Username) {
		if d < cutoff {
			_ = tenantfs.Remove(dir + d + ".json")
		}
	}
	return nil
}

// Dates lists the saved days, newest first
func Dates(username string) []string {
	matches, _ := filepath.Glob(summaryDir(username) + "*.json")
	var out []string
	for _, m := range matches {
		d := filepath.Base(m)
		d = d[:len(d)-len(".json")]
		if dateRE.MatchString(d) {
			out = append(out, d)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(out)))
	return out
}

// Load reads one saved day, date must look like 2006-01-02 so it can't leave the folder
func Load(username, date string) (Summary, bool) {
	var s Summary
	if !dateRE.MatchString(date) {
		return s, false
	}
	raw, err := tenantfs.ReadFile(summaryDir(username) + date + ".json")
	if err != nil || json.Unmarshal(raw, &s) != nil {
		return s, false
	}
	return s, true
}
