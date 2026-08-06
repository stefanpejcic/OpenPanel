// Package wordpress handles WordPress site
// list/install/clone/remove/detach/reload/scan, backup listing/run/restore,
// the wp-cli passthrough endpoint, and the security-rules page. Drupal and
// Mautic are out of scope for this pass.
package wordpress

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

// wordpressFiles lists every top-level file/dir a stock WordPress
// install creates, used by cleanup, remove, and detach to know what to
// delete.
var wordpressFiles = []string{
	".htaccess", "index.php", "license.txt", "readme.html", "wp-activate.php",
	"wp-admin", "wp-blog-header.php", "wp-comments-post.php", "wp-config-sample.php",
	"wp-config.php", "wp-content", "wp-cron.php", "wp-includes", "wp-links-opml.php",
	"wp-load.php", "wp-login.php", "wp-mail.php", "wp-settings.php", "wp-signup.php",
	"wp-trackback.php", "error_log", "xmlrpc.php",
}

// skipDirs are directories reload/scan never descend into while walking
// the html volume for wp-config.php files.
var skipDirs = map[string]bool{"wp-content": true, "node_modules": true, ".git": true, "backups": true}

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

const randomStringAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateRandomString generates a throwaway db name/user/password when
// the install/clone form leaves one blank. Uses crypto/rand rather than a
// non-cryptographic RNG since the result ends up as a real database
// credential.
func generateRandomString(length int) string {
	return generateRandomStringFromAlphabet(length, randomStringAlphabet)
}

// generateRandomStringFromAlphabet is generateRandomString() parameterized
// over the character set - used by generateSaltsLocally() with WordPress's
// much wider salt alphabet (including punctuation).
func generateRandomStringFromAlphabet(length int, alphabet string) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		b[i] = alphabet[n.Int64()]
	}
	return string(b)
}

var (
	validDomainRE = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	validDBRE     = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

// validateDomain/validateDB check a domain/db-name-like value for the
// restricted character set these routes accept as user input.
func validateDomain(name string) bool { return name != "" && validDomainRE.MatchString(name) }
func validateDB(name string) bool     { return name != "" && validDBRE.MatchString(name) }

// validateDocroot rejects path traversal and a leading slash, since the
// value is joined onto the account's html volume root.
func validateDocroot(path string) bool {
	return path != "" && !strings.Contains(path, "..") && !strings.HasPrefix(path, "/")
}

// lockFilePath returns the per-user krompir.lock path used to serialize WordPress operations.
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

// domainRow is the shape every route reads from the domains table.
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

func lookupDomainByURL(ctx context.Context, a *appctx.App, domainURL string) (domainRow, bool, error) {
	var d domainRow
	row := a.DB.QueryRowContext(ctx, "SELECT domain_url, docroot, php_version FROM domains WHERE domain_url = ?", domainURL)
	err := row.Scan(&d.DomainURL, &d.Docroot, &d.PHPVersion)
	if err == sql.ErrNoRows {
		return domainRow{}, false, nil
	}
	if err != nil {
		return domainRow{}, false, err
	}
	return d, true, nil
}

// countUserWebsites counts the user's sites, capped at 1000. This is the
// same query used by internal/modules/appinstall, duplicated here since
// that package doesn't export it - a tiny, one-line query, so a small
// per-package duplicate beats cross-package coupling for something this
// small.
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
