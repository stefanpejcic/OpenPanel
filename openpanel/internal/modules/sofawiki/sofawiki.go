// Package sofawiki installs and manages SofaWiki (github.com/bellenuit/sofawiki)
// inside an existing domain's docroot, run in the domain's existing php-fpm
// container - same shape as internal/modules/drupal and internal/modules/flarum,
// but considerably simpler since SofaWiki needs no database at all.
//
// Install is a plain download-and-extract of the master branch (there are
// no tagged releases - confirmed against the repo - and no composer.json,
// so this isn't a Composer package): no database to provision, no admin
// account to create, no CLI installer to drive. SofaWiki's own install
// wizard (a 4-step browser flow: folder rights -> create configuration ->
// login -> write main page, found by inspecting a live instance's output -
// see install.go) is left for the site owner to complete themselves on
// first visit, exactly as it would if they'd uploaded the files by FTP.
//
// PHP compatibility is the real catch, confirmed by running a live copy
// under every php-fpm version this box has: PHP 7.4 and below work (with
// deprecation noise from the PHP4/5-era code); PHP 8.0 and above throw a
// fatal error in inc/async.php (fwrite() on a failed fsockopen() started
// returning false instead of a resource). install.go refuses to install
// onto a domain configured for PHP 8.0+.
//
// Drupal-parity features NOT implemented here, since there's nothing to
// hook them to:
//   - Maintenance mode, admin auto-login, cache-clear: no such concepts
//     exist in SofaWiki at all.
//   - Version tracking: no tagged releases exist upstream, so there is no
//     "installed version" to report or compare against a "latest" - the
//     manager page just always shows the branch install was made from.
package sofawiki

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

// generateRandomString is used only for the (rare) target_db-shaped clone
// form fields cmsclone-style modules conventionally take - SofaWiki has no
// database, but keeping the same helper name/behavior as drupal/flarum
// keeps clone.go's shape consistent for anyone comparing the three.
func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(randomStringAlphabet))))
		b[i] = randomStringAlphabet[n.Int64()]
	}
	return string(b)
}

// lockFilePath returns the per-user krompir.lock path shared with the
// wordpress/phpapp/drupal/joomla/flarum modules to serialize any one "app
// install" operation per user at a time - not a SofaWiki-specific lock.
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

// countUserWebsites counts the user's sites, capped at 1000.
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
