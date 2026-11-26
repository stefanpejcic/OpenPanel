package account

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterMCPAPI wires the /api/account/mcp routes onto mux, gated behind the "mcp" feature flag
func RegisterMCPAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "mcp", "GET /api/account/mcp", func(w http.ResponseWriter, r *http.Request) { apiMCPList(a, w, r) })
	apiregistry.Handle(mux, a, "mcp", "POST /api/account/mcp", func(w http.ResponseWriter, r *http.Request) { apiMCPCreate(a, w, r) })
	apiregistry.Handle(mux, a, "mcp", "DELETE /api/account/mcp/{id}", func(w http.ResponseWriter, r *http.Request) { apiMCPRevoke(a, w, r) })
	if a.DB != nil {
		// give the database a moment after boot, tokens used before then are pinned on first use anyway
		time.AfterFunc(20*time.Second, func() { freezeLegacyTokens(a) })
	}
}

// freezeLegacyTokens pins every old no-list token to its account's current features, once, right after an update
func freezeLegacyTokens(a *appctx.App) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	owners, err := mcptokens.LegacyOwners(ctx, a.DB)
	if err != nil || len(owners) == 0 {
		return
	}
	total := int64(0)
	for _, o := range owners {
		features, err := a.LoadUserFeatures(ctx, o.Username, o.Context)
		if err != nil {
			log.Printf("MCP - could not load features for %s, its older tokens get pinned on first use instead: %v", o.Username, err)
			continue
		}
		n, err := mcptokens.FreezeLegacy(ctx, a.DB, o.UserID, features)
		if err == nil {
			total += n
		}
	}
	log.Printf("MCP - pinned %d older token(s) to their account's current features", total)
}

type apiMCPTokenEntry struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	TokenPrefix string   `json:"token_prefix"`
	ReadOnly    bool     `json:"read_only"`
	Mode        string   `json:"mode"`
	Features    []string `json:"features"`
	CreatedAt   string   `json:"created_at"`
	LastUsedAt  string   `json:"last_used_at,omitempty"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
}

// apiMCPList returns the caller's MCP tokens, never the raw secret (only ever known at creation time)
func apiMCPList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	if _, features, err := mcpUserFeatures(a, r.Context(), userID); err == nil {
		_, _ = mcptokens.FreezeLegacy(r.Context(), a.DB, userID, features)
	}
	tokens, err := mcptokens.GetTokensForUser(r.Context(), a.DB, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	entries := make([]apiMCPTokenEntry, len(tokens))
	for i, t := range tokens {
		entries[i] = apiMCPTokenEntry{
			ID: t.ID, Name: t.Name, TokenPrefix: t.TokenPrefix, ReadOnly: t.ReadOnly, Mode: t.Mode, Features: t.Features,
			CreatedAt: t.CreatedAt, LastUsedAt: t.LastUsedAt.String, ExpiresAt: t.ExpiresAt.String,
		}
	}
	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"tokens": entries})
}

// apiMCPCreate mints a new MCP token for the caller; the raw token is returned exactly once, in this response
func apiMCPCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	var body struct {
		Name          string   `json:"name"`
		ReadOnly      bool     `json:"read_only"`
		Mode          string   `json:"mode"`
		Features      []string `json:"features"`
		ExpiresInDays int      `json:"expires_in_days"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Name = r.Form.Get("name")
		body.ReadOnly = r.Form.Get("read_only") == "true" || r.Form.Get("read_only") == "on"
		body.Mode = r.Form.Get("mode")
		body.Features = r.Form["features"]
		body.ExpiresInDays, _ = strconv.Atoi(r.Form.Get("expires_in_days"))
	}
	body.Mode = strings.ToLower(strings.TrimSpace(body.Mode))
	if body.Mode != "" && !mcptokens.ValidMode(body.Mode) {
		writeAPIAccountJSON(w, http.StatusBadRequest, map[string]string{"error": "mode must be readonly, modify or full"})
		return
	}
	scope := mcptokens.Scope{Mode: mcptokens.ParseMode(body.Mode, body.ReadOnly), Features: mcptokens.FreezableFeatures(body.Features)}
	username, userFeatures, err := mcpUserFeatures(a, ctx, userID)
	if err != nil {
		writeAPIAccountJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not load the account's features"})
		return
	}
	// no list is "all features", minus the ones that can take over the account
	if len(scope.Features) == 0 {
		scope.Features = mcptokens.DefaultFeatures(userFeatures)
	}
	if missing := mcptokens.Missing(scope.Features, userFeatures); len(missing) > 0 {
		writeAPIAccountJSON(w, http.StatusBadRequest, map[string]any{"error": "This account doesn't have these features", "features": missing})
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "MCP token"
	}
	if len(name) > 100 {
		name = name[:100]
	}

	rawToken, err := mcptokens.CreateTokenForUser(ctx, a.DB, userID, name, scope, body.ExpiresInDays)
	if errors.Is(err, mcptokens.ErrScopesUnavailable) {
		writeAPIAccountJSON(w, http.StatusConflict, map[string]string{"error": "Tokens need the mode and allowed_features columns on mcp_tokens, which couldn't be added"})
		return
	}
	if err != nil {
		writeAPIAccountJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not create token"})
		return
	}

	expiryNote := ""
	if body.ExpiresInDays > 0 {
		expiryNote = ", expires in " + strconv.Itoa(body.ExpiresInDays) + "d"
	}
	_ = logger.RecordUserAction(a.Config, username, "created a new MCP token ("+name+tokenScopeNote(scope, userFeatures)+expiryNote+") via API", reqip.ClientIP(r))
	checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_api_token_change", mcpTokenEmail(a, r, username, name, tokenScope(scope, userFeatures, body.ExpiresInDays)+" through the OpenPanel API", true))

	writeAPIAccountJSON(w, http.StatusCreated, map[string]any{
		"token": rawToken, "name": name, "read_only": scope.Mode == mcptokens.ModeReadOnly, "mode": scope.Mode, "features": scope.Features, "expires_in_days": body.ExpiresInDays,
	})
}

// apiMCPRevoke deletes one of the caller's MCP tokens.
func apiMCPRevoke(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	tokenID, convErr := strconv.Atoi(r.PathValue("id"))
	if convErr != nil {
		writeAPIAccountJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid token id"})
		return
	}

	name := rowName(ctx, a, "SELECT name FROM mcp_tokens WHERE id = ? AND user_id = ?", tokenID, userID)
	deleted, err := mcptokens.RevokeToken(ctx, a.DB, tokenID, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !deleted {
		writeAPIAccountJSON(w, http.StatusNotFound, map[string]string{"error": "Token not found"})
		return
	}

	data, injectErr := a.InjectData(ctx, userID)
	if injectErr == nil {
		username, _ := data["current_username"].(string)
		_ = logger.RecordUserAction(a.Config, username, "revoked MCP token #"+strconv.Itoa(tokenID)+" via API", reqip.ClientIP(r))
		checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_api_token_change", mcpTokenEmail(a, r, username, name, "", false))
	}

	writeAPIAccountJSON(w, http.StatusOK, map[string]string{"message": "Token revoked"})
}
