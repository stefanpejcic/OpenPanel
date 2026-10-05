package mediawiki

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	mediawikiDBNameRE   = regexp.MustCompile(`\$wgDBname\s*=\s*"([^"]*)"`)
	mediawikiDBUserRE   = regexp.MustCompile(`\$wgDBuser\s*=\s*"([^"]*)"`)
	mediawikiDBPrefixRE = regexp.MustCompile(`\$wgDBprefix\s*=\s*"([^"]*)"`)
)

// extractMediaWikiDatabaseInfoForLogin is a local copy of websites.extractMediaWikiDatabaseInfo (unexported in another package, so duplicated here, same pattern every other CMS module already uses rather than sharing across module packages) - reads LocalSettings.php directly, keyed off the docroot the same way install.go/manage.go derive it
func extractMediaWikiDatabaseInfoForLogin(userContext, directory string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(directory, wwwPrefix) {
		return map[string]string{"error": "invalid docroot"}
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(directory, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "LocalSettings.php"))
	if err != nil {
		return map[string]string{"error": "LocalSettings.php not found"}
	}
	text := string(content)

	nameMatch := mediawikiDBNameRE.FindStringSubmatch(text)
	userMatch := mediawikiDBUserRE.FindStringSubmatch(text)
	if nameMatch == nil || userMatch == nil {
		return map[string]string{"error": "No database information found in LocalSettings.php"}
	}
	info := map[string]string{"database_name": nameMatch[1], "database_user": userMatch[1], "database_prefix": ""}
	if prefixMatch := mediawikiDBPrefixRE.FindStringSubmatch(text); prefixMatch != nil {
		info["database_prefix"] = prefixMatch[1]
	}
	return info
}
