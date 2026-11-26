package account

import (
	"context"
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

	// older no-list tokens get pinned here too, so the list never shows one that still grows with the plan
	if _, features, err := mcpUserFeatures(a, ctx, userID); err == nil {
		_, _ = mcptokens.FreezeLegacy(ctx, a.DB, userID, features)
	}
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
	scope := mcptokens.Scope{Mode: mcptokens.ModeFull, Features: mcptokens.FreezableFeatures(r.Form["features"])}
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
	username, userFeatures, err := mcpUserFeatures(a, ctx, userID)
	if err != nil {
		fail("Could not create token: your account's features couldn't be loaded. Please try again.")
		return
	}
	scope.Features = formFeatures(r.Form, userFeatures)
	if len(scope.Features) == 0 {
		fail("Could not create token: pick at least one feature it can use.")
		return
	}
	if missing := mcptokens.Missing(scope.Features, userFeatures); len(missing) > 0 {
		fail("Could not create token: your account doesn't have " + strings.Join(featureLabelList(missing), ", ") + ".")
		return
	}

	rawToken, err := mcptokens.CreateTokenForUser(ctx, a.DB, userID, name, scope, expiresInDays)
	if errors.Is(err, mcptokens.ErrScopesUnavailable) {
		fail("Could not create token: the panel database is missing the columns tokens need. Ask your hosting provider to restart OpenPanel.")
		return
	}
	if err != nil {
		fail("Could not create token: database is unavailable. Please try again.")
		return
	}

	expiryNote := ""
	if expiresInDays > 0 {
		expiryNote = ", expires in " + strconv.Itoa(expiresInDays) + "d"
	}
	_ = logger.RecordUserAction(a.Config, username, "created a new MCP token ("+name+tokenScopeNote(scope, userFeatures)+expiryNote+")", reqip.ClientIP(r))
	checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_api_token_change", mcpTokenEmail(a, r, username, name, tokenScope(scope, userFeatures, expiresInDays), true))

	sess.Values["mcp_new_token"] = rawToken
	flash.Add(sess, "success", web.Tr(a, r, "Token \"%(name)s\" created. Copy it below - it will not be shown again.", "name", name))
	_ = a.Sessions.Save(r, w, sess)

	http.Redirect(w, r, "/account/mcp", http.StatusFound)
}

// formFeatures turns the token form into a feature list, "all features" is the default list plus any sensitive ones ticked
func formFeatures(form url.Values, userFeatures []string) []string {
	picked := mcptokens.FreezableFeatures(form["features"])
	// older templates and themes have no all-features box: ticked boxes are the limit, none ticked meant everything
	if form.Get("features_form") == "" {
		if len(picked) == 0 {
			return mcptokens.DefaultFeatures(userFeatures)
		}
		return picked
	}
	if form.Get("all_features") != "on" {
		return picked
	}
	out := mcptokens.DefaultFeatures(userFeatures)
	for _, f := range picked {
		if mcptokens.IsSensitive(f) {
			out = append(out, f)
		}
	}
	return mcptokens.CleanFeatures(out)
}

// mcpUserFeatures is the caller's username and the features their account has right now
func mcpUserFeatures(a *appctx.App, ctx context.Context, userID int) (string, []string, error) {
	details, err := a.GetUserDetailsWithPlan(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	if details.Username == "" {
		return "", nil, errors.New("no username for user")
	}
	features, err := a.LoadUserFeatures(ctx, details.Username, details.Context)
	return details.Username, features, err
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
