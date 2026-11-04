package account

import (
	"context"
	"maps"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/sessions"
)

func TestTwofaAttemptLimitSurvivesOlderCookie(t *testing.T) {
	const user = 41
	t.Cleanup(func() { forgetTwofaGuesses(user) })

	sess := sessions.NewSession(nil, "s")
	setTwofaPending(sess, user)
	saved := maps.Clone(sess.Values)
	for i := 0; i < twofaMaxAttempts; i++ {
		sess.Values = maps.Clone(saved)
		if _, ok := takeTwofaAttempt(sess); !ok {
			t.Fatalf("try %d refused", i+1)
		}
	}
	sess.Values = maps.Clone(saved)
	if _, ok := takeTwofaAttempt(sess); ok {
		t.Fatal("try past the limit was allowed with the same cookie sent again")
	}
}

func TestAllowTwofaGuessWindow(t *testing.T) {
	const user, other = 42, 43
	t.Cleanup(func() { forgetTwofaGuesses(user); forgetTwofaGuesses(other) })

	start := time.Now()
	for i := 0; i < twofaMaxAttempts; i++ {
		if !allowTwofaGuess(user, start) {
			t.Fatalf("guess %d refused", i+1)
		}
	}
	if allowTwofaGuess(user, start) {
		t.Error("guess past the limit allowed")
	}
	if !allowTwofaGuess(other, start) {
		t.Error("another user's guess refused")
	}
	if !allowTwofaGuess(user, start.Add(twofaPendingTTL)) {
		t.Error("guess refused after the window ended")
	}

	forgetTwofaGuesses(user)
	for i := 0; i < twofaMaxAttempts; i++ {
		if !allowTwofaGuess(user, start) {
			t.Fatalf("guess %d refused after a successful login reset the count", i+1)
		}
	}
}

func TestAPITwofaGuessesShareTheLimit(t *testing.T) {
	const user = 44
	t.Cleanup(func() { forgetTwofaGuesses(user) })

	// a few tries on the login form first, the API must keep counting from there
	for i := 0; i < 2; i++ {
		if !allowTwofaGuess(user, time.Now()) {
			t.Fatalf("form guess %d refused", i+1)
		}
	}
	for i := 2; i < twofaMaxAttempts; i++ {
		if status, code := checkAPITwofa(context.Background(), nil, user, "000000", ""); status != http.StatusUnauthorized || code != "invalid_2fa" {
			t.Fatalf("API guess %d = %d %q, want 401 invalid_2fa", i+1, status, code)
		}
	}
	if status, code := checkAPITwofa(context.Background(), nil, user, "000000", ""); status != http.StatusTooManyRequests || code != "too_many_2fa_attempts" {
		t.Fatalf("API guess past the limit = %d %q, want 429 too_many_2fa_attempts", status, code)
	}
}
