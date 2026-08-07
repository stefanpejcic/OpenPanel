package auth

import (
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
)

// LoadUser resolves the caller's identity from a Bearer token (MCP token
// or JWT) or, failing that, the session cookie, then enforces the
// configured access domain.
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

		userID, err := parseJWT(token, a.SecretKey)
		if err != nil {
			log.Printf("APP - JWT auth failed: %v", err)
			return withAuth(r, 0, ViaNone)
		}
		return withAuth(r, userID, ViaJWT)
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)
	userID, _ := sess.Values["user_id"].(int)
	token, _ := sess.Values["session_token"].(string)
	r = withAuth(r, userID, ViaSession)
	return withSessionToken(r, token)
}

func parseJWT(tokenString string, secret []byte) (int, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwtAlgorithm}))
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}
	sub, ok := claims[jwtSubClaim]
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}

	switch v := sub.(type) {
	case string:
		return strconv.Atoi(v)
	case float64:
		return int(v), nil
	default:
		return 0, jwt.ErrTokenInvalidClaims
	}
}

// enforceAccessDomain redirects the request to the configured access
// domain when the request's Host doesn't match it. Returns true if it
// already wrote a redirect response.
//
// If ForceDomain is empty (e.g. the `opencli domain` lookup failed or
// isn't installed, which is the normal case outside a real panel
// install), the redirect is skipped entirely: a real deployment always
// has opencli configured with a domain, so an empty ForceDomain only
// happens in environments where domain enforcement isn't meaningful
// anyway.
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

	// a.ForcePort comes from `opencli port`, which can fail to resolve in
	// some deployments; when that happens, fall back to the port this
	// very request came in on rather than omitting the port entirely -
	// see dynamicdns.publicBaseURL for the same fallback and why.
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
