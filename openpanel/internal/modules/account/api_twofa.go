package account

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/pquerna/otp/totp"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
)

// RegisterTwofaAPI wires the /api/account/2fa routes onto mux, gated
// behind the "twofa" feature flag. Setup is two calls, mirroring the web
// page's flow: POST .../setup generates and stores a pending secret (2FA
// still off), then POST .../confirm validates a code against it and turns
// 2FA on.
func RegisterTwofaAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "twofa", "GET /api/account/2fa", func(w http.ResponseWriter, r *http.Request) { apiTwofaStatus(a, w, r) })
	apiregistry.Handle(mux, a, "twofa", "POST /api/account/2fa/setup", func(w http.ResponseWriter, r *http.Request) { apiTwofaSetup(a, w, r) })
	apiregistry.Handle(mux, a, "twofa", "POST /api/account/2fa/confirm", func(w http.ResponseWriter, r *http.Request) { apiTwofaConfirm(a, w, r) })
	apiregistry.Handle(mux, a, "twofa", "DELETE /api/account/2fa", func(w http.ResponseWriter, r *http.Request) { apiTwofaDisable(a, w, r) })
}

func otpauthURL(username, secret, issuer string) string {
	return "otpauth://totp/" + url.PathEscape(username) + "?secret=" + secret + "&issuer=" + url.QueryEscape(issuer)
}

// twofaIssuerName is what shows up as the account issuer in the user's
// authenticator app: the configured brand name, falling back to the panel's
// own domain, falling back to "OpenPanel" if neither is set.
func twofaIssuerName(a *appctx.App, ctx context.Context) string {
	if brandName := a.Config.Get("brand_name", ""); brandName != "" {
		return brandName
	}
	if domain := sysinfo.GetOpenPanelDomain(ctx, a.Cache); domain != "" {
		return domain
	}
	return "OpenPanel"
}

// apiTwofaStatus reports whether 2FA is currently enabled for the caller.
func apiTwofaStatus(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	status, err := a.Get2FAStatusForUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeAPIAccountJSON(w, http.StatusOK, map[string]bool{"enabled": status.Enabled})
}

// apiTwofaSetup generates a new pending TOTP secret. 2FA remains disabled
// until apiTwofaConfirm validates a code against it.
func apiTwofaSetup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	secret, err := randomBase32Secret()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, execErr := a.DB.ExecContext(ctx, "UPDATE users SET twofa_enabled = 0, otp_secret = ? WHERE id = ?", secret, userID); execErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	invalidate2FAStatus(a, r, userID)

	data, injectErr := a.InjectData(ctx, userID)
	username, _ := data["current_username"].(string)
	if injectErr != nil {
		username = ""
	}

	writeAPIAccountJSON(w, http.StatusOK, map[string]string{
		"secret":      secret,
		"otpauth_url": otpauthURL(username, secret, twofaIssuerName(a, ctx)),
	})
}

// apiTwofaConfirm validates a code against the pending secret from
// apiTwofaSetup and, on success, enables 2FA.
func apiTwofaConfirm(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	var body struct {
		OTPCode string `json:"otp_code"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.OTPCode = r.Form.Get("otp_code")
	}
	otpCode := strings.TrimSpace(body.OTPCode)

	status, err := a.Get2FAStatusForUser(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if status.OTPSecret == "" || !totp.Validate(otpCode, status.OTPSecret) {
		writeAPIAccountJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid verification code"})
		return
	}

	if _, execErr := a.DB.ExecContext(ctx, "UPDATE users SET twofa_enabled = 1 WHERE id = ?", userID); execErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	invalidate2FAStatus(a, r, userID)

	data, injectErr := a.InjectData(ctx, userID)
	if injectErr == nil {
		username, _ := data["current_username"].(string)
		checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_twofactorauth_change", "Two-Factor Authentication was enabled via API.")
		_ = logger.RecordUserAction(a.Config, username, "enabled 2FA for account via API", reqip.ClientIP(r))
	}

	writeAPIAccountJSON(w, http.StatusOK, map[string]bool{"enabled": true})
}

// apiTwofaDisable turns 2FA off and clears the stored secret.
func apiTwofaDisable(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	if _, execErr := a.DB.ExecContext(ctx, "UPDATE users SET twofa_enabled = 0, otp_secret = NULL WHERE id = ?", userID); execErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	invalidate2FAStatus(a, r, userID)

	data, injectErr := a.InjectData(ctx, userID)
	if injectErr == nil {
		username, _ := data["current_username"].(string)
		checkIfUserShouldBeNotified(a, ctx, userID, username, "notify_twofactorauth_change", "Two-Factor Authentication was disabled via API.")
		_ = logger.RecordUserAction(a.Config, username, "disabled 2FA for account via API", reqip.ClientIP(r))
	}

	writeAPIAccountJSON(w, http.StatusOK, map[string]bool{"enabled": false})
}
