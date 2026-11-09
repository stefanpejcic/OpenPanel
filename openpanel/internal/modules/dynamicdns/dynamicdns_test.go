package dynamicdns

import (
	"strings"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	tok := generateToken(16)
	if len(tok) != 16 {
		t.Fatalf("expected 16-char token, got %d: %q", len(tok), tok)
	}
	for _, r := range tok {
		if !strings.ContainsRune(tokenAlphabet, r) {
			t.Errorf("token contains unexpected char %q", r)
		}
	}
}

func TestValidateSubdomain(t *testing.T) {
	cases := map[string]bool{
		"home": true, "my-router": true, "a": true,
		"": false, "-bad": false, "bad-": false, "has space": false, "under_score": false,
	}
	for in, want := range cases {
		if got := validateSubdomain(in); got != want {
			t.Errorf("validateSubdomain(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestValidateIPv4(t *testing.T) {
	cases := map[string]bool{
		"1.2.3.4": true, "0.0.0.0": true, "198.51.100.23": true,
		"::1": false, "2001:db8::1": false, "::ffff:1.2.3.4": false, "1.2.3": false, "256.1.1.1": false, "not-an-ip": false, "": false,
	}
	for in, want := range cases {
		if got := validateIPv4(in); got != want {
			t.Errorf("validateIPv4(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIPFitsRecordType(t *testing.T) {
	cases := []struct {
		ip, typ string
		want    bool
	}{
		{"1.2.3.4", "A", true},
		{"2001:db8::1", "A", false},
		{"::ffff:1.2.3.4", "A", false},
		{"2001:db8::1", "AAAA", true},
		{"1.2.3.4", "AAAA", false},
		{"1.2.3.4", "CNAME", false},
	}
	for _, c := range cases {
		if got := ipFitsRecordType(c.ip, c.typ); got != c.want {
			t.Errorf("ipFitsRecordType(%q, %q) = %v, want %v", c.ip, c.typ, got, c.want)
		}
	}
	if wrongFamilyMessage("A") != "IPv4 address required" || wrongFamilyMessage("AAAA") != "IPv6 address required" {
		t.Error("unexpected wrong-family message")
	}
}

const testZone = `$TTL 14400
@ IN SOA ns1.example.com. admin.example.com. (
	1 ; serial
	)
home 300 IN A 1.2.3.4 ; webcall=abc123 updated=2026-01-01T00:00:00Z
www 14400 IN A 5.6.7.8
shop.example.com. IN CNAME example.com.
mail 14400 IN MX 10 mx.example.com.
txt IN TXT "home"
other.example.net. IN A 9.9.9.9
`

func TestHostConflict(t *testing.T) {
	cases := []struct {
		sub  string
		skip int
		want string
	}{
		{"home", 0, conflictDDNS},
		{"HOME", 0, conflictDDNS},
		{"home", 5, conflictNone},
		{"www", 0, conflictRecord},
		{"shop", 0, conflictRecord},
		{"mail", 0, conflictNone},
		{"txt", 0, conflictNone},
		{"other", 0, conflictNone},
		{"nas", 0, conflictNone},
	}
	for _, c := range cases {
		if got := hostConflict(testZone, "example.com", c.sub, c.skip); got != c.want {
			t.Errorf("hostConflict(%q, skip %d) = %q, want %q", c.sub, c.skip, got, c.want)
		}
	}
}

func TestHostProblem(t *testing.T) {
	cases := []struct {
		sub, ip string
		skip    int
		want    string
	}{
		{"nas", "0.0.0.0", 0, ""},
		{"nas", "198.51.100.23", 0, ""},
		{"", "1.2.3.4", 0, "Enter a subdomain, like home."},
		{"-bad", "1.2.3.4", 0, "Use letters, numbers and hyphens (up to 63), not starting or ending with a hyphen."},
		{"nas", "2001:db8::1", 0, "Enter an IPv4 address like 198.51.100.23. Dynamic DNS hosts are A records."},
		{"nas", "", 0, "Enter an IPv4 address like 198.51.100.23. Dynamic DNS hosts are A records."},
		{"home", "1.2.3.4", 0, "%(host)s is already a Dynamic DNS host."},
		{"home", "1.2.3.4", 5, ""},
		{"www", "1.2.3.4", 0, "%(host)s already has a DNS record. Pick another subdomain or remove that record first."},
	}
	for _, c := range cases {
		if got := hostProblem(testZone, "example.com", c.sub, c.ip, c.skip); got != c.want {
			t.Errorf("hostProblem(%q, %q, %d) = %q, want %q", c.sub, c.ip, c.skip, got, c.want)
		}
	}
}

func TestZoneOwnerLabel(t *testing.T) {
	cases := map[string]string{"@": "", "home": "home", "Home.Example.com.": "home", "example.com.": "", "x.example.net.": "x.example.net."}
	for in, want := range cases {
		if got := zoneOwnerLabel(in, "example.com"); got != want {
			t.Errorf("zoneOwnerLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildZoneLine(t *testing.T) {
	line := buildZoneLine("home", "A", "1.2.3.4", "tok123", "2026-01-01T00:00:00Z")
	want := "home 300 IN A 1.2.3.4 ; webcall=tok123 updated=2026-01-01T00:00:00Z"
	if line != want {
		t.Errorf("buildZoneLine = %q, want %q", line, want)
	}
}

func TestParseDynamicDNSFromZoneContent(t *testing.T) {
	content := "$TTL 14400\n" +
		"@ IN SOA ns1.example.com. admin.example.com. ( 1 )\n" +
		"home 300 IN A 1.2.3.4 ; webcall=abc123 updated=2026-01-01T00:00:00Z\n" +
		"www 14400 IN A 5.6.7.8\n"

	entries := parseDynamicDNSFromZoneContent(content)
	if len(entries) != 1 {
		t.Fatalf("expected 1 dynamic dns entry, got %d: %+v", len(entries), entries)
	}
	e := entries[0]
	if e.Subdomain != "home" || e.Record != "1.2.3.4" || e.Token != "abc123" || e.LastUpdated != "2026-01-01T00:00:00Z" {
		t.Errorf("unexpected entry: %+v", e)
	}
	if e.LineNumber != 3 {
		t.Errorf("LineNumber = %d, want 3", e.LineNumber)
	}
}

func TestRegeneratedLineKeepsEverythingButTheToken(t *testing.T) {
	zone := "home 300 IN A 198.51.100.7 ; webcall=OldTokenOldToken updated=2026-10-01T08:00:00Z\n"
	entries := parseDynamicDNSFromZoneContent(zone)
	if len(entries) != 1 {
		t.Fatalf("parsed %d entries", len(entries))
	}
	line := regeneratedLine(entries[0])
	again := parseDynamicDNSFromZoneContent(line + "\n")
	if len(again) != 1 {
		t.Fatalf("the new line doesn't parse: %q", line)
	}
	got, old := again[0], entries[0]
	if got.Token == old.Token || !dynDNSTokenRE.MatchString(got.Token) || len(got.Token) != 16 {
		t.Errorf("token %q, old %q", got.Token, old.Token)
	}
	if got.Subdomain != old.Subdomain || got.Type != old.Type || got.Record != old.Record || got.LastUpdated != old.LastUpdated {
		t.Errorf("record changed: %+v -> %+v", old, got)
	}
}
