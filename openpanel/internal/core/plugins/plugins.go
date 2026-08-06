// Package plugins scans /etc/openpanel/modules/<folder>/readme.txt for
// admin-installed plugin announcements and serves them for the
// sidebar/dashboard injection already wired into base.html.
//
// It deliberately does not port dynamic plugin code-loading - plugins in
// this codebase are metadata-only; executing arbitrary per-plugin code is
// out of scope. Go has no safe, practical equivalent for loading arbitrary
// compiled code into a running binary at runtime (the stdlib `plugin`
// package is Linux-only and requires an exact toolchain/build match with
// the host binary - GO_MIGRATION_PLAN.md already rules it out for this
// reason), and no plugin exists anywhere in this codebase or on any test
// host to load regardless. A plugin's "link" field is just a URL - it was
// never required to be served by routes registered inside this same
// process, so the sidebar/dashboard integration below works the same
// whether a plugin's backend is a separate service or nothing at all yet.
package plugins

import (
	"os"
	"path/filepath"
	"strings"
)

// BaseDir is the default directory plugins are scanned from.
const BaseDir = "/etc/openpanel/modules/"

// Plugin is one parsed readme.txt, plus its folder name - a generic map
// since readme.txt's key=value format has no fixed schema. Marshaled
// directly to JSON for the /plugins response.
type Plugin map[string]string

// parseReadme parses a generic "key = value" file, ignoring blank lines
// and lines starting with "#".
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

// List returns every immediate subdirectory of baseDir that has a
// readme.txt, parsed plus its folder name. Re-scans the filesystem on
// every call so a plugin dropped in while the app is running appears
// without a restart.
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

// Names returns the set of plugin folder names, computed once at startup
// (unlike List, which re-scans per request) since it drives PluginNames,
// which is baked into the per-user feature/menu visibility set for the
// process lifetime.
func Names(baseDir string) map[string]bool {
	names := map[string]bool{}
	for _, p := range List(baseDir) {
		if folder := p["folder"]; folder != "" {
			names[folder] = true
		}
	}
	return names
}
