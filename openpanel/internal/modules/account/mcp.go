package account

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleMCPSettings renders the MCP token management page - the actual /mcp JSON-RPC endpoint lives in mcp_rpc.go
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
	expiresInDays, _ := strconv.Atoi(r.Form.Get("expires_in_days"))
	scope := mcptokens.Scope{Mode: mcptokens.ModeFull, Features: mcptokens.CleanFeatures(r.Form["features"])}
	switch {
	case r.Form.Get("read_only") == "on":
		scope.Mode = mcptokens.ModeReadOnly
	// scope_form is only on the new form, so an old cached page without the box still makes a full token
	case r.Form.Get("scope_form") != "" && r.Form.Get("allow_destructive") != "on":
		scope.Mode = mcptokens.ModeModify
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)
	fail := func(msg string) {
		flash.Add(sess, "error", msg)
		_ = a.Sessions.Save(r, w, sess)
		http.Redirect(w, r, "/account/mcp", http.StatusFound)
	}
	if bad := unknownMCPFeatures(scope.Features); len(bad) > 0 {
		fail("Could not create token: unknown feature " + strings.Join(bad, ", ") + ".")
		return
	}

	rawToken, err := mcptokens.CreateTokenForUser(ctx, a.DB, userID, name, scope, expiresInDays)
	if errors.Is(err, mcptokens.ErrScopesUnavailable) {
		fail("Could not create a limited token: the panel database is missing the columns for it. Ask your hosting provider to restart OpenPanel, or create a read-only or full-access token.")
		return
	}
	if err != nil {
		fail("Could not create token: database is unavailable. Please try again.")
		return
	}

	data, injectErr := a.InjectData(ctx, userID)
	if injectErr == nil {
		username, _ := data["current_username"].(string)
		expiryNote := ""
		if expiresInDays > 0 {
			expiryNote = ", expires in " + strconv.Itoa(expiresInDays) + "d"
		}
		_ = logger.RecordUserAction(a.Config, username, "created a new MCP token ("+name+tokenScopeNote(scope)+expiryNote+")", reqip.ClientIP(r))
		checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_api_token_change", mcpTokenEmail(a, r, username, name, tokenScope(scope, expiresInDays), true))
	}

	sess.Values["mcp_new_token"] = rawToken
	flash.Add(sess, "success", web.Tr(a, r, "Token \"%(name)s\" created. Copy it below - it will not be shown again.", "name", name))
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

	name := rowName(ctx, a, "SELECT name FROM mcp_tokens WHERE id = ? AND user_id = ?", tokenID, userID)
	deleted, err := mcptokens.RevokeToken(ctx, a.DB, tokenID, userID)
	sess, _ := a.Sessions.Get(r, session.CookieName)

	if err == nil && deleted {
		data, injectErr := a.InjectData(ctx, userID)
		if injectErr == nil {
			username, _ := data["current_username"].(string)
			_ = logger.RecordUserAction(a.Config, username, "revoked MCP token #"+strconv.Itoa(tokenID), reqip.ClientIP(r))
			checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_api_token_change", mcpTokenEmail(a, r, username, name, "", false))
		}
		flash.Add(sess, "success", "Token revoked.")
	} else {
		flash.Add(sess, "error", "Token not found.")
	}
	_ = a.Sessions.Save(r, w, sess)

	http.Redirect(w, r, "/account/mcp", http.StatusFound)
}

// RegisterMCP wires the /account/mcp* token-management routes onto mux, gated behind the "mcp" feature flag
func RegisterMCP(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "mcp")(h)
	}
	mux.Handle("GET /account/mcp", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMCPSettings(a, w, r) }))
	mux.Handle("POST /account/mcp/create", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMCPCreateToken(a, w, r) }))
	mux.Handle("POST /account/mcp/{id}/revoke", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleMCPRevokeToken(a, w, r) }))
	mux.Handle("POST /account/mcp/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		web.ServeBulkDispatch(a, mux, w, r, mcpBulkActions(web.RequestTranslator(a, r)), func(_, _, id string) (*web.BulkCall, web.BulkResult) {
			return web.Call(web.BulkCall{Method: http.MethodPost, Path: "/account/mcp/" + url.PathEscape(id) + "/revoke"})
		})
	}))
}

func mcpBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "revoke", Label: t.Get("Revoke"), Confirm: t.Get("Revoke the selected tokens? MCP clients using them stop working immediately."), Danger: true},
	}
}
