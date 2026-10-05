package nextcloud

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// extractNextcloudDatabaseInfoForLogin is a local copy of websites.extractNextcloudDatabaseInfo (unexported there, duplicated here like the other CMS modules do), only populates the fields handleNextcloudLogin needs
func extractNextcloudDatabaseInfoForLogin(userContext, directory string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(directory, wwwPrefix) {
		return map[string]string{"error": "invalid docroot"}
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(directory, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "config", "config.php"))
	if err != nil {
		return map[string]string{"error": "config/config.php not found"}
	}
	text := string(content)

	nameRE := regexp.MustCompile(`'dbname'\s*=>\s*'([^']*)'`)
	prefixRE := regexp.MustCompile(`'dbtableprefix'\s*=>\s*'([^']*)'`)
	nameMatch := nameRE.FindStringSubmatch(text)
	if nameMatch == nil {
		return map[string]string{"error": "No database information found in config/config.php"}
	}
	info := map[string]string{"database_name": nameMatch[1], "database_prefix": "oc_"}
	if prefixMatch := prefixRE.FindStringSubmatch(text); prefixMatch != nil {
		info["database_prefix"] = prefixMatch[1]
	}
	return info
}
