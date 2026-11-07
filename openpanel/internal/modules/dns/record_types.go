package dns

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

// DMARC isn't a real record type, the form uses it as a helper and it gets saved as TXT
const dmarcHelperType = "DMARC"

const defaultDNSRecordTypes = "A AAAA CAA CNAME MX SRV TLSA TXT"

// form order, openadmin's dns_record_types option has to list the same names
var knownDNSRecordTypes = []string{"A", "AAAA", "CAA", "CNAME", "MX", "SRV", "TLSA", "TXT", "HTTPS", "SVCB", "NS", "SSHFP", "URI", dmarcHelperType}

// parseDNSRecordTypes keeps known types in form order, a value with nothing usable falls back to the defaults so the form is never empty
func parseDNSRecordTypes(raw string) []string {
	wanted := map[string]bool{}
	for _, t := range strings.FieldsFunc(raw, func(r rune) bool { return r == ' ' || r == ',' || r == '\t' }) {
		wanted[strings.ToUpper(t)] = true
	}
	var types []string
	for _, t := range knownDNSRecordTypes {
		if wanted[t] {
			types = append(types, t)
		}
	}
	if len(types) == 0 && raw != defaultDNSRecordTypes {
		return parseDNSRecordTypes(defaultDNSRecordTypes)
	}
	return types
}

func allowedDNSRecordTypes(cfg config.Config) []string {
	return parseDNSRecordTypes(cfg.Get("dns_record_types", defaultDNSRecordTypes))
}

func recordTypeSet(types []string) map[string]bool {
	set := make(map[string]bool, len(types))
	for _, t := range types {
		set[t] = true
	}
	return set
}

// editableRecordTypes is what a saved row can be switched to, so without the DMARC helper
func editableRecordTypes(types []string) []string {
	out := make([]string, 0, len(types))
	for _, t := range types {
		if t != dmarcHelperType {
			out = append(out, t)
		}
	}
	return out
}

var (
	nsTargetRE = regexp.MustCompile(`^[A-Za-z0-9._-]+\.$`)
	svcbRE     = regexp.MustCompile(`^\d+ \S+( \S.*)?$`)
	sshfpRE    = regexp.MustCompile(`^\d+ \d+ [0-9A-Fa-f]+$`)
	uriRE      = regexp.MustCompile(`^\d+ \d+ "[^"]+"$`)
	uriPlainRE = regexp.MustCompile(`^(\d+ \d+) ([^"\s]+)$`)
)

func isZoneApex(name, domain string) bool {
	return name == "" || name == "@" || name == domain || name == domain+"."
}

// normalizeNewRecord turns form or API input into what goes in the zone, zone writes don't run named-checkzone so the new types get checked here
func normalizeNewRecord(domain, name, recordType, record string) (string, string, string, error) {
	if strings.HasSuffix(name, domain) {
		name += "."
	}

	switch recordType {
	case dmarcHelperType:
		if name != "_dmarc" && !strings.HasPrefix(name, "_dmarc.") {
			return "", "", "", fmt.Errorf("A DMARC record has to be named _dmarc.")
		}
		if !strings.HasPrefix(record, "v=DMARC1") || strings.Contains(record, `"`) {
			return "", "", "", fmt.Errorf("A DMARC record has to start with v=DMARC1 and can't contain quotes.")
		}
		recordType = "TXT"
	case "NS":
		if isZoneApex(name, domain) {
			return "", "", "", fmt.Errorf("NS records can only be added for a subdomain, not for the domain itself.")
		}
		if strings.Contains(record, ".") && !strings.HasSuffix(record, ".") {
			record += "."
		}
		if !nsTargetRE.MatchString(record) {
			return "", "", "", fmt.Errorf("Enter a nameserver hostname, for example ns1.example.net.")
		}
	case "CNAME", "MX":
		// without the dot BIND would append the zone name to other.com
		if strings.Contains(record, ".") && !strings.HasSuffix(record, ".") {
			record += "."
		}
	case "HTTPS", "SVCB":
		if !svcbRE.MatchString(record) {
			return "", "", "", fmt.Errorf("Use priority, target and optional parameters, for example: 1 . alpn=\"h2,h3\"")
		}
		// without the dot BIND would append the zone name to the target
		fields := strings.SplitN(record, " ", 3)
		if fields[1] != "." && strings.Contains(fields[1], ".") && !strings.HasSuffix(fields[1], ".") {
			fields[1] += "."
			record = strings.Join(fields, " ")
		}
	case "SSHFP":
		if !sshfpRE.MatchString(record) {
			return "", "", "", fmt.Errorf("Use algorithm, type and fingerprint, for example: 4 2 followed by the 64 character SHA-256 fingerprint")
		}
	case "URI":
		if m := uriPlainRE.FindStringSubmatch(record); m != nil {
			record = m[1] + ` "` + m[2] + `"`
		}
		if !uriRE.MatchString(record) {
			return "", "", "", fmt.Errorf("Use priority, weight and target, for example: 10 1 https://example.com/")
		}
	}

	// a dot on quoted text would become part of the value
	if recordType == "TXT" {
		if !(strings.HasPrefix(record, `"`) && strings.HasSuffix(record, `"`)) {
			record = `"` + record + `"`
		}
	} else if strings.HasSuffix(record, domain) {
		record += "."
	}
	return name, recordType, record, nil
}

