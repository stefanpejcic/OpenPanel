package waf

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
)

// LogRule is one rule that matched a request, in plain words for the Basic log view
type LogRule struct {
	ID       string `json:"id"`
	Msg      string `json:"msg"`
	Category string `json:"category"`
	Data     string `json:"data,omitempty"`
}

// LogEvent is one request the firewall flagged
type LogEvent struct {
	Time   string    `json:"time"`
	IP     string    `json:"ip"`
	Method string    `json:"method"`
	URI    string    `json:"uri"`
	Status int       `json:"status"`
	Result string    `json:"result"` // blocked, would_block (monitor only) or flagged
	Rules  []LogRule `json:"rules"`
}

// LogRuleGroup is every event one rule caused, for the "top reasons" list
type LogRuleGroup struct {
	LogRule
	Count    int      `json:"count"`
	Blocked  int      `json:"blocked"`
	Paths    []string `json:"paths"`
	LastSeen string   `json:"last_seen"`
}

// LogSummary is the Basic view of the WAF log page
type LogSummary struct {
	Events     []LogEvent     `json:"events"`
	Groups     []LogRuleGroup `json:"groups"`
	Blocked    int            `json:"blocked"`
	WouldBlock int            `json:"would_block"`
	Flagged    int            `json:"flagged"`
	// logged requests with no rule match, like plain 404s, only shown in Advanced
	Other int `json:"other"`
}

var (
	logRuleIDRE  = regexp.MustCompile(`\[id "(\d+)"\]`)
	logRuleMsgRE = regexp.MustCompile(`\[msg "([^"]*)"\]`)
	logDataRE    = regexp.MustCompile(`\[data "([^"]*)"\]`)
	logTagRE     = regexp.MustCompile(`\[tag "([^"]+)"\]`)
)

// attackCategories maps CRS attack-* tags to words a site owner understands
var attackCategories = map[string]string{
	"attack-sqli":                 "SQL injection",
	"attack-xss":                  "Cross-site scripting (XSS)",
	"attack-rce":                  "Remote command execution",
	"attack-lfi":                  "Access to restricted files",
	"attack-rfi":                  "Remote file inclusion",
	"attack-php":                  "PHP code injection",
	"attack-java":                 "Java code injection",
	"attack-injection-php":        "PHP code injection",
	"attack-injection-generic":    "Code injection",
	"attack-generic":              "Code injection",
	"attack-protocol":             "Malformed request",
	"attack-reputation-scanner":   "Vulnerability scanner",
	"attack-reputation-scripting": "Automated script",
	"attack-fixation":             "Session hijacking",
	"attack-disclosure":           "Information leak",
	"attack-multipart-header":     "Malformed upload",
	"attack-ssrf":                 "Server-side request forgery",
	"attack-deserialization":      "Unsafe data deserialization",
}

func ruleCategory(tags []string) string {
	for _, t := range tags {
		if c, ok := attackCategories[t]; ok {
			return c
		}
	}
	for _, t := range tags {
		if strings.HasPrefix(t, "attack-") {
			return strings.ReplaceAll(strings.TrimPrefix(t, "attack-"), "-", " ")
		}
	}
	return "Suspicious request"
}

type rawLogEntry struct {
	Transaction struct {
		Timestamp string `json:"timestamp"`
		ClientIP  string `json:"client_ip"`
		Request   struct {
			Method string `json:"method"`
			URI    string `json:"uri"`
		} `json:"request"`
		Response struct {
			Status int `json:"status"`
		} `json:"response"`
		IsInterrupted bool `json:"is_interrupted"`
	} `json:"transaction"`
	Messages []struct {
		ErrorMessage string `json:"error_message"`
	} `json:"messages"`
}

