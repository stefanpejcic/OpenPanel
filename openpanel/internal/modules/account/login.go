// Package account implements the panel's own login form, admin
// autologin/impersonation, and 2FA challenge. CAPTCHA and the login
// notification email are deferred - both are separate modules this one
// merely calls into; their absence doesn't change whether login itself
// works.
package account

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/pquerna/otp/totp"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const isoLayout = "2006-01-02T15:04:05.999999"

var loginPage = web.MustLoadPage("user/_login.html", "user/login.html")

var validAutologinUsername = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// Register wires /login, /login_autologin and /logout onto mux. These
// routes are always available regardless of which optional modules are
// enabled.
func Register(mux *http.ServeMux, a *appctx.App) {
	limiter := newLoginRateLimiter(a)

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if !limiter.Allow(reqip.ClientIP(r)) {
				handleRateLimitExceeded(a, w, r)
				return
			}
		}
		handleLogin(a, w, r)
	})
	mux.HandleFunc("/login_autologin", func(w http.ResponseWriter, r *http.Request) {
		handleAutologin(a, w, r)
	})
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		handleLogout(a, w, r)
	})
}

type localeOption struct {
	Code     string
	FlagCode string
}

// flagOverrides maps locale codes to a different flag image code, for
// locales whose ISO code doesn't match the corresponding country flag.
var flagOverrides = map[string]string{"en": "gb", "zh": "cn", "uk": "ua"}

func localeOptions(codes []string) []localeOption {
	opts := make([]localeOption, len(codes))
	for i, c := range codes {
		flag := strings.ToLower(c)
		if primary, _, ok := strings.Cut(flag, "_"); ok {
			flag = primary
		}
		if override, ok := flagOverrides[flag]; ok {
			flag = override
		}
		opts[i] = localeOption{Code: c, FlagCode: flag}
	}
	return opts
}

type loginPageData struct {
	Title             string
	BrandName         string
	Logo              string
	Favicon           string
	CSRFToken         string
	Locales           []localeOption
	PasswordReset     string
	IsEnterprise      bool
	TwofaEnabled      bool
	UserID            int
	Username          string
	ErrorMessage      string
	UnrecognizedError bool
	PasswordError     bool
	FirstFlash        *flash.Message
	T                 i18n.Translator

	// Email and PasswordStrength are only used by the reset_password pages
	// (forgotpassword.go), which reuse this struct rather than defining
	// their own near-identical one.
	Email            string
	PasswordStrength int
}

func basePageData(a *appctx.App, r *http.Request, t i18n.Translator) loginPageData {
	return loginPageData{
		Title:         "Login",
		BrandName:     a.Config.Get("brand_name", ""),
		Logo:          a.Config.Get("logo", ""),
		PasswordReset: a.Config.Get("password_reset", "yes"),
		IsEnterprise:  strings.HasPrefix(a.LicenseKey, "enterprise"),
		CSRFToken:     csrf.Token(r),
		T:             t,
	}
}

func handleLogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	locale := resolveLocale(a, r, sess)
	t := a.I18n.Translator(locale)

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		switch {
		case r.Form.Get("username") != "":
			handleLoginPassword(a, w, r, sess, t)
			return
		case r.Form.Get("twofa_code") != "":
			handleLoginTwofa(a, w, r, sess, t)
			return
		case r.Form.Get("locale") != "":
			sess.Values["locale"] = r.Form.Get("locale")
			_ = a.Sessions.Save(r, w, sess)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
	}

	data := basePageData(a, r, t)
	data.Locales = localeOptions(a.I18n.AvailableLocales(r.Context()))
	if messages := flash.Pop(a.Sessions, w, r, sess); len(messages) > 0 {
		data.FirstFlash = &messages[0]
	}
	renderLogin(w, http.StatusOK, data)
}

func handleLoginPassword(a *appctx.App, w http.ResponseWriter, r *http.Request, sess *sessions.Session, t i18n.Translator) {
	username := strings.TrimSpace(r.Form.Get("username"))
	password := r.Form.Get("password")

	data := basePageData(a, r, t)
	data.Username = username

	if username == "" || password == "" {
		data.ErrorMessage = t.Get("Username and password are required.")
		renderLogin(w, http.StatusOK, data)
		return
	}

	result, errMsg := verifyPassword(a, r.Context(), username, password, t)
	if errMsg != "" {
		data.ErrorMessage = errMsg
		data.UnrecognizedError = strings.Contains(errMsg, "Unrecognized")
		data.PasswordError = strings.Contains(strings.ToLower(errMsg), "password")
		renderLogin(w, http.StatusOK, data)
		return
	}

	if result.TwofaEnabled {
		data.TwofaEnabled = true
		data.UserID = result.UserID
		renderLogin(w, http.StatusOK, data)
		return
	}

	completeLogin(a, w, r, sess, result.UserID, result.Username, "logged in with a password", "password")
}

