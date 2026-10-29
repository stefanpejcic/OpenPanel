package webserver

import (
	"regexp"
	"strconv"
	"strings"
)

// failed test results that come from running the test, not from the web server
const (
	TestNoOutput = "config test failed without any output"
	TestTimedOut = "config test did not finish in time"
)

// ConfigProblem is one problem from a failed config test, File and Line stay empty when the output doesn't name them
type ConfigProblem struct {
	File    string
	Line    int
	Message string
}

var (
	// nginx puts "2026/10/27 13:10:17 [emerg] 14#14: " or "nginx: [emerg] " in front of each message
	nginxPrefix = regexp.MustCompile(`^(?:\d{4}/\d\d/\d\d \d\d:\d\d:\d\d |nginx: )\[(?:emerg|alert|crit|error)\] (?:\d+#\d+: )?`)
	nginxWhere  = regexp.MustCompile(`^(.*) in (\S+):(\d+)$`)
	apacheWhere = regexp.MustCompile(`Syntax error on line (\d+) of ([^:]+):`)
	apacheDup   = regexp.MustCompile(`^\S+:\d+:\s*`)
	// openlitespeed puts "2026-10-27 13:15:39.829106 [ERROR] [29] " in front of each message
	olsPrefix = regexp.MustCompile(`^(?:\S+ \S+ )?\[ERROR\] (?:\[\d+\] )?`)
	olsWhere  = regexp.MustCompile(`in file (\S+?):(\d+)`)
)

// ConfigTestProblems picks the problems out of a failed config test's output, without the timestamps, repeats and warnings around them
func ConfigTestProblems(serviceName, output string) []ConfigProblem {
	lines := strings.Split(output, "\n")
	var problems []ConfigProblem
	seen := map[ConfigProblem]bool{}
	add := func(p ConfigProblem) {
		if p.Message != "" && !seen[p] {
			seen[p] = true
			problems = append(problems, p)
		}
	}

	switch serviceName {
	case "nginx", "openresty":
		for _, l := range lines {
			l = strings.TrimSpace(l)
			loc := nginxPrefix.FindStringIndex(l)
			if loc == nil {
				continue
			}
			msg := l[loc[1]:]
			if m := nginxWhere.FindStringSubmatch(msg); m != nil {
				n, _ := strconv.Atoi(m[3])
				add(ConfigProblem{File: m[2], Line: n, Message: m[1]})
			} else {
				add(ConfigProblem{Message: msg})
			}
		}
	case "apache":
		for i, l := range lines {
			l = strings.TrimSpace(l)
			all := apacheWhere.FindAllStringSubmatchIndex(l, -1)
			if all == nil {
				continue
			}
			// an error in an included file is nested, the last file named is where it really is
			m := all[len(all)-1]
			n, _ := strconv.Atoi(l[m[2]:m[3]])
			msg := apacheDup.ReplaceAllString(strings.TrimSpace(l[m[1]:]), "")
			// AH00526 puts the reason on the next line
			if msg == "" && i+1 < len(lines) {
				msg = strings.TrimSpace(lines[i+1])
			}
			add(ConfigProblem{File: l[m[4]:m[5]], Line: n, Message: msg})
		}
	case "openlitespeed":
		for _, l := range lines {
			l = strings.TrimSpace(l)
			loc := olsPrefix.FindStringIndex(l)
			if loc == nil {
				continue
			}
			p := ConfigProblem{Message: l[loc[1]:]}
			if m := olsWhere.FindStringSubmatch(p.Message); m != nil {
				p.File = m[1]
				p.Line, _ = strconv.Atoi(m[2])
			}
			add(p)
		}
	}

	if len(problems) == 0 {
		// nothing recognized, the last lines are the likeliest to say what went wrong
		var rest []string
		for _, l := range lines {
			if l = strings.TrimSpace(l); l != "" {
				rest = append(rest, l)
			}
		}
		if len(rest) > 5 {
			rest = rest[len(rest)-5:]
		}
		for _, l := range rest {
			add(ConfigProblem{Message: l})
		}
	}
	return problems
}
