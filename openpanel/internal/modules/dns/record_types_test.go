package dns

import (
	"slices"
	"strings"
	"testing"
)

func TestParseDNSRecordTypes(t *testing.T) {
	defaults := []string{"A", "AAAA", "CAA", "CNAME", "MX", "SRV", "TLSA", "TXT"}
	for raw, want := range map[string][]string{
		"":                    defaults,
		"   ":                 defaults,
		"BOGUS":               defaults,
		defaultDNSRecordTypes: defaults,
		"txt,a  Dmarc NOPE":   {"A", "TXT", "DMARC"},
	} {
		if got := parseDNSRecordTypes(raw); !slices.Equal(got, want) {
			t.Errorf("%q: got %v, want %v", raw, got, want)
		}
	}
	if slices.Contains(editableRecordTypes([]string{"A", "DMARC"}), "DMARC") {
		t.Error("DMARC can't be picked on a saved row")
	}
}

func TestNormalizeNewRecord(t *testing.T) {
	const d = "example.com"
	ok := []struct{ name, typ, rec, wantName, wantType, wantRec string }{
		{"www", "A", "1.2.3.4", "www", "A", "1.2.3.4"},
		{"www", "CNAME", "example.com", "www", "CNAME", "example.com."},
		{"www", "CNAME", "other.com", "www", "CNAME", "other.com."},
		{"www", "CNAME", "other.com.", "www", "CNAME", "other.com."},
		{"www", "CNAME", "host", "www", "CNAME", "host"},
		{"@", "MX", "mail.other.com", "@", "MX", "mail.other.com."},
		{"@", "TXT", "v=spf1 include:example.com", "@", "TXT", `"v=spf1 include:example.com"`},
		{"_dmarc", "DMARC", "v=DMARC1; p=none; rua=mailto:a@example.com", "_dmarc", "TXT", `"v=DMARC1; p=none; rua=mailto:a@example.com"`},
		{"sub", "NS", "ns1.other.net", "sub", "NS", "ns1.other.net."},
		{"@", "HTTPS", `1 . alpn="h2,h3"`, "@", "HTTPS", `1 . alpn="h2,h3"`},
		{"_x", "SVCB", "1 svc.example.com", "_x", "SVCB", "1 svc.example.com."},
		{"_y", "SVCB", "1 svc.other.net port=8443", "_y", "SVCB", "1 svc.other.net. port=8443"},
		{"@", "HTTPS", "1 . port=443", "@", "HTTPS", "1 . port=443"},
		{"h", "SSHFP", "4 2 abcdef01", "h", "SSHFP", "4 2 abcdef01"},
		{"_ftp._tcp", "URI", "10 1 https://example.com/", "_ftp._tcp", "URI", `10 1 "https://example.com/"`},
	}
	for _, c := range ok {
		n, ty, r, err := normalizeNewRecord(d, c.name, c.typ, c.rec)
		if err != nil || n != c.wantName || ty != c.wantType || r != c.wantRec {
			t.Errorf("%s %s: got %q %q %q err=%v", c.typ, c.rec, n, ty, r, err)
		}
	}
	bad := []struct{ name, typ, rec string }{
		{"@", "NS", "ns1.other.net"},
		{"example.com", "NS", "ns1.other.net"},
		{"sub", "NS", "bad host"},
		{"www", "DMARC", "v=DMARC1; p=none"},
		{"_dmarc", "DMARC", "p=none"},
		{"@", "HTTPS", "nonsense"},
		{"h", "SSHFP", "4 2 not-hex"},
		{"u", "URI", "https://example.com/"},
	}
	for _, c := range bad {
		if _, _, _, err := normalizeNewRecord(d, c.name, c.typ, c.rec); err == nil {
			t.Errorf("%s %s %s should be refused", c.name, c.typ, c.rec)
		}
	}
}

func TestCheckRecordEdit(t *testing.T) {
	allowed := recordTypeSet([]string{"A", "TXT", "NS"})
	const d = "example.com"
	for _, c := range []struct {
		old, new string
		ok       bool
	}{
		{"www 3600 IN A 1.1.1.1\n", "www 3600 IN A 2.2.2.2", true},
		{"www 3600 IN A 1.1.1.1\n", "www 3600 A 2.2.2.2", true},
		{"www 3600 IN A 1.1.1.1\n", "www IN 3600 URI 1 1 \"x\"", false},
		{"www 3600 IN URI 1 1 \"x\"\n", "www 3600 IN A 2.2.2.2", false},
		{"www 3600 IN A 1.1.1.1\n", "; note\nwww 3600 IN URI 1 1 \"x\"", false},
		{"www 3600 IN A 1.1.1.1\n", "www 3600 IN TXT \"URI ; SSHFP\"", true},
		{"www 3600 IN A 1.1.1.1\n", "@ 3600 IN NS ns1.other.net.", false},
		{"www 3600 IN A 1.1.1.1\n", "sub 3600 IN NS ns1.other.net.", true},
		{"www 3600 IN TXT ( \"a\"\n  \"URI\" )\n", "www 3600 IN A 2.2.2.2", true},
	} {
		if err := checkRecordEdit(allowed, d, c.old, c.new); (err == nil) != c.ok {
			t.Errorf("%q -> %q: err=%v, want ok=%v", c.old, c.new, err, c.ok)
		}
	}
}

func TestZoneWriteError(t *testing.T) {
	orig := checkZoneContent
	t.Cleanup(func() { checkZoneContent = orig })
	bad := "dns_rdata_fromtext: /tmp/tmp.abc:32: near eol: unexpected end of input\nzone x/IN: not loaded due to errors.\n"
	checkZoneContent = func(_, content string) string {
		if strings.Contains(content, "BAD") {
			return bad
		}
		return ""
	}

	if err := zoneWriteError("x", "good", "good"); err != nil {
		t.Errorf("good zone: %v", err)
	}
	err := zoneWriteError("x", "good", "good BAD")
	if err == nil || err.Error() != "The DNS server refused this record: near eol: unexpected end of input" {
		t.Errorf("got %v", err)
	}
	if err := zoneWriteError("x", "already BAD", "still BAD"); err != nil {
		t.Errorf("an already broken zone has to stay editable: %v", err)
	}
}

func TestDmarcExistsAndAppendedZone(t *testing.T) {
	zone := "www 3600 IN A 1.1.1.1\n_dmarc 14400 IN TXT \"v=DMARC1; p=none\"\n"
	if !dmarcExists(zone, "_dmarc") {
		t.Error("should find the policy")
	}
	if dmarcExists(zone, "_dmarc.sub") || dmarcExists(zone, "www") {
		t.Error("only the same name counts")
	}
	if got := appendedZone("a", "b"); got != "a\nb\n" {
		t.Errorf("got %q", got)
	}
}
