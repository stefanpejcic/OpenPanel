package account

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/sessions"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

const (
	twofaPendingTTL      = 5 * time.Minute
	twofaMaxAttempts     = 5
	twofaUsedCodeTTL     = 90 * time.Second
	sessTwofaPendingUser = "twofa_pending_user"
	sessTwofaPendingAt   = "twofa_pending_at"
	sessTwofaAttempts    = "twofa_attempts"
)

// setTwofaPending marks that userID passed the password step, the 2FA step only trusts this and never the form's user_id
func setTwofaPending(sess *sessions.Session, userID int) {
	sess.Values[sessTwofaPendingUser] = userID
	sess.Values[sessTwofaPendingAt] = time.Now().Unix()
	sess.Values[sessTwofaAttempts] = 0
}

func clearTwofaPending(sess *sessions.Session) {
	delete(sess.Values, sessTwofaPendingUser)
	delete(sess.Values, sessTwofaPendingAt)
	delete(sess.Values, sessTwofaAttempts)
}

// takeTwofaAttempt returns the pending user and counts one attempt, ok is false when there's no fresh pending login or it ran out of attempts
func takeTwofaAttempt(sess *sessions.Session) (int, bool) {
	userID, _ := sess.Values[sessTwofaPendingUser].(int)
	at, _ := sess.Values[sessTwofaPendingAt].(int64)
	attempts, _ := sess.Values[sessTwofaAttempts].(int)
	if userID <= 0 || time.Since(time.Unix(at, 0)) > twofaPendingTTL || attempts >= twofaMaxAttempts {
		clearTwofaPending(sess)
		return 0, false
	}
	sess.Values[sessTwofaAttempts] = attempts + 1
	return userID, true
}

// claimTOTPCode makes a valid code single-use per user, it fails open if Redis is down
func claimTOTPCode(ctx context.Context, a *appctx.App, userID int, code string) bool {
	ok, err := a.Cache.Raw().SetNX(ctx, fmt.Sprintf("twofa_used:%d:%s", userID, code), 1, twofaUsedCodeTTL).Result()
	return err != nil || ok
}
