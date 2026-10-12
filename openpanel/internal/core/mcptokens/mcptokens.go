// Package mcptokens implements long-lived bearer tokens for MCP clients and other non-interactive /api/ consumers, as an alternative to session cookies or JWTs.
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

const lastUsedInterval = time.Minute

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

type AuthResult struct {
	UserID   int
	ReadOnly bool
}

// Authenticate looks up rawToken by hash, rejects/deletes it if expired, and bumps last_used_at on success
func Authenticate(db *sql.DB, rawToken string) (*AuthResult, error) {
	if rawToken == "" || !strings.HasPrefix(rawToken, TokenPrefix) {
		return nil, nil
	}
	var (
		tokenID   int
		userID    int
		readOnly  bool
		expiresAt sql.NullTime
		lastUsed  sql.NullTime
	)
	row := db.QueryRow(
		"SELECT id, user_id, read_only, expires_at, last_used_at FROM mcp_tokens WHERE token_hash = ?",
		hashToken(rawToken),
	)
	if err := row.Scan(&tokenID, &userID, &readOnly, &expiresAt, &lastUsed); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		_, _ = db.Exec("DELETE FROM mcp_tokens WHERE id = ?", tokenID)
		return nil, nil
	}

	// only bump last_used_at once a minute, a busy MCP client would otherwise write on every call
	if !lastUsed.Valid || time.Since(lastUsed.Time) > lastUsedInterval {
		_, _ = db.Exec("UPDATE mcp_tokens SET last_used_at = NOW() WHERE id = ?", tokenID)
	}

	return &AuthResult{UserID: userID, ReadOnly: readOnly}, nil
}

// Token is one row of mcp_tokens, the shape the /account/mcp settings page needs - never the raw token itself, only its hash is stored
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

// randomURLSafeToken generates nBytes random bytes, base64url-encoded without padding
func randomURLSafeToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// CreateTokenForUser generates and persists a new token, returning the raw token (shown to the caller exactly once)
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

	_, err = db.ExecContext(ctx,
		"INSERT INTO mcp_tokens (user_id, name, token_prefix, token_hash, read_only, expires_at) VALUES (?, ?, ?, ?, ?, ?)",
		userID, name, tokenPrefix, hashToken(rawToken), readOnly, expiresAt)
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

// RevokeToken deletes the token row matching tokenID and userID; false means "not found", not an error
func RevokeToken(ctx context.Context, db *sql.DB, tokenID, userID int) (bool, error) {
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
