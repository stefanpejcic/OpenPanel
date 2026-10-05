package opencart

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// extractOpenCartDatabaseInfoForLogin is a local copy of websites.extractOpenCartDatabaseInfo (unexported there, duplicated here like the other CMS modules do), only populates the field handleOpenCartLogin needs
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
