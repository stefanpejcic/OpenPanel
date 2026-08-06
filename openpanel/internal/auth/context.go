// Package auth implements request-level authentication: identity
// resolution from a session cookie, JWT, or MCP token, the login/API
// access-gating middleware, and the license gate.
package auth

import (
	"context"
	"net/http"
)

type ctxKey int

const (
	userIDKey ctxKey = iota
	authViaKey
	mcpReadOnlyKey
	sessionTokenKey
)

// AuthVia identifies how the caller's identity was established.
type AuthVia string

const (
	ViaNone    AuthVia = ""
	ViaSession AuthVia = "session"
	ViaJWT     AuthVia = "jwt"
	ViaMCP     AuthVia = "mcp_token"
)

// UserID returns the authenticated user's ID and whether one is set.
func UserID(r *http.Request) (int, bool) {
	v, ok := r.Context().Value(userIDKey).(int)
	return v, ok
}

func Via(r *http.Request) AuthVia {
	if v, ok := r.Context().Value(authViaKey).(AuthVia); ok {
		return v
	}
	return ViaNone
}

func MCPReadOnly(r *http.Request) bool {
	v, _ := r.Context().Value(mcpReadOnlyKey).(bool)
	return v
}

// SessionToken returns the session_token value read from the session
// cookie during LoadUser, so downstream middleware doesn't need to
// re-decode the session.
func SessionToken(r *http.Request) string {
	v, _ := r.Context().Value(sessionTokenKey).(string)
	return v
}

func withAuth(r *http.Request, userID int, via AuthVia) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	ctx = context.WithValue(ctx, authViaKey, via)
	return r.WithContext(ctx)
}

func withMCPReadOnly(r *http.Request, readOnly bool) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), mcpReadOnlyKey, readOnly))
}

func withSessionToken(r *http.Request, token string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), sessionTokenKey, token))
}
