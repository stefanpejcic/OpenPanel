// Package dns implements BIND zone file editing (table and raw code
// views), record add/update/delete, serial-number bumping plus rndc
// reload, zone export, and zone reset.
package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// HealthIssue is the {id, severity, message} shape consumed client-side
// by reportHealthIssues() to render dismissible health toasts.
type HealthIssue struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeText(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(text))
}

// readLinesKeepEnds splits content into lines, each keeping its trailing
// "\n" (except possibly the last), and a trailing "\n" in the source
// doesn't produce a spurious empty trailing element the way
// strings.SplitAfter alone would.
func readLinesKeepEnds(content string) []string {
	if content == "" {
		return nil
	}
	lines := strings.SplitAfter(content, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// ZoneFileDir is where BIND zone files live on disk.
const ZoneFileDir = "/etc/bind/zones/"

func zoneFilePath(domain string) string {
	return ZoneFileDir + domain + ".zone"
}

func injected(a *appctx.App, r *http.Request) (username, userContext string, err error) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return username, userContext, nil
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

func flashSess(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

// validateZoneFile runs named-checkzone inside the shared openpanel_dns
// container (bind-mounted at the same /etc/bind path, so the on-disk file
// can be checked directly) and returns its error output, or "" if valid.
func validateZoneFile(domain, zoneFilePath string) string {
	cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cctx, "podman", "exec", "openpanel_dns", "named-checkzone", domain, zoneFilePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

// serialNumberRE matches an 8-digit date plus a 2-digit daily counter,
// immediately followed by the "; Serial number" marker comment.
var serialNumberRE = regexp.MustCompile(`(\d{8})(\d{2})\s*;\s*Serial number`)

// RestartDNSService bumps the zone's serial number (same date ->
// increment the daily counter, rolling over past 99; new date -> reset to
// 01) and reloads the zone via rndc.
func RestartDNSService(domainURL string) {
	path := zoneFilePath(domainURL)
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	currentDate := time.Now().UTC().Format("20060102")
	lines := readLinesKeepEnds(string(content))
	updated := false
	for i, line := range lines {
		m := serialNumberRE.FindStringSubmatchIndex(line)
		if m == nil {
			continue
		}
		fileDate := line[m[2]:m[3]]
		fileCounter, _ := strconv.Atoi(line[m[4]:m[5]])

		newCounter := 1
		if fileDate == currentDate {
			newCounter = fileCounter + 1
			if newCounter > 99 {
				newCounter = 1
			}
		}
		newSerial := currentDate + fmt.Sprintf("%02d", newCounter)
		lines[i] = serialNumberRE.ReplaceAllString(line, newSerial+" ; Serial number")
		updated = true
		break
	}

	if updated {
		_ = os.WriteFile(path, []byte(strings.Join(lines, "")), 0o644)
	}

	cctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = exec.CommandContext(cctx, "podman", "exec", "openpanel_dns", "rndc", "reload", domainURL).Run()
}

// cnameNameRE builds the per-call regex used to detect an existing CNAME
// with the given name: `^<name>\s+\d+\s+IN\s+CNAME`.
func cnameNameRE(name string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(name) + `\s+\d+\s+IN\s+CNAME`)
}

// cnameRecordExists checks whether a CNAME record with the given name
// already exists in the zone file, memoized for 10s since it's called
// repeatedly during record validation.
func cnameRecordExists(ctx context.Context, a *appctx.App, zoneFilePath, name string) bool {
	exists, _ := cache.Memoize(ctx, a.Cache, "cname_record_exists:"+zoneFilePath+":"+name, 10*time.Second, func() (bool, error) {
		content, err := os.ReadFile(zoneFilePath)
		if err != nil {
			return false, nil
		}
		re := cnameNameRE(name)
		for _, line := range strings.Split(string(content), "\n") {
			if re.MatchString(line) {
				return true, nil
			}
		}
		return false, nil
	})
	return exists
}

// splitMaxN splits on runs of whitespace, stopping after n splits so the
// final field keeps any remaining whitespace-separated content intact (up
// to n+1 elements).
func splitMaxN(s string, n int) []string {
	var fields []string
	rest := s
	for len(fields) < n {
		rest = strings.TrimLeft(rest, " \t\r\n\f\v")
		if rest == "" {
			return fields
		}
		idx := strings.IndexAny(rest, " \t\r\n\f\v")
		if idx == -1 {
			fields = append(fields, rest)
			return fields
		}
		fields = append(fields, rest[:idx])
		rest = rest[idx:]
	}
	rest = strings.TrimLeft(rest, " \t\r\n\f\v")
	if rest != "" {
		fields = append(fields, rest)
	}
	return fields
}

// extractComment returns the text after the last ';' that isn't inside a
// quoted string. Since RE2 has no lookaround, this walks the line
// tracking quote state instead of using a lookaround-based regex.
func extractComment(line string) string {
	lastSemicolon := -1
	inQuotes := false
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '"':
			inQuotes = !inQuotes
		case ';':
			if !inQuotes {
				lastSemicolon = i
			}
		}
	}
	if lastSemicolon == -1 {
		return ""
	}
	return strings.TrimSpace(line[lastSemicolon+1:])
}

// readSerialNumber finds the first line containing "; Serial number" and
// returns its leading token (the serial number itself).
func readSerialNumber(lines []string) string {
	for _, line := range lines {
		if strings.Contains(line, "; Serial number") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0]
			}
			return ""
		}
	}
	return ""
}
