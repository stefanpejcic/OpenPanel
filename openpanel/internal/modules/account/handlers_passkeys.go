package account

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

func writeJSONPasskeys(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONErrorPasskeys(w http.ResponseWriter, status int, msg string) {
	writeJSONPasskeys(w, status, map[string]string{"error": msg})
}

// handlePasskeysSettings renders the passkey management page.
func handlePasskeysSettings(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	passkeys, err := getPasskeysForUser(r.Context(), a, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	renderPasskeysPage(a, w, r, passkeys, webAuthnUnavailableReason(r))
}

// handlePasskeysRegisterBegin starts WebAuthn registration for a new passkey.
func handlePasskeysRegisterBegin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	if ensureErr := ensurePasskeysTable(ctx, a); ensureErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rows, queryErr := a.DB.QueryContext(ctx, "SELECT credential_id FROM user_passkeys WHERE user_id = ?", userID)
	if queryErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var exclusions []protocol.CredentialDescriptor
	for rows.Next() {
		var credID string
		if scanErr := rows.Scan(&credID); scanErr != nil {
			rows.Close()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if raw, decodeErr := decodeCredentialID(credID); decodeErr == nil {
			exclusions = append(exclusions, protocol.CredentialDescriptor{Type: protocol.PublicKeyCredentialType, CredentialID: raw})
		}
	}
	rows.Close()

	wa, waErr := newWebAuthnForRequest(r, a.Config.Get("brand_name", ""))
	if waErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := &passkeyUser{userID: userID, username: username}
	creation, sessionData, beginErr := wa.BeginRegistration(user,
		webauthn.WithExclusions(exclusions),
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementPreferred,
			UserVerification: protocol.VerificationPreferred,
		}),
	)
	if beginErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sessionJSON, marshalErr := json.Marshal(sessionData)
	if marshalErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sess, _ := a.Sessions.Get(r, session.CookieName)
	sess.Values["passkey_reg_session"] = string(sessionJSON)
	_ = a.Sessions.Save(r, w, sess)

	writeJSONPasskeys(w, http.StatusOK, creation.Response)
}

// handlePasskeysRegisterComplete verifies and stores a newly registered passkey.
func handlePasskeysRegisterComplete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	sess, _ := a.Sessions.Get(r, session.CookieName)
	rawSession, _ := sess.Values["passkey_reg_session"].(string)
	delete(sess.Values, "passkey_reg_session")
	_ = a.Sessions.Save(r, w, sess)

	if rawSession == "" {
		writeJSONErrorPasskeys(w, http.StatusBadRequest, "Registration session expired. Please try again.")
		return
	}
	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(rawSession), &sessionData); err != nil {
		writeJSONErrorPasskeys(w, http.StatusBadRequest, "Registration session expired. Please try again.")
		return
	}

	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		name = "Passkey"
	}
	if len(name) > 100 {
		name = name[:100]
	}

	data, injectErr := a.InjectData(ctx, userID)
	if injectErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	wa, waErr := newWebAuthnForRequest(r, a.Config.Get("brand_name", ""))
	if waErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := &passkeyUser{userID: userID, username: username}
	credential, finishErr := wa.FinishRegistration(user, sessionData, r)
	if finishErr != nil {
		writeJSONErrorPasskeys(w, http.StatusBadRequest, finishErr.Error())
		return
	}

	credentialID := encodeCredentialID(credential.ID)
	publicKey := encodeCredentialID(credential.PublicKey)

	if ensureErr := ensurePasskeysTable(ctx, a); ensureErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if _, execErr := a.DB.ExecContext(ctx,
		"INSERT INTO user_passkeys (user_id, credential_id, public_key, sign_count, name) VALUES (?, ?, ?, ?, ?)",
		userID, credentialID, publicKey, credential.Authenticator.SignCount, name); execErr != nil {
		writeJSONErrorPasskeys(w, http.StatusInternalServerError, "Could not save passkey.")
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "registered a new passkey ("+name+")", reqip.ClientIP(r))
	checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_twofactorauth_change", "A new passkey \""+name+"\" was added to your account.")

	writeJSONPasskeys(w, http.StatusOK, map[string]bool{"success": true})
}

// handlePasskeysDelete removes one of the user's registered passkeys.
func handlePasskeysDelete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	passkeyID, convErr := strconv.Atoi(r.PathValue("id"))
	if convErr != nil {
		http.NotFound(w, r)
		return
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)

	result, execErr := a.DB.ExecContext(ctx, "DELETE FROM user_passkeys WHERE id = ? AND user_id = ?", passkeyID, userID)
	deleted := false
	if execErr == nil {
		if n, _ := result.RowsAffected(); n > 0 {
			deleted = true
		}
	}

	data, injectErr := a.InjectData(ctx, userID)
	username, _ := data["current_username"].(string)

	if deleted {
		if injectErr == nil {
			_ = logger.RecordUserAction(a.Config, username, "removed passkey #"+strconv.Itoa(passkeyID), reqip.ClientIP(r))
		}
		flash.Add(sess, "success", "Passkey removed.")
	} else {
		flash.Add(sess, "error", "Passkey not found.")
	}
	_ = a.Sessions.Save(r, w, sess)

	http.Redirect(w, r, "/account/passkeys", http.StatusFound)
}

