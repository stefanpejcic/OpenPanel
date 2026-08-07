package account

import (
	"crypto/rand"
	"encoding/base32"
	"net/http"
	"strconv"
	"strings"

	"github.com/pquerna/otp/totp"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

func invalidate2FAStatus(a *appctx.App, r *http.Request, userID int) {
	_ = a.Cache.Delete(r.Context(), "get_2fa_status_for_user:"+strconv.Itoa(userID))
}

// randomBase32Secret generates a TOTP secret. Deliberately stronger than
// a common TOTP library default (16 base32 chars / 10 random bytes / 80
// bits): some client-side TOTP libraries (e.g. otplib, used by this
// project's Playwright tests) reject anything under the RFC
// 4226-recommended 128-bit minimum, so this uses the RFC's recommended 160
// bits (20 random bytes, encoding to exactly 32 base32 characters with no
// padding needed) instead.
func randomBase32Secret() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// handleTwofaSettings views or updates the user's 2FA enrollment.
func handleTwofaSettings(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	status, err := a.Get2FAStatusForUser(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	twofaEnabled, otpSecret := status.Enabled, status.OTPSecret

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		setupConfirmed := r.Form.Get("setup_confirmed") != ""
		twofaActive := r.Form.Get("twofa_active") != ""

		var message string

		switch {
		case setupConfirmed:
			otpCode := strings.TrimSpace(r.Form.Get("otp_code"))
			if otpSecret == "" || !totp.Validate(otpCode, otpSecret) {
				sess, _ := a.Sessions.Get(r, session.CookieName)
				flash.Add(sess, "error", "Invalid verification code. Please try again.")
				_ = a.Sessions.Save(r, w, sess)
				http.Redirect(w, r, "/account/2fa", http.StatusFound)
				return
			}
			twofaEnabled = true
			message = "OTP code removed and Two-Factor Authentication is now required on new logins."

		case twofaActive:
			otpSecret, err = randomBase32Secret()
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			twofaEnabled = false
			message = "Setup started: Set up the app or save the OTP code, then confirm to remove it and activate 2FA."

		default:
			otpSecret = ""
			twofaEnabled = false
			message = "Two-Factor Authentication is now disabled."
		}

		if _, execErr := a.DB.ExecContext(ctx, "UPDATE users SET twofa_enabled = ?, otp_secret = ? WHERE id = ?",
			twofaEnabled, nullableString(otpSecret), userID); execErr != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		invalidate2FAStatus(a, r, userID)

		data, injectErr := a.InjectData(ctx, userID)
		currentUsername, _ := data["current_username"].(string)
		if injectErr == nil {
			logAction := "disabled 2FA for account"
			if twofaEnabled {
				logAction = "enabled 2FA for account"
			}
			checkIfUserShouldBeNotified(a, ctx, userID, currentUsername, "notify_twofactorauth_change", message)
			_ = logger.RecordUserAction(a.Config, currentUsername, logAction, reqip.ClientIP(r))
		}

		sess, _ := a.Sessions.Get(r, session.CookieName)
		flash.Add(sess, "success", message)
		_ = a.Sessions.Save(r, w, sess)
		http.Redirect(w, r, "/account/2fa", http.StatusFound)
		return
	}

	renderTwofaPage(a, w, r, twofaEnabled, otpSecret)
}

// nullableString maps an empty Go string ("no secret") to SQL NULL rather
// than storing an empty string in the otp_secret column.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// RegisterTwofa wires the 2FA settings route onto mux, gated behind the
// "twofa" feature flag.
func RegisterTwofa(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "twofa")(h)
	}
	mux.Handle("/account/2fa", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleTwofaSettings(a, w, r) }))
}
