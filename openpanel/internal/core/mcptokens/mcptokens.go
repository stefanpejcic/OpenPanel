// Package mcptokens implements long-lived bearer tokens for MCP clients and other non-interactive /api/ consumers, as an alternative to session cookies or JWTs.
package mcptokens

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

const TokenPrefix = "op_mcp_"

const lastUsedInterval = time.Minute

// what a token may do: readonly looks things up, modify can't run destructive routes, full is everything the plan allows
const (
	ModeReadOnly = "readonly"
	ModeModify   = "modify"
	ModeFull     = "full"
)

// ErrScopesUnavailable means the mode/allowed_features columns couldn't be added, so only plain tokens can be made
var ErrScopesUnavailable = errors.New("token scopes need the mode and allowed_features columns on mcp_tokens")

// Scope is what a token is limited to, no Features means every feature the plan has
type Scope struct {
	Mode     string
	Features []string
}

// Limited is true for anything narrower than a plain full-access token
func (s Scope) Limited() bool {
	return s.Mode != ModeFull || len(s.Features) > 0
}

// ParseMode maps a stored or submitted mode to a known one, unknown values fail closed to read-only
func ParseMode(mode string, readOnly bool) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "":
		if readOnly {
			return ModeReadOnly
		}
		return ModeFull
	case ModeReadOnly, "read-only", "read_only":
		return ModeReadOnly
	case ModeModify:
		return ModeModify
	case ModeFull:
		return ModeFull
	}
	return ModeReadOnly
}

// ValidMode reports whether mode is one of the three, for rejecting bad API input instead of guessing
func ValidMode(mode string) bool {
	switch mode {
	case ModeReadOnly, ModeModify, ModeFull:
		return true
	}
	return false
}

