package php

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// ConfigEntry is one parsed line of a php.ini file. Value is only
// meaningful when HasValue is true - a bare directive name with no '='
// still counts as an entry, just with no value.
type ConfigEntry struct {
	Key      string
	Value    string
	HasValue bool
}

// parseConfigContent skips blank/comment/section lines and splits the rest
// on the first '='.
func parseConfigContent(content string) []ConfigEntry {
	var entries []ConfigEntry
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") || strings.HasPrefix(line, ";") {
			continue
		}
		if idx := strings.Index(line, "="); idx != -1 {
			entries = append(entries, ConfigEntry{
				Key: strings.TrimSpace(line[:idx]), Value: strings.TrimSpace(line[idx+1:]), HasValue: true,
			})
		} else {
			entries = append(entries, ConfigEntry{Key: line})
		}
	}
	return entries
}

// configEntriesToMap collapses parsed entries into a map; later entries
// win on a duplicate key.
func configEntriesToMap(entries []ConfigEntry) map[string]string {
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		m[e.Key] = e.Value
	}
	return m
}

// loadKeysFromFile returns the editable option keys: a per-user override
// file, else a global override file, else the built-in default key list.
func loadKeysFromFile(userContext string) []string {
	userFilePath := "/home/" + userContext + "/php.ini/options.txt"
	if content, err := os.ReadFile(userFilePath); err == nil {
		return splitNonEmptyLines(content)
	}
	if content, err := os.ReadFile("/etc/openpanel/php/options.txt"); err == nil {
		return splitNonEmptyLines(content)
	}
	return []string{
		"allow_url_fopen", "date.timezone", "disable_functions", "display_errors",
		"error_reporting", "file_uploads", "log_errors", "max_execution_time",
		"max_input_time", "max_input_vars", "memory_limit", "open_basedir",
		"output_buffering", "post_max_size", "short_open_tag", "upload_max_filesize",
		"zlib.output_compression",
	}
}

func splitNonEmptyLines(content []byte) []string {
	var keys []string
	for _, line := range strings.Split(string(content), "\n") {
		keys = append(keys, strings.TrimSpace(line))
	}
	return keys
}

// keyExistsAndNotCommented reports whether key has an active (uncommented)
// line in the given PHP version's php.ini.
func keyExistsAndNotCommented(userContext, version, key string) bool {
	content, err := os.ReadFile("/home/" + userContext + "/php.ini/" + version + ".ini")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, key+"=") || strings.HasPrefix(line, key+" ") {
			return true
		}
	}
	return false
}

