package auth

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apirisk"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// isoLayout is the timestamp format for "created_at"/"last_active" in the Redis session hash - must match what internal/modules/account/login.go writes
// Always write these in UTC: time.Parse reads a zone-less string as UTC, so local time made sessions look hours old on hosts west of UTC (CST -6h > 5h max lifetime = instant logout)
const isoLayout = "2006-01-02T15:04:05.999999"

// RequireLogin is middleware that enforces an authenticated session (existence, max lifetime, IP binding, 2FA, demo-mode writes) and gates access by feature - granted if the caller has any one of featureNames, so a route shared by two modules passes both
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

			// picking a theme only changes how the panel looks, demo visitors can try the themes
			if r.Method != http.MethodGet && a.DemoMode && r.URL.Path != "/account/theme" {
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

			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, userContextKey{}, details.Context)))
		})
	}
}

type userContextKey struct{}

// UserContext is the logged-in account's context (its /home folder), "" outside RequireLogin
func UserContext(r *http.Request) string {
	c, _ := r.Context().Value(userContextKey{}).(string)
	return c
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
	pipe.HSet(ctx, sessionKey, "last_active", time.Now().UTC().Format(isoLayout))
	pipe.Expire(ctx, sessionKey, a.SessionDuration)
	_, _ = pipe.Exec(ctx)
}

// clearSession empties the session's values map but doesn't save - callers add a flash message after this and save once, so it survives the clear
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

