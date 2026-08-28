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
	"errors"
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
			if errors.Is(err, errSMTPNotConfigured) {
				data := basePageData(a, r, t)
				data.Title = t.Get("Forgot Password")
				data.Locales = localeOptions(a.I18n.AvailableLocales(r.Context()))
				data.Email = email
				data.ErrorMessage = t.Get("Email sending is not configured. Please contact your administrator.")
				renderPage(w, requestEmailPage, http.StatusOK, data)
				return
			}
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

// errSMTPNotConfigured is returned when openpanel.config has no [SMTP]
// mail_server set - the same section OpenAdmin's notification settings page
// (Settings > Notifications) writes to. handleForgotPassword surfaces this
// to the user instead of silently pretending the email was sent.
var errSMTPNotConfigured = errors.New("smtp is not configured")

// sendPasswordResetEmail stashes the token in Redis (default TTL, 300s/5m -
// shorter than the token's own 15-minute signature window and shorter than
// what the email tells the user, so the link silently stops working after
// 5 minutes; a known quirk, not something to "fix" here), then emails the
// reset link via the same SMTP relay OpenAdmin's own notification emails
// use (see loadSMTPConfig).
func sendPasswordResetEmail(a *appctx.App, ctx context.Context, t i18n.Translator, userID int, username, email string) error {
	cfg, ok := loadSMTPConfig(a)
	if !ok {
		return errSMTPNotConfigured
	}

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
		"Title":        t.Get("Password Reset Requested"),
		"ResetLink":    resetLink,
		"Hostname":     forceDomain,
		"LoginURL":     loginURL,
		"IsEnterprise": strings.HasPrefix(a.LicenseKey, "enterprise"),
		"T":            t,
	}); err != nil {
		return err
	}

	subject := t.Get("OpenPanel [%(domain)s] - Password Reset Requested", "domain", forceDomain)
	return sendMail(cfg, email, subject, buf.String())
}

// smtpConfig is the [SMTP] section of openpanel.config, as configured via
// OpenAdmin's Settings > Notifications page - the very same keys OpenAdmin's
// own mailer (openadmin/internal/handlers/mailer.go) reads, so a single set
// of SMTP credentials/relay serves both apps' outgoing mail.
type smtpConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	UseSSL   bool
	UseTLS   bool
}

// loadSMTPConfig reads openpanel.config for mail_server et al. openpanel's
// Config.Load has no notion of ini [section] headers, but its key=value
// parser still picks these up regardless - a "[SMTP]" line simply doesn't
// match its key=value regex and is skipped, so no separate config plumbing
// is needed to share these values with OpenAdmin. Returns ok=false when no
// mail_server is set, i.e. nobody has configured SMTP yet.
func loadSMTPConfig(a *appctx.App) (smtpConfig, bool) {
	host := a.Config.Get("mail_server", "")
	if host == "" {
		return smtpConfig{}, false
	}
	username := a.Config.Get("mail_username", "")
	from := a.Config.Get("mail_default_sender", "")
	if from == "" {
		from = username
	}
	if from == "" {
		return smtpConfig{}, false
	}
	return smtpConfig{
		Host:     host,
		Port:     a.Config.Get("mail_port", "465"),
		Username: username,
		Password: a.Config.Get("mail_password", ""),
		From:     from,
		UseSSL:   strings.EqualFold(a.Config.Get("mail_use_ssl", ""), "true"),
		UseTLS:   strings.EqualFold(a.Config.Get("mail_use_tls", ""), "true"),
	}, true
}

// sendMail sends an HTML email over cfg's relay: implicit TLS when UseSSL,
// STARTTLS when UseTLS (and the server advertises it), plaintext otherwise -
// mirroring OpenAdmin's own mailerSendRun so both apps behave the same way
// for the same [SMTP] settings.
func sendMail(cfg smtpConfig, to, subject, htmlBody string) error {
	msg := "From: " + cfg.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		htmlBody

	addr := cfg.Host + ":" + cfg.Port

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	var client *smtp.Client
	var err error
	if cfg.UseSSL {
		var conn *tls.Conn
		conn, err = tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host, MinVersion: tls.VersionTLS12})
		if err != nil {
			return err
		}
		client, err = smtp.NewClient(conn, cfg.Host)
	} else {
		client, err = smtp.Dial(addr)
	}
	if err != nil {
		return err
	}
	defer client.Close()

	if !cfg.UseSSL && cfg.UseTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: cfg.Host}); err != nil {
				return err
			}
		}
	}

	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err := client.Mail(cfg.From); err != nil {
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
