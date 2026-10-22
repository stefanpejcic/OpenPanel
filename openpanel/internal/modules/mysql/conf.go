package mysql

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	mysqldriver "github.com/go-sql-driver/mysql"
)

// defaultConfKeys is the fallback list of admin-editable my.cnf keys used when no custom keys file is present
var defaultConfKeys = []string{
	"max_allowed_packet", "max_connect_errors", "max_connections", "open_files_limit",
	"performance_schema", "sql_mode", "thread_cache_size", "interactive_timeout",
	"wait_timeout", "log_output", "log_error", "log_error_verbosity", "general_log",
	"general_log_file", "long_query_time", "slow_query_log", "slow_query_log_file",
	"join_buffer_size", "key_buffer_size", "read_buffer_size", "read_rnd_buffer_size",
	"sort_buffer_size", "innodb_log_buffer_size", "innodb_log_file_size", "innodb_sort_buffer_size",
	"innodb_buffer_pool_chunk_size", "innodb_buffer_pool_instances", "innodb_buffer_pool_size",
	"max_heap_table_size", "tmp_table_size", "table_open_cache", "table_definition_cache", "innodb_flush_log_at_trx_commit",
}

const confKeysFile = "/etc/openpanel/mysql/keys.txt"

// availableConfKeys holds the admin-editable keys file if present, else falls back to defaultConfKeys - loaded once at Register() time
var availableConfKeys = defaultConfKeys

