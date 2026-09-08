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

// rpIDAndOrigin computes the WebAuthn Relying Party ID (host without port) and origin (scheme+host), which must match the browser's clientDataJSON.origin byte-for-byte. Trusts X-Forwarded-Proto first since this backend usually sits behind a Caddy TLS-terminating proxy.
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

// isIPHost reports whether host (already stripped of any port) is a literal IP address rather than a domain name
func isIPHost(host string) bool {
	return net.ParseIP(host) != nil
}

// webAuthnUnavailableReason reports why WebAuthn won't work for this request, or "" if it should work fine - browsers only expose navigator.credentials over HTTPS/localhost, and refuse an IP-address domain, so we explain that up front instead of a cryptic "navigator.credentials is undefined" in the console
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

// newWebAuthnForRequest builds a *webauthn.WebAuthn scoped to the current request's host, recomputed on every call since the RP ID/origin depend on it
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

// passkeyUser adapts a user row to go-webauthn's webauthn.User interface. WebAuthnID returns the decimal-ASCII user id rather than a random opaque handle - changing this would break every already-registered passkey.
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

// getPasskeysForUser returns id, name, created_at, last_used_at only - the shape the settings page table needs
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

// decodeCredentialID decodes the raw credential ID bytes stored as a base64url string (no padding) in user_passkeys.credential_id
func decodeCredentialID(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func encodeCredentialID(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
