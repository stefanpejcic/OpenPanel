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

func TestValidateIP(t *testing.T) {
	if !validateIP("1.2.3.4") {
		t.Error("expected valid IPv4")
	}
	if !validateIP("::1") {
		t.Error("expected valid IPv6")
	}
	if validateIP("not-an-ip") {
		t.Error("expected invalid IP to fail")
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
