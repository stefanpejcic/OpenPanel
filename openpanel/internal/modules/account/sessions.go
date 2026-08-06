package account

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// ActiveSession is one row of the /account/sessions table.
type ActiveSession struct {
	SessionToken string
	IPAddress    string
	CreatedAt    string
	ExpiresIn    string
}

// handleActiveSessions lists this account's currently active sessions.
func handleActiveSessions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	var sessionsList []ActiveSession
	pattern := fmt.Sprintf("session:%d:*", userID)
	iter := a.Cache.Raw().Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		data, err := a.Cache.Raw().HGetAll(ctx, key).Result()
		if err != nil || len(data) == 0 {
			continue
		}

		ttl, _ := a.Cache.Raw().TTL(ctx, key).Result()
		parts := strings.Split(key, ":")
		token := parts[len(parts)-1]

		expiresIn := "Expiring..."
		if ttl > 0 {
			expiresIn = fmt.Sprintf("%dm", int64(ttl/time.Minute))
		}

		ip := data["ip_address"]
		if ip == "" {
			ip = "Unknown"
		}
		createdAt := data["created_at"]
		if createdAt == "" {
			createdAt = "N/A"
		}

		sessionsList = append(sessionsList, ActiveSession{
			SessionToken: token, IPAddress: ip, CreatedAt: createdAt, ExpiresIn: expiresIn,
		})
	}

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	renderActiveSessionsPage(a, w, r, sessionsList)
}

// terminateUserSession deletes the Redis session record, and if it's the
// caller's own current session,
// also clears the local session cookie so this request's own session
// dies immediately rather than lingering until the (now-deleted) Redis
// record would have naturally expired.
func terminateUserSession(a *appctx.App, r *http.Request, sess *sessions.Session, sessionToken string, userID int) bool {
	sessionKey := fmt.Sprintf("session:%d:%s", userID, sessionToken)
	n, err := a.Cache.Raw().Del(r.Context(), sessionKey).Result()
	if err != nil {
		return false
	}

	if auth.SessionToken(r) == sessionToken {
		clearSessionValues(sess)
		return true
	}
	return n > 0
}

// handleTerminateSession ends one of this account's active sessions.
func handleTerminateSession(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	sessionToken := r.PathValue("session_token")

	sess, _ := a.Sessions.Get(r, session.CookieName)
	success := terminateUserSession(a, r, sess, sessionToken, userID)

	if success {
		flash.Add(sess, "success", "Session terminated successfully.")

		data, err := a.InjectData(r.Context(), userID)
		if err == nil {
			username, _ := data["current_username"].(string)
			_ = logger.RecordUserAction(a.Config, username, "terminated session "+sessionToken, reqip.ClientIP(r))
		}

		_ = a.Sessions.Save(r, w, sess)

		if _, ok := sess.Values["session_token"]; !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		http.Redirect(w, r, "/account/sessions", http.StatusFound)
		return
	}

	flash.Add(sess, "error", "Failed to terminate session or session already expired.")
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, "/account/sessions", http.StatusFound)
}

// RegisterSessions wires the session-management routes onto mux, gated
// behind the "sessions" feature flag.
func RegisterSessions(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "sessions")(h)
	}
	mux.Handle("GET /account/sessions", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleActiveSessions(a, w, r) }))
	mux.Handle("POST /account/sessions/terminate/{session_token}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleTerminateSession(a, w, r) }))
}
