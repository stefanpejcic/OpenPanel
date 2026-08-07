package account

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRPIDAndOrigin(t *testing.T) {
	cases := []struct {
		name       string
		host       string
		forwarded  string
		tls        bool
		wantRPID   string
		wantOrigin string
	}{
		{"plain http, no proxy header", "185.7.32.112:5000", "", false, "185.7.32.112", "http://185.7.32.112:5000"},
		{"behind https-terminating proxy", "example.com:2083", "https", false, "example.com", "https://example.com:2083"},
		{"forwarded proto with extra hops", "example.com", "https,http", false, "example.com", "https://example.com"},
		{"direct tls, no proxy header", "example.com", "", true, "example.com", "https://example.com"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Host = c.host
			if c.forwarded != "" {
				r.Header.Set("X-Forwarded-Proto", c.forwarded)
			}
			if c.tls {
				r.TLS = &tls.ConnectionState{}
			}
			gotRPID, gotOrigin := rpIDAndOrigin(r)
			if gotRPID != c.wantRPID {
				t.Errorf("rpID = %q, want %q", gotRPID, c.wantRPID)
			}
			if gotOrigin != c.wantOrigin {
				t.Errorf("origin = %q, want %q", gotOrigin, c.wantOrigin)
			}
		})
	}
}

func TestEncodeDecodeCredentialID(t *testing.T) {
	raw := []byte{0x01, 0x02, 0xff, 0x00, 0xab}
	encoded := encodeCredentialID(raw)
	decoded, err := decodeCredentialID(encoded)
	if err != nil {
		t.Fatalf("decodeCredentialID: %v", err)
	}
	if string(decoded) != string(raw) {
		t.Errorf("round-trip mismatch: got %v, want %v", decoded, raw)
	}
}

func TestPasskeyUserWebAuthnID(t *testing.T) {
	u := &passkeyUser{userID: 42, username: "bob"}
	if string(u.WebAuthnID()) != "42" {
		t.Errorf("WebAuthnID() = %q, want %q (decimal-ASCII, matching the scheme existing passkeys were registered under)", u.WebAuthnID(), "42")
	}
	if u.WebAuthnName() != "bob" || u.WebAuthnDisplayName() != "bob" {
		t.Errorf("WebAuthnName/DisplayName = %q/%q, want %q", u.WebAuthnName(), u.WebAuthnDisplayName(), "bob")
	}
}