func handleLoginTwofa(a *appctx.App, w http.ResponseWriter, r *http.Request, sess *sessions.Session, t i18n.Translator) {
	userIDStr := r.Form.Get("user_id")
	code := r.Form.Get("twofa_code")
	userID, _ := strconv.Atoi(userIDStr)

	data := basePageData(a, r, t)
	data.TwofaEnabled = true
	data.UserID = userID

	var (
		username string
		otp      sql.NullString // otp_secret is NULL until 2FA is set up
	)
	row := a.DB.QueryRowContext(r.Context(), "SELECT username, otp_secret FROM users WHERE id = ?", userID)
	if err := row.Scan(&username, &otp); err != nil {
		data.ErrorMessage = t.Get("Invalid 2FA code. Please try again.")
		renderLogin(w, http.StatusOK, data)
		return
	}

	if !totp.Validate(code, otp.String) {
		data.ErrorMessage = t.Get("Invalid 2FA code. Please try again.")
		renderLogin(w, http.StatusOK, data)
		return
	}

	completeLogin(a, w, r, sess, userID, username, "logged in with 2FA code", "2fa")
}

// loginNotifyMessage builds the "New login to OpenPanel" notification body
// for each of the three login flows (password, 2FA, passkey - see
// passkeys.go), passed to checkIfUserShouldBeNotified.
func loginNotifyMessage(kind, ip string) string {
	switch kind {
	case "password":
		return "New login to OpenPanel\n New login from IP: <a href='https://www.abuseipdb.com/check/" + ip + "' target='_blank'>" + ip + "</a>"
	case "2fa":
		return "New login to OpenPanel\n New 2FA login from IP " + ip
	case "passkey":
		return "New login to OpenPanel\n New passkey login from IP " + ip
	default:
		return "New login to OpenPanel\n New login from IP " + ip
	}
}

