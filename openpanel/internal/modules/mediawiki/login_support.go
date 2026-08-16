package mediawiki

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// toStringCell converts one mysqlmanager.Exec() result cell to a string.
// Mirrors every other CMS module's identical helper - see
// wordpress/backups.go's comment for why every numeric driver type needs
// its own case (a missing one silently becomes "" rather than a compile
// error).
func toStringCell(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case uint64:
		return strconv.FormatUint(t, 10)
	case int:
		return strconv.Itoa(t)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return ""
	}
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func itoa(n int) string { return strconv.Itoa(n) }

var (
	mediawikiDBNameRE   = regexp.MustCompile(`\$wgDBname\s*=\s*"([^"]*)"`)
	mediawikiDBUserRE   = regexp.MustCompile(`\$wgDBuser\s*=\s*"([^"]*)"`)
	mediawikiDBPrefixRE = regexp.MustCompile(`\$wgDBprefix\s*=\s*"([^"]*)"`)
)

// extractMediaWikiDatabaseInfoForLogin is a local copy of
// websites.extractMediaWikiDatabaseInfo (unexported in another package, so
// duplicated here - same small-helper-duplication pattern every other CMS
// module already uses rather than sharing across module packages). Reads
// LocalSettings.php directly, keyed off the docroot the same way
// install.go/manage.go derive it.
func extractMediaWikiDatabaseInfoForLogin(userContext, directory string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(directory, wwwPrefix) {
		return map[string]string{"error": "invalid docroot"}
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(directory, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "LocalSettings.php"))
	if err != nil {
		return map[string]string{"error": "LocalSettings.php not found"}
	}
	text := string(content)

	nameMatch := mediawikiDBNameRE.FindStringSubmatch(text)
	userMatch := mediawikiDBUserRE.FindStringSubmatch(text)
	if nameMatch == nil || userMatch == nil {
		return map[string]string{"error": "No database information found in LocalSettings.php"}
	}
	info := map[string]string{"database_name": nameMatch[1], "database_user": userMatch[1], "database_prefix": ""}
	if prefixMatch := mediawikiDBPrefixRE.FindStringSubmatch(text); prefixMatch != nil {
		info["database_prefix"] = prefixMatch[1]
	}
	return info
}
