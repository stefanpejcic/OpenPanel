package account

import (
	"maps"
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
