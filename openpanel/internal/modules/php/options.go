package php

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// ConfigEntry is one parsed line of a php.ini file - Value is only meaningful when HasValue is true, since a bare directive with no '=' still counts as an entry
type ConfigEntry struct {
	Key      string
	Value    string
	HasValue bool
}

// parseConfigContent skips blank/comment/section lines and splits the rest on the first '='
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

// configEntriesToMap collapses parsed entries into a map; later entries win on a duplicate key
func configEntriesToMap(entries []ConfigEntry) map[string]string {
	m := make(map[string]string, len(entries))
	for _, e := range entries {
		m[e.Key] = e.Value
	}
	return m
}

// loadKeysFromFile returns the editable option keys: a per-user override file, else a global override file, else the built-in default key list
func loadKeysFromFile(userContext string) []string {
	userFilePath := "/home/" + userContext + "/php.ini/options.txt"
	if content, err := tenantfs.ReadFile(userFilePath); err == nil {
		return splitNonEmptyLines(content)
	}
	if content, err := tenantfs.ReadFile("/etc/openpanel/php/options.txt"); err == nil {
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
		if line = strings.TrimSpace(line); line != "" {
			keys = append(keys, line)
		}
	}
	return keys
}

// activeIniKeyLine reports whether line is an uncommented "key = ..." directive for exactly this key
func activeIniKeyLine(line, key string) bool {
	rest, ok := strings.CutPrefix(strings.TrimLeft(line, " \t"), key)
	return ok && strings.HasPrefix(strings.TrimLeft(rest, " \t"), "=")
}

var iniValueNewlines = strings.NewReplacer("\r", "", "\n", "")

// updatePHPConfigFile deletes, updates, or appends each key in keyOrder, then overwrites the file in place so the bind mount in the php container sees it without a restart
func updatePHPConfigFile(userContext, version string, keyOrder []string, values map[string]string) error {
	phpIniFile := "/home/" + userContext + "/php.ini/" + version + ".ini"
	content, err := tenantfs.ReadFile(phpIniFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	lines := strings.Split(string(content), "\n")

	for _, key := range keyOrder {
		if key == "" {
			continue
		}
		value := iniValueNewlines.Replace(values[key])
		found := false
		out := make([]string, 0, len(lines)+1)
		for _, line := range lines {
			if activeIniKeyLine(line, key) {
				found = true
				if value == "" {
					continue
				}
				line = key + " = " + value
			}
			out = append(out, line)
		}
		if !found && value != "" {
			// keep the trailing newline at the end of the file
			if n := len(out); n > 0 && out[n-1] == "" {
				out = out[:n-1]
			}
			out = append(out, key+" = "+value, "")
		}
		lines = out
	}

	return tenantfs.WriteFile(phpIniFile, []byte(strings.Join(lines, "\n")), 0o644)
}

// AvailableTimezones lists the IANA timezones available for the date.timezone option - the Go stdlib has no bundled zone list, so this walks /usr/share/zoneinfo and filters out the non-zone metadata files it also contains
var timezoneSkipNames = map[string]bool{
	"posixrules": true, "Factory": true, "iso3166.tab": true, "zone.tab": true,
	"zone1970.tab": true, "leapseconds": true, "tzdata.zi": true, "leap-seconds.list": true,
}

func AvailableTimezones() []string {
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

// OptionField is one rendered row of options.html's key/value table, precomputed server-side so the template just switches on Kind instead of re-deriving the value-classification logic itself
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

// handlePHPOptions gets or updates the PHP options table for one version, versionSeg is "" for the bare /php/options route (version picker only)
func handlePHPOptions(a *appctx.App, w http.ResponseWriter, r *http.Request, versionSeg string) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	version := phpVersionFromSegment(versionSeg)
	title := "Options"

	if version == "" {
		installedVersions := FetchPHPVersions(ctx, a, userContext)
		if r.URL.Query().Get("output") == "json" {
			web.WriteJSON(w, http.StatusOK, map[string]any{"installed_versions": installedVersions})
			return
		}
		renderPHPOptionsSelectPage(a, w, r, title, installedVersions)
		return
	}

	currentContent, readErr := tenantfs.ReadFile("/home/" + userContext + "/php.ini/" + version + ".ini")
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
		if writeErr := updatePHPConfigFile(userContext, version, availableKeys, values); writeErr != nil {
			web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Could not save PHP %(version)s configuration: %(err_msg)s", "version", version, "err_msg", writeErr.Error()), "/php/php"+version+"/options")
			return
		}

		webServer := webserver.GetEnvFileValue(userContext, "WEB_SERVER")
		isLitespeed := strings.Contains(strings.ToLower(webServer), "litespeed")
		var phpContainer, text string
		if isLitespeed {
			phpContainer, text = webServer, "Litespeed"
		} else {
			phpContainer, text = "php-fpm-"+version, "PHP-FPM"
		}
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "edited PHP "+version+" configuration using PHP Selector", ipAddress)
		if !isLitespeed && signalPHPFPMReload(ctx, userContext, phpContainer) {
			web.Flash(a, w, r, "success", web.Tr(a, r, "Configuration edited successfully and %(text)s reloaded to apply new settings.", "text", text))
		} else if result := docker.RestartContainer(ctx, userContext, phpContainer); result.Success {
			web.Flash(a, w, r, "success", web.Tr(a, r, "Configuration edited successfully and %(text)s service restarted to apply new settings.", "text", text))
		} else {
			web.Flash(a, w, r, "error", web.Tr(a, r, "Configuration edited successfully, but %(text)s failed to restart. Try restarting it manually from Services.", "text", text))
		}
		http.Redirect(w, r, "/php/php"+version+"/options", http.StatusFound)
		return
	}

	timezones := AvailableTimezones()

	if r.URL.Query().Get("output") == "json" {
		web.WriteJSON(w, http.StatusOK, map[string]any{
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

	// LiteSpeed runs PHP itself, the startup scripts that need these are only in the php-fpm image
	var startupFuncs []string
	if !strings.Contains(strings.ToLower(webserver.GetEnvFileValue(userContext, "WEB_SERVER")), "litespeed") {
		startupFuncs = fpmStartupFunctions
		if blocked := blockedStartupFunctions(currentConfig["disable_functions"]); len(blocked) > 0 {
			phpIniIssues = append(phpIniIssues, HealthIssue{
				ID: "php-disable-functions:" + version, Severity: "error",
				Message: fmt.Sprintf("disable_functions includes %s. PHP-FPM %s needs these to start, remove them from disable_functions.", strings.Join(blocked, ", "), version),
			})
		}
	}

	fields := make([]OptionField, 0, len(availableKeys))
	for _, key := range availableKeys {
		fields = append(fields, buildOptionField(key, currentConfig[key], timezones))
	}

	renderPHPOptionsPage(a, w, r, version, title, fields, phpIniIssues, startupFuncs)
}

// fpmStartupFunctions are called by the php-fpm image's startup scripts, disabling any of them stops PHP-FPM from starting
var fpmStartupFunctions = []string{"exec", "shell_exec", "ini_get_all"}

// blockedStartupFunctions returns which of fpmStartupFunctions a disable_functions value lists
func blockedStartupFunctions(value string) []string {
	listed := map[string]bool{}
	for _, f := range strings.FieldsFunc(strings.ToLower(strings.Trim(value, `"'`)), func(r rune) bool { return r == ',' || r == ' ' }) {
		listed[f] = true
	}
	var blocked []string
	for _, f := range fpmStartupFunctions {
		if listed[f] {
			blocked = append(blocked, f)
		}
	}
	return blocked
}
