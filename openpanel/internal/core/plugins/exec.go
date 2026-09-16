package plugins

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// BinaryPath returns a plugin's own binary path under baseDir if it's installed, or "" otherwise. The plugin contract puts a plugin's binary at <baseDir>/<name>/<name> (folder and binary share the plugin's name) - see any first-party plugin's README for the full contract.
func BinaryPath(baseDir, name string) string {
	path := filepath.Join(baseDir, name, name)
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return path
	}
	return ""
}

// Exec runs an installed plugin's binary with args under a timeout, returning its stdout. Callers get os.ErrNotExist when the plugin isn't installed, so they can distinguish "not installed" from "installed but failed".
func Exec(ctx context.Context, baseDir, name string, timeout time.Duration, args ...string) ([]byte, error) {
	bin := BinaryPath(baseDir, name)
	if bin == "" {
		return nil, os.ErrNotExist
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return exec.CommandContext(cctx, bin, args...).Output()
}

// FindByLink returns the installed plugin (from baseDir) whose readme.txt "link" field exactly matches path, used to route a request to the plugin that claims it - see pluginpage.TryServe.
func FindByLink(baseDir, path string) (Plugin, bool) {
	for _, p := range List(baseDir) {
		if p["link"] == path {
			return p, true
		}
	}
	return Plugin{}, false
}
