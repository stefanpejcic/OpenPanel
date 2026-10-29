// Package appkit holds the small helpers every one-click app installer shares
package appkit

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

const randomStringAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomString uses crypto/rand since results end up as real db credentials and login tokens
func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(randomStringAlphabet))))
		b[i] = randomStringAlphabet[n.Int64()]
	}
	return string(b)
}

// lockBaseDir is where per-user lock files live, tests point it at a temp dir
var lockBaseDir = "/etc/openpanel/openpanel/core/users/"

// LockFilePath is the per-user krompir.lock so only one app install runs per user at a time
func LockFilePath(username string) string {
	return lockBaseDir + username + "/krompir.lock"
}

// ErrInstallInProgress means another install or update already holds this user's lock
var ErrInstallInProgress = errors.New("another installation is already in progress for this account")

// a lock left behind by a crashed install would block every later one, so one older than this is treated as stale and taken over
const installLockStaleAfter = 15 * time.Minute

// CreateLockFile atomically takes the per-user install lock, returning ErrInstallInProgress if a fresh one is already held
func CreateLockFile(username string) error {
	dir := lockBaseDir + username
	if err := tenantfs.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := LockFilePath(username)
	if fi, err := tenantfs.Stat(path); err == nil && time.Since(fi.ModTime()) > installLockStaleAfter {
		_ = tenantfs.Remove(path)
	}
	f, err := tenantfs.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return ErrInstallInProgress
		}
		return err
	}
	return f.Close()
}

func RemoveLockFile(username string) {
	_ = tenantfs.Remove(LockFilePath(username))
}

const backupsHtaccess = "# added by OpenPanel: backups hold database dumps and copies of wp-config.php, never serve them\nRequire all denied\n"

// ProtectBackupsDir puts a deny-all .htaccess in htmlVolume/backups unless one is already there, so apache and litespeed won't serve dumps even if a docroot ends up pointing at that folder
func ProtectBackupsDir(htmlVolume string) {
	dir := filepath.Join(htmlVolume, "backups")
	path := filepath.Join(dir, ".htaccess")
	if _, err := tenantfs.Lstat(path); err == nil {
		return
	}
	if tenantfs.MkdirAll(dir, 0o755) != nil {
		return
	}
	_ = tenantfs.WriteFile(path, []byte(backupsHtaccess), 0o644)
}

// DomainRow is the shape installers read from the domains table
type DomainRow struct {
	DomainURL  string
	Docroot    sql.NullString
	PHPVersion sql.NullString
}

func LookupDomainByID(ctx context.Context, a *appctx.App, domainID string) (DomainRow, bool, error) {
	var d DomainRow
	row := a.DB.QueryRowContext(ctx, "SELECT domain_url, docroot, php_version FROM domains WHERE domain_id = ?", domainID)
	err := row.Scan(&d.DomainURL, &d.Docroot, &d.PHPVersion)
	if err == sql.ErrNoRows {
		return DomainRow{}, false, nil
	}
	if err != nil {
		return DomainRow{}, false, err
	}
	return d, true, nil
}

// CountUserWebsites counts the user's sites, capped at 1000
func CountUserWebsites(a *appctx.App, userID int) (int, error) {
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

// InvalidateMySQLCaches drops the cached database list/count after an install or remove touches them
func InvalidateMySQLCaches(ctx context.Context, a *appctx.App, userContext, currentUsername string) {
	_ = a.Cache.Delete(ctx, "databases_info:"+userContext)
	_ = a.Cache.Delete(ctx, "get_database_count:"+currentUsername)
}

// SiteSlug turns "domain.tld/sub" into a filesystem-safe "domain_tld_sub"
func SiteSlug(selectedDomain string) string {
	slug := strings.ReplaceAll(selectedDomain, "/", "_")
	slug = strings.ReplaceAll(slug, ".", "_")
	return slug
}

func SHA256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

var safeDocrootRE = regexp.MustCompile(`^/var/www/html/[A-Za-z0-9._/-]+$`)

// SafeDocroot is true for a container path under /var/www/html/ that can't climb out of it or break out of a shell command
func SafeDocroot(docroot string) bool {
	return safeDocrootRE.MatchString(docroot) && !strings.Contains(docroot, "..")
}
