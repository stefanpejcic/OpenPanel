// Package account (this file) implements the pre-login "forgot your
// password" email-link flow. Distinct from passwords.go/settings.go's
// self-service password change, which requires being already logged in.
package account

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var (
	requestEmailPage   = web.MustLoadPage("user/_login.html", "user/reset_password/request_email.html")
	linkSentPage       = web.MustLoadPage("user/_login.html", "user/reset_password/link_is_sent.html")
	setNewPasswordPage = web.MustLoadPage("user/_login.html", "user/reset_password/set_new_password.html")
	resetEmailFragment = web.MustLoadFragment("user/email_template.html")
)

// passwordResetTokenTTL: a reset token's signature is only accepted within
// 15 minutes of being minted.
const passwordResetTokenTTL = 15 * time.Minute

// RegisterPasswordReset wires the reset-password routes onto mux, unless
// disabled via the password_reset config value - in which case the routes
// are simply never registered.
func RegisterPasswordReset(mux *http.ServeMux, a *appctx.App) {
	if strings.ToLower(a.Config.Get("password_reset", "yes")) != "yes" {
		return
	}
	mux.HandleFunc("/reset_password", func(w http.ResponseWriter, r *http.Request) {
		handleForgotPassword(a, w, r)
	})
	mux.HandleFunc("/reset_password/{token}", func(w http.ResponseWriter, r *http.Request) {
		handleResetPasswordToken(a, w, r)
	})
}

// signResetToken produces a tamper-evident, timestamped token embedding the
// user ID: nothing outside this app ever needs to parse it, only
// sign/verify round-trip through this same process, so an HMAC-SHA256
// payload+signature pair is all that's needed (can't be forged or replayed
// past its age).
func signResetToken(secretKey []byte, userID int) string {
	payload := fmt.Sprintf("%d.%d", userID, time.Now().Unix())
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(payload))
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// resetTokenStatus distinguishes the three outcomes of verifying a reset
// token: valid, expired, or malformed/tampered.
type resetTokenStatus int

const (
	resetTokenValid resetTokenStatus = iota
	resetTokenExpired
	resetTokenInvalid
)

// verifyResetToken validates the HMAC signature and the embedded
// timestamp's age.
func verifyResetToken(secretKey []byte, token string) (userID int, status resetTokenStatus) {
	payloadPart, sigPart, found := strings.Cut(token, ".")
	if !found {
		return 0, resetTokenInvalid
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return 0, resetTokenInvalid
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return 0, resetTokenInvalid
	}

	mac := hmac.New(sha256.New, secretKey)
	mac.Write(payloadBytes)
	if !hmac.Equal(mac.Sum(nil), sigBytes) {
		return 0, resetTokenInvalid
	}

	idPart, tsPart, found := strings.Cut(string(payloadBytes), ".")
	if !found {
		return 0, resetTokenInvalid
	}
	uid, err := strconv.Atoi(idPart)
	if err != nil {
		return 0, resetTokenInvalid
	}
	ts, err := strconv.ParseInt(tsPart, 10, 64)
	if err != nil {
		return 0, resetTokenInvalid
	}
	if time.Since(time.Unix(ts, 0)) > passwordResetTokenTTL {
		return uid, resetTokenExpired
	}
	return uid, resetTokenValid
}

// resetTokenCacheKey is the cache keyspace for a pending reset token.
func resetTokenCacheKey(token string) string {
	return cache.KeyPrefix + "reset_token:" + token
}

func handleForgotPassword(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	locale := resolveLocale(a, r, sess)
	t := a.I18n.Translator(locale)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		email := r.Form.Get("email")

		var userID int
		var username string
		row := a.DB.QueryRowContext(r.Context(), "SELECT id, username FROM users WHERE email = ?", email)
		if err := row.Scan(&userID, &username); err != nil {
			data := basePageData(a, r, t)
			data.Title = t.Get("Forgot Password")
			data.Locales = localeOptions(a.I18n.AvailableLocales(r.Context()))
			data.Email = email
			data.ErrorMessage = t.Get("Email %(email)s not found", "email", email)
			renderPage(w, requestEmailPage, http.StatusOK, data)
			return
		}

		if err := sendPasswordResetEmail(a, r.Context(), t, userID, username, email); err != nil {
			log.Printf("FORGOT_PASSWORD - failed to send reset email to %s: %v", email, err)
		}

		_ = logger.RecordUserAction(a.Config, username, t.Get("password reset requested"), reqip.ClientIP(r))

		data := basePageData(a, r, t)
		data.Title = t.Get("Password Reset Link Sent")
		data.Locales = localeOptions(a.I18n.AvailableLocales(r.Context()))
		renderPage(w, linkSentPage, http.StatusOK, data)
		return
	}

	data := basePageData(a, r, t)
	data.Title = t.Get("Forgot Password")
	data.Locales = localeOptions(a.I18n.AvailableLocales(r.Context()))
	renderPage(w, requestEmailPage, http.StatusOK, data)
}

