// Package ojs installs and manages Open Journal Systems (OJS, pkp/ojs)
// sites inside an existing domain's docroot, run in the domain's existing
// php-fpm container - same overall shape as internal/modules/moodle.
//
// OJS ships no ready-made release tarball on GitHub itself: the repo uses
// git submodules (lib/pkp, several plugins, ui-library) that a plain GitHub
// archive/zip silently omits (confirmed: .gitmodules present at every
// released tag), which would produce a broken install. The real
// already-bundled release packages are hosted directly by PKP at
// https://pkp.sfu.ca/ojs/download/ojs-{dotted-version}.tar.gz (see
// version.go).
//
// Like Moodle, OJS needs its file-storage directory (submission uploads
// etc.) to live outside the public web root - unlike Moodle, though, OJS
// itself is *not* split into an approot+public/ pair; the release tarball's
// top-level directory (after stripping the wrapper) is itself the full web
// root (index.php, tools/, lib/, etc. all live there, same flat shape as
// this codebase's joomla module). So this module still uses a Moodle-style
// "app root" sibling directory + docroot symlink (not a flat "extract
// straight into docroot" layout like Joomla) purely to give update.go an
// atomic swap-and-rollback target (extract the new version into a fresh
// sibling directory, then repoint the symlink - see update.go), with a
// second, separate sibling directory ("_ojsfiles") for the files_dir OJS
// installer prompts for, kept outside the docroot/approot tree entirely so
// it's never web-accessible.
package ojs

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
// and cron-comment-safe token, used to name the approot/files sibling
// directories and the cron job's unique comment.
func siteSlug(selectedDomain string) string {
	slug := strings.ReplaceAll(selectedDomain, "/", "_")
	slug = strings.ReplaceAll(slug, ".", "_")
	return slug
}

// ojsCronComment is the crons.ini comment used both by install.go's
// crons.AddJob call and manage.go's crons.RemoveJobByComment call - it must
// be derived identically in both places to actually find/remove the same
// job later. Registers lib/pkp/tools/scheduler.php run, the documented way
// to run OJS's scheduled tasks outside of its (discouraged for real
// traffic) built-in end-of-request task runner.
func ojsCronComment(selectedDomain string) string {
	return "ojs-" + siteSlug(selectedDomain)
}
