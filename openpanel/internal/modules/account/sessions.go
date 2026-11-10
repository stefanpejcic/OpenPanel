package account

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/sessions"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// ActiveSession is one row of the /account/sessions table.
type ActiveSession struct {
	SessionToken string
	IPAddress    string
	CreatedAt    string
	ExpiresIn    string
	Current      bool
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
		createdAt := sessionTime(data["created_at"])
		if createdAt == "" {
			createdAt = "N/A"
		}

		sessionsList = append(sessionsList, ActiveSession{
			SessionToken: token, IPAddress: ip, CreatedAt: createdAt, ExpiresIn: expiresIn,
			Current: token == auth.SessionToken(r),
		})
	}

	// the scan comes back in no order, show this session first and then the newest
	sort.SliceStable(sessionsList, func(i, j int) bool {
		if sessionsList[i].Current != sessionsList[j].Current {
			return sessionsList[i].Current
		}
		return sessionsList[i].CreatedAt > sessionsList[j].CreatedAt
	})

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	renderActiveSessionsPage(a, w, r, sessionsList)
}

// sessionTime turns the stored "2026-11-05T10:03:54.071078" into "2026-11-05 10:03:54", anything it can't read is kept as is
func sessionTime(raw string) string {
	for _, layout := range []string{"2006-01-02T15:04:05.999999999", time.RFC3339Nano, "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}
	return raw
}

// terminateUserSession deletes the Redis session record, and if it's the caller's own current session, also clears the local cookie so it dies immediately instead of lingering until the deleted record would've expired
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

// RegisterSessions wires the session-management routes onto mux, gated behind the "sessions" feature flag
func RegisterSessions(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "sessions")(h)
	}
	mux.Handle("GET /account/sessions", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleActiveSessions(a, w, r) }))
	mux.Handle("POST /account/sessions/terminate/{session_token}", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleTerminateSession(a, w, r) }))
	mux.Handle("POST /account/sessions/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		web.ServeBulkDispatch(a, mux, w, r, sessionsBulkActions(web.RequestTranslator(a, r)), func(_, _, token string) (*web.BulkCall, web.BulkResult) {
			// ending your own session halfway through would log you out
			if token == auth.SessionToken(r) {
				return web.Skip("This is your current session.")
			}
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/account/sessions/terminate/" + url.PathEscape(token)})
		})
	}))
}

func sessionsBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "terminate", Label: t.Get("Terminate"), Confirm: t.Get("Terminate the selected sessions? Anyone using them will be logged out."), Danger: true},
	}
}
