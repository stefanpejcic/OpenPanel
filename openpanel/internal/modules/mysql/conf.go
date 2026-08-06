package mysql

import (
	"net/http"
	"os"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// defaultConfKeys is the fallback list of admin-editable my.cnf keys used
// when no custom keys file is present.
var defaultConfKeys = []string{
	"max_allowed_packet", "max_connect_errors", "max_connections", "open_files_limit",
	"performance_schema", "sql_mode", "thread_cache_size", "interactive_timeout",
	"wait_timeout", "log_output", "log_error", "log_error_verbosity", "general_log",
	"general_log_file", "long_query_time", "slow_query_log", "slow_query_log_file",
	"join_buffer_size", "key_buffer_size", "read_buffer_size", "read_rnd_buffer_size",
	"sort_buffer_size", "innodb_log_buffer_size", "innodb_log_file_size", "innodb_sort_buffer_size",
	"innodb_buffer_pool_chunk_size", "innodb_buffer_pool_instances", "innodb_buffer_pool_size",
	"max_heap_table_size", "tmp_table_size",
}

const confKeysFile = "/etc/openpanel/mysql/keys.txt"

// availableConfKeys holds the admin-editable keys file if present, else
// falls back to defaultConfKeys. It's loaded once at Register() time.
var availableConfKeys = defaultConfKeys

func loadConfKeys() {
	content, err := os.ReadFile(confKeysFile)
	if err != nil {
		availableConfKeys = defaultConfKeys
		return
	}
	var keys []string
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			keys = append(keys, line)
		}
	}
	if len(keys) > 0 {
		availableConfKeys = keys
	}
}

var safeContextRE = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// mysqlConfPath returns the path to a user's custom.cnf override file.
func mysqlConfPath(userContext string) string {
	safeContext := safeContextRE.ReplaceAllString(userContext, "")
	return "/home/" + safeContext + "/custom.cnf"
}

// readMySQLConfigFile returns the contents of a user's custom.cnf, or ""
// if it doesn't exist.
func readMySQLConfigFile(userContext string) string {
	content, err := os.ReadFile(mysqlConfPath(userContext))
	if err != nil {
		return ""
	}
	return string(content)
}

// parseMySQLConfigContent parses custom.cnf lines into key/value entries,
// skipping blanks, comments, section headers, and skip-log-bin.
func parseMySQLConfigContent(content string) []ConfigEntry {
	var entries []ConfigEntry
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "skip-log-bin") {
			continue
		}
		if idx := strings.Index(line, "="); idx != -1 {
			entries = append(entries, ConfigEntry{Key: strings.TrimSpace(line[:idx]), Value: strings.TrimSpace(line[idx+1:]), HasValue: true})
		} else {
			entries = append(entries, ConfigEntry{Key: line})
		}
	}
	return entries
}

// ConfigEntry is one parsed line of custom.cnf.
type ConfigEntry struct {
	Key      string
	Value    string
	HasValue bool
}

func configEntriesToMap(entries []ConfigEntry) map[string]string {
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		m[e.Key] = e.Value
	}
	return m
}

// updateMySQLConfigFile rewrites
// existing lines whose key is in newConfig (dropping ones set to an empty
// value), leave every other line untouched, then append any newConfig keys
// that weren't already present.
func updateMySQLConfigFile(userContext string, newConfig map[string]string, keyOrder []string) {
	path := mysqlConfPath(userContext)
	var existingLines []string
	if content, err := os.ReadFile(path); err == nil {
		existingLines = strings.SplitAfter(string(content), "\n")
		if len(existingLines) > 0 && existingLines[len(existingLines)-1] == "" {
			existingLines = existingLines[:len(existingLines)-1]
		}
	}

	var newLines []string
	keysHandled := map[string]bool{}

	for _, line := range existingLines {
		stripped := strings.TrimSpace(line)
		if stripped == "" {
			newLines = append(newLines, line)
			continue
		}
		if strings.Contains(stripped, "=") && !strings.HasPrefix(stripped, "#") {
			key := strings.TrimSpace(strings.SplitN(stripped, "=", 2)[0])
			if value, ok := newConfig[key]; ok {
				if strings.TrimSpace(value) == "" {
					keysHandled[key] = true
					continue
				}
				safeValue := strings.ReplaceAll(strings.ReplaceAll(value, "\n", ""), "\r", "")
				newLines = append(newLines, key+" = "+safeValue+"\n")
				keysHandled[key] = true
			} else {
				newLines = append(newLines, line)
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	for _, key := range keyOrder {
		if keysHandled[key] {
			continue
		}
		value, ok := newConfig[key]
		if !ok || strings.TrimSpace(value) == "" {
			continue
		}
		safeValue := strings.ReplaceAll(strings.ReplaceAll(value, "\n", ""), "\r", "")
		newLines = append(newLines, key+" = "+safeValue+"\n")
	}

	_ = os.WriteFile(path, []byte(strings.Join(newLines, "")), 0o644)
}

// handleEditMySQLConfig renders and processes the MySQL configuration
// editor: on POST, writes the submitted keys to custom.cnf and restarts
// the database service.
func handleEditMySQLConfig(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	mysqlVersion := webserver.GetEnvFileValue(userContext, "MYSQL_TYPE")

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		newConfig := map[string]string{}
		for _, key := range availableConfKeys {
			if r.Form.Has(key) {
				newConfig[key] = r.Form.Get(key)
			}
		}
		updateMySQLConfigFile(userContext, newConfig, availableConfKeys)
		_ = logger.RecordUserAction(a.Config, currentUsername, "edited MySQL configuration", reqip.ClientIP(r))

		if mysqlVersion != "mysql" && mysqlVersion != "mariadb" {
			flashSess(a, w, r, "error", "Unknown database service, cannot restart")
		} else {
			argv := podmanmanager.PodmanArgv(userContext, "restart", mysqlVersion)
			if runErr := podmanmanager.Command(ctx, userContext, argv).Run(); runErr != nil {
				flashSess(a, w, r, "error", mysqlVersion+" configuration saved but service failed to restart.")
			} else {
				flashSess(a, w, r, "success", mysqlVersion+" configuration updated and service restarted.")
			}
		}
	}

	if !docker.IsServiceRunning(ctx, userContext, mysqlVersion) {
		flashSess(a, w, r, "warning", mysqlVersion+" container is not running. Please wait for initialization.")
		docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")
	}

	currentContent := readMySQLConfigFile(userContext)
	currentConfig := configEntriesToMap(parseMySQLConfigContent(currentContent))

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{"current_config": currentConfig, "default_keys": availableConfKeys})
		return
	}

	renderConfigurationPage(a, w, r, mysqlVersion, currentConfig, availableConfKeys)
}
