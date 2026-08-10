package account

import (
	"net/http"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterPasskeysAPI wires the /api/account/passkeys routes onto mux,
// gated behind the "passkeys" feature flag. Registration itself
// (BeginRegistration/FinishRegistration) isn't exposed here - the WebAuthn
// ceremony is a browser-native ceremony (navigator.credentials) tied to a
// session cookie for the in-progress challenge, which doesn't fit a
// stateless Bearer-token API client. Listing and revoking existing
// passkeys does fit, so only those two are exposed.
func RegisterPasskeysAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "passkeys", "GET /api/account/passkeys", func(w http.ResponseWriter, r *http.Request) { apiPasskeysList(a, w, r) })
	apiregistry.Handle(mux, a, "passkeys", "DELETE /api/account/passkeys/{id}", func(w http.ResponseWriter, r *http.Request) { apiPasskeysDelete(a, w, r) })
}

type apiPasskeyEntry struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"`
	LastUsedAt string `json:"last_used_at,omitempty"`
}

// apiPasskeysList returns the caller's registered passkeys.
func apiPasskeysList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	passkeys, err := getPasskeysForUser(r.Context(), a, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	entries := make([]apiPasskeyEntry, len(passkeys))
	for i, pk := range passkeys {
		entries[i] = apiPasskeyEntry{ID: pk.ID, Name: pk.Name, CreatedAt: pk.CreatedAt, LastUsedAt: pk.LastUsedAt.String}
	}
	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"passkeys": entries})
}

// apiPasskeysDelete removes one of the caller's registered passkeys.
func apiPasskeysDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	passkeyID, convErr := strconv.Atoi(r.PathValue("id"))
	if convErr != nil {
		writeAPIAccountJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid passkey id"})
		return
	}

	result, execErr := a.DB.ExecContext(ctx, "DELETE FROM user_passkeys WHERE id = ? AND user_id = ?", passkeyID, userID)
	if execErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		writeAPIAccountJSON(w, http.StatusNotFound, map[string]string{"error": "Passkey not found"})
		return
	}

	data, injectErr := a.InjectData(ctx, userID)
	if injectErr == nil {
		username, _ := data["current_username"].(string)
		_ = logger.RecordUserAction(a.Config, username, "removed passkey #"+strconv.Itoa(passkeyID)+" via API", reqip.ClientIP(r))
	}

	writeAPIAccountJSON(w, http.StatusOK, map[string]string{"message": "Passkey removed"})
}
