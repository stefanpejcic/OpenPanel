package account

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cpanelpw"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/werkzeugpw"
)

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

type loginResult struct {
	UserID       int
	Username     string
	TwofaEnabled bool
}

// verifyPassword looks up the user and verifies their password against
// whichever hash scheme is stored (cPanel SHA-512 crypt, CyberPanel
// SHA-256+salt, or the werkzeugpw scheme), transparently upgrading legacy
// hashes to the current scheme on success. Returns a non-empty errMsg
// (already translated) on any failure.
func verifyPassword(a *appctx.App, ctx context.Context, username, password string, t i18n.Translator) (loginResult, string) {
	var (
		userID int
		twofa  sql.NullBool   // twofa_enabled: NULL for most accounts, not just 0/1
		otp    sql.NullString // otp_secret: NULL until 2FA is set up
		pwHash string
	)

	row := a.DB.QueryRowContext(ctx,
		"SELECT id, twofa_enabled, otp_secret, password FROM users WHERE username = ?", username)
	err := row.Scan(&userID, &twofa, &otp, &pwHash)
	if err == sql.ErrNoRows {
		var suspendedID int
		suspRow := a.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE username LIKE ?", "SUSPENDED_%"+username)
		if suspRow.Scan(&suspendedID) == nil {
			return loginResult{}, t.Get("Your account is suspended. Please contact support.")
		}
		return loginResult{}, t.Get("Unrecognized account. Please check username.")
	}
	if err != nil {
		log.Printf("LOGIN - database error looking up user %q: %v", username, err)
		return loginResult{}, t.Get("Unable to connect to database.")
	}

	valid := false

	switch {
	case strings.HasPrefix(pwHash, "$6$"):
		valid = cpanelpw.VerifySHA512Crypt(password, pwHash)
		if valid {
			upgradeHash(a, ctx, userID, password)
		}

	case isCyberPanelHash(pwHash):
		valid = verifyCyberPanelPassword(password, pwHash)
		if valid {
			upgradeHash(a, ctx, userID, password)
		}

	default:
		valid = werkzeugpw.CheckPasswordHash(pwHash, password)
	}

	if !valid {
		return loginResult{}, t.Get("Invalid password. Please try again.")
	}

	return loginResult{UserID: userID, Username: username, TwofaEnabled: twofa.Bool}, ""
}

// isCyberPanelHash reports whether pwHash looks like a CyberPanel-style
// hash: a colon-separated hash part followed by a salt, where the hash
// part is exactly 64 characters (a hex SHA-256 digest).
func isCyberPanelHash(pwHash string) bool {
	before, _, found := strings.Cut(pwHash, ":")
	return found && len(before) == 64
}

// verifyCyberPanelPassword computes SHA-256(password+salt) hex and compares
// it against the stored hash part.
func verifyCyberPanelPassword(password, storedHash string) bool {
	hashPart, salt, ok := strings.Cut(storedHash, ":")
	if !ok || len(hashPart) != 64 || len(salt) != 32 {
		return false
	}
	return sha256Hex(password+salt) == hashPart
}

func upgradeHash(a *appctx.App, ctx context.Context, userID int, password string) {
	newHash, err := werkzeugpw.GeneratePasswordHash(password)
	if err != nil {
		return
	}
	_, _ = a.DB.ExecContext(ctx, "UPDATE users SET password = ? WHERE id = ?", newHash, userID)
}