// RequireAPI is middleware for API routes: Bearer-token auth only (JWT or MCP) or an AI Assistant replay, feature-gated like RequireLogin
func RequireAPI(a *appctx.App, featureName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID, hasUser := UserID(r)
			via := Via(r)

			if (via != ViaJWT && via != ViaMCP && via != ViaAI) || !hasUser || userID == 0 {
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

			if !details.Found || strings.HasPrefix(details.Username, "SUSPENDED_") {
				log.Printf("APP - API user %d (%s): account suspended or deleted, rejecting token", userID, details.Username)
				writeJSONError(w, http.StatusUnauthorized, "Account is suspended or no longer exists", "")
				return
			}
			if via == ViaJWT {
				fp, err := PasswordFingerprint(ctx, a, userID)
				if err != nil || !hmac.Equal([]byte(fp), []byte(jwtPwd(r))) {
					log.Printf("APP - API user %d (%s): JWT issued before the last password change, rejecting", userID, details.Username)
					writeJSONError(w, http.StatusUnauthorized, "Token is no longer valid", "Log in again to get a new token")
					return
				}
			}

			userFeatures, err := a.LoadUserFeatures(ctx, details.Username, details.Context)
			if err != nil {
				log.Printf("APP - API user %d (%s): failed to load enabled features: %v", userID, details.Username, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			// the ai assistant acts for a logged in user, so it only needs the route's own feature
			if featureName != "api" && via != ViaAI && !contains(userFeatures, "api") {
				log.Printf("APP - API user %d (%s): API access not enabled for this account", userID, details.Username)
				writeJSONError(w, http.StatusForbidden, "Access denied", "API access is not enabled for your account")
				return
			}
			if via == ViaAI && !contains(userFeatures, "ai_assistant") {
				writeJSONError(w, http.StatusForbidden, "Access denied", "")
				return
			}
			if !contains(userFeatures, featureName) {
				log.Printf("APP - API user %d (%s): access denied to %s (feature %q not enabled for this account)", userID, details.Username, r.URL.Path, featureName)
				writeJSONError(w, http.StatusForbidden, "Access denied", "This account doesn't have the '"+featureName+"' feature")
				return
			}

			// tokens from before feature lists get pinned to what the account has today, from then on they don't grow
			if via == ViaMCP {
				if tok, ok := MCPTokenInfo(r); ok && tok.Scope.AllFeatures {
					if n, err := mcptokens.FreezeLegacy(ctx, a.DB, userID, userFeatures); err == nil && n > 0 {
						log.Printf("APP - API user %d (%s): pinned %d older MCP token(s) to the account's current features", userID, details.Username, n)
					}
				}
			}

			// same 2fa enforcement as RequireLogin, bearer tokens only get the 2fa enrollment routes until it's on
			if a.TwofaEnforce && a.ModuleEnabled("twofa") && featureName != "twofa" {
				status, err := a.Get2FAStatusForUser(ctx, userID)
				if err != nil || !status.Enabled {
					log.Printf("APP - API user %d (%s): 2FA required but not enabled, blocked %s %s", userID, details.Username, r.Method, r.URL.Path)
					writeJSONError(w, http.StatusForbidden, "Two-Factor Authentication is required", "Enable 2FA via /api/account/2fa/setup first")
					return
				}
			}

			// only the json-rpc transport itself is exempt, token create/revoke under /api/account/mcp is not
			isMCPTransport := r.Method == http.MethodPost && r.URL.Path == "/mcp"
			// a few GET routes change things too (wp-cli update_now), so it's the route's risk that counts, not just the method
			if via == ViaMCP && MCPReadOnly(r) && !isMCPTransport && (r.Method != http.MethodGet && r.Method != http.MethodHead || apirisk.ForRequest(r) != apirisk.Read) {
				log.Printf("APP - API user %d (%s): read-only MCP token blocked %s %s", userID, details.Username, r.Method, r.URL.Path)
				writeJSONError(w, http.StatusForbidden, "This token is read-only", "Generate a token without the read-only option to make changes")
				return
			}

			if via == ViaMCP && !isMCPTransport {
				tok, _ := MCPTokenInfo(r)
				if msg, hint := MCPScopeBlock(tok.Scope, featureName, r); msg != "" {
					log.Printf("APP - API user %d (%s): MCP token #%d blocked %s %s: %s", userID, details.Username, tok.ID, r.Method, r.URL.Path, msg)
					writeJSONError(w, http.StatusForbidden, msg, hint)
					return
				}
			}

			if r.Method != http.MethodGet && a.DemoMode {
				log.Printf("APP - API user %d (%s): write blocked by demo mode (%s %s)", userID, details.Username, r.Method, r.URL.Path)
				writeJSONError(w, http.StatusForbidden, "Disabled in demo mode", "")
				return
			}

			// handlers log what changed, this adds which token did it
			if via == ViaMCP && !isMCPTransport && r.Method != http.MethodGet && r.Method != http.MethodHead {
				tok, _ := MCPTokenInfo(r)
				_ = logger.RecordUserAction(a.Config, details.Username, fmt.Sprintf("used MCP token \"%s\" (#%d) for %s %s", tok.Name, tok.ID, r.Method, r.URL.Path), reqip.ClientIP(r))
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MCPScopeBlock says why a token may not make this request, empty when it may
func MCPScopeBlock(scope mcptokens.Scope, featureName string, r *http.Request) (string, string) {
	// a token minting or revoking tokens could hand itself more than it was given
	if TokenManagementWrite(r.Method, r.URL.Path) {
		return "Tokens can't create or revoke tokens", "Manage tokens on Account > AI Assistant (MCP) in OpenPanel"
	}
	if featureName != "api" && featureName != "mcp" && !scope.Allows(featureName) {
		return "This token isn't allowed to use the '" + featureName + "' feature", "Make a new token that includes '" + featureName + "' on Account > AI Assistant (MCP)"
	}
	// a database login works around read-only and no-deletes, you can drop tables with it
	if scope.Mode != mcptokens.ModeFull && apirisk.ExposesSecrets(r) {
		return "This token can't read database passwords or log in to phpMyAdmin", "Use a full-access token that includes the feature"
	}
	if scope.Mode == mcptokens.ModeModify && apirisk.ForRequest(r) == apirisk.Destructive {
		return "This token can't delete or overwrite things", "Use a full-access token for deletes, restores, password changes and similar actions"
	}
	return "", ""
}

// TokenManagementWrite is true for creating or revoking MCP tokens
func TokenManagementWrite(method, path string) bool {
	return strings.HasPrefix(path, "/api/account/mcp") && method != http.MethodGet && method != http.MethodHead
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
