// Package appkit holds the small helpers every one-click app installer shares
package appkit

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"math/big"
	"os"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
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

// LockFilePath is the per-user krompir.lock so only one app install runs per user at a time
func LockFilePath(username string) string {
	return "/etc/openpanel/openpanel/core/users/" + username + "/krompir.lock"
}

func CreateLockFile(username string) error {
	dir := "/etc/openpanel/openpanel/core/users/" + username
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(LockFilePath(username), nil, 0o644)
}

func RemoveLockFile(username string) {
	_ = os.Remove(LockFilePath(username))
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
