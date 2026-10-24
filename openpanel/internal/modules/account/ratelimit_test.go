package account

import (
	"strings"
	"testing"
	"time"
)

func TestLoginRateLimiterAllow(t *testing.T) {
	l := &loginRateLimiter{limit: 3, windows: map[string]*window{}}

	for i := 0; i < 3; i++ {
		if !l.Allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed within the limit", i+1)
		}
	}
	if l.Allow("1.2.3.4") {
		t.Error("4th request within the window should be rejected")
	}
}

func TestLoginRateLimiterPerIP(t *testing.T) {
	l := &loginRateLimiter{limit: 1, windows: map[string]*window{}}

	if !l.Allow("1.1.1.1") {
		t.Error("first request from 1.1.1.1 should be allowed")
	}
	if !l.Allow("2.2.2.2") {
		t.Error("first request from a different IP should be allowed independently")
	}
	if l.Allow("1.1.1.1") {
		t.Error("second request from 1.1.1.1 should be rejected")
	}
}

func TestClearFailedAttempts(t *testing.T) {
	failedAttemptsMu.Lock()
	failedAttempts["9.9.9.9"] = 5
	failedAttemptsMu.Unlock()

	clearFailedAttempts("9.9.9.9")

	failedAttemptsMu.Lock()
	_, ok := failedAttempts["9.9.9.9"]
	failedAttemptsMu.Unlock()

	if ok {
		t.Error("expected failed attempts counter to be cleared")
	}
}

func TestTempbanEntry(t *testing.T) {
	now := time.Unix(1700000000, 0)
	if got, ok := tempbanEntry("203.0.113.7", now); !ok || got != "1700000000|203.0.113.7||in|3600|Too many failed login attempts" {
		t.Errorf("valid ip: got %q, %v", got, ok)
	}
	if got, ok := tempbanEntry("::ffff:203.0.113.7", now); !ok || !strings.HasPrefix(got, "1700000000|203.0.113.7|") {
		t.Errorf("mapped ipv4: got %q, %v", got, ok)
	}
	for _, bad := range []string{"", "not-an-ip", "203.0.113.7|22", "203.0.113.0/24", "127.0.0.1", "::1", "0.0.0.0", "fe80::1%eth0"} {
		if got, ok := tempbanEntry(bad, now); ok {
			t.Errorf("%q should be refused, got %q", bad, got)
		}
	}
}