// sendPasswordResetEmail stashes the token in Redis (default TTL, 300s/5m -
// shorter than the token's own 15-minute signature window and shorter than
// what the email tells the user, so the link silently stops working after
// 5 minutes; a known quirk, not something to "fix" here), then emails the
// reset link via the hardcoded openpanel.org SMTP relay (see smtpDefaults -
// openpanel.config has no nested SMTP section, so these defaults are the
// only values ever used).
func sendPasswordResetEmail(a *appctx.App, ctx context.Context, t i18n.Translator, userID int, username, email string) error {
	token := signResetToken(a.SecretKey, userID)
	if err := a.Cache.Raw().Set(ctx, resetTokenCacheKey(token), userID, cache.DefaultTTL).Err(); err != nil {
		log.Printf("FORGOT_PASSWORD - failed to store reset token in Redis: %v", err)
	}

	forceDomain := sysinfo.GetOpenPanelDomain(ctx, a.Cache)
	protocol, host := "http", sysinfo.FetchPublicIP(ctx, a.Cache)
	if forceDomain != "" && sysinfo.HasSSL(ctx, a.Cache, forceDomain) {
		protocol, host = "https", forceDomain
	}
	base := protocol + "://" + host
	resetLink := base + "/reset_password/" + token
	loginURL := base + "/login"

	var buf bytes.Buffer
	if err := resetEmailFragment.ExecuteTemplate(&buf, "reset_password_email", map[string]any{
		"Title":     t.Get("Password Reset Requested"),
		"ResetLink": resetLink,
		"Hostname":  forceDomain,
		"LoginURL":  loginURL,
		"T":         t,
	}); err != nil {
		return err
	}

	subject := t.Get("OpenPanel [%(domain)s] - Password Reset Requested", "domain", forceDomain)
	return sendMail(email, subject, buf.String())
}

// smtpDefaults are the panel's own hardcoded SMTP relay credentials - see
// sendPasswordResetEmail's comment on why these are effectively the only
// values ever used.
const (
	smtpHost = "mail.openpanel.com"
	smtpPort = "465"
	smtpUser = "no-reply@openpanel.org"
	smtpPass = "96e2ygBDrThxl9zI"
	smtpFrom = "no-reply@openpanel.org"
)

// sendMail sends an HTML email over implicit TLS (port 465).
func sendMail(to, subject, htmlBody string) error {
	msg := "From: " + smtpFrom + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		htmlBody

	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, &tls.Config{ServerName: smtpHost, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Auth(smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)); err != nil {
		return err
	}
	if err := client.Mail(smtpFrom); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	wc, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write([]byte(msg)); err != nil {
		_ = wc.Close()
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func handleResetPasswordToken(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	locale := resolveLocale(a, r, sess)
	t := a.I18n.Translator(locale)
	token := r.PathValue("token")
	ctx := r.Context()

	userID, tokenStatus := verifyResetToken(a.SecretKey, token)
	if tokenStatus != resetTokenValid {
		errMsg := t.Get("Reset link is invalid")
		if tokenStatus == resetTokenExpired {
			errMsg = t.Get("Reset link has expired")
		}
		data := basePageData(a, r, t)
		data.Title = errMsg
		data.Locales = localeOptions(a.I18n.AvailableLocales(ctx))
		data.ErrorMessage = errMsg
		renderPage(w, requestEmailPage, http.StatusOK, data)
		return
	}

	cachedUserID, cacheErr := a.Cache.Raw().Get(ctx, resetTokenCacheKey(token)).Int()
	if cacheErr != nil || cachedUserID != userID {
		errMsg := t.Get("Link has already been used or is invalid")
		data := basePageData(a, r, t)
		data.Title = errMsg
		data.Locales = localeOptions(a.I18n.AvailableLocales(ctx))
		data.ErrorMessage = errMsg
		renderPage(w, requestEmailPage, http.StatusOK, data)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		password := r.Form.Get("password")
		confirmPassword := r.Form.Get("confirm_password")

		if password != confirmPassword {
			errMsg := t.Get("Passwords do not match")
			data := basePageData(a, r, t)
			data.Title = errMsg
			data.Locales = localeOptions(a.I18n.AvailableLocales(ctx))
			data.PasswordStrength = passwordStrengthThreshold(a)
			data.ErrorMessage = errMsg
			renderPage(w, setNewPasswordPage, http.StatusOK, data)
			return
		}

		if !updatePasswordByID(ctx, a, sess, userID, password) {
			errMsg := t.Get("Password does not meet the required strength")
			data := basePageData(a, r, t)
			data.Title = errMsg
			data.Locales = localeOptions(a.I18n.AvailableLocales(ctx))
			data.PasswordStrength = passwordStrengthThreshold(a)
			data.ErrorMessage = errMsg
			renderPage(w, setNewPasswordPage, http.StatusOK, data)
			return
		}

		var username string
		if err := a.DB.QueryRowContext(ctx, "SELECT username FROM users WHERE id = ?", userID).Scan(&username); err != nil {
			username = ""
		}
		_ = logger.RecordUserAction(a.Config, username, t.Get("password changed using email confirmation"), reqip.ClientIP(r))

		a.Cache.Raw().Del(ctx, resetTokenCacheKey(token))

		flash.Add(sess, "success", t.Get("Your password has been reset. Please log in."))
		_ = a.Sessions.Save(r, w, sess)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	data := basePageData(a, r, t)
	data.Title = t.Get("Reset Password")
	data.Locales = localeOptions(a.I18n.AvailableLocales(ctx))
	data.PasswordStrength = passwordStrengthThreshold(a)
	renderPage(w, setNewPasswordPage, http.StatusOK, data)
}

func passwordStrengthThreshold(a *appctx.App) int {
	return validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)
}

func renderPage(w http.ResponseWriter, page *web.Page, status int, data loginPageData) {
	if err := page.Render(w, status, data); err != nil {
		log.Printf("FORGOT_PASSWORD - template render error: %v", err)
	}
}
