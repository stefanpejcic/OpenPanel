package moodle

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
	moodleDBNameRE   = regexp.MustCompile(`CFG->dbname\s*=\s*'([^']*)'`)
	moodleDBPrefixRE = regexp.MustCompile(`CFG->prefix\s*=\s*'([^']*)'`)
)

// extractMoodleDatabaseInfoForBackup is a local copy of
// websites.extractMoodleDatabaseInfo (unexported in another package, so
// duplicated here - same small-helper-duplication pattern every other CMS
// module already uses rather than sharing across module packages). Reads
// the approot's config.php directly (docroot is a symlink to
// <approot>/public, not where config.php lives - see moodle.go's package
// doc comment), keyed off domain the same way install.go/manage.go derive
// the approot directory.
func extractMoodleDatabaseInfoForBackup(userContext, domain string) map[string]string {
	approotHostPath := filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", siteSlug(domain)+"_moodleapp")
	content, err := os.ReadFile(filepath.Join(approotHostPath, "config.php"))
	if err != nil {
		return map[string]string{"error": "config.php not found"}
	}
	text := string(content)

	nameMatch := moodleDBNameRE.FindStringSubmatch(text)
	if nameMatch == nil {
		return map[string]string{"error": "No database information found in config.php"}
	}
	info := map[string]string{"database_name": nameMatch[1], "database_prefix": "mdl_"}
	if prefixMatch := moodleDBPrefixRE.FindStringSubmatch(text); prefixMatch != nil {
		info["database_prefix"] = prefixMatch[1]
	}
	return info
}
