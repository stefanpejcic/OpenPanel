package postgresql

import (
	"net/http"
	"os"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// defaultConfKeys is the built-in set of postgresql.conf keys exposed for
// editing when no admin-provided keys file exists.
var defaultConfKeys = []string{
	"max_connections", "shared_buffers", "work_mem", "maintenance_work_mem",
	"effective_cache_size", "max_worker_processes", "max_parallel_workers",
	"wal_level", "synchronous_commit", "checkpoint_timeout",
	"checkpoint_completion_target", "logging_collector", "log_directory",
	"log_filename", "log_min_duration_statement", "log_statement",
	"autovacuum", "autovacuum_max_workers", "autovacuum_naptime",
	"autovacuum_vacuum_scale_factor",
}

const confKeysFile = "/etc/openpanel/postgres/keys.txt"

// availableConfKeys is the admin-editable keys file if present, else
// defaultConfKeys, loaded once at Register() time.
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

// postgresConfPath is the per-user custom PostgreSQL config file path.
func postgresConfPath(userContext string) string {
	return "/home/" + userContext + "/postgre_custom.conf"
}

// readPostgresConfigFile reads the user's custom PostgreSQL config file.
func readPostgresConfigFile(userContext string) (string, error) {
	content, err := os.ReadFile(postgresConfPath(userContext))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// parsePostgresConfigContent parses "key = value" lines into a map,
// skipping blank lines, full-line comments, and section headers, and
// stripping trailing inline comments.
func parsePostgresConfigContent(content string) map[string]string {
	config := map[string]string{}
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}
		if idx := strings.Index(line, "="); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			config[key] = value
		} else {
			config[line] = ""
		}
	}
	return config
}

// updatePostgresConfigFile: for every key in newConfig, replaces its
// existing (possibly commented-out) line in place, or appends a new line
// if the key isn't present yet.
func updatePostgresConfigFile(userContext string, newConfig map[string]string, keyOrder []string) {
	path := postgresConfPath(userContext)
	content, _ := os.ReadFile(path)
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	for _, key := range keyOrder {
		value, ok := newConfig[key]
		if !ok {
			continue
		}
		linePattern := regexp.MustCompile(`^[ \t]*#*[ \t]*` + regexp.QuoteMeta(key) + `[ \t]*=.*`)

		var newLine string
		if value == "" {
			newLine = "# " + key + " = "
		} else {
			newLine = key + " = " + value
		}

		replaced := false
		for i, line := range lines {
			if linePattern.MatchString(line) {
				lines[i] = newLine
				replaced = true
				break
			}
		}
		if !replaced {
			lines = append(lines, newLine)
		}
	}

	_ = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

// handleEditPostgresConfig saves the submitted config values and restarts
// the postgres container to apply them, then renders the current config.
func handleEditPostgresConfig(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		newConfig := map[string]string{}
		for _, key := range availableConfKeys {
			newConfig[key] = r.Form.Get(key)
		}
		updatePostgresConfigFile(userContext, newConfig, availableConfKeys)
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "edited PostgreSQL configuration", ipAddress)

		argv := podmanmanager.PodmanArgv(userContext, "restart", "postgres")
		if runErr := podmanmanager.Command(ctx, userContext, argv).Run(); runErr != nil {
			flashSess(a, w, r, "error", "PostgreSQL configuration was saved successfully but postgres service failed to start! Please revert the changes to restart postgres.")
		} else {
			flashSess(a, w, r, "success", "PostgreSQL Configuration was edited successfully and service restarted to apply new settings.")
		}
	}

	var currentConfig map[string]string
	if !docker.IsServiceRunning(ctx, userContext, "postgres") {
		flashSess(a, w, r, "warning", "postgres container is not running. Please allow a few moments for initialization..")
		docker.StartComposeServiceIfNotRunning(ctx, userContext, "postgres")
		currentConfig = map[string]string{}
	} else {
		content, readErr := readPostgresConfigFile(userContext)
		if readErr != nil {
			currentConfig = map[string]string{}
		} else {
			currentConfig = parsePostgresConfigContent(content)
		}
	}

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{"current_config": currentConfig, "default_keys": availableConfKeys})
		return
	}

	renderConfigurationPage(a, w, r, currentConfig, availableConfKeys)
}