// summarizeLogs turns raw Coraza JSON lines (newest first) into plain events and per-rule groups; 949/959/980 only add up scores, so they decide the result instead of showing as a reason
func summarizeLogs(lines []json.RawMessage) LogSummary {
	var s LogSummary
	groups := map[string]*LogRuleGroup{}
	var order []string

	for _, line := range lines {
		var e rawLogEntry
		if json.Unmarshal(line, &e) != nil {
			continue
		}
		ev := LogEvent{
			Time: e.Transaction.Timestamp, IP: e.Transaction.ClientIP, Method: e.Transaction.Request.Method,
			URI: e.Transaction.Request.URI, Status: e.Transaction.Response.Status,
		}
		scored := false
		for _, m := range e.Messages {
			id := firstSubmatch(logRuleIDRE, m.ErrorMessage)
			if id == "" {
				continue
			}
			if isScoringRule(id) {
				if strings.HasPrefix(id, "949") || strings.HasPrefix(id, "959") {
					scored = true
				}
				continue
			}
			if hasRule(ev.Rules, id) {
				continue
			}
			var tags []string
			for _, t := range logTagRE.FindAllStringSubmatch(m.ErrorMessage, -1) {
				tags = append(tags, t[1])
			}
			ev.Rules = append(ev.Rules, LogRule{
				ID: id, Msg: firstSubmatch(logRuleMsgRE, m.ErrorMessage), Category: ruleCategory(tags),
				Data: strings.TrimPrefix(firstSubmatch(logDataRE, m.ErrorMessage), "Matched Data: "),
			})
		}
		if len(ev.Rules) == 0 && !e.Transaction.IsInterrupted {
			s.Other++
			continue
		}
		switch {
		case e.Transaction.IsInterrupted:
			ev.Result = "blocked"
			s.Blocked++
		case scored:
			ev.Result = "would_block"
			s.WouldBlock++
		default:
			ev.Result = "flagged"
			s.Flagged++
		}
		s.Events = append(s.Events, ev)

		for _, r := range ev.Rules {
			g, ok := groups[r.ID]
			if !ok {
				g = &LogRuleGroup{LogRule: r, LastSeen: ev.Time}
				groups[r.ID] = g
				order = append(order, r.ID)
			}
			g.Count++
			if ev.Result != "flagged" {
				g.Blocked++
			}
			path := ev.URI
			if i := strings.IndexByte(path, '?'); i >= 0 {
				path = path[:i]
			}
			if len(g.Paths) < 3 && !containsString(g.Paths, path) {
				g.Paths = append(g.Paths, path)
			}
		}
	}

	for _, id := range order {
		s.Groups = append(s.Groups, *groups[id])
	}
	// rules behind actual blocks first, then the noisiest
	sort.SliceStable(s.Groups, func(i, j int) bool {
		if s.Groups[i].Blocked != s.Groups[j].Blocked {
			return s.Groups[i].Blocked > s.Groups[j].Blocked
		}
		return s.Groups[i].Count > s.Groups[j].Count
	})
	return s
}

func firstSubmatch(re *regexp.Regexp, s string) string {
	if m := re.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return ""
}

func hasRule(rules []LogRule, id string) bool {
	for _, r := range rules {
		if r.ID == id {
			return true
		}
	}
	return false
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// recentLogLines is how many of the newest log entries the domain page summarizes
const recentLogLines = 500

// readLogTail returns up to max of the newest lines of a Coraza JSON log, newest first, reading backwards so big logs stay cheap
func readLogTail(path string, max int) []json.RawMessage {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil
	}
	const blockSize = 64 * 1024
	var out []json.RawMessage
	var rest []byte
	for pos := info.Size(); pos > 0 && len(out) < max; {
		size := int64(blockSize)
		if size > pos {
			size = pos
		}
		pos -= size
		buf := make([]byte, size)
		if _, err := f.ReadAt(buf, pos); err != nil {
			return out
		}
		buf = append(buf, rest...)
		lines := bytes.Split(buf, []byte("\n"))
		// the first piece may be cut in half, keep it for the next block unless we're at the start of the file
		rest = lines[0]
		if pos == 0 {
			rest = nil
		} else {
			lines = lines[1:]
		}
		for i := len(lines) - 1; i >= 0 && len(out) < max; i-- {
			if l := bytes.TrimSpace(lines[i]); len(l) > 0 {
				out = append(out, json.RawMessage(append([]byte(nil), l...)))
			}
		}
	}
	return out
}

// latestEventsShown is how many rows the domain page's "Latest requests" table has
const latestEventsShown = 10

// recentSummary is the domain page's summary: groups count every recent entry, only the newest events are sent to the page
func recentSummary(domain string) LogSummary {
	s := summarizeLogs(readLogTail(wafLogPath(domain), recentLogLines))
	if len(s.Events) > latestEventsShown {
		s.Events = s.Events[:latestEventsShown]
	}
	// empty lists, not null, the page reads .length on them
	if s.Events == nil {
		s.Events = []LogEvent{}
	}
	if s.Groups == nil {
		s.Groups = []LogRuleGroup{}
	}
	return s
}

// purgeDomainLog empties the domain's WAF log in place, Coraza keeps its file open and appends, so new entries start at the top again
func purgeDomainLog(domain string) error {
	err := os.Truncate(wafLogPath(domain), 0)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
