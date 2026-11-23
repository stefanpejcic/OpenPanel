// Package auth handles request-level authentication: identity from a session cookie/JWT/MCP token, the access-gating middleware, and the license gate.
package auth

import (
	"context"
	"net/http"

	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
)

type ctxKey int

const (
	userIDKey ctxKey = iota
	authViaKey
	mcpTokenKey
	sessionTokenKey
	jwtPwdKey
)

// AuthVia identifies how the caller's identity was established.
type AuthVia string

const (
	ViaNone    AuthVia = ""
	ViaSession AuthVia = "session"
	ViaJWT     AuthVia = "jwt"
	ViaMCP     AuthVia = "mcp_token"
	ViaAI      AuthVia = "ai_assistant"
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

// MCPToken is the token behind a ViaMCP request, so handlers can see its scope and name
type MCPToken struct {
	ID    int
	Name  string
	Scope mcptokens.Scope
}

// MCPTokenInfo returns the token this request authenticated with, false for anything that isn't an MCP token
func MCPTokenInfo(r *http.Request) (MCPToken, bool) {
	v, ok := r.Context().Value(mcpTokenKey).(MCPToken)
	return v, ok
}

func MCPReadOnly(r *http.Request) bool {
	t, ok := MCPTokenInfo(r)
	return ok && t.Scope.Mode == mcptokens.ModeReadOnly
}

// SessionToken returns the session_token read during LoadUser, so downstream middleware doesn't need to re-decode the session
func SessionToken(r *http.Request) string {
	v, _ := r.Context().Value(sessionTokenKey).(string)
	return v
}

func withAuth(r *http.Request, userID int, via AuthVia) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	ctx = context.WithValue(ctx, authViaKey, via)
	return r.WithContext(ctx)
}

// WithAIAssistant marks an internal request the AI Assistant replays for a logged in user, it can only be set in code, never from a header
func WithAIAssistant(r *http.Request, userID int) *http.Request {
	return withAuth(r, userID, ViaAI)
}

func withMCPToken(r *http.Request, t MCPToken) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), mcpTokenKey, t))
}

func withSessionToken(r *http.Request, token string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), sessionTokenKey, token))
}

func withJWTPwd(r *http.Request, fp string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), jwtPwdKey, fp))
}

func jwtPwd(r *http.Request) string {
	v, _ := r.Context().Value(jwtPwdKey).(string)
	return v
}