// handlePasskeysLoginBegin starts WebAuthn login - unauthenticated.
func handlePasskeysLoginBegin(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	wa, waErr := newWebAuthnForRequest(r, a.Config.Get("brand_name", ""))
	if waErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	assertion, sessionData, beginErr := wa.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationPreferred))
	if beginErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sessionJSON, marshalErr := json.Marshal(sessionData)
	if marshalErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	sess, _ := a.Sessions.Get(r, session.CookieName)
	sess.Values["passkey_auth_session"] = string(sessionJSON)
	_ = a.Sessions.Save(r, w, sess)

	writeJSONPasskeys(w, http.StatusOK, assertion.Response)
}

// handlePasskeysLoginComplete verifies a WebAuthn login assertion and
// establishes the session - unauthenticated. Uses the same
// finishLoginSession tail as password/2FA login in login.go, but responds
// with JSON (this endpoint is called via fetch(), not a form submit)
// instead of a redirect.
func handlePasskeysLoginComplete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, _ := a.Sessions.Get(r, session.CookieName)
	rawSession, _ := sess.Values["passkey_auth_session"].(string)
	delete(sess.Values, "passkey_auth_session")

	if rawSession == "" {
		writeJSONErrorPasskeys(w, http.StatusBadRequest, "Login session expired. Please try again.")
		return
	}
	var sessionData webauthn.SessionData
	if err := json.Unmarshal([]byte(rawSession), &sessionData); err != nil {
		writeJSONErrorPasskeys(w, http.StatusBadRequest, "Login session expired. Please try again.")
		return
	}

	wa, waErr := newWebAuthnForRequest(r, a.Config.Get("brand_name", ""))
	if waErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var matchedPasskeyID int

	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		targetUserID, convErr := strconv.Atoi(string(userHandle))
		if convErr != nil {
			return nil, convErr
		}

		if ensureErr := ensurePasskeysTable(ctx, a); ensureErr != nil {
			return nil, ensureErr
		}

		credentialID := encodeCredentialID(rawID)
		var (
			passkeyRowID int
			publicKeyB64 string
			signCount    uint32
			username     string
		)
		row := a.DB.QueryRowContext(ctx,
			"SELECT p.id, p.public_key, p.sign_count, u.username FROM user_passkeys p JOIN users u ON u.id = p.user_id WHERE p.user_id = ? AND p.credential_id = ?",
			targetUserID, credentialID)
		if err := row.Scan(&passkeyRowID, &publicKeyB64, &signCount, &username); err != nil {
			return nil, err
		}

		publicKey, decodeErr := decodeCredentialID(publicKeyB64)
		if decodeErr != nil {
			return nil, decodeErr
		}

		matchedPasskeyID = passkeyRowID

		cred := webauthn.Credential{
			ID:        rawID,
			PublicKey: publicKey,
		}
		cred.Authenticator.SignCount = signCount

		return &passkeyUser{userID: targetUserID, username: username, credentials: []webauthn.Credential{cred}}, nil
	}

	user, credential, finishErr := wa.FinishPasskeyLogin(handler, sessionData, r)
	if finishErr != nil {
		_ = a.Sessions.Save(r, w, sess)
		writeJSONErrorPasskeys(w, http.StatusBadRequest, finishErr.Error())
		return
	}

	pu, ok := user.(*passkeyUser)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, _ = a.DB.ExecContext(ctx, "UPDATE user_passkeys SET sign_count = ?, last_used_at = NOW() WHERE id = ?",
		credential.Authenticator.SignCount, matchedPasskeyID)

	finishLoginSession(a, w, r, sess, pu.userID, pu.username, "logged in with a passkey", "passkey")

	writeJSONPasskeys(w, http.StatusOK, map[string]any{"success": true, "redirect": "/dashboard"})
}

// RegisterPasskeys wires the passkey routes onto mux, gated behind the
// "passkeys" feature flag - including the unauthenticated
// /login/passkey/* endpoints.
func RegisterPasskeys(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "passkeys")(h)
	}

	mux.Handle("GET /account/passkeys", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePasskeysSettings(a, w, r) }))
	mux.Handle("POST /account/passkeys/register/begin", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePasskeysRegisterBegin(a, w, r) }))
	mux.Handle("POST /account/passkeys/register/complete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePasskeysRegisterComplete(a, w, r) }))
	mux.Handle("POST /account/passkeys/{id}/delete", requireLogin(func(w http.ResponseWriter, r *http.Request) { handlePasskeysDelete(a, w, r) }))

	mux.HandleFunc("POST /login/passkey/begin", func(w http.ResponseWriter, r *http.Request) { handlePasskeysLoginBegin(a, w, r) })
	mux.HandleFunc("POST /login/passkey/complete", func(w http.ResponseWriter, r *http.Request) { handlePasskeysLoginComplete(a, w, r) })
}
