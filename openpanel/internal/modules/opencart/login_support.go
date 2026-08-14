package opencart

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
// Mirrors wordpress/backups.go's identical helper - see that file's
// comment for why every numeric driver type needs its own case (a missing
// one silently becomes "" rather than a compile error).
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

// extractOpenCartDatabaseInfoForLogin is a local copy of
// websites.extractOpenCartDatabaseInfo (unexported in another package, so
// duplicated here - same small-helper-duplication pattern the other CMS
// modules already use rather than sharing across module packages). Only
// the field handleOpenCartLogin actually needs is populated.
func extractOpenCartDatabaseInfoForLogin(userContext, directory string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(directory, wwwPrefix) {
		return map[string]string{"error": "invalid docroot"}
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(directory, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "config.php"))
	if err != nil {
		return map[string]string{"error": "config.php not found"}
	}
	text := string(content)

	re := regexp.MustCompile(`DB_DATABASE'\s*,\s*'([^']*)'`)
	m := re.FindStringSubmatch(text)
	if m == nil {
		return map[string]string{"error": "No database information found in config.php"}
	}
	return map[string]string{"database_name": m[1]}
}
