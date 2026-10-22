package postgresql

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/podmanmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/postgresmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// defaultConfKeys is the built-in set of postgresql.conf keys exposed for editing when no admin-provided keys file exists
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

// availableConfKeys is the admin-editable keys file if present, else defaultConfKeys, loaded once at Register() time
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

// postgresConfPath is the per-user custom PostgreSQL config file path.
func postgresConfPath(userContext string) string {
	return "/home/" + userContext + "/postgre_custom.conf"
}

// readPostgresConfigFile reads the user's custom PostgreSQL config file.
func readPostgresConfigFile(userContext string) (string, error) {
	content, err := tenantfs.ReadFile(postgresConfPath(userContext))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// parsePostgresConfigContent parses "key = value" lines into a map, skipping blank lines, full-line comments, and section headers, and stripping trailing inline comments
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

// updatePostgresConfigFile replaces each key in newConfig's existing (possibly commented-out) line in place, or appends a new line if the key isn't present yet
func updatePostgresConfigFile(userContext string, newConfig map[string]string, keyOrder []string) {
	path := postgresConfPath(userContext)
	content, _ := tenantfs.ReadFile(path)
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

	_ = tenantfs.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

var pgSettingNameRE = regexp.MustCompile(`^[a-z_]+$`)

// applyPostgresConfig stores the keys with ALTER SYSTEM and reloads, restarting only when postgres marks a setting pending_restart; a rejected value puts back what postgresql.auto.conf had and comes back as the error
func applyPostgresConfig(ctx context.Context, userContext string, newConfig map[string]string) (restarted bool, err error) {
	// compared against postgresql.auto.conf rather than postgre_custom.conf, the file was never mounted so it doesn't say what's running
	rows, err := postgresmanager.Exec(ctx, userContext, "SELECT name, setting FROM pg_file_settings WHERE sourcefile LIKE '%postgresql.auto.conf'", "postgres")
	if err != nil {
		return false, errors.New("PostgreSQL is not running, start it and save again")
	}
	stored := map[string]string{}
	for _, row := range rows {
		stored[fmt.Sprint(row[0])] = fmt.Sprint(row[1])
	}
	alter := func(key, value string) error {
		if value == "" {
			_, err := postgresmanager.Exec(ctx, userContext, "ALTER SYSTEM RESET "+key, "postgres")
			return err
		}
		_, err := postgresmanager.Exec(ctx, userContext, "ALTER SYSTEM SET "+key+" = '"+strings.ReplaceAll(value, "'", "''")+"'", "postgres")
		return err
	}

	var applied []string
	for key, value := range newConfig {
		value = strings.Trim(strings.TrimSpace(value), "'")
		if value == stored[key] {
			continue
		}
		if !pgSettingNameRE.MatchString(key) {
			return false, fmt.Errorf("invalid setting name %q", key)
		}
		if err := alter(key, value); err != nil {
			for _, done := range applied {
				_ = alter(done, stored[done])
			}
			return false, fmt.Errorf("%s: %v", key, err)
		}
		applied = append(applied, key)
	}
	if len(applied) == 0 {
		return false, nil
	}

	if _, err := postgresmanager.Exec(ctx, userContext, "SELECT pg_reload_conf()", "postgres"); err != nil {
		return false, err
	}
	// the reload is only signalled, give postgres a moment before asking what it couldn't apply
	time.Sleep(time.Second)
	rows, err = postgresmanager.Exec(ctx, userContext, "SELECT count(*) FROM pg_settings WHERE pending_restart", "postgres")
	if err != nil || len(rows) == 0 || postgresmanager.ToInt(rows[0][0]) == 0 {
		return false, nil
	}

	if err := podmanmanager.Command(ctx, userContext, podmanmanager.PodmanArgv(userContext, "restart", "postgres")).Run(); err != nil {
		return false, err
	}
	postgresmanager.InvalidatePool(userContext, "postgres")
	deadline := time.Now().Add(time.Minute)
	for time.Now().Before(deadline) {
		if _, err := postgresmanager.Exec(ctx, userContext, "SELECT 1", "postgres"); err == nil {
			return true, nil
		}
		time.Sleep(2 * time.Second)
	}
	return true, errors.New("PostgreSQL did not come back after the restart")
}

// handleEditPostgresConfig saves the submitted config values and applies them with a reload, restarting only for settings that need it, then renders the current config
func handleEditPostgresConfig(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		newConfig := map[string]string{}
		// only fields the form sent, a missing one would otherwise be commented out of postgre_custom.conf
		for _, key := range availableConfKeys {
			if r.Form.Has(key) {
				newConfig[key] = r.Form.Get(key)
			}
		}
		restarted, applyErr := applyPostgresConfig(ctx, userContext, newConfig)
		switch {
		case applyErr != nil && !restarted:
			web.Flash(a, w, r, "error", "PostgreSQL configuration was not changed: "+applyErr.Error())
		case applyErr != nil:
			updatePostgresConfigFile(userContext, newConfig, availableConfKeys)
			web.Flash(a, w, r, "error", "PostgreSQL configuration was saved but the service did not come back after the restart: "+applyErr.Error())
		default:
			updatePostgresConfigFile(userContext, newConfig, availableConfKeys)
			_ = logger.RecordUserAction(a.Config, currentUsername, "edited PostgreSQL configuration", reqip.ClientIP(r))
			if restarted {
				web.Flash(a, w, r, "success", "PostgreSQL Configuration was edited successfully and service restarted to apply new settings.")
			} else {
				web.Flash(a, w, r, "success", "PostgreSQL Configuration was edited successfully and applied without a restart.")
			}
		}
	}

	var currentConfig map[string]string
	if !docker.IsServiceRunning(ctx, userContext, "postgres") {
		web.Flash(a, w, r, "warning", "Postgres container is not running. Please allow a few moments for the initialization..")
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
		web.WriteJSON(w, http.StatusOK, map[string]any{"current_config": currentConfig, "default_keys": availableConfKeys})
		return
	}

	renderConfigurationPage(a, w, r, currentConfig, availableConfKeys)
}
