// Package sieveparser implements a tokenizer + recursive-descent parser
// that converts a .dovecot.sieve file into the structured filter list the
// emails/filter.html GUI editor expects. Only the read side (Parse) is
// implemented - a Sieve-generating counterpart is unnecessary dead code:
// the GUI's "Save filters" button regenerates Sieve client-side in JS
// instead, and posts the raw text directly.
//
// Supported subset (covers ~95% of real-world Dovecot personal filters):
//   - require [...]
//   - if / elsif / else blocks (flattened to independent filters for the GUI)
//   - anyof / allof / single test
//   - Tests: header, address, body :text, exists
//   - Actions: fileinto, redirect (incl. :copy), discard, reject, stop, addflag, vacation
//   - Leading # comment on an if block is used as the filter name
package sieveparser

import (
	"regexp"
	"strconv"
	"strings"
)

// Rule is one condition inside a Filter.
type Rule struct {
	Field string `json:"field"`
	Match string `json:"match"`
	Value string `json:"value"`
}

// AutoresponderValue is the structured value of an "autoresponder" action.
type AutoresponderValue struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
	Days    int    `json:"days"`
}

// Action is one action inside a Filter. Value is a plain string for every
// action type except "autoresponder", where it's an *AutoresponderValue.
type Action struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

// Filter is one if/elsif block, flattened for the GUI editor.
type Filter struct {
	Name    string   `json:"name"`
	Logic   string   `json:"logic"`
	Rules   []Rule   `json:"rules"`
	Actions []Action `json:"actions"`
}

// Parse tokenizes and parses a .dovecot.sieve file into a flattened list
// of filters.
func Parse(source string) []Filter {
	if strings.TrimSpace(source) == "" {
		return nil
	}

	tokens := tokenize(source)
	var filters []Filter
	i := 0
	pendingComment := ""
	hasPendingComment := false

	for i < len(tokens) {
		tok := tokens[i]

		if strings.HasPrefix(tok, "#") {
			pendingComment = strings.TrimSpace(tok[1:])
			hasPendingComment = true
			i++
			continue
		}

		lower := strings.ToLower(tok)
		if lower == "if" || lower == "elsif" || lower == "else" {
			name := ""
			if hasPendingComment {
				name = pendingComment
			}
			block, consumed := parseIfBlock(tokens, i, name)
			if block != nil {
				filters = append(filters, *block)
			}
			pendingComment = ""
			hasPendingComment = false
			i += consumed
			continue
		}

		pendingComment = ""
		hasPendingComment = false
		i++
	}

	return filters
}

func tokenize(s string) []string {
	var tokens []string
	i := 0
	n := len(s)

	for i < n {
		c := s[i]

		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			i++
			continue
		}

		if c == '#' {
			end := strings.IndexByte(s[i:], '\n')
			if end == -1 {
				end = n
			} else {
				end += i
			}
			tokens = append(tokens, s[i:end])
			i = end
			continue
		}

		if c == '"' {
			j := i + 1
			for j < n {
				if s[j] == '\\' {
					j += 2
					continue
				}
				if s[j] == '"' {
					j++
					break
				}
				j++
			}
			if j > n {
				j = n
			}
			tokens = append(tokens, s[i:j])
			i = j
			continue
		}

		if c == '[' {
			depth := 0
			j := i
			for j < n {
				if s[j] == '[' {
					depth++
				} else if s[j] == ']' {
					depth--
					if depth == 0 {
						j++
						break
					}
				}
				j++
			}
			tokens = append(tokens, s[i:j])
			i = j
			continue
		}

		if c == '{' {
			depth := 0
			j := i
			for j < n {
				if s[j] == '"' {
					j++
					for j < n {
						if s[j] == '\\' {
							j += 2
							continue
						}
						if s[j] == '"' {
							j++
							break
						}
						j++
					}
					continue
				}
				if s[j] == '{' {
					depth++
				} else if s[j] == '}' {
					depth--
					if depth == 0 {
						j++
						break
					}
				}
				j++
			}
			tokens = append(tokens, s[i:j])
			i = j
			continue
		}

		j := i
		for j < n && !strings.ContainsRune(" \t\r\n\"[]{}#;,", rune(s[j])) {
			j++
		}
		if j == i {
			j++
		}
		tokens = append(tokens, s[i:j])
		i = j
	}

	return tokens
}

func parseIfBlock(tokens []string, start int, name string) (*Filter, int) {
	i := start
	i++ // consume if/elsif/else keyword

	var testTokens []string
	for i < len(tokens) && !strings.HasPrefix(tokens[i], "{") {
		testTokens = append(testTokens, tokens[i])
		i++
	}

	blockTok := "{}"
	if i < len(tokens) {
		blockTok = tokens[i]
	}
	i++

	logic, rules := parseTest(testTokens)
	actions := parseActions(blockTok)

	if len(rules) == 0 || len(actions) == 0 {
		return nil, i - start
	}

	return &Filter{Name: name, Logic: logic, Rules: rules, Actions: actions}, i - start
}

