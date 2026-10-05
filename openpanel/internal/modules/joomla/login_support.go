package joomla

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// extractJoomlaDatabaseInfoForLogin is a local copy of websites.extractJoomlaDatabaseInfo (unexported in another package, so duplicated here, same pattern wordpress/drupal already use rather than sharing across module packages) - only the fields handleJoomlaLogin actually needs are populated
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
