// Package joomla installs and manages a Joomla site (downloaded from
// GitHub releases + Joomla's own CLI installer, `installation/joomla.php
// install`) inside an existing domain's docroot, run in the domain's
// existing php-fpm container - same shape as internal/modules/drupal.
// Deliberately minimal, matching drupal's scope: no cloning, no
// scan-for-existing-installs, no hardening rules, no dedicated
// backup/restore system - just install, a small read-only manage/overview
// page (files, database, PHP/DB versions, screenshot), a Logs tab, cache
// clearing, a one-time admin login link, and uninstall. MySQL/MariaDB only.
package joomla

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"math/big"
	"net/http"
	"os"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

const randomStringAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateRandomString generates a throwaway db name/user/password or login
// token when needed. Uses crypto/rand since results end up as real
// credentials/tokens (same approach as drupal.generateRandomString).
func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(randomStringAlphabet))))
		b[i] = randomStringAlphabet[n.Int64()]
	}
	return string(b)
}

func injected(a *appctx.App, r *http.Request) (userID int, username, userContext string, err error) {
	userID, _ = auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return userID, "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return userID, username, userContext, nil
}

func flashSess(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	flashSess(a, w, r, category, message)
	http.Redirect(w, r, path, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeNDJSON(w http.ResponseWriter, flusher http.Flusher, canFlush bool, v map[string]any) {
	b, _ := json.Marshal(v)
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n"))
	if canFlush {
		flusher.Flush()
	}
}

// lockFilePath returns the per-user krompir.lock path shared with the
// wordpress/phpapp/drupal modules to serialize any one "app install"
// operation per user at a time - not a Joomla-specific lock.
func lockFilePath(username string) string {
	return "/etc/openpanel/openpanel/core/users/" + username + "/krompir.lock"
}

func createLockFile(username string) error {
	dir := "/etc/openpanel/openpanel/core/users/" + username
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(lockFilePath(username), nil, 0o644)
}

func removeLockFile(username string) {
	_ = os.Remove(lockFilePath(username))
}

// domainRow is the shape install/remove read from the domains table.
type domainRow struct {
	DomainURL  string
	Docroot    sql.NullString
	PHPVersion sql.NullString
}

func lookupDomainByID(ctx context.Context, a *appctx.App, domainID string) (domainRow, bool, error) {
	var d domainRow
	row := a.DB.QueryRowContext(ctx, "SELECT domain_url, docroot, php_version FROM domains WHERE domain_id = ?", domainID)
	err := row.Scan(&d.DomainURL, &d.Docroot, &d.PHPVersion)
	if err == sql.ErrNoRows {
		return domainRow{}, false, nil
	}
	if err != nil {
		return domainRow{}, false, err
	}
	return d, true, nil
}

// countUserWebsites counts the user's sites, capped at 1000 - same query
// wordpress/phpapp/drupal each duplicate locally rather than sharing across
// packages for something this small.
func countUserWebsites(a *appctx.App, userID int) (int, error) {
	rows, err := a.DB.Query(
		"SELECT site_name FROM sites WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?) LIMIT 1000", userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		n++
	}
	return n, rows.Err()
}

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}