func parseTest(tokens []string) (string, []Rule) {
	if len(tokens) == 0 {
		return "anyof", nil
	}

	first := strings.ToLower(tokens[0])

	if first == "anyof" || first == "allof" {
		logic := first
		joined := strings.TrimSpace(strings.Join(tokens[1:], " "))
		if strings.HasPrefix(joined, "(") && strings.HasSuffix(joined, ")") {
			joined = joined[1 : len(joined)-1]
		}
		subTests := splitTopLevelCommas(joined)
		var rules []Rule
		for _, st := range subTests {
			if rule := parseSingleTest(strings.Fields(strings.TrimSpace(st))); rule != nil {
				rules = append(rules, *rule)
			}
		}
		return logic, rules
	}

	if rule := parseSingleTest(tokens); rule != nil {
		return "anyof", []Rule{*rule}
	}
	return "anyof", nil
}

func splitTopLevelCommas(s string) []string {
	var parts []string
	depth := 0
	inQuote := false
	var buf strings.Builder

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch {
		case c == '"' && !inQuote:
			inQuote = true
			buf.WriteRune(c)
		case c == '"' && inQuote:
			inQuote = false
			buf.WriteRune(c)
		case c == '\\' && inQuote:
			buf.WriteRune(c)
			i++
			if i < len(runes) {
				buf.WriteRune(runes[i])
			}
		case (c == '(' || c == '[') && !inQuote:
			depth++
			buf.WriteRune(c)
		case (c == ')' || c == ']') && !inQuote:
			depth--
			buf.WriteRune(c)
		case c == ',' && depth == 0 && !inQuote:
			parts = append(parts, buf.String())
			buf.Reset()
		default:
			buf.WriteRune(c)
		}
	}
	if buf.Len() > 0 {
		parts = append(parts, buf.String())
	}
	return parts
}

func parseSingleTest(tokens []string) *Rule {
	if len(tokens) == 0 {
		return nil
	}

	negated := false
	i := 0
	if strings.ToLower(tokens[i]) == "not" {
		negated = true
		i++
	}
	if i >= len(tokens) {
		return nil
	}

	testName := strings.ToLower(tokens[i])
	i++
	rest := tokens[i:]

	switch testName {
	case "header":
		return parseHeaderTest(rest, negated)
	case "address":
		return parseAddressTest(rest, negated)
	case "body":
		return parseBodyTest(rest, negated)
	case "exists":
		return parseExistsTest(rest, negated)
	}
	return nil
}

func matchTagToOp(tag string, negated bool) string {
	tag = strings.TrimPrefix(strings.ToLower(tag), ":")
	base := "contains"
	switch tag {
	case "contains", "is", "matches":
		base = tag
	}
	if negated {
		return "not_" + base
	}
	return base
}

func unquote(s string) string {
	if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) && len(s) >= 2 {
		s = s[1 : len(s)-1]
	}
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}

func headerNameToField(hdr string) string {
	h := strings.TrimSpace(strings.ToLower(hdr))
	switch h {
	case "from":
		return "from"
	case "subject":
		return "subject"
	case "to":
		return "to"
	case "x-spam-status":
		return "spam_status"
	case "x-spam-score":
		return "spam_score"
	case "list-id":
		return "list_id"
	}
	return h
}

var quotedStringRE = regexp.MustCompile(`"([^"]*)"`)

func parseHeaderTest(tokens []string, negated bool) *Rule {
	i := 0
	matchTag := ":contains"
	for i < len(tokens) && strings.HasPrefix(tokens[i], ":") {
		switch strings.ToLower(tokens[i]) {
		case ":contains", ":is", ":matches", ":regex", ":count", ":value":
			matchTag = tokens[i]
		}
		i++
	}
	if i >= len(tokens) {
		return nil
	}
	headerTok := tokens[i]
	i++
	if i >= len(tokens) {
		return nil
	}
	valueTok := tokens[i]

	var headerName string
	if strings.HasPrefix(headerTok, "[") {
		m := quotedStringRE.FindStringSubmatch(headerTok)
		if m != nil {
			headerName = m[1]
		} else {
			headerName = "unknown"
		}
	} else {
		headerName = unquote(headerTok)
	}

	var value string
	if !strings.HasPrefix(valueTok, "[") {
		value = unquote(valueTok)
	} else if m := quotedStringRE.FindStringSubmatch(valueTok); m != nil {
		value = unquote(m[1])
	}

	field := headerNameToField(headerName)
	match := matchTagToOp(matchTag, negated)

	if (match == "matches" || match == "not_matches") && strings.HasSuffix(value, "*") && !strings.HasPrefix(value, "*") {
		if negated {
			match = "not_begins"
		} else {
			match = "begins"
		}
		value = strings.TrimRight(value, "*")
	} else if (match == "matches" || match == "not_matches") && strings.HasPrefix(value, "*") && !strings.HasSuffix(value, "*") {
		if negated {
			match = "not_ends"
		} else {
			match = "ends"
		}
		value = strings.TrimLeft(value, "*")
	}

	return &Rule{Field: field, Match: match, Value: value}
}

