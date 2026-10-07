// Package wordpress handles WordPress site list/install/clone/remove/detach/reload/scan, backup listing/run/restore, the wp-cli passthrough endpoint, and the security-rules page. Drupal and Mautic are out of scope for this pass.
package wordpress

import (
	"context"
	"crypto/rand"
	"database/sql"
	"math/big"
	"regexp"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
)

// wordpressFiles lists every top-level file/dir a stock WordPress install creates, used by cleanup, remove, and detach to know what to delete.
var wordpressFiles = []string{
	".htaccess", "index.php", "license.txt", "readme.html", "wp-activate.php",
	"wp-admin", "wp-blog-header.php", "wp-comments-post.php", "wp-config-sample.php",
	"wp-config.php", "wp-content", "wp-cron.php", "wp-includes", "wp-links-opml.php",
	"wp-load.php", "wp-login.php", "wp-mail.php", "wp-settings.php", "wp-signup.php",
	"wp-trackback.php", "error_log", "xmlrpc.php",
}

// skipDirs are directories reload/scan never descend into while walking the html volume for wp-config.php files.
var skipDirs = map[string]bool{"wp-content": true, "node_modules": true, ".git": true, "backups": true}

const randomStringAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateRandomString generates a throwaway db name/user/password when the install/clone form leaves one blank, uses crypto/rand since the result ends up as a real database credential.
func generateRandomString(length int) string {
	return generateRandomStringFromAlphabet(length, randomStringAlphabet)
}

// generateRandomStringFromAlphabet is generateRandomString() parameterized over the character set - used by generateSaltsLocally() with WordPress's much wider salt alphabet (including punctuation).
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

// validateDomain/validateDB check a domain/db-name-like value for the restricted character set these routes accept as user input.
func validateDomain(name string) bool { return name != "" && validDomainRE.MatchString(name) }
func validateDB(name string) bool     { return name != "" && validDBRE.MatchString(name) }

func lookupDomainByURL(ctx context.Context, a *appctx.App, domainURL string) (appkit.DomainRow, bool, error) {
	var d appkit.DomainRow
	row := a.DB.QueryRowContext(ctx, "SELECT domain_url, docroot, php_version FROM domains WHERE domain_url = ?", domainURL)
	err := row.Scan(&d.DomainURL, &d.Docroot, &d.PHPVersion)
	if err == sql.ErrNoRows {
		return appkit.DomainRow{}, false, nil
	}
	if err != nil {
		return appkit.DomainRow{}, false, err
	}
	return d, true, nil
}
