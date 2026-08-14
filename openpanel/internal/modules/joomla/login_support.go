package joomla

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
// one silently becomes "" rather than a compile error, which broke WP's
// autologin token flow the same way once already).
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

// extractJoomlaDatabaseInfoForLogin is a local copy of
// websites.extractJoomlaDatabaseInfo (unexported in another package, so
// duplicated here - same small-helper-duplication pattern wordpress/drupal
// already use rather than sharing across module packages). Only the fields
// handleJoomlaLogin actually needs are populated.
func extractJoomlaDatabaseInfoForLogin(userContext, directory string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(directory, wwwPrefix) {
		return map[string]string{"error": "invalid docroot"}
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(directory, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "configuration.php"))
	if err != nil {
		return map[string]string{"error": "configuration.php not found"}
	}
	text := string(content)

	info := map[string]string{}
	for key, field := range map[string]string{
		"database_name": "db", "database_prefix": "dbprefix",
	} {
		re := regexp.MustCompile(`\$` + field + `\s*=\s*'([^']*)'`)
		if m := re.FindStringSubmatch(text); m != nil {
			info[key] = m[1]
		}
	}
	if info["database_name"] == "" || info["database_prefix"] == "" {
		return map[string]string{"error": "No database information found in configuration.php"}
	}
	return info
}
