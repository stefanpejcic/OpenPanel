// Package plugins scans /etc/openpanel/modules/<folder>/readme.txt for admin-installed plugin announcements and serves them for the sidebar/dashboard - metadata-only, no dynamic code-loading, since Go's stdlib `plugin` package needs an exact toolchain match and isn't practical here.
package plugins

import (
	"os"
	"path/filepath"
	"strings"
)

// BaseDir is the default directory plugins are scanned from.
const BaseDir = "/etc/openpanel/modules/"

// Plugin is one parsed readme.txt plus its folder name - a generic map since readme.txt has no fixed schema, marshaled directly to JSON for /plugins
type Plugin map[string]string

// parseReadme parses a generic "key = value" file, ignoring blank lines and "#" comments
func parseReadme(path string) (Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	meta := Plugin{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		meta[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return meta, nil
}

// List returns every immediate subdirectory of baseDir with a readme.txt, parsed plus its folder name. Re-scans on every call so a plugin dropped in while running shows up without a restart.
func List(baseDir string) []Plugin {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil
	}

	var plugins []Plugin
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		folder := entry.Name()
		readmePath := filepath.Join(baseDir, folder, "readme.txt")
		meta, err := parseReadme(readmePath)
		if err != nil {
			continue
		}
		meta["folder"] = folder
		plugins = append(plugins, meta)
	}
	return plugins
}

// Names returns the set of plugin folder names, computed once at startup (unlike List) since it drives PluginNames, fixed for the process lifetime
func Names(baseDir string) map[string]bool {
	names := map[string]bool{}
	for _, p := range List(baseDir) {
		if folder := p["folder"]; folder != "" {
			names[folder] = true
		}
	}
	return names
}
