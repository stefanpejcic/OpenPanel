package ojs

import (
	"os"
	"path/filepath"
	"regexp"

	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
)

// config.inc.php is INI format ("key = value" under "[section]"), not a PHP $CFG-> assignment file like Moodle/WordPress/Joomla, so these match a bare "key = ..." line - safe without anchoring on the section since key names don't collide across the sections this module touches
var (
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

// ojsApprootDir maps a site's docroot (a symlink to <slug>_ojsapp, see ojs.go) to its backing app-root directory where config.inc.php/tools/index.php live, via the same appkit.SiteSlug() install.go used to create it
func ojsApprootDir(userContext, directory string) string {
	const wwwPrefix = "/var/www/html/"
	relPath := directory
	if len(relPath) >= len(wwwPrefix) && relPath[:len(wwwPrefix)] == wwwPrefix {
		relPath = relPath[len(wwwPrefix):]
	}
	slug := appkit.SiteSlug(relPath)
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
