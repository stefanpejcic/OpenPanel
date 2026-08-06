package account

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

const jwtExpSeconds = 3600 // API token lifetime: 1h

// RegisterAPILogin wires POST /api/login onto mux. It lives in the account
// package (rather than a generic api package) because it reuses
// verifyPassword/logUserLogin/checkIfUserShouldBeNotified/
// clearFailedAttempts directly rather than duplicating that logic.
func RegisterAPILogin(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Add("POST /api/login")
	limiter := newLoginRateLimiter(a)
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow(reqip.ClientIP(r)) {
			handleRateLimitExceeded(a, w, r)
			return
		}
		handleAPILogin(a, w, r)
	})
}

func writeAPIJSON(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// handleAPILogin authenticates a username/password (and optional 2FA code)
// request and, on success, issues a signed JWT for subsequent API calls.
func handleAPILogin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, _ := a.Sessions.Get(r, session.CookieName)
	t := a.I18n.Translator(resolveLocale(a, r, sess))

	var body struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		TwofaCode string `json:"twofa_code"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if body.Username == "" || body.Password == "" {
		writeAPIJSON(w, http.StatusBadRequest, map[string]any{"error": t.Get("Username and password are required.")})
		return
	}

	result, errMsg := verifyPassword(a, ctx, body.Username, body.Password, t)
	if errMsg != "" {
		writeAPIJSON(w, http.StatusUnauthorized, map[string]any{"error": apiLoginErrorCode(errMsg)})
		return
	}

	if result.TwofaEnabled {
		var otp sql.NullString
		row := a.DB.QueryRowContext(ctx, "SELECT otp_secret FROM users WHERE id = ?", result.UserID)
		_ = row.Scan(&otp)

		if body.TwofaCode == "" {
			writeAPIJSON(w, http.StatusUnauthorized, map[string]any{"twofa_required": true, "user_id": result.UserID})
			return
		}
		if !totp.Validate(body.TwofaCode, otp.String) {
			writeAPIJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid_2fa"})
			return
		}
	}

	ip := reqip.ClientIP(r)
	if err := logUserLogin(a, r, sess, result.UserID, result.Username, ip); err != nil {
		writeAPIJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal_error"})
		return
	}
	_ = a.Sessions.Save(r, w, sess)
	_ = logger.RecordUserAction(a.Config, result.Username, "logged in via user API", ip)
	clearFailedAttempts(ip)
	checkIfUserShouldBeNotified(a, ctx, result.UserID, result.Username, "notify_account_login",
		loginNotifyMessage("password", ip))

	claims := jwt.MapClaims{
		"sub": strconv.Itoa(result.UserID),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(jwtExpSeconds * time.Second).Unix(),
	}
	token, signErr := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.SecretKey)
	if signErr != nil {
		writeAPIJSON(w, http.StatusInternalServerError, map[string]any{"error": "token_generation_failed"})
		return
	}

	writeAPIJSON(w, http.StatusOK, map[string]any{
		"token": token, "user_id": result.UserID, "expires_in": jwtExpSeconds,
	})
}

// apiLoginErrorCode maps verifyPassword's translated error message back to
// a stable machine-readable code for API clients. verifyPassword returns a
// single translated string (shared with the session-login path), so this
// pattern-matches on its content to recover a specific error category.
func apiLoginErrorCode(errMsg string) string {
	switch {
	case strings.Contains(errMsg, "suspended"):
		return "suspended"
	case strings.Contains(errMsg, "Unrecognized"):
		return "unknown_user"
	case strings.Contains(errMsg, "Unable to connect"):
		return "db_error"
	case strings.Contains(errMsg, "Invalid password"):
		return "invalid_password"
	default:
		return "login_failed"
	}
}