func loadConfKeys() {
	content, err := tenantfs.ReadFile(confKeysFile)
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

// readMySQLConfigFile returns the contents of a user's custom.cnf, or "" if it doesn't exist
func readMySQLConfigFile(userContext string) string {
	content, err := tenantfs.ReadFile(mysqlConfPath(userContext))
	if err != nil {
		return ""
	}
	return string(content)
}

// parseMySQLConfigContent parses custom.cnf lines into key/value entries, skipping blanks, comments, section headers, and skip-log-bin
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

// updateMySQLConfigFile rewrites existing lines whose key is in newConfig (dropping ones set to an empty value), leaves every other line untouched, then appends any newConfig keys that weren't already present
func updateMySQLConfigFile(userContext string, newConfig map[string]string, keyOrder []string) {
	path := mysqlConfPath(userContext)
	var existingLines []string
	if content, err := tenantfs.ReadFile(path); err == nil {
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

	_ = tenantfs.WriteFile(path, []byte(strings.Join(newLines, "")), 0o644)
}

// applyResult is how a config save went, Running tells whether the database is up after a rollback, Live that it applied without a restart, Rejected that the server refused a value before anything restarted
type applyResult struct {
	OK       bool
	Running  bool
	Live     bool
	Rejected bool
	Reason   string
}

// applyMySQLConfig writes the settings and applies them with SET GLOBAL, restarting only when one can't change at runtime, and puts the old custom.cnf back when the database doesn't come up with them
func applyMySQLConfig(ctx context.Context, userContext, service string, newConfig map[string]string) applyResult {
	path := mysqlConfPath(userContext)
	previous, readErr := tenantfs.ReadFile(path)
	updateMySQLConfigFile(userContext, newConfig, availableConfKeys)
	live, rejected := setGlobals(ctx, userContext, configEntriesToMap(parseMySQLConfigContent(string(previous))), newConfig)
	if live {
		return applyResult{OK: true, Running: true, Live: true}
	}
	if rejected != nil {
		restoreFile(path, previous, readErr)
		return applyResult{Running: true, Rejected: true, Reason: rejected.Error()}
	}
	if restartAndWait(ctx, userContext, service) {
		return applyResult{OK: true, Running: true}
	}
	// grab the reason before the next restart replaces the log
	reason := startupErrors(ctx, userContext, service)
	restoreFile(path, previous, readErr)
	return applyResult{Running: restartAndWait(ctx, userContext, service), Reason: reason}
}

var (
	globalNameRE = regexp.MustCompile(`^[a-z_]+$`)
	sizeValueRE  = regexp.MustCompile(`^(\d+)([KkMmGg])$`)
	plainValueRE = regexp.MustCompile(`^(\d+(\.\d+)?|(?i:on|off|true|false))$`)
)

// setGlobals applies every changed key with SET GLOBAL; live false means a restart is needed (database down, a key removed, read-only or unknown to SET), a non-nil error means the server rejected a value and every key it already set was put back
func setGlobals(ctx context.Context, userContext string, previous, newConfig map[string]string) (live bool, rejected error) {
	if _, err := mysqlmanager.Exec(ctx, userContext, "SELECT 1", ""); err != nil {
		return false, nil
	}
	undo := map[string]string{}
	revert := func() {
		for key, old := range undo {
			_, _ = mysqlmanager.Exec(ctx, userContext, "SET GLOBAL "+key+" = "+globalValue(old), "")
		}
	}
	for key, value := range newConfig {
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if value == strings.Trim(previous[key], `"'`) {
			continue
		}
		// a removed key falls back to whatever the other cnf files say, only a restart knows that
		if value == "" || !globalNameRE.MatchString(key) {
			return false, nil
		}
		rows, err := mysqlmanager.Exec(ctx, userContext, "SELECT @@GLOBAL."+key, "")
		if err != nil || len(rows) == 0 {
			// not a system variable SET knows, a restart will tell whether the server takes it
			return false, nil
		}
		if _, err := mysqlmanager.Exec(ctx, userContext, "SET GLOBAL "+key+" = "+globalValue(value), ""); err != nil {
			var mysqlErr *mysqldriver.MySQLError
			// 1238: read-only, only settable at startup
			if errors.As(err, &mysqlErr) && mysqlErr.Number != 1238 {
				revert()
				return false, fmt.Errorf("%s: %s", key, mysqlErr.Message)
			}
			return false, nil
		}
		undo[key] = fmt.Sprint(rows[0][0])
	}
	return true, nil
}

// globalValue turns a cnf value into a SET GLOBAL literal, cnf size suffixes like 64M aren't valid in SQL
func globalValue(value string) string {
	if m := sizeValueRE.FindStringSubmatch(value); m != nil {
		n, _ := strconv.ParseInt(m[1], 10, 64)
		shift := map[string]uint{"k": 10, "m": 20, "g": 30}[strings.ToLower(m[2])]
		return strconv.FormatInt(n<<shift, 10)
	}
	if plainValueRE.MatchString(value) {
		return value
	}
	return "'" + escapeMySQLString(value) + "'"
}

// restoreFile puts custom.cnf back the way it was, removing it if it didn't exist
func restoreFile(path string, previous []byte, readErr error) {
	if readErr == nil {
		_ = tenantfs.WriteFile(path, previous, 0o644)
	} else {
		_ = tenantfs.Remove(path)
	}
}

// restartAndWait restarts the database container and waits up to a minute for it to answer queries again
func restartAndWait(ctx context.Context, userContext, service string) bool {
	argv := podmanmanager.PodmanArgv(userContext, "restart", service)
	if podmanmanager.Command(ctx, userContext, argv).Run() != nil {
		return false
	}
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := mysqlmanager.Exec(ctx, userContext, "SELECT 1", ""); err == nil {
			return true
		}
		if st := docker.GetContainerStatus(ctx, userContext, service).State; st == "exited" || st == "stopped" {
			return false
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

// startupErrors returns the [ERROR] lines from the failed start, which usually name the bad setting
func startupErrors(ctx context.Context, userContext, service string) string {
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, _ := podmanmanager.Command(cctx, userContext, podmanmanager.PodmanArgv(userContext, "logs", "--tail", "50", service)).CombinedOutput()
	var errs []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "[ERROR]") {
			errs = append(errs, strings.TrimSpace(line))
		}
	}
	if len(errs) > 3 {
		errs = errs[len(errs)-3:]
	}
	return strings.Join(errs, " ")
}

// handleEditMySQLConfig renders and processes the MySQL configuration editor: on POST, writes the submitted keys to custom.cnf and applies them, live when possible
func handleEditMySQLConfig(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
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
		_ = logger.RecordUserAction(a.Config, currentUsername, "edited MySQL configuration", reqip.ClientIP(r))
		if mysqlVersion != "mysql" && mysqlVersion != "mariadb" {
			updateMySQLConfigFile(userContext, newConfig, availableConfKeys)
			web.Flash(a, w, r, "error", "Unknown database service, cannot restart")
		} else if res := applyMySQLConfig(ctx, userContext, mysqlVersion, newConfig); res.OK && res.Live {
			web.Flash(a, w, r, "success", web.Tr(a, r, "%(mysql_version)s configuration updated and applied without a restart.", "mysql_version", mysqlVersion))
		} else if res.OK {
			web.Flash(a, w, r, "success", web.Tr(a, r, "%(mysql_version)s configuration updated and service restarted.", "mysql_version", mysqlVersion))
		} else {
			_ = logger.RecordUserAction(a.Config, currentUsername, "MySQL configuration rolled back, service did not start with it", reqip.ClientIP(r))
			msg := web.Tr(a, r, "%(mysql_version)s did not start with the new settings, so the previous configuration was restored.", "mysql_version", mysqlVersion)
			if res.Rejected {
				msg = web.Tr(a, r, "%(mysql_version)s rejected the new settings, nothing was changed.", "mysql_version", mysqlVersion)
			} else if !res.Running {
				msg = web.Tr(a, r, "%(mysql_version)s did not start with the new settings. The previous configuration was restored but the service is still not running.", "mysql_version", mysqlVersion)
			}
			if res.Reason != "" {
				msg += " " + res.Reason
			}
			web.Flash(a, w, r, "error", msg)
		}
	}

	if !docker.IsServiceRunning(ctx, userContext, mysqlVersion) {
		web.Flash(a, w, r, "warning", web.Tr(a, r, "%(mysql_version)s container is not running. Please wait for initialization.", "mysql_version", mysqlVersion))
		docker.StartComposeServiceIfNotRunning(ctx, userContext, "sql")
	}

	currentContent := readMySQLConfigFile(userContext)
	currentConfig := configEntriesToMap(parseMySQLConfigContent(currentContent))

	if r.URL.Query().Get("output") == "json" {
		web.WriteJSON(w, http.StatusOK, map[string]any{"current_config": currentConfig, "default_keys": availableConfKeys})
		return
	}

	renderConfigurationPage(a, w, r, mysqlVersion, currentConfig, availableConfKeys)
}
