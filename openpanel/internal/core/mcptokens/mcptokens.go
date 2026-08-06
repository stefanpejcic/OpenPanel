// Package mcptokens implements long-lived bearer tokens used to
// authenticate MCP clients (and other non-interactive /api/ consumers) via
// the Authorization header, as an alternative to session cookies or
// short-lived JWTs.
package mcptokens

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"
)

const TokenPrefix = "op_mcp_"

const createTableSQL = `
CREATE TABLE IF NOT EXISTS mcp_tokens (
	id INT AUTO_INCREMENT PRIMARY KEY,
	user_id INT NOT NULL,
	name VARCHAR(100) NOT NULL,
	token_prefix VARCHAR(20) NOT NULL,
	token_hash CHAR(64) NOT NULL UNIQUE,
	read_only TINYINT(1) NOT NULL DEFAULT 0,
	expires_at TIMESTAMP NULL DEFAULT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	last_used_at TIMESTAMP NULL,
	INDEX idx_user_id (user_id)
) ENGINE=InnoDB`

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func EnsureTable(db *sql.DB) error {
	_, err := db.Exec(createTableSQL)
	return err
}

type AuthResult struct {
	UserID   int
	ReadOnly bool
}

// Authenticate looks up rawToken by hash, rejects/deletes it if expired,
// and bumps last_used_at on success.
func Authenticate(db *sql.DB, rawToken string) (*AuthResult, error) {
	if rawToken == "" || !strings.HasPrefix(rawToken, TokenPrefix) {
		return nil, nil
	}
	if err := EnsureTable(db); err != nil {
		return nil, err
	}

	var (
		tokenID   int
		userID    int
		readOnly  bool
		expiresAt sql.NullTime
	)
	row := db.QueryRow(
		"SELECT id, user_id, read_only, expires_at FROM mcp_tokens WHERE token_hash = ?",
		hashToken(rawToken),
	)
	if err := row.Scan(&tokenID, &userID, &readOnly, &expiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		_, _ = db.Exec("DELETE FROM mcp_tokens WHERE id = ?", tokenID)
		return nil, nil
	}

	if _, err := db.Exec("UPDATE mcp_tokens SET last_used_at = NOW() WHERE id = ?", tokenID); err != nil {
		return nil, err
	}

	return &AuthResult{UserID: userID, ReadOnly: readOnly}, nil
}

// Token is one row of mcp_tokens, the shape the /account/mcp settings page
// needs (never the token itself, which is only ever known at creation
// time - only its hash is stored).
type Token struct {
	ID          int
	Name        string
	TokenPrefix string
	CreatedAt   string
	LastUsedAt  sql.NullString
	ReadOnly    bool
	ExpiresAt   sql.NullString
}

// GetTokensForUser returns every token row belonging to userID.
func GetTokensForUser(ctx context.Context, db *sql.DB, userID int) ([]Token, error) {
	if err := EnsureTable(db); err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx,
		"SELECT id, name, token_prefix, created_at, last_used_at, read_only, expires_at FROM mcp_tokens WHERE user_id = ? ORDER BY created_at DESC",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		var t Token
		if err := rows.Scan(&t.ID, &t.Name, &t.TokenPrefix, &t.CreatedAt, &t.LastUsedAt, &t.ReadOnly, &t.ExpiresAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// randomURLSafeToken generates nBytes random bytes, base64url-encoded
// without padding.
func randomURLSafeToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// CreateTokenForUser generates and persists a new token, returning the raw
// token (shown to the caller exactly once) or an error if it couldn't be
// persisted.
func CreateTokenForUser(ctx context.Context, db *sql.DB, userID int, name string, readOnly bool, expiresInDays int) (string, error) {
	suffix, err := randomURLSafeToken(32)
	if err != nil {
		return "", err
	}
	rawToken := TokenPrefix + suffix
	tokenPrefix := rawToken[:len(TokenPrefix)+6]

	var expiresAt sql.NullTime
	if expiresInDays > 0 {
		expiresAt = sql.NullTime{Time: time.Now().Add(time.Duration(expiresInDays) * 24 * time.Hour), Valid: true}
	}

	if err := EnsureTable(db); err != nil {
		return "", err
	}
	_, err = db.ExecContext(ctx,
		"INSERT INTO mcp_tokens (user_id, name, token_prefix, token_hash, read_only, expires_at) VALUES (?, ?, ?, ?, ?, ?)",
		userID, name, tokenPrefix, hashToken(rawToken), readOnly, expiresAt)
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

// RevokeToken deletes the token row matching tokenID and userID, returning
// whether a matching row was deleted (false means "not found", not an
// error).
func RevokeToken(ctx context.Context, db *sql.DB, tokenID, userID int) (bool, error) {
	if err := EnsureTable(db); err != nil {
		return false, err
	}
	result, err := db.ExecContext(ctx, "DELETE FROM mcp_tokens WHERE id = ? AND user_id = ?", tokenID, userID)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
