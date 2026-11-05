package dynamicdns

import (
	"testing"
	"time"
)

func TestTokenNeedsLength(t *testing.T) {
	for _, tok := range []string{"a", "abc123", "short15charsxxx", "bad-token-with-dash"} {
		if dynDNSTokenRE.MatchString(tok) {
			t.Errorf("accepted %q", tok)
		}
	}
	if !dynDNSTokenRE.MatchString(generateToken(16)) {
		t.Fatal("generated token is refused")
	}
}

func TestUpdateLimiter(t *testing.T) {
	l := &ipLimiter{limit: 3, window: time.Minute, hits: map[string]*ipWindow{}}
	for i := 0; i < 3; i++ {
		if !l.allow("192.0.2.1") {
			t.Fatalf("hit %d refused", i)
		}
	}
	if l.allow("192.0.2.1") {
		t.Fatal("4th hit allowed")
	}
	if !l.allow("192.0.2.2") {
		t.Fatal("other IP refused")
	}
}
