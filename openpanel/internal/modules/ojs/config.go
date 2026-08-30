package ojs

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// toStringCell converts one mysqlmanager.Exec() result cell to a string.
// Mirrors every other CMS module's identical helper.
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

// config.inc.php is OJS's config file - INI format (php.ini-style
// "key = value" lines under "[section]" headers), not a PHP $CFG-> variable
// assignment file the way Moodle/WordPress/Joomla's config files are, so
// every regex here matches a bare "key = ..." line instead. Since key names
// are unique across the sections this module actually touches (driver/
// host/username/password/name only live under [database]; base_url/
// time_zone only under [general]), a plain "^key\s*=" match is safe without
// also anchoring on the preceding "[section]" line.
var (
	iniDatabaseDriverRE   = regexp.MustCompile(`(?m)^driver\s*=.*$`)
	iniDatabaseHostRE     = regexp.MustCompile(`(?m)^host\s*=.*$`)
	iniDatabaseUsernameRE = regexp.MustCompile(`(?m)^username\s*=.*$`)
	iniDatabasePasswordRE = regexp.MustCompile(`(?m)^password\s*=.*$`)
	iniDatabaseNameRE     = regexp.MustCompile(`(?m)^name\s*=.*$`)
	iniBaseURLRE          = regexp.MustCompile(`(?m)^base_url\s*=.*$`)
	iniTimeZoneRE         = regexp.MustCompile(`(?m)^time_zone\s*=.*$`)
	iniFilesDirRE         = regexp.MustCompile(`(?m)^files_dir\s*=.*$`)
	iniAllowedHostsRE     = regexp.MustCompile(`(?m)^allowed_hosts\s*=.*$`)

	iniReadDatabaseNameRE     = regexp.MustCompile(`(?m)^name\s*=\s*"?([^"\r\n]*)"?\s*$`)
	iniReadDatabaseUsernameRE = regexp.MustCompile(`(?m)^username\s*=\s*"?([^"\r\n]*)"?\s*$`)
)

func iniQuoted(key, value string) string {
	return key + ` = "` + value + `"`
}

func iniBare(key, value string) string {
	return key + " = " + value
}

// ojsApprootDir maps an OJS site's docroot (a symlink to <slug>_ojsapp -
// see ojs.go's package doc comment) to its backing app-root directory,
// where config.inc.php/tools/index.php actually live, the same
// domain/subdirectory-derived siteSlug() install.go used to create it.
func ojsApprootDir(userContext, directory string) string {
	const wwwPrefix = "/var/www/html/"
	relPath := directory
	if len(relPath) >= len(wwwPrefix) && relPath[:len(wwwPrefix)] == wwwPrefix {
		relPath = relPath[len(wwwPrefix):]
	}
	slug := siteSlug(relPath)
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + slug + "_ojsapp"
}

// extractOJSDatabaseInfoForLogin reads database name/username straight out
// of the approot's config.inc.php - only the fields the autologin/backup
// flows actually need (OJS has no per-site table prefix the way
// Moodle/Joomla/WordPress do, so there's no "database_prefix" to extract).
func extractOJSDatabaseInfoForLogin(userContext, domain string) map[string]string {
	approot := ojsApprootDir(userContext, domain)
	content, err := os.ReadFile(filepath.Join(approot, "config.inc.php"))
	if err != nil {
		return map[string]string{"error": "config.inc.php not found"}
	}
	text := string(content)
	nameMatch := iniReadDatabaseNameRE.FindStringSubmatch(text)
	userMatch := iniReadDatabaseUsernameRE.FindStringSubmatch(text)
	if nameMatch == nil {
		return map[string]string{"error": "No database information found in config.inc.php"}
	}
	info := map[string]string{"database_name": nameMatch[1]}
	if userMatch != nil {
		info["database_username"] = userMatch[1]
	}
	return info
}
