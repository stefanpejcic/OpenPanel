package errorlogs

import (
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// Entry is one normalized error line, whatever container or file it came from
type Entry struct {
	Time    time.Time `json:"time"`
	Source  string    `json:"source"`
	Service string    `json:"service"`
	Domain  string    `json:"domain,omitempty"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
	uri     string
}

// levelRank orders levels for the "warnings and up" / "errors only" filters and the severity column sort
var levelRank = map[string]int{"info": 0, "notice": 1, "warning": 2, "error": 3, "critical": 4}

var (
	fpmPrefixRE    = regexp.MustCompile(`^(NOTICE|WARNING|ERROR|ALERT): `)
	phpLevelRE     = regexp.MustCompile(`^PHP ([A-Za-z ]+?):\s+(.*)$`)
	nginxLineRE    = regexp.MustCompile(`^\d{4}/\d\d/\d\d \d\d:\d\d:\d\d \[(\w+)\] (?:\d+#\d+: )?(?:\*\d+ )?(.*)$`)
	apacheLineRE   = regexp.MustCompile(`^\[[^\]]+\] \[(?:[\w-]+:)?(\w+)\] (?:\[pid [^\]]+\] )?(?:\[client [^\]]+\] )?(.*)$`)
	mysqlLineRE    = regexp.MustCompile(`\[(ERROR|Warning|Note|System)\] (.*)$`)
	postgresLineRE = regexp.MustCompile(`\b(PANIC|FATAL|ERROR|WARNING|NOTICE|LOG|DETAIL|HINT|STATEMENT|CONTEXT):\s+(.*)$`)
	olsLineRE      = regexp.MustCompile(`^(\d{4}-\d\d-\d\d \d\d:\d\d:\d\d(?:\.\d+)?) \[(\w+)\] \[[^\]]*\] (?:\[[^\]#]*#([^\]]+)\] )?(.*)$`)
	ansiRE         = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]`)
	// podman splits a long write into records and prints each record's stamp, so it lands mid-message; RFC3339Nano keeps 7-9 digits almost always
	innerStampRE = regexp.MustCompile(`\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d\.\d{7,9}Z `)
	// php-fpm puts its own "[29-Oct-2026 11:35:06] " in front of NOTICE/WARNING lines when it isn't logging to a pipe
	fpmStampRE = regexp.MustCompile(`^\[\d\d-[A-Za-z]{3}-\d{4} \d\d:\d\d:\d\d(?:\.\d+)?\] `)
	// lines that carry on a PHP error: the stack trace, its "thrown in" line, chained exceptions and xdebug frames
	phpContRE = regexp.MustCompile(`^(Stack trace:|PHP Stack trace:|#\d+ |\s+\S|Next [\w\\]+|PHP\s+\d+\. )`)
)

// parseOLS reads OpenLiteSpeed's own error.log, which has its own timestamps and carries lsphp output as [STDERR] lines
func parseOLS(service, content string, from, to time.Time) []Entry {
	var entries []Entry
	keep := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimRight(line, "\r")
		m := olsLineRE.FindStringSubmatch(line)
		if m == nil {
			// stack traces come through as bare lines after their [STDERR] line
			if keep && strings.TrimSpace(line) != "" {
				entries[len(entries)-1].Message += "\n" + line
			}
			continue
		}
		keep = false
		ts, err := time.Parse("2006-01-02 15:04:05.999999", m[1])
		if err != nil || ts.Before(from) || (!to.IsZero() && ts.After(to)) {
			continue
		}
		e := Entry{Time: ts, Source: "webserver", Service: service, Domain: m[3], Level: webserverLevel(m[2]), Message: m[4]}
		if rest, ok := strings.CutPrefix(m[4], "[STDERR] "); ok {
			e.Source, e.Level, e.Message = "php", "info", rest
			if pm := phpLevelRE.FindStringSubmatch(rest); pm != nil {
				e.Level, e.Message = phpLevel(pm[1]), pm[1]+": "+pm[2]
			}
		} else if e.Level == "notice" {
			// OLS logs dozens of NOTICE lines on every start and reload, only warnings and up are worth showing
			continue
		}
		entries = append(entries, e)
		keep = true
	}
	return entries
}

// splitPodmanTS peels the RFC3339 stamp `podman logs --timestamps` puts in front of every line and cleans the rest
func splitPodmanTS(line string) (time.Time, string) {
	if i := strings.IndexByte(line, ' '); i > 0 {
		if ts, err := time.Parse(time.RFC3339Nano, line[:i]); err == nil {
			return ts, cleanText(line[i+1:])
		}
	}
	return time.Time{}, cleanText(line)
}

// cleanText drops terminal colour codes and the podman stamps glued into the middle of a split line
func cleanText(s string) string {
	if strings.IndexByte(s, 0x1b) >= 0 {
		s = ansiRE.ReplaceAllString(s, "")
	}
	if strings.Contains(s, "Z ") {
		s = innerStampRE.ReplaceAllString(s, "")
	}
	return s
}

func phpLevel(label string) string {
	l := strings.ToLower(label)
	switch {
	case strings.Contains(l, "fatal"), strings.Contains(l, "parse"):
		return "critical"
	case strings.Contains(l, "warning"):
		return "warning"
	case strings.Contains(l, "notice"), strings.Contains(l, "deprecated"), strings.Contains(l, "strict"):
		return "notice"
	}
	return "error"
}

func webserverLevel(l string) string {
	switch strings.ToLower(l) {
	case "emerg", "alert", "crit":
		return "critical"
	case "error":
		return "error"
	case "warn", "warning":
		return "warning"
	}
	return "notice"
}

// parseLines turns one container's `podman logs --timestamps` output into entries, kind picks the line format
func parseLines(kind, service, out string) []Entry {
	var entries []Entry
	appendTo := func(text string) {
		if n := len(entries); n > 0 {
			entries[n-1].Message += "\n" + text
		}
	}
	// open is true while the last php line was an entry or one of its continuation lines
	open := false
	for _, line := range strings.Split(out, "\n") {
		ts, text := splitPodmanTS(strings.TrimRight(line, "\r"))
		if strings.TrimSpace(text) == "" {
			continue
		}
		switch kind {
		case "php":
			text = fpmStampRE.ReplaceAllString(text, "")
			fpmLevel := ""
			if m := fpmPrefixRE.FindStringSubmatch(text); m != nil {
				fpmLevel, text = m[1], text[len(m[0]):]
			}
			if rest, ok := strings.CutPrefix(text, "PHP message: "); ok {
				e := Entry{Time: ts, Source: "php", Service: service, Level: "info", Message: rest}
				if m := phpLevelRE.FindStringSubmatch(rest); m != nil {
					e.Level, e.Message = phpLevel(m[1]), m[1]+": "+m[2]
				}
				entries = append(entries, e)
				open = true
				continue
			}
			switch fpmLevel {
			case "":
				// only stack-trace style lines right after an entry belong to it, the image's start banner doesn't
				if open && phpContRE.MatchString(text) {
					appendTo(text)
					continue
				}
			case "WARNING", "ERROR", "ALERT":
				entries = append(entries, Entry{Time: ts, Source: "php", Service: service, Level: webserverLevel(fpmLevel), Message: text})
				open = true
				continue
			}
			// fpm NOTICE lines are start/stop chatter, skip them
			open = false
		case "nginx":
			if m := nginxLineRE.FindStringSubmatch(text); m != nil {
				entries = append(entries, Entry{Time: ts, Source: "webserver", Service: service, Level: webserverLevel(m[1]), Message: m[2]})
			}
		case "apache":
			if m := apacheLineRE.FindStringSubmatch(text); m != nil {
				entries = append(entries, Entry{Time: ts, Source: "webserver", Service: service, Level: webserverLevel(m[1]), Message: m[2]})
			}
		case "mysql":
			if m := mysqlLineRE.FindStringSubmatch(text); m != nil {
				lvl := map[string]string{"ERROR": "error", "Warning": "warning"}[m[1]]
				if lvl == "" {
					// Note and System lines are startup chatter, a first start alone logs dozens
					continue
				}
				entries = append(entries, Entry{Time: ts, Source: "database", Service: service, Level: lvl, Message: m[2]})
			}
		case "postgres":
			if m := postgresLineRE.FindStringSubmatch(text); m != nil {
				switch m[1] {
				case "DETAIL", "HINT", "STATEMENT", "CONTEXT":
					appendTo(m[1] + ": " + m[2])
					continue
				}
				lvl := map[string]string{"PANIC": "critical", "FATAL": "critical", "ERROR": "error", "WARNING": "warning"}[m[1]]
				if lvl == "" {
					lvl = "notice"
				}
				entries = append(entries, Entry{Time: ts, Source: "database", Service: service, Level: lvl, Message: m[2]})
			}
		}
	}
	return entries
}

// caddyLine is the subset of a Caddy JSON access-log line needed to spot 5xx responses
type caddyLine struct {
	TS      float64 `json:"ts"`
	Status  int     `json:"status"`
	Request struct {
		ClientIP string `json:"client_ip"`
		Method   string `json:"method"`
		Host     string `json:"host"`
		URI      string `json:"uri"`
	} `json:"request"`
}

// parseCaddy5xx keeps only the 5xx responses from a chunk of a domain's Caddy access log
func parseCaddy5xx(domain, content string, from, to time.Time) []Entry {
	var entries []Entry
	for _, line := range strings.Split(content, "\n") {
		if !strings.Contains(line, `"status":5`) {
			continue
		}
		var l caddyLine
		if json.Unmarshal([]byte(line), &l) != nil || l.Status < 500 {
			continue
		}
		sec := int64(l.TS)
		ts := time.Unix(sec, int64((l.TS-float64(sec))*1e9)).UTC()
		if ts.Before(from) || (!to.IsZero() && ts.After(to)) {
			continue
		}
		entries = append(entries, Entry{
			Time: ts, Source: "http", Service: "caddy", Domain: domain, Level: "error",
			Message: "HTTP " + strconv.Itoa(l.Status) + " " + l.Request.Method + " " + l.Request.URI + " (client " + l.Request.ClientIP + ")",
			uri:     l.Request.URI,
		})
	}
	return entries
}

// domainMatcher works out which of the user's domains a log line is about
type domainMatcher struct {
	docroots []appctx.Domain
	names    []string
	// backends maps an app container name to its site, for web server proxy errors that only name the upstream
	backends map[string]string
}

// proxyBackendRE pulls the upstream name out of Apache's mod_proxy errors (AH01114, AH00957, AH00898)
var proxyBackendRE = regexp.MustCompile(`(?:connection to backend: |attempt to connect to \S+ \(|DNS lookup failure for: )([A-Za-z0-9][\w.-]*)`)

func newDomainMatcher(domains []appctx.Domain) domainMatcher {
	m := domainMatcher{}
	for _, d := range domains {
		if d.Docroot != "" {
			m.docroots = append(m.docroots, d)
		}
		m.names = append(m.names, d.DomainURL)
	}
	// longest first so a subdomain's docroot or name wins over its parent's
	sort.Slice(m.docroots, func(i, j int) bool { return len(m.docroots[i].Docroot) > len(m.docroots[j].Docroot) })
	sort.Slice(m.names, func(i, j int) bool { return len(m.names[i]) > len(m.names[j]) })
	return m
}

// containsBounded is strings.Contains but needle can't be glued to more hostname/path characters on either side
func containsBounded(s, needle string, isPart func(byte) bool) bool {
	for off := 0; ; {
		i := strings.Index(s[off:], needle)
		if i < 0 {
			return false
		}
		start, end := off+i, off+i+len(needle)
		if (start == 0 || !isPart(s[start-1])) && (end == len(s) || !isPart(s[end])) {
			return true
		}
		off = start + 1
	}
}

func isHostChar(c byte) bool {
	return c == '.' || c == '-' || c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// isPathChar only stops a docroot match from bleeding into a sibling dir, "/" after it is fine
func isPathChar(c byte) bool { return c != '/' && isHostChar(c) }

func (m domainMatcher) match(msg string) string {
	for _, d := range m.docroots {
		if containsBounded(msg, strings.TrimRight(d.Docroot, "/"), isPathChar) {
			return d.DomainURL
		}
	}
	for _, n := range m.names {
		if containsBounded(msg, n, isHostChar) {
			return n
		}
	}
	if len(m.backends) > 0 {
		if b := proxyBackendRE.FindStringSubmatch(msg); b != nil {
			return m.backends[strings.ToLower(b[1])]
		}
	}
	return ""
}
