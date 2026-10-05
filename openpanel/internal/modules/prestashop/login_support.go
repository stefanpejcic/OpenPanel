package prestashop

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// extractPrestashopDatabaseInfoForLogin is a local copy of websites.extractPrestashopDatabaseInfo (unexported there, duplicated here like the other CMS modules do), only populates the fields handlePrestashopLogin needs
func extractPrestashopDatabaseInfoForLogin(userContext, directory string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(directory, wwwPrefix) {
		return map[string]string{"error": "invalid docroot"}
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(directory, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "app", "config", "parameters.php"))
	if err != nil {
		return map[string]string{"error": "app/config/parameters.php not found"}
	}
	text := string(content)

	nameRE := regexp.MustCompile(`'database_name'\s*=>\s*'([^']*)'`)
	prefixRE := regexp.MustCompile(`'database_prefix'\s*=>\s*'([^']*)'`)
	nameMatch := nameRE.FindStringSubmatch(text)
	if nameMatch == nil {
		return map[string]string{"error": "No database information found in app/config/parameters.php"}
	}
	info := map[string]string{"database_name": nameMatch[1], "database_prefix": "ps_"}
	if prefixMatch := prefixRE.FindStringSubmatch(text); prefixMatch != nil {
		info["database_prefix"] = prefixMatch[1]
	}
	return info
}
