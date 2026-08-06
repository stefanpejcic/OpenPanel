package dns

import (
	"strings"
)

// ZoneLineEntry is one logical DNS record, which may span multiple
// physical lines (parenthesized multi-line TXT/DKIM records).
type ZoneLineEntry struct {
	LineNumber    int
	EndLineNumber int
	Line          string
	Comment       string
	Multiline     bool
}

// parseZoneWithLineNumbers skips indented/comment/directive lines and the
// apex SOA/NS lines, merges parenthesized continuation lines into one
// logical record, and extracts any trailing (non-quoted) comment.
func parseZoneWithLineNumbers(content string) []ZoneLineEntry {
	rawLines := strings.Split(content, "\n")
	totalLines := len(rawLines)

	var entries []ZoneLineEntry
	for i := 0; i < totalLines; i++ {
		line := rawLines[i]
		lineNumber := i + 1

		if isSkippedZoneLine(line) {
			continue
		}

		strippedLine := strings.TrimSpace(line)
		if strippedLine == "" {
			continue
		}

		fullLine := strippedLine
		openCount := strings.Count(fullLine, "(")
		closeCount := strings.Count(fullLine, ")")
		endLineNumber := lineNumber
		j := i
		for openCount > closeCount && j+1 < totalLines {
			j++
			cont := strings.TrimSpace(rawLines[j])
			fullLine += " " + cont
			openCount += strings.Count(cont, "(")
			closeCount += strings.Count(cont, ")")
			endLineNumber = j + 1
		}

		entries = append(entries, ZoneLineEntry{
			LineNumber: lineNumber, EndLineNumber: endLineNumber, Line: fullLine,
			Comment: extractComment(fullLine), Multiline: endLineNumber != lineNumber,
		})
		i = j
	}

	return entries
}

// isSkippedZoneLine reports whether a line should be skipped when parsing
// a zone: indented/comment/directive lines, and the apex SOA/NS
// declaration lines.
func isSkippedZoneLine(line string) bool {
	if line == "" {
		return false
	}
	switch line[0] {
	case ' ', '\t', '#', '$', ';':
		return true
	}
	if strings.HasPrefix(line, "@") && (strings.Contains(line, "SOA") || strings.HasPrefix(line, "@ IN NS")) {
		return true
	}
	return false
}

// ZoneRow is one rendered row of the zone table view - a record split
// into up to 5 whitespace-separated fields, with the SOA line excluded
// and the display value comment/quote-stripped, precomputed server-side.
type ZoneRow struct {
	LineNumber    int
	EndLineNumber int
	Name          string
	TTL           string
	Type          string
	RawLine       string
	DisplayValue  string
	Comment       string
	Multiline     bool
}

// buildZoneRows splits each entry's merged line into fields, drops the
// SOA line, and computes the display value (quote-stripped,
// comment-stripped).
func buildZoneRows(entries []ZoneLineEntry) []ZoneRow {
	var rows []ZoneRow
	for _, item := range entries {
		values := splitMaxN(item.Line, 4)
		if len(values) < 4 {
			continue
		}
		recordType := values[3]
		if strings.Contains(recordType, "SOA") {
			continue
		}

		var displayValue string
		if len(values) > 4 {
			last := values[4]
			if strings.HasPrefix(last, `"`) && strings.HasSuffix(last, `"`) && len(last) >= 2 {
				displayValue = strings.Trim(last, `"`)
			} else {
				displayValue = last
			}
		}
		if item.Comment != "" {
			displayValue = strings.TrimSpace(strings.Replace(displayValue, " ; "+item.Comment, "", 1))
		}

		rows = append(rows, ZoneRow{
			LineNumber: item.LineNumber, EndLineNumber: item.EndLineNumber,
			Name: values[0], TTL: values[1], Type: recordType, RawLine: item.Line,
			DisplayValue: displayValue, Comment: item.Comment, Multiline: item.Multiline,
		})
	}
	return rows
}
