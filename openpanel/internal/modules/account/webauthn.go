package account

import (
	"context"
	"database/sql"
	"encoding/base64"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-webauthn/webauthn/webauthn"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

const createPasskeysTableSQL = `
CREATE TABLE IF NOT EXISTS user_passkeys (
	id INT AUTO_INCREMENT PRIMARY KEY,
	user_id INT NOT NULL,
	credential_id VARCHAR(255) NOT NULL UNIQUE,
	public_key TEXT NOT NULL,
	sign_count INT NOT NULL DEFAULT 0,
	name VARCHAR(100) NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	last_used_at TIMESTAMP NULL,
	INDEX idx_user_id (user_id)
) ENGINE=InnoDB`

func ensurePasskeysTable(ctx context.Context, a *appctx.App) error {
	_, err := a.DB.ExecContext(ctx, createPasskeysTableSQL)
	return err
}

// rpIDAndOrigin computes the WebAuthn Relying Party ID (the request host
// without a port) and origin (the full scheme+host) - which the browser's
// own clientDataJSON.origin must match byte-for-byte for registration/login
// to verify. This panel's backend listens on plain HTTP behind a Caddy
// TLS-terminating reverse proxy, so the raw, unproxied connection scheme
// would almost always read "http" even though the browser is on https -
// this trusts X-Forwarded-Proto first since that's what actually reflects
// what the browser used.
func rpIDAndOrigin(r *http.Request) (rpID, origin string) {
	host := r.Host
	rpID = host
	if idx := strings.Index(host, ":"); idx != -1 {
		rpID = host[:idx]
	}

	scheme := "http"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = strings.ToLower(strings.TrimSpace(strings.SplitN(proto, ",", 2)[0]))
	} else if r.TLS != nil {
		scheme = "https"
	}

	return rpID, scheme + "://" + host
}

// isIPHost reports whether host (already stripped of any port) is a literal
// IP address rather than a domain name.
func isIPHost(host string) bool {
	return net.ParseIP(host) != nil
}

// webAuthnUnavailableReason reports why the browser's WebAuthn API won't be
// usable for the current request, or "" if it should work fine. Browsers
// only expose navigator.credentials in a secure context (HTTPS, or
// localhost), and additionally refuse to register/verify a credential when
// the effective domain is a literal IP address rather than a real domain -
// hitting either case leaves navigator.credentials undefined client-side,
// which without this check surfaces only as a cryptic "can't access
// property 'create', navigator.credentials is undefined" in the browser
// console instead of an explanation on the page.
func webAuthnUnavailableReason(r *http.Request) string {
	rpID, origin := rpIDAndOrigin(r)
	switch {
	case !strings.HasPrefix(origin, "https://") && rpID != "localhost":
		return "Passkeys require the panel to be accessed over HTTPS. Ask your administrator to configure a domain and SSL certificate for this panel."
	case isIPHost(rpID):
		return "Passkeys require the panel to be accessed via a domain name, not an IP address. Ask your administrator to configure a domain for this panel."
	default:
		return ""
	}
}

// newWebAuthnForRequest builds a *webauthn.WebAuthn scoped to the current
// request's host, matching rp_id_and_origin() being recomputed on every
// call in Python rather than configured once.
func newWebAuthnForRequest(r *http.Request, brandName string) (*webauthn.WebAuthn, error) {
	rpID, origin := rpIDAndOrigin(r)
	if brandName == "" {
		brandName = "OpenPanel"
	}
	return webauthn.New(&webauthn.Config{
		RPID:          rpID,
		RPDisplayName: brandName,
		RPOrigins:     []string{origin},
	})
}

// passkeyUser adapts a user row to go-webauthn's webauthn.User interface.
// WebAuthnID intentionally returns the decimal-ASCII encoding of the
// numeric user id rather than the library's recommended random opaque
// handle - the handle is embedded by the authenticator into
// already-registered credentials, so changing this scheme would break
// every passkey a user already registered.
type passkeyUser struct {
	userID      int
	username    string
	credentials []webauthn.Credential
}

func (u *passkeyUser) WebAuthnID() []byte                         { return []byte(strconv.Itoa(u.userID)) }
func (u *passkeyUser) WebAuthnName() string                       { return u.username }
func (u *passkeyUser) WebAuthnDisplayName() string                { return u.username }
func (u *passkeyUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

// passkeyRow is one row of user_passkeys.
type passkeyRow struct {
	ID           int
	Name         string
	CreatedAt    string
	LastUsedAt   sql.NullString
	CredentialID string
	PublicKey    string
	SignCount    uint32
}

// getPasskeysForUser returns id, name, created_at, last_used_at only - the
// shape the settings page table needs.
func getPasskeysForUser(ctx context.Context, a *appctx.App, userID int) ([]passkeyRow, error) {
	if err := ensurePasskeysTable(ctx, a); err != nil {
		return nil, err
	}
	rows, err := a.DB.QueryContext(ctx,
		"SELECT id, name, created_at, last_used_at FROM user_passkeys WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []passkeyRow
	for rows.Next() {
		var pk passkeyRow
		if err := rows.Scan(&pk.ID, &pk.Name, &pk.CreatedAt, &pk.LastUsedAt); err != nil {
			return nil, err
		}
		result = append(result, pk)
	}
	return result, rows.Err()
}

// decodeCredentialID decodes the raw credential ID bytes stored (as a
// base64url string, no padding) in user_passkeys.credential_id.
func decodeCredentialID(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func encodeCredentialID(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
