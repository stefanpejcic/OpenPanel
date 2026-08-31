package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// isoLayout is the timestamp format used for the "created_at" and
// "last_active" fields stored in the Redis session hash - it must match
// the format internal/modules/account/login.go writes when it creates
// the session ("2024-01-02T15:04:05.678901", fractional seconds omitted
// entirely when zero, which Go's ".999999" layout element also handles).
const isoLayout = "2006-01-02T15:04:05.999999"

// RequireLogin is middleware that enforces an authenticated session -
// checking session existence, max lifetime, IP binding, 2FA enrollment,
// and demo-mode write restrictions - and gates access by feature.
// featureNames are the feature-gate keys checked against the caller's
// enabled features; each route registration passes its own feature
// name(s) explicitly. Access is granted if the caller has any one of
// them - a route shared by two modules (e.g. docker's and services'
// /json/services) passes both so it works whenever either is enabled.
func RequireLogin(a *appctx.App, featureNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID, hasUser := UserID(r)
			token := SessionToken(r)
			if !hasUser || userID == 0 || token == "" {
				log.Printf("APP - %s %s: no active session, redirecting to login", r.Method, r.URL.Path)
				redirectToLogin(w, r)
				return
			}

			sess, _ := a.Sessions.Get(r, session.CookieName)
			sessionKey := fmt.Sprintf("session:%d:%s", userID, token)

			data, err := a.Cache.Raw().HGetAll(ctx, sessionKey).Result()
			if err != nil || len(data) == 0 {
				log.Printf("APP - user %d: session %s not found in cache (expired or invalidated), redirecting to login", userID, sessionKey)
				clearSession(sess)
				_ = a.Sessions.Save(r, w, sess)
				redirectToLogin(w, r)
				return
			}

			if a.ValidateIPAddressCookie {
				currentIP := reqip.ClientIP(r)
				if data["ip_address"] != currentIP {
					log.Printf("APP - user %d: IP mismatch (session=%s, request=%s), invalidating session", userID, data["ip_address"], currentIP)
					a.Cache.Raw().Del(ctx, sessionKey)
					clearSession(sess)
					flash.Add(sess, "danger", a.I18n.Get(a.I18n.SystemDefaultLocale(ctx), "IP address mismatch. Please login again."))
					_ = a.Sessions.Save(r, w, sess)
					redirectToLogin(w, r)
					return
				}
			}

			createdAt, err := time.Parse(isoLayout, data["created_at"])
			if err == nil && time.Since(createdAt) > a.MaxSessionLifetime {
				log.Printf("APP - user %d: session reached max lifetime (%s), expiring", userID, a.MaxSessionLifetime)
				a.Cache.Raw().Del(ctx, sessionKey)
				clearSession(sess)
				_ = a.Sessions.Save(r, w, sess)
				redirectToLogin(w, r)
				return
			}

			refreshLastActive(ctx, a, sessionKey, data["last_active"])

			details, err := a.GetUserDetailsWithPlan(ctx, userID)
			if err != nil {
				log.Printf("APP - user %d: failed to load user details: %v", userID, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if strings.HasPrefix(details.Username, "SUSPENDED_") {
				log.Printf("APP - user %d (%s): account suspended, invalidating session", userID, details.Username)
				a.Cache.Raw().Del(ctx, sessionKey)
				clearSession(sess)
				_ = a.Sessions.Save(r, w, sess)
				redirectToLogin(w, r)
				return
			}

			if a.TwofaEnforce && a.ModuleEnabled("twofa") && !contains(featureNames, "twofa") {
				status, err := a.Get2FAStatusForUser(ctx, userID)
				if err == nil && !status.Enabled {
					log.Printf("APP - user %d (%s): 2FA required but not enabled, redirecting to enrollment", userID, details.Username)
					flash.Add(sess, "warning", a.I18n.Get(a.I18n.SystemDefaultLocale(ctx),
						"Two-Factor Authentication is required on this panel. Please enable it to continue."))
					_ = a.Sessions.Save(r, w, sess)
					http.Redirect(w, r, "/account/2fa", http.StatusFound)
					return
				}
			}

			userFeatures, err := a.LoadUserFeatures(ctx, details.Username, details.Context)
			if err != nil {
				log.Printf("APP - user %d (%s): failed to load enabled features: %v", userID, details.Username, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if !containsAny(userFeatures, featureNames) {
				log.Printf("APP - user %d (%s): access denied to %s (none of %q enabled for this account)", userID, details.Username, r.URL.Path, featureNames)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if r.Method != http.MethodGet && a.DemoMode {
				log.Printf("APP - user %d (%s): write blocked by demo mode (%s %s)", userID, details.Username, r.Method, r.URL.Path)
				flash.Add(sess, "warning", a.I18n.Get(a.I18n.SystemDefaultLocale(ctx), "Disabled in demo mode."))
				_ = a.Sessions.Save(r, w, sess)
				referrer := r.Referer()
				if referrer == "" {
					referrer = "/dashboard"
				}
				http.Redirect(w, r, referrer, http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func refreshLastActive(ctx context.Context, a *appctx.App, sessionKey, lastActiveRaw string) {
	needsRefresh := lastActiveRaw == ""
	if !needsRefresh {
		if lastActive, err := time.Parse(isoLayout, lastActiveRaw); err == nil {
			needsRefresh = time.Since(lastActive) > 60*time.Second
		}
	}
	if !needsRefresh {
		return
	}

	pipe := a.Cache.Raw().Pipeline()
	pipe.HSet(ctx, sessionKey, "last_active", time.Now().Format(isoLayout))
	pipe.Expire(ctx, sessionKey, a.SessionDuration)
	_, _ = pipe.Exec(ctx)
}

// clearSession empties the session's values map. It does not save -
// callers add any flash message after calling this and save once at the
// end, so the flash message survives the clear.
func clearSession(sess *sessions.Session) {
	for k := range sess.Values {
		delete(sess.Values, k)
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

func containsAny(list []string, wants []string) bool {
	for _, want := range wants {
		if contains(list, want) {
			return true
		}
	}
	return false
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

// RequireAPI is middleware for API routes: Bearer-token auth only (JWT or
// MCP token), feature-gated the same way as RequireLogin.
func RequireAPI(a *appctx.App, featureName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID, hasUser := UserID(r)
			via := Via(r)

			if (via != ViaJWT && via != ViaMCP) || !hasUser || userID == 0 {
				log.Printf("APP - API %s %s: missing or invalid bearer token (via=%v)", r.Method, r.URL.Path, via)
				writeJSONError(w, http.StatusUnauthorized, "Authentication required", "Provide a valid Bearer token")
				return
			}

			details, err := a.GetUserDetailsWithPlan(ctx, userID)
			if err != nil {
				log.Printf("APP - API user %d: failed to load user details: %v", userID, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			userFeatures, err := a.LoadUserFeatures(ctx, details.Username, details.Context)
			if err != nil {
				log.Printf("APP - API user %d (%s): failed to load enabled features: %v", userID, details.Username, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			if featureName != "api" && !contains(userFeatures, "api") {
				log.Printf("APP - API user %d (%s): API access not enabled for this account", userID, details.Username)
				writeJSONError(w, http.StatusForbidden, "Access denied", "API access is not enabled for your account")
				return
			}
			if !contains(userFeatures, featureName) {
				log.Printf("APP - API user %d (%s): access denied to %s (feature %q not enabled for this account)", userID, details.Username, r.URL.Path, featureName)
				writeJSONError(w, http.StatusForbidden, "Access denied", "")
				return
			}

			if via == ViaMCP && MCPReadOnly(r) && featureName != "mcp" && r.Method != http.MethodGet && r.Method != http.MethodHead {
				log.Printf("APP - API user %d (%s): read-only MCP token blocked %s %s", userID, details.Username, r.Method, r.URL.Path)
				writeJSONError(w, http.StatusForbidden, "This token is read-only", "Generate a token without the read-only option to make changes")
				return
			}

			if r.Method != http.MethodGet && a.DemoMode {
				log.Printf("APP - API user %d (%s): write blocked by demo mode (%s %s)", userID, details.Username, r.Method, r.URL.Path)
				writeJSONError(w, http.StatusForbidden, "Disabled in demo mode", "")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeJSONError(w http.ResponseWriter, status int, errMsg, hint string) {
	body := map[string]string{"error": errMsg}
	if hint != "" {
		body["hint"] = hint
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
