package account

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mcptokens"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterMCPAPI wires the /api/account/mcp routes onto mux, gated behind
// the "mcp" feature flag.
func RegisterMCPAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "mcp", "GET /api/account/mcp", func(w http.ResponseWriter, r *http.Request) { apiMCPList(a, w, r) })
	apiregistry.Handle(mux, a, "mcp", "POST /api/account/mcp", func(w http.ResponseWriter, r *http.Request) { apiMCPCreate(a, w, r) })
	apiregistry.Handle(mux, a, "mcp", "DELETE /api/account/mcp/{id}", func(w http.ResponseWriter, r *http.Request) { apiMCPRevoke(a, w, r) })
}

type apiMCPTokenEntry struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	TokenPrefix string `json:"token_prefix"`
	ReadOnly    bool   `json:"read_only"`
	CreatedAt   string `json:"created_at"`
	LastUsedAt  string `json:"last_used_at,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
}

// apiMCPList returns the caller's MCP tokens (never the raw secret, which
// is only ever known at creation time).
func apiMCPList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	tokens, err := mcptokens.GetTokensForUser(r.Context(), a.DB, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	entries := make([]apiMCPTokenEntry, len(tokens))
	for i, t := range tokens {
		entries[i] = apiMCPTokenEntry{
			ID: t.ID, Name: t.Name, TokenPrefix: t.TokenPrefix, ReadOnly: t.ReadOnly,
			CreatedAt: t.CreatedAt, LastUsedAt: t.LastUsedAt.String, ExpiresAt: t.ExpiresAt.String,
		}
	}
	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"tokens": entries})
}

// apiMCPCreate mints a new MCP token for the caller. The raw token is
// returned exactly once, in this response.
func apiMCPCreate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	var body struct {
		Name          string `json:"name"`
		ReadOnly      bool   `json:"read_only"`
		ExpiresInDays int    `json:"expires_in_days"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Name = r.Form.Get("name")
		body.ReadOnly = r.Form.Get("read_only") == "true" || r.Form.Get("read_only") == "on"
		body.ExpiresInDays, _ = strconv.Atoi(r.Form.Get("expires_in_days"))
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		name = "MCP token"
	}
	if len(name) > 100 {
		name = name[:100]
	}

	rawToken, err := mcptokens.CreateTokenForUser(ctx, a.DB, userID, name, body.ReadOnly, body.ExpiresInDays)
	if err != nil {
		writeAPIAccountJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not create token"})
		return
	}

	data, injectErr := a.InjectData(ctx, userID)
	if injectErr == nil {
		username, _ := data["current_username"].(string)
		scopeNote := ""
		if body.ReadOnly {
			scopeNote = ", read-only"
		}
		expiryNote := ""
		if body.ExpiresInDays > 0 {
			expiryNote = ", expires in " + strconv.Itoa(body.ExpiresInDays) + "d"
		}
		_ = logger.RecordUserAction(a.Config, username, "created a new MCP token ("+name+scopeNote+expiryNote+") via API", reqip.ClientIP(r))
	}

	writeAPIAccountJSON(w, http.StatusCreated, map[string]any{
		"token": rawToken, "name": name, "read_only": body.ReadOnly, "expires_in_days": body.ExpiresInDays,
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
	}

	writeAPIAccountJSON(w, http.StatusOK, map[string]string{"message": "Token revoked"})
}