// updatePHPConfigFile deletes, updates, or appends each key in keyOrder in
// php.ini by shelling out to sed rather than reimplementing the line
// matching in Go, since a hand-rolled parser could diverge from the exact
// line-matching behavior admins may already depend on for hand-edited ini
// files.
func updatePHPConfigFile(ctx context.Context, userContext, version string, keyOrder []string, values map[string]string) {
	phpIniFile := "/home/" + userContext + "/php.ini/" + version + ".ini"

	for _, key := range keyOrder {
		value := values[key]
		exists := keyExistsAndNotCommented(userContext, version, key)

		switch {
		case value == "":
			sedCmd := fmt.Sprintf("/^[[:space:]]*%s[[:space:]]*[^;]/d", key)
			_ = exec.CommandContext(ctx, "sed", "-i", sedCmd, phpIniFile).Run()
		case exists:
			safeValue := strings.ReplaceAll(value, "&", `\&`)
			sedCmd := fmt.Sprintf("s/^[[:space:]]*%s[[:space:]]*=.*/%s = %s/", key, key, safeValue)
			_ = exec.CommandContext(ctx, "sed", "-i", sedCmd, phpIniFile).Run()
		default:
			f, err := os.OpenFile(phpIniFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err == nil {
				_, _ = f.WriteString(key + " = " + value + "\n")
				_ = f.Close()
			}
		}
	}
}

// availableTimezones lists the IANA timezones available for the
// date.timezone option. The Go stdlib has no bundled zone list, so this
// walks /usr/share/zoneinfo (present on every host this panel targets) and
// filters out the non-zone metadata files that directory also contains.
var timezoneSkipNames = map[string]bool{
	"posixrules": true, "Factory": true, "iso3166.tab": true, "zone.tab": true,
	"zone1970.tab": true, "leapseconds": true, "tzdata.zi": true, "leap-seconds.list": true,
}

func availableTimezones() []string {
	const root = "/usr/share/zoneinfo"
	var zones []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		if strings.HasPrefix(filepath.Base(rel), ".") || timezoneSkipNames[rel] {
			return nil
		}
		if strings.HasPrefix(rel, "right"+string(filepath.Separator)) || strings.HasPrefix(rel, "posix"+string(filepath.Separator)) {
			return nil
		}
		zones = append(zones, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(zones)
	return zones
}

// binaryCheckboxKeys mirrors options.html's hardcoded 0/1-checkbox key list.
var binaryCheckboxKeys = map[string]bool{
	"allow_url_fopen": true, "display_errors": true, "file_uploads": true,
	"log_errors": true, "short_open_tag": true, "zlib.output_compression": true,
}

var optionUnitValueRE = regexp.MustCompile(`^(-?\d+(?:\.\d+)?)([a-zA-Z]+)$`)
var optionNumericValueRE = regexp.MustCompile(`^-?\d+(?:\.\d+)?$`)

// TimezoneOption is one <option> in options.html's date.timezone <select>.
type TimezoneOption struct {
	Name     string
	Selected bool
}

// OptionField is one rendered row of options.html's key/value table -
// precomputed server-side so the template just switches on Kind instead of
// re-deriving the value-classification logic itself.
type OptionField struct {
	Key        string
	Kind       string // "checkbox_binary" | "timezone" | "checkbox_onoff" | "unit" | "number" | "text"
	Value      string
	Checked    bool
	NumberPart string
	UnitPart   string
	Timezones  []TimezoneOption
}

func buildOptionField(key, value string, timezones []string) OptionField {
	if binaryCheckboxKeys[key] {
		return OptionField{Key: key, Kind: "checkbox_binary", Value: value, Checked: value == "1" || value == "On"}
	}

	if key == "date.timezone" {
		found := false
		for _, tz := range timezones {
			if tz == value {
				found = true
				break
			}
		}
		var opts []TimezoneOption
		if !found && value != "" {
			opts = append(opts, TimezoneOption{Name: value, Selected: true})
		}
		for _, tz := range timezones {
			opts = append(opts, TimezoneOption{Name: tz, Selected: tz == value})
		}
		return OptionField{Key: key, Kind: "timezone", Value: value, Timezones: opts}
	}

	if value == "On" || value == "Off" {
		return OptionField{Key: key, Kind: "checkbox_onoff", Value: value, Checked: value == "On"}
	}

	if m := optionUnitValueRE.FindStringSubmatch(value); m != nil {
		return OptionField{Key: key, Kind: "unit", Value: value, NumberPart: m[1], UnitPart: m[2]}
	}

	if optionNumericValueRE.MatchString(value) {
		return OptionField{Key: key, Kind: "number", Value: value}
	}

	return OptionField{Key: key, Kind: "text", Value: value}
}

// handlePHPOptions gets or updates the PHP options table for one version.
// versionSeg is "" for the bare /php/options route (version picker only).
func handlePHPOptions(a *appctx.App, w http.ResponseWriter, r *http.Request, versionSeg string) {
	ctx := r.Context()
	currentUsername, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	version := phpVersionFromSegment(versionSeg)
	title := "Options"

	if version == "" {
		installedVersions := FetchPHPVersions(ctx, a, userContext)
		if r.URL.Query().Get("output") == "json" {
			writeJSON(w, http.StatusOK, map[string]any{"installed_versions": installedVersions})
			return
		}
		renderPHPOptionsSelectPage(a, w, r, title, installedVersions)
		return
	}

	currentContent, readErr := os.ReadFile("/home/" + userContext + "/php.ini/" + version + ".ini")
	if readErr != nil {
		writeJSONError(w, http.StatusInternalServerError, readErr.Error())
		return
	}
	currentConfig := configEntriesToMap(parseConfigContent(string(currentContent)))
	title = version + " Options"
	availableKeys := loadKeysFromFile(userContext)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		values := make(map[string]string, len(availableKeys))
		for _, key := range availableKeys {
			values[key] = r.Form.Get(key)
		}
		updatePHPConfigFile(ctx, userContext, version, availableKeys, values)

		webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
		var phpContainer, text string
		if strings.Contains(strings.ToLower(webServer), "litespeed") {
			phpContainer, text = webServer, "Litespeed"
		} else {
			phpContainer, text = "php-fpm-"+version, "PHP-FPM"
		}
		docker.StartComposeServiceIfNotRunning(ctx, userContext, phpContainer)

		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "edited PHP "+version+" configuration using PHP Selector", ipAddress)
		flashSess(a, w, r, "success", "Configuration edited successfully and "+text+" service restarted to apply new settings.")
		http.Redirect(w, r, "/php/php"+version+"/options", http.StatusFound)
		return
	}

	timezones := availableTimezones()

	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusOK, map[string]any{
			"available_keys":      availableKeys,
			"current_config":      currentConfig,
			"available_timezones": timezones,
		})
		return
	}

	var phpIniIssues []HealthIssue
	if syntaxError := checkPHPIniSyntax(ctx, userContext, version); syntaxError != "" {
		phpIniIssues = append(phpIniIssues, HealthIssue{
			ID: "php-ini-syntax:" + version, Severity: "error",
			Message: fmt.Sprintf("Syntax error in the PHP %s ini file: %s", version, syntaxError),
		})
	}

	fields := make([]OptionField, 0, len(availableKeys))
	for _, key := range availableKeys {
		fields = append(fields, buildOptionField(key, currentConfig[key], timezones))
	}

	renderPHPOptionsPage(a, w, r, version, title, fields, phpIniIssues)
}
