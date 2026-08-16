package matomo

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

func itoa(n int) string { return strconv.Itoa(n) }

// generateSecretToken returns a long random hex string used as the
// login-helper's shared secret (see login_php.go).
func generateSecretToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// matomoCredentials is what saveMatomoCredentials/loadMatomoCredentials
// persist per install, root-only, outside the webroot - unlike the other
// CMS modules (which need only a one-time hashed token because Matomo's own
// CLI/DB offers no equivalent of Drush's `user:login` or a bcrypt-free
// session bootstrap they could reuse), the auto-login flow here needs the
// real admin password to replay Matomo's own login form server-side (see
// login_php.go) - kept in the same trust boundary as every other CMS
// module's plaintext DB password already sitting in its own config file.
type matomoCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func credentialsPath(selectedDomain string) string {
	safe := strings.NewReplacer("/", "__", " ", "_").Replace(selectedDomain)
	return filepath.Join("/etc/openpanel/matomo/credentials", safe+".json")
}

func saveMatomoCredentials(selectedDomain, login, password string) (matomoCredentials, error) {
	creds := matomoCredentials{Login: login, Password: password, Token: generateSecretToken()}
	dir := filepath.Dir(credentialsPath(selectedDomain))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return creds, err
	}
	b, err := json.Marshal(creds)
	if err != nil {
		return creds, err
	}
	return creds, os.WriteFile(credentialsPath(selectedDomain), b, 0o600)
}

func loadMatomoCredentials(selectedDomain string) (matomoCredentials, error) {
	var creds matomoCredentials
	b, err := os.ReadFile(credentialsPath(selectedDomain))
	if err != nil {
		return creds, err
	}
	err = json.Unmarshal(b, &creds)
	return creds, err
}

func removeMatomoCredentials(selectedDomain string) {
	_ = os.Remove(credentialsPath(selectedDomain))
}

var (
	matomoDBNameRE   = regexp.MustCompile(`(?m)^dbname\s*=\s*"([^"]*)"`)
	matomoDBPrefixRE = regexp.MustCompile(`(?m)^tables_prefix\s*=\s*"([^"]*)"`)
)

// extractMatomoDatabaseInfoForBackup reads config/config.ini.php straight
// off the host filesystem - same "read config from host path" pattern as
// every other CMS module's login_support.go, just parsing Matomo's INI
// format (a regex-based line match, matching the rest of this codebase's
// preference for small regexes over pulling in an INI-parsing library)
// instead of a PHP array/define() list.
func extractMatomoDatabaseInfoForBackup(userContext, docroot string) map[string]string {
	const wwwPrefix = "/var/www/html/"
	if !strings.HasPrefix(docroot, wwwPrefix) {
		return map[string]string{"error": "invalid docroot"}
	}
	mappedDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/" + strings.TrimPrefix(docroot, wwwPrefix)
	content, err := os.ReadFile(filepath.Join(mappedDir, "config", "config.ini.php"))
	if err != nil {
		return map[string]string{"error": "config/config.ini.php not found"}
	}
	text := string(content)

	nameMatch := matomoDBNameRE.FindStringSubmatch(text)
	if nameMatch == nil {
		return map[string]string{"error": "No database information found in config/config.ini.php"}
	}
	info := map[string]string{"database_name": nameMatch[1], "database_prefix": "matomo_"}
	if prefixMatch := matomoDBPrefixRE.FindStringSubmatch(text); prefixMatch != nil {
		info["database_prefix"] = prefixMatch[1]
	}
	return info
}
