package account

import "testing"

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
