package web

import (
	"io/fs"
	"os"
	"path"
	"strings"
	"sync/atomic"
)

// panel themes, picked server-wide by ui_theme in openpanel.config
const (
	ThemeModern  = "modern"
	ThemeClassic = "classic"
)

// classicTemplatesDir holds the pre-redesign templates, a page falls back to templates/ for any file missing here
const classicTemplatesDir = "templates_classic/"

// OPENPANEL_UI_THEME only sets the starting theme, so tests can run the render suite under classic
var classicTheme atomic.Bool

func init() {
	SetTheme(os.Getenv("OPENPANEL_UI_THEME"))
}

// SetTheme switches every page to the given theme, anything but "classic" is modern
func SetTheme(name string) {
	classicTheme.Store(NormalizeTheme(name) == ThemeClassic)
}

// Theme is the current panel theme
func Theme() string {
	if classicTheme.Load() {
		return ThemeClassic
	}
	return ThemeModern
}

// NormalizeTheme maps an ui_theme value to a known theme, modern by default
func NormalizeTheme(name string) string {
	if strings.EqualFold(strings.Trim(strings.TrimSpace(name), `"'`), ThemeClassic) {
		return ThemeClassic
	}
	return ThemeModern
}

// classicPaths swaps each file for its classic copy, nil when a page file has none (a page that only exists in the redesign stays modern)
func classicPaths(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		rel := strings.TrimPrefix(p, "templates/")
		if _, err := fs.Stat(templateFS, classicTemplatesDir+rel); err == nil {
			out[i] = classicTemplatesDir + rel
			continue
		}
		// a new partial only adds defines, so the modern one can sit next to classic files
		if strings.HasPrefix(path.Base(rel), "_") {
			out[i] = p
			continue
		}
		return nil
	}
	return out
}

// classicStaticURL points classic pages at the pre-redesign copy of a static file when there is one under static/classic/
func classicStaticURL(rel string) string {
	if assetHash("classic/"+rel) != "" {
		return StaticURL("classic/" + rel)
	}
	return StaticURL(rel)
}
