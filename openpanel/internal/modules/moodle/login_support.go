package moodle

import (
	"os"
	"path/filepath"
	"regexp"

	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
)

var (
	moodleDBNameRE   = regexp.MustCompile(`CFG->dbname\s*=\s*'([^']*)'`)
	moodleDBPrefixRE = regexp.MustCompile(`CFG->prefix\s*=\s*'([^']*)'`)
)

// extractMoodleDatabaseInfoForBackup is a local copy of websites.extractMoodleDatabaseInfo (unexported in another package, so duplicated here, same pattern every other CMS module already uses rather than sharing across module packages) - reads the approot's config.php directly (docroot is a symlink to <approot>/public, not where config.php lives, see moodle.go's package doc comment), keyed off domain the same way install.go/manage.go derive the approot directory
func extractMoodleDatabaseInfoForBackup(userContext, domain string) map[string]string {
	approotHostPath := filepath.Join("/home/"+userContext+"/docker-data/volumes", userContext+"_html_data/_data", appkit.SiteSlug(domain)+"_moodleapp")
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
