package ojs

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// toStringCell converts one mysqlmanager.Exec() result cell to a string, mirrors every other CMS module's identical helper
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

// config.inc.php is INI format ("key = value" under "[section]"), not a PHP $CFG-> assignment file like Moodle/WordPress/Joomla, so these match a bare "key = ..." line - safe without anchoring on the section since key names don't collide across the sections this module touches
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

// ojsApprootDir maps a site's docroot (a symlink to <slug>_ojsapp, see ojs.go) to its backing app-root directory where config.inc.php/tools/index.php live, via the same siteSlug() install.go used to create it
func ojsApprootDir(userContext, directory string) string {
	const wwwPrefix = "/var/www/html/"
	relPath := directory
	if len(relPath) >= len(wwwPrefix) && relPath[:len(wwwPrefix)] == wwwPrefix {
		relPath = relPath[len(wwwPrefix):]
	}
	slug := siteSlug(relPath)
	return "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + slug + "_ojsapp"
}

// extractOJSDatabaseInfoForLogin reads database name/username out of the approot's config.inc.php - just what autologin/backup need, no "database_prefix" since OJS has no per-site table prefix
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
