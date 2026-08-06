package account

import (
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// handleMCPSettings renders the MCP token management page. The actual /mcp
// JSON-RPC endpoint (tool registry, tools/list, tools/call) lives in mcp_rpc.go.
func handleMCPSettings(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	tokens, err := mcptokens.GetTokensForUser(ctx, a.DB, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, origin := rpIDAndOrigin(r)
	mcpURL := origin + "/mcp"

	sess, _ := a.Sessions.Get(r, session.CookieName)
	newToken, _ := sess.Values["mcp_new_token"].(string)
	if newToken != "" {
		delete(sess.Values, "mcp_new_token")
		_ = a.Sessions.Save(r, w, sess)
	}

	renderMCPPage(a, w, r, tokens, mcpURL, newToken)
}

// handleMCPCreateToken mints a new MCP token for this account.
func handleMCPCreateToken(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	_ = r.ParseForm()
	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" {
		name = "MCP token"
	}
	if len(name) > 100 {
		name = name[:100]
	}
	readOnly := r.Form.Get("read_only") == "on"
	expiresInDays, _ := strconv.Atoi(r.Form.Get("expires_in_days"))

	sess, _ := a.Sessions.Get(r, session.CookieName)

	rawToken, err := mcptokens.CreateTokenForUser(ctx, a.DB, userID, name, readOnly, expiresInDays)
	if err != nil {
		flash.Add(sess, "error", "Could not create token: database is unavailable. Please try again.")
		_ = a.Sessions.Save(r, w, sess)
		http.Redirect(w, r, "/account/mcp", http.StatusFound)
		return
	}

	data, injectErr := a.InjectData(ctx, userID)
	if injectErr == nil {
		username, _ := data["current_username"].(string)
		scopeNote := ""
		if readOnly {
			scopeNote = ", read-only"
		}
		expiryNote := ""
		if expiresInDays > 0 {
			expiryNote = ", expires in " + strconv.Itoa(expiresInDays) + "d"
		}
		_ = logger.RecordUserAction(a.Config, username, "created a new MCP token ("+name+scopeNote+expiryNote+")", reqip.ClientIP(r))
	}

	sess.Values["mcp_new_token"] = rawToken
	flash.Add(sess, "success", "Token \""+name+"\" created. Copy it below - it will not be shown again.")
	_ = a.Sessions.Save(r, w, sess)

	http.Redirect(w, r, "/account/mcp", http.StatusFound)
}

// handleMCPRevokeToken revokes one of this account's MCP tokens.
func handleMCPRevokeToken(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	tokenID, convErr := strconv.Atoi(r.PathValue("id"))
	if convErr != nil {
		http.NotFound(w, r)
		return
	}

	deleted, err := mcptokens.RevokeToken(ctx, a.DB, tokenID, userID)
	sess, _ := a.Sessions.Get(r, session.CookieName)

	if err == nil && deleted {
		data, injectErr := a.InjectData(ctx, userID)
		if injectErr == nil {
			username, _ := data["current_username"].(string)
			_ = logger.RecordUserAction(a.Config, username, "revoked MCP token #"+strconv.Itoa(tokenID), reqip.ClientIP(r))
		}
		flash.Add(sess, "success", "Token revoked.")
	} else {
		flash.Add(sess, "error", "Token not found.")
	}
	_ = a.Sessions.Save(r, w, sess)

	http.Redirect(w, r, "/account/mcp", http.StatusFound)
}

// RegisterMCP wires the /account/mcp* token-management routes onto mux,
// gated behind the "mcp" feature flag.
func RegisterMCP(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "mcp")(h)
	}
	mux.Handle("GET /account/mcp", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMCPSettings(a, w, r) }))
	mux.Handle("POST /account/mcp/create", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMCPCreateToken(a, w, r) }))
	mux.Handle("POST /account/mcp/{id}/revoke", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMCPRevokeToken(a, w, r) }))
}
