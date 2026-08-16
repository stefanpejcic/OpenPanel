// Package moodle installs and manages a Moodle LMS site (downloaded from
// download.moodle.org's packaged tarballs + Moodle's own genuine
// non-interactive `admin/cli/install.php` CLI installer) inside an existing
// domain's docroot, run in the domain's existing php-fpm container - same
// shape as internal/modules/prestashop and internal/modules/matomo.
//
// Moodle 5.x restructured its tree so the actual web-served files live
// under a `public/` subdirectory, with config.php, admin/, lib/, etc. one
// level above it (outside the web root, by design - public/config.php is a
// tiny shim that does `require __DIR__.'/../config.php'`). This module
// extracts the full release into a sibling "app root" directory next to the
// domain's docroot, then replaces the (otherwise-empty) docroot with a
// symlink to <approot>/public - the docroot the panel/webserver already
// serves ends up being exactly Moodle's own public/ folder, and the shim's
// relative `../config.php` resolves straight to the approot's config.php.
// This is the standard supported deployment shape for Moodle 5.x (not an
// OpenPanel-specific workaround).
//
// Moodle also requires a periodic `admin/cli/cron.php` run (nothing works
// without it - no emails, no enrolments, no scheduled tasks): install.go
// registers a per-minute job via internal/modules/crons.AddJob, and
// manage.go's uninstall handler removes it via crons.RemoveJobByComment.
package moodle

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

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

const randomStringAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateRandomString generates a throwaway db name/user/password when
// needed. Uses crypto/rand since results end up as real credentials (same
// approach as every other CMS module's identical helper).
func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(randomStringAlphabet))))
		b[i] = randomStringAlphabet[n.Int64()]
	}
	return string(b)
}

// lockFilePath returns the per-user krompir.lock path shared with every
// other "app install" module to serialize one install operation per user.
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

// countUserWebsites counts the user's sites, capped at 1000 - duplicated
// locally per established convention (every CMS module does this rather
// than sharing across packages).
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

// siteSlug turns a "domain.com/sub/dir"-style site name into a filesystem-
// and cron-comment-safe token, used to name the approot/dataroot sibling
// directories and the cron job's unique comment.
func siteSlug(selectedDomain string) string {
	slug := strings.ReplaceAll(selectedDomain, "/", "_")
	slug = strings.ReplaceAll(slug, ".", "_")
	return slug
}

// moodleCronComment is the crons.ini comment used both by install.go's
// crons.AddJob call and manage.go's crons.RemoveJobByComment call - it must
// be derived identically in both places to actually find/remove the same
// job later.
func moodleCronComment(selectedDomain string) string {
	return "moodle-" + siteSlug(selectedDomain)
}