// completeLogin establishes the session, logs/notifies, and redirects to
// the dashboard - the tail shared by all three login flows (password,
// 2FA, and the browser-redirect form of passkey login). The fetch()-based
// passkey completion endpoint in passkeys.go needs the same session/log/
// notify work but a JSON response instead of a redirect, so that shared
// part lives in finishLoginSession below.
func completeLogin(a *appctx.App, w http.ResponseWriter, r *http.Request, sess *sessions.Session, userID int, username, action, notifyKind string) {
	finishLoginSession(a, w, r, sess, userID, username, action, notifyKind)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// finishLoginSession is the common tail shared by all three success paths:
// establish the session, append to .lastlogin, record the activity log
// entry, clear rate-limit state, and fire the "new login" notification
// email. Does not write any HTTP response - callers do that themselves
// (a redirect for password/2FA, a JSON body for passkey login).
func finishLoginSession(a *appctx.App, w http.ResponseWriter, r *http.Request, sess *sessions.Session, userID int, username, action, notifyKind string) {
	ip := reqip.ClientIP(r)

	sess.Values["user_id"] = userID
	sess.Values["user_ip"] = ip
	if err := logUserLogin(a, r, sess, userID, username, ip); err != nil {
		log.Printf("LOGIN - failed to record login session: %v", err)
	}
	if err := a.Sessions.Save(r, w, sess); err != nil {
		log.Printf("LOGIN - failed to save session: %v", err)
	}

	_ = logger.RecordUserAction(a.Config, username, action, ip)
	clearFailedAttempts(ip)
	checkIfUserShouldBeNotified(a, r.Context(), userID, username, "notify_account_login", loginNotifyMessage(notifyKind, ip))
}

// logUserLogin appends to the user's .lastlogin history file, then creates
// the Redis-backed session record RequireLogin validates on every
// subsequent request.
func logUserLogin(a *appctx.App, r *http.Request, sess *sessions.Session, userID int, username, ip string) error {
	userFolder := filepath.Join("/etc/openpanel/openpanel/core/users", username)
	if err := os.MkdirAll(userFolder, 0o755); err != nil {
		return err
	}
	lastLoginFile := filepath.Join(userFolder, ".lastlogin")

	var records []string
	if data, err := os.ReadFile(lastLoginFile); err == nil {
		records = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}

	country := getCountryCode(a, r.Context(), ip)
	records = append(records, fmt.Sprintf("IP: %s - Country: %s - Login Time: %s", ip, country, time.Now().Format("2006-01-02 15:04:05")))

	maxRecords := atoiDefault(a.Config.Get("max_login_records", ""), 20)
	if len(records) > maxRecords {
		records = records[len(records)-maxRecords:]
	}
	if err := os.WriteFile(lastLoginFile, []byte(strings.Join(records, "\n")+"\n"), 0o644); err != nil {
		return err
	}

	sessionToken := uuid.NewString()
	sessionKey := fmt.Sprintf("session:%d:%s", userID, sessionToken)
	data := map[string]any{
		"ip_address": ip,
		"created_at": time.Now().Format(isoLayout),
		"username":   username,
	}

	ctx := r.Context()
	if err := a.Cache.Raw().HSet(ctx, sessionKey, data).Err(); err != nil {
		log.Printf("LOGIN - Failed to write to Redis: %v", err)
		return nil // log and continue - a Redis hiccup shouldn't fail the login
	}
	a.Cache.Raw().Expire(ctx, sessionKey, a.SessionDuration)

	sess.Values["session_token"] = sessionToken
	return nil
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// getCountryCode looks up the country for an IP via a public geolocation
// API, cached for 6 minutes to avoid hitting it on every login.
func getCountryCode(a *appctx.App, ctx context.Context, ip string) string {
	code, _ := cache.Memoize(ctx, a.Cache, "get_country_code:"+ip, 6*time.Minute, func() (string, error) {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("https://api.country.is/" + ip)
		if err != nil {
			return "UNKNOWN", nil //nolint:nilerr // request failure shouldn't fail the login; fall back to UNKNOWN
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "UNKNOWN", nil
		}
		var body struct {
			Country string `json:"country"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.Country == "" {
			return "UNKNOWN", nil
		}
		return body.Country, nil
	})
	return code
}

func handleAutologin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	locale := resolveLocale(a, r, sess)
	t := a.I18n.Translator(locale)
	data := basePageData(a, r, t)

	_ = r.ParseForm()

	adminToken := firstNonEmpty(r.Form.Get("admin_token"), r.URL.Query().Get("admin_token"))
	usernameToLogin := firstNonEmpty(r.Form.Get("username"), r.URL.Query().Get("username"))
	impersonate := firstNonEmpty(r.Form.Get("impersonate"), r.URL.Query().Get("impersonate"))
	adminPort := firstNonEmpty(r.Form.Get("admin_port"), r.URL.Query().Get("admin_port"))

	if usernameToLogin == "" || !validAutologinUsername.MatchString(usernameToLogin) {
		data.ErrorMessage = t.Get("Autologin failed: Invalid username.")
		renderLogin(w, http.StatusOK, data)
		return
	}

	tokenPath := fmt.Sprintf("/etc/openpanel/openpanel/core/users/%s/logintoken.txt", usernameToLogin)
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		data.ErrorMessage = t.Get("Error reading token from user files.")
		renderLogin(w, http.StatusOK, data)
		return
	}
	loginToken := strings.TrimSpace(string(tokenBytes))

	if adminToken == "" || !hmac.Equal([]byte(adminToken), []byte(loginToken)) {
		data.ErrorMessage = t.Get("Autologin failed: Invalid or expired token.")
		renderLogin(w, http.StatusOK, data)
		return
	}

	var userID int
	row := a.DB.QueryRowContext(r.Context(), "SELECT id FROM users WHERE username = ?", usernameToLogin)
	if err := row.Scan(&userID); err != nil {
		data.ErrorMessage = t.Get("Autologin failed: User does not exist.")
		renderLogin(w, http.StatusOK, data)
		return
	}

	sess.Values["user_id"] = userID
	if impersonate != "" {
		sess.Values["impersonate"] = impersonate
		sess.Values["admin_port"] = adminPort
	} else {
		delete(sess.Values, "impersonate")
	}

	ip := reqip.ClientIP(r)
	loginUsername := usernameToLogin
	if rest, ok := strings.CutPrefix(loginUsername, "SUSPENDED_"); ok {
		loginUsername = rest
	} else if rest, ok := strings.CutPrefix(loginUsername, "suspended_"); ok {
		loginUsername = rest
	}

	_ = os.Remove(tokenPath)
	if err := logUserLogin(a, r, sess, userID, loginUsername, ip); err != nil {
		log.Printf("LOGIN - failed to record autologin session: %v", err)
	}
	_ = a.Sessions.Save(r, w, sess)
	_ = logger.RecordUserAction(a.Config, loginUsername, "logged in via admin API", ip)

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// handleLogout only clears the browser-side session cookie - the
// server-side Redis session record is left to expire on its own TTL
// rather than being deleted immediately.
func handleLogout(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	sess, _ := a.Sessions.Get(r, session.CookieName)

	var username string
	if userID, ok := auth.UserID(r); ok && userID != 0 {
		if data, err := a.InjectData(r.Context(), userID); err == nil {
			username, _ = data["current_username"].(string)
		}
	}
	log.Printf("LOGOUT - /logout opened, terminating session for user: %s", username)
	if username != "" {
		_ = logger.RecordUserAction(a.Config, username, "logged out.", reqip.ClientIP(r))
	}

	for k := range sess.Values {
		delete(sess.Values, k)
	}
	_ = a.Sessions.Save(r, w, sess)

	http.Redirect(w, r, a.Config.Get("logout_redirect", "/login"), http.StatusFound)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func renderLogin(w http.ResponseWriter, status int, data loginPageData) {
	if err := loginPage.Render(w, status, data); err != nil {
		log.Printf("LOGIN - template render error: %v", err)
	}
}

// resolveLocale determines the locale for the pre-auth login page: session
// value first, then Accept-Language/default. Login is reachable pre-auth,
// so there's no per-account locale file to check yet.
func resolveLocale(a *appctx.App, r *http.Request, sess *sessions.Session) string {
	sessionLocale, _ := sess.Values["locale"].(string)
	return a.I18n.ResolveLocale(r.Context(), sessionLocale, "", r.Header.Get("Accept-Language"))
}