var quotedTextRE = regexp.MustCompile(`"[^"]*"`)

// zoneRecordsIn lists each record's owner and type in a block of zone lines, skipping comments, directives and the inside of ( ) groups
func zoneRecordsIn(block string) (names, types []string) {
	depth := 0
	for _, line := range strings.Split(block, "\n") {
		clean := quotedTextRE.ReplaceAllString(line, `""`)
		if i := strings.Index(clean, ";"); i >= 0 {
			clean = clean[:i]
		}
		insideGroup := depth > 0
		depth += strings.Count(clean, "(") - strings.Count(clean, ")")
		if insideGroup {
			continue
		}
		fields := strings.Fields(clean)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "$") {
			continue
		}
		name := ""
		if clean[0] != ' ' && clean[0] != '\t' {
			name, fields = fields[0], fields[1:]
		}
		for _, f := range fields {
			upper := strings.ToUpper(f)
			if upper == "IN" || upper == "CH" || upper == "HS" || upper == "CS" || (f[0] >= '0' && f[0] <= '9') {
				continue
			}
			names, types = append(names, name), append(types, upper)
			break
		}
	}
	return names, types
}

// checkRecordEdit refuses replacing or writing a record of a type the admin hasn't allowed
func checkRecordEdit(allowed map[string]bool, domain, oldBlock, newBlock string) error {
	_, oldTypes := zoneRecordsIn(oldBlock)
	for _, t := range oldTypes {
		if !allowed[t] {
			return fmt.Errorf("Editing %s records is not allowed on this server.", t)
		}
	}
	newNames, newTypes := zoneRecordsIn(newBlock)
	for i, t := range newTypes {
		if !allowed[t] {
			return fmt.Errorf("%s records are not allowed on this server.", t)
		}
		if t == "NS" && isZoneApex(newNames[i], domain) {
			return fmt.Errorf("NS records can only be added for a subdomain, not for the domain itself.")
		}
	}
	return nil
}

// named-checkzone reads a temp copy so the live zone is only touched once it passes
const checkZoneScript = `f=$(mktemp); cat > "$f"; named-checkzone "$1" "$f"; rc=$?; rm -f "$f"; exit $rc`

// checkZoneContent returns BIND's complaint, "" when it's fine or when the check itself can't run
var checkZoneContent = func(domain, content string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "podman", "exec", "-i", "openpanel_dns", "sh", "-c", checkZoneScript, "sh", domain)
	cmd.Stdin = strings.NewReader(content)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	var exitErr *exec.ExitError
	// named-checkzone exits 1 on a bad zone, anything else means podman or the container had a problem
	if err := cmd.Run(); errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return out.String()
	}
	return ""
}

var checkZonePathRE = regexp.MustCompile(`^.*?/\S+:\d+: `)

// zoneWriteError stops a record BIND would refuse, since one bad line keeps the whole zone from loading
func zoneWriteError(domain, oldContent, newContent string) error {
	complaint := checkZoneContent(domain, newContent)
	if complaint == "" {
		return nil
	}
	// a zone that's already broken has to stay editable or the user can never fix it
	if checkZoneContent(domain, oldContent) != "" {
		return nil
	}
	first, _, _ := strings.Cut(strings.TrimSpace(complaint), "\n")
	return fmt.Errorf("The DNS server refused this record: %s", checkZonePathRE.ReplaceAllString(first, ""))
}

// dmarcExists finds a DMARC policy already published on name, a second one makes receivers ignore both
func dmarcExists(content, name string) bool {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name && strings.Contains(line, "v=DMARC1") {
			return true
		}
	}
	return false
}

// appendedZone is the zone as it would look with one more line, same newline handling as AppendZoneRecord
func appendedZone(content, line string) string {
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + line + "\n"
}