// CleanFeatures trims, dedupes and sorts a feature list
func CleanFeatures(features []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range features {
		f = strings.TrimSpace(f)
		if f == "" || seen[f] || strings.ContainsAny(f, ", ") {
			continue
		}
		seen[f] = true
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

func joinFeatures(features []string) sql.NullString {
	features = CleanFeatures(features)
	if len(features) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: strings.Join(features, ","), Valid: true}
}

func splitFeatures(raw sql.NullString) []string {
	if !raw.Valid || raw.String == "" {
		return nil
	}
	return CleanFeatures(strings.Split(raw.String, ","))
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// schemaState remembers per db whether the scope columns exist, and when we last tried to add them
type schemaState struct {
	ready     bool
	lastCheck time.Time
}

var (
	schemaMu    sync.Mutex
	schemaByDB  = map[*sql.DB]*schemaState{}
	schemaRetry = time.Minute
)

func scopeColumnCount(ctx context.Context, db *sql.DB) int {
	var n int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'mcp_tokens' AND COLUMN_NAME IN ('mode', 'allowed_features')").Scan(&n)
	if err != nil {
		return -1
	}
	return n
}

// ScopesReady adds the mode and allowed_features columns on first use, plain ADD COLUMN so it works on both MySQL and MariaDB
func ScopesReady(ctx context.Context, db *sql.DB) bool {
	schemaMu.Lock()
	defer schemaMu.Unlock()
	st := schemaByDB[db]
	if st == nil {
		st = &schemaState{}
		schemaByDB[db] = st
	}
	if st.ready || (!st.lastCheck.IsZero() && time.Since(st.lastCheck) < schemaRetry) {
		return st.ready
	}
	st.lastCheck = time.Now()

	if scopeColumnCount(ctx, db) == 2 {
		st.ready = true
		return true
	}
	for col, def := range map[string]string{"mode": "VARCHAR(10) NULL DEFAULT NULL", "allowed_features": "TEXT NULL DEFAULT NULL"} {
		var has int
		if err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'mcp_tokens' AND COLUMN_NAME = ?", col).Scan(&has); err != nil || has > 0 {
			continue
		}
		if _, err := db.ExecContext(ctx, "ALTER TABLE mcp_tokens ADD COLUMN "+col+" "+def); err != nil {
			log.Printf("MCP - could not add mcp_tokens.%s, scoped tokens are off until it exists: %v", col, err)
		}
	}
	st.ready = scopeColumnCount(ctx, db) == 2
	return st.ready
}

type AuthResult struct {
	UserID   int
	TokenID  int
	Name     string
	ReadOnly bool
	Scope    Scope
}

// Authenticate looks up rawToken by hash, rejects/deletes it if expired, and bumps last_used_at on success
func Authenticate(db *sql.DB, rawToken string) (*AuthResult, error) {
	if rawToken == "" || !strings.HasPrefix(rawToken, TokenPrefix) {
		return nil, nil
	}
	var (
		tokenID   int
		userID    int
		name      string
		readOnly  bool
		mode      sql.NullString
		features  sql.NullString
		expiresAt sql.NullTime
		lastUsed  sql.NullTime
	)
	var err error
	if ScopesReady(context.Background(), db) {
		err = db.QueryRow(
			"SELECT id, user_id, name, read_only, mode, allowed_features, expires_at, last_used_at FROM mcp_tokens WHERE token_hash = ?",
			hashToken(rawToken),
		).Scan(&tokenID, &userID, &name, &readOnly, &mode, &features, &expiresAt, &lastUsed)
	} else {
		err = db.QueryRow(
			"SELECT id, user_id, read_only, expires_at, last_used_at FROM mcp_tokens WHERE token_hash = ?",
			hashToken(rawToken),
		).Scan(&tokenID, &userID, &readOnly, &expiresAt, &lastUsed)
	}
	if err != nil {
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

	scope := Scope{Mode: ParseMode(mode.String, readOnly), Features: splitFeatures(features)}
	return &AuthResult{UserID: userID, TokenID: tokenID, Name: name, ReadOnly: scope.Mode == ModeReadOnly, Scope: scope}, nil
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
	Mode        string
	Features    []string
}

// GetTokensForUser returns every token row belonging to userID.
func GetTokensForUser(ctx context.Context, db *sql.DB, userID int) ([]Token, error) {
	scoped := ScopesReady(ctx, db)
	query := "SELECT id, name, token_prefix, created_at, last_used_at, read_only, expires_at FROM mcp_tokens WHERE user_id = ? ORDER BY created_at DESC"
	if scoped {
		query = "SELECT id, name, token_prefix, created_at, last_used_at, read_only, expires_at, mode, allowed_features FROM mcp_tokens WHERE user_id = ? ORDER BY created_at DESC"
	}
	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		var t Token
		var mode, features sql.NullString
		dest := []any{&t.ID, &t.Name, &t.TokenPrefix, &t.CreatedAt, &t.LastUsedAt, &t.ReadOnly, &t.ExpiresAt}
		if scoped {
			dest = append(dest, &mode, &features)
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		t.Mode = ParseMode(mode.String, t.ReadOnly)
		t.ReadOnly = t.Mode == ModeReadOnly
		t.Features = splitFeatures(features)
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
func CreateTokenForUser(ctx context.Context, db *sql.DB, userID int, name string, scope Scope, expiresInDays int) (string, error) {
	scope.Mode = ParseMode(scope.Mode, false)
	scoped := ScopesReady(ctx, db)
	// a plain read-only or full token still fits in read_only alone
	if !scoped && (scope.Mode == ModeModify || len(scope.Features) > 0) {
		return "", ErrScopesUnavailable
	}

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

	// read_only stays in step with mode so an older panel binary reading this row still gets it right
	readOnly := scope.Mode == ModeReadOnly
	if scoped {
		_, err = db.ExecContext(ctx,
			"INSERT INTO mcp_tokens (user_id, name, token_prefix, token_hash, read_only, expires_at, mode, allowed_features) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			userID, name, tokenPrefix, hashToken(rawToken), readOnly, expiresAt, scope.Mode, joinFeatures(scope.Features))
	} else {
		_, err = db.ExecContext(ctx,
			"INSERT INTO mcp_tokens (user_id, name, token_prefix, token_hash, read_only, expires_at) VALUES (?, ?, ?, ?, ?, ?)",
			userID, name, tokenPrefix, hashToken(rawToken), readOnly, expiresAt)
	}
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

// RevokeAllForUser deletes every token the user has, returns how many were removed
func RevokeAllForUser(ctx context.Context, db *sql.DB, userID int) (int64, error) {
	result, err := db.ExecContext(ctx, "DELETE FROM mcp_tokens WHERE user_id = ?", userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
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
