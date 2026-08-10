package docker

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// homeDirOverride lets tests point "/home/<context>/..." paths at a
// temp directory instead of the real (root-owned) /home. Production code
// never sets this; it's a test-only seam.
var homeDirOverride string

// homePath builds /home/<context>/<rel>, the per-user home directory
// convention used throughout this package. In tests, homeDirOverride
// stands in for the whole "/home/<context>" directory (tests don't care
// about multi-user isolation, just a writable temp dir).
func homePath(userContext string, rel ...string) string {
	if homeDirOverride != "" {
		return filepath.Join(append([]string{homeDirOverride}, rel...)...)
	}
	parts := append([]string{"/home", userContext}, rel...)
	return filepath.Join(parts...)
}

// LoadCompose reads and parses /home/<context>/docker-compose.yml, or
// returns an empty map if it doesn't exist.
func LoadCompose(userContext string) (map[string]any, error) {
	path := homePath(userContext, "docker-compose.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

// SaveCompose writes data back to /home/<context>/docker-compose.yml.
func SaveCompose(userContext string, data map[string]any) error {
	path := homePath(userContext, "docker-compose.yml")
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// servicesOf reads composeData["services"] as a map[string]map[string]any,
// tolerating a missing or wrong-typed key by returning an empty map.
func servicesOf(composeData map[string]any) map[string]map[string]any {
	result := map[string]map[string]any{}
	raw, _ := composeData["services"].(map[string]any)
	for k, v := range raw {
		if m, ok := v.(map[string]any); ok {
			result[k] = m
		}
	}
	return result
}

// GetAvailableVolumes returns the top-level volume names declared in the
// compose file.
func GetAvailableVolumes(composeData map[string]any) []string {
	return mapKeys(composeData, "volumes")
}

// GetAvailableNetworks returns the top-level network names declared in
// the compose file.
func GetAvailableNetworks(composeData map[string]any) []string {
	return mapKeys(composeData, "networks")
}

func mapKeys(composeData map[string]any, key string) []string {
	raw, _ := composeData[key].(map[string]any)
	keys := make([]string, 0, len(raw))
	for k := range raw {
		keys = append(keys, k)
	}
	return keys
}

// LoadEnvFile parses /home/<context>/.env into a key->value map (values
// with surrounding double quotes stripped).
func LoadEnvFile(userContext string) map[string]string {
	env := map[string]string{}
	data, err := os.ReadFile(homePath(userContext, ".env"))
	if err != nil {
		return env
	}
	for _, line := range strings.Split(string(data), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok {
			env[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"`)
		}
	}
	return env
}

// SaveEnvFile writes key->value pairs back to /home/<context>/.env, one
// KEY="value" per line. Line order is not guaranteed (Go maps don't
// preserve insertion order), but callers that read env values look them
// up by name, not position, so file order doesn't matter functionally.
func SaveEnvFile(userContext string, env map[string]string) error {
	var b strings.Builder
	for k, v := range env {
		safeValue := strings.ReplaceAll(v, `"`, `\"`)
		b.WriteString(k + `="` + safeValue + "\"\n")
	}
	return os.WriteFile(homePath(userContext, ".env"), []byte(b.String()), 0o644)
}

// GetEnvValue looks up a single KEY=value from /home/<context>/.env. The
// bool return distinguishes "key not found" from "key present with an
// empty value" - "" is a valid value.
func GetEnvValue(userContext, variableName string) (string, bool) {
	variableName = strings.ToUpper(variableName)
	data, err := os.ReadFile(homePath(userContext, ".env"))
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok && strings.TrimSpace(k) == variableName {
			return strings.Trim(strings.TrimSpace(v), `'"`), true
		}
	}
	return "", false
}

// restrictedEnvKeys are .env keys SetEnvValue refuses to modify.
var restrictedEnvKeys = map[string]bool{
	"USERNAME": true, "USER_ID": true, "CONTEXT": true,
	"TOTAL_CPU": true, "TOTAL_RAM": true, "HOSTNAME": true, "OS": true,
}

// SetEnvValue updates (or appends) one KEY=value line in
// /home/<context>/.env, preserving every other line's exact formatting.
// Returns a non-empty error message for a restricted key, or "" on
// success.
func SetEnvValue(userContext, variableName, newValue string) string {
	variableName = strings.ToUpper(variableName)
	if restrictedEnvKeys[variableName] {
		return "restricted"
	}

	path := homePath(userContext, ".env")
	var lines []string
	if data, err := os.ReadFile(path); err == nil {
		lines = strings.Split(string(data), "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
	}

	updated := false
	for i, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			if k, _, ok := strings.Cut(trimmed, "="); ok && strings.TrimSpace(k) == variableName {
				lines[i] = variableName + `="` + newValue + `"`
				updated = true
				break
			}
		}
	}
	if !updated {
		lines = append(lines, variableName+`="`+newValue+`"`)
	}

	_ = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
	return ""
}

var validServiceNameRE = regexp.MustCompile(`^[a-z][a-z0-9]{2,}$`)

// IsValidServiceName reports whether name starts with a-z, contains only
// a-z0-9, and is at least 3 characters.
func IsValidServiceName(name string) bool {
	return validServiceNameRE.MatchString(name)
}

var ramRE = regexp.MustCompile(`^\d+(\.\d+)?[gGmM]$`)

// IsValidCPULimit reports whether cpu parses as a positive number.
func IsValidCPULimit(cpu string) bool {
	val, err := strconv.ParseFloat(cpu, 64)
	return err == nil && val > 0
}

// IsValidRAMLimit reports whether ram is a positive number followed by
// 'M' or 'G'.
func IsValidRAMLimit(ram string) bool {
	return ramRE.MatchString(ram)
}

var envKeyReplacer = strings.NewReplacer("-", "_", ".", "_")

// ServiceKeyPrefix uppercases serviceName and replaces '-' and '.' with
// '_', producing the env-var prefix used for that service's settings.
func ServiceKeyPrefix(serviceName string) string {
	return envKeyReplacer.Replace(strings.ToUpper(serviceName))
}

// ParseEnvVars turns each "key: value" line from the form's freeform
// environment textarea into both a compose-file `${SERVICE_KEY}`
// placeholder and a `SERVICE_KEY=value` .env entry.
func ParseEnvVars(rawLines []string, serviceKeyPrefix string) (environmentVars map[string]string, newEnvVars map[string]string) {
	environmentVars = map[string]string{}
	newEnvVars = map[string]string{}
	for _, line := range rawLines {
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key := strings.TrimSpace(k)
		value := strings.TrimSpace(v)
		envKey := serviceKeyPrefix + "_" + strings.ToUpper(key)
		environmentVars[key] = "${" + envKey + "}"
		newEnvVars[envKey] = value
	}
	return environmentVars, newEnvVars
}

// UpdateEnvFileWithVars merges newEnvVars into the existing .env file, if
// there's anything to merge.
func UpdateEnvFileWithVars(userContext string, newEnvVars map[string]string) {
	if len(newEnvVars) == 0 {
		return
	}
	current := LoadEnvFile(userContext)
	for k, v := range newEnvVars {
		current[k] = v
	}
	_ = SaveEnvFile(userContext, current)
}

// VolumeEntry is one (name, mount, readonly) form row.
type VolumeEntry struct {
	Name     string
	Mount    string
	ReadOnly bool
}

// ComposeVolumeString renders a VolumeEntry as a compose-file volume
// string, e.g. "name:/mount" or "name:/mount:ro".
func (v VolumeEntry) ComposeVolumeString() string {
	s := v.Name + ":" + v.Mount
	if v.ReadOnly {
		s += ":ro"
	}
	return s
}

// ParseVolumeEntries zips the three parallel form-array fields
// (volume_name[], volume_mount[], volume_readonly[]) into compose-file
// volume strings, optionally appending the podman-socket bind mount.
func ParseVolumeEntries(names, mounts, readonlyFlags []string, addSocket bool) []string {
	var entries []string
	for i := 0; i < len(names); i++ {
		name := strings.TrimSpace(names[i])
		var mount string
		if i < len(mounts) {
			mount = strings.TrimSpace(mounts[i])
		}
		if name == "" || mount == "" {
			continue
		}
		readonly := i < len(readonlyFlags) && readonlyFlags[i] == "on"
		entries = append(entries, VolumeEntry{Name: name, Mount: mount, ReadOnly: readonly}.ComposeVolumeString())
	}

	const dockerSockEntry = "/run/user/${USER_ID}/podman/podman.sock:/var/run/docker.sock:ro"
	if addSocket && !containsString(entries, dockerSockEntry) {
		entries = append(entries, dockerSockEntry)
	}
	return entries
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

var envPlaceholderRE = regexp.MustCompile(`^\$\{([A-Z0-9_]+)(:-([^}]+))?\}$`)

// ResolveEnvPlaceholder takes a compose-file value like
// "${SERVICE_CPU:-1}" and resolves it against the .env map, falling back
// to the placeholder's own default.
func ResolveEnvPlaceholder(value string, env map[string]string) string {
	m := envPlaceholderRE.FindStringSubmatch(value)
	if m == nil {
		return value
	}
	key, def := m[1], m[3]
	if v, ok := env[key]; ok {
		return v
	}
	return def
}

// parseYAMLString parses a small YAML snippet (e.g. a healthcheck block
// from a form textarea) into a generic value.
func parseYAMLString(s string) (any, error) {
	var v any
	if err := yaml.Unmarshal([]byte(s), &v); err != nil {
		return nil, err
	}
	return v, nil
}

// dumpYAML serializes v back to YAML text, used to redisplay a stored
// healthcheck block in the edit form's textarea.
func dumpYAML(v any) string {
	out, err := yaml.Marshal(v)
	if err != nil {
		return ""
	}
	return string(out)
}
