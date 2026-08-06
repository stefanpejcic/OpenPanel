// Package dynamicdns implements a CRUD panel for token-authenticated
// dynamic DNS entries (stored as specially-tagged comments on ordinary
// A/AAAA records in the same BIND zone files the dns package manages)
// plus the public, unauthenticated webcall endpoint routers/IoT devices
// hit to push their current IP.
package dynamicdns

import (
	"crypto/rand"
	"math/big"
	"net"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/dns"
)

const tokenAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// generateToken returns a random alphanumeric token of the given length.
func generateToken(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(tokenAlphabet))))
		if err != nil {
			continue
		}
		b[i] = tokenAlphabet[n.Int64()]
	}
	return string(b)
}

var subdomainRE = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$`)

// validateSubdomain reports whether s is a valid DNS label.
func validateSubdomain(s string) bool {
	return subdomainRE.MatchString(s)
}

// validateIP reports whether ip parses as a valid IPv4 or IPv6 address.
func validateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// nowUTCStr returns the current UTC time in the format stored in each
// entry's "updated=" marker.
func nowUTCStr() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// DynDNSEntry is one dynamic DNS entry parsed from a zone file. Its JSON
// tags are read back out in the browser by dynamic_dns.html's Alpine
// component, which consumes a JSON-serialized flat array of these.
type DynDNSEntry struct {
	LineNumber  int    `json:"line_number"`
	Subdomain   string `json:"subdomain"`
	TTL         string `json:"ttl"`
	Type        string `json:"type"`
	Record      string `json:"record"`
	Token       string `json:"token"`
	LastUpdated string `json:"last_updated"`
	RawLine     string `json:"raw_line"`
	// Index is this entry's 0-based position in the flattened entry list
	// that dynamic_dns.html's openEdit()/openDelete() index into. Computed
	// when building the page data, since text/template has no ergonomic
	// running counter across nested range loops.
	Index int `json:"-"`
}

var webcallTokenRE = regexp.MustCompile(`;\s*webcall=(\S+)`)
var webcallUpdatedRE = regexp.MustCompile(`updated=(\S+)`)

func zoneFilePath(domain string) string {
	return dns.ZoneFileDir + domain + ".zone"
}

// parseDynamicDNSFromZone scans a domain's zone file for lines carrying a
// "; webcall=TOKEN updated=TS" marker comment - those are this feature's
// entries, distinguishable from ordinary records by that comment alone.
func parseDynamicDNSFromZone(domain string) []DynDNSEntry {
	content, err := os.ReadFile(zoneFilePath(domain))
	if err != nil {
		return nil
	}
	return parseDynamicDNSFromZoneContent(string(content))
}

// parseDynamicDNSFromZoneContent is parseDynamicDNSFromZone()'s pure
// parsing half, split out for testability without touching the filesystem.
func parseDynamicDNSFromZoneContent(content string) []DynDNSEntry {
	var entries []DynDNSEntry
	lines := strings.Split(content, "\n")
	for i, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "$") || strings.HasPrefix(line, ";") {
			continue
		}
		m := webcallTokenRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		token := m[1]
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		lastUpdated := ""
		if um := webcallUpdatedRE.FindStringSubmatch(line); um != nil {
			lastUpdated = um[1]
		}

		entries = append(entries, DynDNSEntry{
			LineNumber: i + 1, Subdomain: fields[0], TTL: fields[1],
			Type: fields[3], Record: fields[4], Token: token,
			LastUpdated: lastUpdated, RawLine: line,
		})
	}
	return entries
}

// buildZoneLine formats a dynamic DNS entry as a zone-file record line
// carrying its "; webcall=TOKEN updated=TS" marker comment.
func buildZoneLine(subdomain, recordType, ip, token, updated string) string {
	ts := updated
	if ts == "" {
		ts = nowUTCStr()
	}
	return subdomain + " 300 IN " + recordType + " " + ip + " ; webcall=" + token + " updated=" + ts
}

// writeZoneLines writes lines back to domain's zone file under an
// exclusive flock, guarding against concurrent writers - this route and
// the public update webcall both write the same file.
func writeZoneLines(domain string, lines []string) error {
	path := zoneFilePath(domain)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	_, err = f.WriteString(strings.Join(lines, "\n"))
	return err
}

func readZoneLines(domain string) ([]string, error) {
	content, err := os.ReadFile(zoneFilePath(domain))
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

// updateZoneLine replaces one zone-file line by number (1-indexed).
func updateZoneLine(domain string, lineNumber int, newLine string) bool {
	lines, err := readZoneLines(domain)
	if err != nil {
		return false
	}
	if lineNumber <= 0 || lineNumber > len(lines) {
		return false
	}
	lines[lineNumber-1] = newLine
	if err := writeZoneLines(domain, lines); err != nil {
		return false
	}
	dns.RestartDNSService(domain)
	return true
}

// deleteZoneLine removes one zone-file line by number (1-indexed).
func deleteZoneLine(domain string, lineNumber int) (deleted string, ok bool) {
	lines, err := readZoneLines(domain)
	if err != nil {
		return "", false
	}
	if lineNumber <= 0 || lineNumber > len(lines) {
		return "", false
	}
	deleted = lines[lineNumber-1]
	lines = append(lines[:lineNumber-1], lines[lineNumber:]...)
	if err := writeZoneLines(domain, lines); err != nil {
		return "", false
	}
	dns.RestartDNSService(domain)
	return strings.TrimSpace(deleted), true
}

// addDynamicDNSEntry appends a new dynamic DNS record to domain's zone
// file and returns its generated token.
func addDynamicDNSEntry(domain, subdomain, recordType, ip string) (token string, ok bool) {
	token = generateToken(16)
	newLine := buildZoneLine(subdomain, recordType, ip, token, "")

	content, err := os.ReadFile(zoneFilePath(domain))
	if err != nil {
		return "", false
	}
	updated := string(content)
	if !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	updated += newLine + "\n"

	if err := os.WriteFile(zoneFilePath(domain), []byte(updated), 0o644); err != nil {
		return "", false
	}
	dns.RestartDNSService(domain)
	return token, true
}
