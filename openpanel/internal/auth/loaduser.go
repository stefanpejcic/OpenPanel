package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
)

const (
	jwtAlgorithm = "HS256"
	jwtSubClaim  = "sub"
	// JWTPwdClaim ties a token to the password it was issued under, so a password change revokes it
	JWTPwdClaim = "pwd"
)

// LoadUser resolves the caller's identity from a Bearer token (MCP or JWT), falling back to the session cookie, then enforces the configured access domain
func LoadUser(a *appctx.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = resolveIdentity(a, r)

			if redirected := enforceAccessDomain(a, w, r); redirected {
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func resolveIdentity(a *appctx.App, r *http.Request) *http.Request {
	authHeader := r.Header.Get("Authorization")

	if token, ok := strings.CutPrefix(authHeader, "Bearer "); ok {
		token = strings.TrimSpace(token)

		if strings.HasPrefix(token, mcptokens.TokenPrefix) {
			if a.DB != nil {
				if result, err := mcptokens.Authenticate(a.DB, token); err == nil && result != nil {
					r = withAuth(r, result.UserID, ViaMCP)
					r = withMCPReadOnly(r, result.ReadOnly)
					return r
				}
			}
			log.Print("APP - MCP token auth failed: token is invalid, revoked, or expired.")
			return withAuth(r, 0, ViaNone)
		}

		userID, pwd, err := parseJWT(token, a.SecretKey)
		if err != nil {
			log.Printf("APP - JWT auth failed: %v", err)
			return withAuth(r, 0, ViaNone)
		}
		return withJWTPwd(withAuth(r, userID, ViaJWT), pwd)
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)
	userID, _ := sess.Values["user_id"].(int)
	token, _ := sess.Values["session_token"].(string)
	r = withAuth(r, userID, ViaSession)
	return withSessionToken(r, token)
}

func parseJWT(tokenString string, secret []byte) (int, string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwtAlgorithm}), jwt.WithExpirationRequired())
	if err != nil {
		return 0, "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", jwt.ErrTokenInvalidClaims
	}
	sub, ok := claims[jwtSubClaim]
	if !ok {
		return 0, "", jwt.ErrTokenInvalidClaims
	}
	pwd, _ := claims[JWTPwdClaim].(string)

	switch v := sub.(type) {
	case string:
		id, err := strconv.Atoi(v)
		return id, pwd, err
	case float64:
		return int(v), pwd, nil
	default:
		return 0, "", jwt.ErrTokenInvalidClaims
	}
}

// PasswordFingerprint is a keyed hash of the user's current password hash, it changes whenever the password does and leaks nothing about it
func PasswordFingerprint(ctx context.Context, a *appctx.App, userID int) (string, error) {
	var pwHash string
	if err := a.DB.QueryRowContext(ctx, "SELECT password FROM users WHERE id = ?", userID).Scan(&pwHash); err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, a.SecretKey)
	mac.Write([]byte("api-token:" + pwHash))
	return hex.EncodeToString(mac.Sum(nil))[:32], nil
}

// enforceAccessDomain redirects to the configured access domain if the request's Host doesn't match, returning true if it wrote a redirect - skipped entirely when ForceDomain is empty
func enforceAccessDomain(a *appctx.App, w http.ResponseWriter, r *http.Request) bool {
	desiredDomain := strings.TrimSpace(a.ForceDomain)
	if desiredDomain == "" {
		return false
	}

	requestedHost, requestedPort := r.Host, ""
	if h, p, err := net.SplitHostPort(requestedHost); err == nil {
		requestedHost, requestedPort = h, p
	}

	if requestedHost == desiredDomain {
		return false
	}

	via := Via(r)
	if r.URL.Path == "/login" || r.URL.Path == "/login_autologin" || via == ViaJWT || via == ViaMCP {
		return false
	}

	scheme := "http"
	if sysinfo.HasSSL(r.Context(), a.Cache, desiredDomain) {
		scheme = "https"
	}

	// a.ForcePort can fail to resolve, so fall back to the port this request came in on rather than dropping it - see dynamicdns.publicBaseURL for the same trick
	portSuffix := ""
	switch {
	case a.ForcePort != "":
		portSuffix = ":" + a.ForcePort
	case requestedPort != "":
		portSuffix = ":" + requestedPort
	}

	target := scheme + "://" + desiredDomain + portSuffix + r.URL.RequestURI()
	log.Printf("APP - The requested domain does not match the configured domain for panel access, redirecting to: %s", target)
	http.Redirect(w, r, target, http.StatusMovedPermanently)
	return true
}