func parseAddressTest(tokens []string, negated bool) *Rule {
	i := 0
	matchTag := ":contains"
	for i < len(tokens) && strings.HasPrefix(tokens[i], ":") {
		switch strings.ToLower(tokens[i]) {
		case ":contains", ":is", ":matches":
			matchTag = tokens[i]
		}
		i++
	}
	if i+1 >= len(tokens) {
		return nil
	}
	valueTok := tokens[i+1]
	value := ""
	if !strings.HasPrefix(valueTok, "[") {
		value = unquote(valueTok)
	}
	match := matchTagToOp(matchTag, negated)
	return &Rule{Field: "any_recipient", Match: match, Value: value}
}

func parseBodyTest(tokens []string, negated bool) *Rule {
	i := 0
	matchTag := ":contains"
	for i < len(tokens) && strings.HasPrefix(tokens[i], ":") {
		switch strings.ToLower(tokens[i]) {
		case ":contains", ":is", ":matches":
			matchTag = tokens[i]
		}
		i++
	}
	if i >= len(tokens) {
		return nil
	}
	value := unquote(tokens[i])
	match := matchTagToOp(matchTag, negated)
	return &Rule{Field: "body", Match: match, Value: value}
}

func parseExistsTest(tokens []string, negated bool) *Rule {
	if len(tokens) == 0 {
		return nil
	}
	value := unquote(tokens[0])
	match := "contains"
	if negated {
		match = "not_contains"
	}
	return &Rule{Field: "any_header", Match: match, Value: value}
}

func parseActions(block string) []Action {
	inner := strings.TrimSpace(block)
	inner = strings.TrimPrefix(inner, "{")
	inner = strings.TrimSuffix(inner, "}")

	var actions []Action
	toks := tokenize(inner)
	i := 0
	for i < len(toks) {
		tok := strings.TrimSuffix(strings.ToLower(toks[i]), ";")

		switch {
		case tok == "fileinto":
			i++
			val := "INBOX"
			if i < len(toks) {
				val = unquote(strings.TrimSuffix(toks[i], ";"))
			}
			actions = append(actions, Action{Type: "deliver_folder", Value: val})

		case tok == "redirect":
			i++
			nxt := ""
			if i < len(toks) {
				nxt = strings.TrimSuffix(strings.ToLower(toks[i]), ";")
			}
			if nxt == ":copy" {
				i++
				val := ""
				if i < len(toks) {
					val = unquote(strings.TrimSuffix(toks[i], ";"))
				}
				actions = append(actions, Action{Type: "forward", Value: val})
			} else {
				val := ""
				if i < len(toks) {
					val = unquote(strings.TrimSuffix(toks[i], ";"))
				}
				actions = append(actions, Action{Type: "redirect", Value: val})
			}

		case tok == "vacation":
			i++
			days := 1
			subject := ""
			message := ""
			for i < len(toks) {
				t := toks[i]
				tl := strings.TrimSuffix(strings.ToLower(t), ";")
				if tl == ":days" {
					i++
					if i < len(toks) {
						if d, err := strconv.Atoi(strings.TrimSuffix(toks[i], ";")); err == nil {
							days = d
						} else {
							days = 1
						}
					}
					i++
				} else if tl == ":subject" {
					i++
					if i < len(toks) {
						subject = unquote(strings.TrimSuffix(toks[i], ";"))
					}
					i++
				} else if strings.HasPrefix(t, `"`) {
					message = unquote(strings.TrimSuffix(t, ";"))
					break
				} else {
					i++
				}
			}
			actions = append(actions, Action{Type: "autoresponder", Value: &AutoresponderValue{Subject: subject, Message: message, Days: days}})

		case tok == "discard":
			actions = append(actions, Action{Type: "discard", Value: ""})

		case tok == "reject":
			i++
			val := ""
			if i < len(toks) {
				val = unquote(strings.TrimSuffix(toks[i], ";"))
			}
			actions = append(actions, Action{Type: "reject", Value: val})

		case tok == "stop":
			actions = append(actions, Action{Type: "stop", Value: ""})

		case tok == "addflag":
			i++
			flag := ""
			if i < len(toks) {
				flag = unquote(strings.TrimSuffix(toks[i], ";"))
			}
			lower := strings.ToLower(flag)
			if lower == `\seen` || lower == `\\seen` {
				actions = append(actions, Action{Type: "mark_read", Value: ""})
			} else {
				actions = append(actions, Action{Type: "add_flag", Value: flag})
			}

		case strings.HasPrefix(tok, "#"):
			// comments inside blocks are skipped
		}

		i++
	}

	return actions
}
